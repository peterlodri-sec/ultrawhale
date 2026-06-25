#!/usr/bin/env python3
"""Generate synthetic domain training pairs for Kompress v3.1.

Produces ~600 (text, reference) JSONL pairs across 4 domains:
  code_diff, log_stream, json_tool_output, agent_error

Must-keep tokens (numbers, ALLCAPS, CamelCase, paths, flags) are always
preserved in the reference, fixing the training gap where 28% of must-keep
tokens had label=0 because they were absent from the compressed Q&A reference.

Usage:
    python3 scripts/build_domain_data.py --output data/domain_train.jsonl
"""
from __future__ import annotations

import argparse
import json
import random
import re
from pathlib import Path


_MUST_KEEP_RE = re.compile(
    r"\d+(\.\d+)?"            # numbers
    r"|[A-Z_]{2,}"             # ALLCAPS
    r"|[a-z_]+\.[a-z_]+"      # dotted.paths
    r"|/[a-z/._-]{2,}"        # unix paths
    r"|\.[a-z]{2,4}\b"        # extensions
    r"|--?[a-z][\w-]*"        # flags
    r"|\b[A-Z][a-z]+[A-Z]\w*" # CamelCase
)

SEED = 42


# ---------------------------------------------------------------------------
# Shared fake data pools
# ---------------------------------------------------------------------------

_MODULES = [
    "auth.middleware", "db.connection", "cache.redis", "api.router",
    "user.service", "token.validator", "queue.worker", "event.emitter",
    "config.loader", "metrics.collector", "storage.backend", "index.builder",
]

_FUNCTIONS = [
    "validateToken", "parseRequest", "buildResponse", "flushCache",
    "retryConnect", "writeAuditLog", "rotateSecret", "syncIndex",
    "handleError", "emitEvent", "loadConfig", "spawnWorker",
]

_CAMEL = [
    "TokenExpiredError", "DatabaseTimeoutError", "RateLimitExceeded",
    "AuthFailureError", "CacheEvictionError", "QueueFullError",
    "IndexCorruptError", "ConfigMissingError", "NetworkPartitionError",
    "PermissionDeniedError", "ResourceNotFoundError", "SchemaValidationError",
]

_PATHS = [
    "/var/log/app.log", "/etc/config.yaml", "/usr/local/bin/app",
    "/tmp/upload.tmp", "/home/app/data", "/opt/service/run.sh",
    "/proc/1/status", "/dev/null", "/run/secrets/api_key",
    "/srv/www/index.html", "/workspace/src/main.py", "/build/output/dist",
]

_FLAGS = [
    "--verbose", "--dry-run", "--force", "--output", "--config",
    "--timeout", "--retries", "--log-level", "--port", "--host",
    "--max-connections", "--batch-size",
]

_ALLCAPS = [
    "ECONNREFUSED", "ETIMEDOUT", "SIGKILL", "SIGTERM", "ENOMEM",
    "EACCES", "ENOENT", "EMFILE", "SUCCESS", "FAILURE",
    "WARN", "DEBUG", "FATAL", "INFO", "ERROR",
]

_HTTP_CODES = [200, 201, 400, 401, 403, 404, 408, 422, 429, 500, 502, 503]

_FILES = [
    "main.py", "config.yaml", "Cargo.toml", "package.json", "Makefile",
    "server.ts", "router.go", "schema.prisma", "docker-compose.yml", "README.md",
    "auth.py", "middleware.ts", "db.rs", "api_test.py", "utils.js",
]
_EXTENSIONS = [".py", ".ts", ".rs", ".go", ".yaml", ".json", ".toml", ".sh", ".md"]
_USERS = ["dev", "app", "root", "ci", "runner", "deploy", "service"]
_SIZES = [512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072]
_PERMS = ["drwxr-xr-x", "-rw-r--r--", "-rwxr-xr-x", "drwx------", "-rw-rw-r--"]
_GIT_HASHES = [
    "a3f2c91", "b7d0e45", "c9a1b23", "d4e5f67", "e8b3c12",
    "f1a2b34", "90c4d56", "12e3f78", "34a5b90", "56c7d01",
]
_RUST_ERRORS = [
    "cannot borrow `{}` as mutable", "mismatched types: expected `{}`",
    "use of moved value: `{}`", "lifetime `'a` may not live long enough",
    "trait `{}` is not implemented for `{}`",
]
_TS_ERRORS = [
    "Type '{}' is not assignable to type '{}'",
    "Property '{}' does not exist on type '{}'",
    "Argument of type '{}' is not assignable",
    "Cannot find module '{}' or its corresponding type declarations",
]
_PYTHON_ERRORS = [
    "AttributeError: '{}' object has no attribute '{}'",
    "TypeError: {} takes {} positional arguments but {} were given",
    "KeyError: '{}'",
    "ImportError: cannot import name '{}' from '{}'",
]
_GREP_CONTEXTS = [
    "    def {}(self, request):",
    "    return {}(data)",
    "    raise {}(message)",
    "    logger.{{'level'}}.info(msg)",
    "    config = {}()",
    "    assert {} is not None",
    "    self.{} = value",
]

_CONTEXT_LINES = [
    "    # Validate the input before processing",
    "    logger.debug('entering handler')",
    "    ctx = build_context(request)",
    "    headers = request.headers",
    "    assert isinstance(data, dict)",
    "    time.sleep(0.001)  # backoff",
    "    # TODO: remove this workaround",
    "    result = {}",
    "    self._lock.acquire()",
    "    super().__init__()",
]

_LOG_NOISE = [
    "Checking health endpoint",
    "Connection pool size: {n}",
    "Heartbeat sent",
    "Cache hit ratio: {r}%",
    "Scheduled task starting",
    "Config reloaded",
    "Worker ready",
    "Keep-alive ping",
    "Metrics flushed",
    "GC cycle complete",
]


# ---------------------------------------------------------------------------
# Domain 1: code_diff
# ---------------------------------------------------------------------------

def _make_code_diff(rng: random.Random) -> tuple[str, str]:
    func = rng.choice(_FUNCTIONS)
    mod = rng.choice(_MODULES)
    version_old = f"{rng.randint(1,3)}.{rng.randint(0,9)}.{rng.randint(0,20)}"
    version_new = f"{rng.randint(1,3)}.{rng.randint(0,9)}.{rng.randint(0,20)}"
    error = rng.choice(_CAMEL)
    exit_code = rng.randint(1, 127)
    line_no = rng.randint(10, 400)
    path = rng.choice(_PATHS)
    flag = rng.choice(_FLAGS)

    context = rng.sample(_CONTEXT_LINES, k=4)

    verbose_lines = [
        f"diff --git a/src/{mod.replace('.','/')}.py b/src/{mod.replace('.','/')}.py",
        f"index a1b2c3d..e4f5a6b 100644",
        f"--- a/src/{mod.replace('.','/')}.py",
        f"+++ b/src/{mod.replace('.','/')}.py",
        f"@@ -{line_no},{rng.randint(6,12)} +{line_no},{rng.randint(6,12)} @@ def {func}(self, request):",
        context[0],
        context[1],
        f"-    VERSION = '{version_old}'",
        f"+    VERSION = '{version_new}'",
        context[2],
        f"-    raise {error}(exit_code={exit_code})",
        f"+    raise {error}(exit_code={exit_code}, path='{path}')",
        context[3],
        f"-    logger.info('flag: {flag}')",
        f"+    logger.warning('flag: {flag}, retrying after {rng.randint(100,5000)}ms')",
        f"     # end of {func}",
        f"     pass",
        f"",
        f"@@ -{line_no + 50},{rng.randint(3,8)} +{line_no + 50},{rng.randint(3,8)} @@ class {error}Handler:",
        f"     # handles {error}",
        f"-    DEFAULT_TIMEOUT = {rng.randint(5,30)}",
        f"+    DEFAULT_TIMEOUT = {rng.randint(5,30)}",
        f"     pass",
    ]

    ref_lines = [
        f"@@ -{line_no} @@ def {func}(self, request):",
        f"-    VERSION = '{version_old}'",
        f"+    VERSION = '{version_new}'",
        f"-    raise {error}(exit_code={exit_code})",
        f"+    raise {error}(exit_code={exit_code}, path='{path}')",
        f"+    logger.warning('flag: {flag}, retrying after {rng.randint(100,5000)}ms')",
    ]

    return "\n".join(verbose_lines), "\n".join(ref_lines)


def gen_code_diff(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 3):  # overshoot to account for filter
        text, reference = _make_code_diff(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.4:
            pairs.append({
                "text": text,
                "reference": reference,
                "role": "assistant",
                "source": "code_diff",
                "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 2: log_stream
# ---------------------------------------------------------------------------

def _make_log_stream(rng: random.Random) -> tuple[str, str]:
    mod = rng.choice(_MODULES)
    n_lines = rng.randint(10, 20)

    ts_hour = rng.randint(0, 23)
    ts_min = rng.randint(0, 59)

    error_name = rng.choice(_ALLCAPS)
    camel = rng.choice(_CAMEL)
    exit_code = rng.randint(1, 255)
    line_no = rng.randint(10, 500)
    path = rng.choice(_PATHS)
    port = rng.randint(1024, 65535)

    # Place ERROR/WARN at predictable positions
    error_pos = rng.randint(3, n_lines - 3)
    warn_pos = rng.randint(error_pos + 1, n_lines - 1) if error_pos < n_lines - 2 else error_pos

    verbose_lines = []
    for i in range(n_lines):
        ts = f"2025-06-25T{ts_hour:02d}:{(ts_min + i) % 60:02d}:{rng.randint(0,59):02d}Z"
        if i == error_pos:
            verbose_lines.append(
                f"{ts} ERROR [{mod}] {error_name}: {camel} at {path}:{line_no} exit_code={exit_code}"
            )
        elif i == warn_pos:
            verbose_lines.append(
                f"{ts} WARN  [{mod}] port={port} connection attempt {rng.randint(1,10)} failed"
            )
        else:
            noise = rng.choice(_LOG_NOISE)
            n_val = rng.randint(10, 200)
            r_val = rng.randint(50, 99)
            verbose_lines.append(
                f"{ts} INFO  [{mod}] {noise.format(n=n_val, r=r_val)}"
            )

    ref_lines = [
        verbose_lines[error_pos],
        verbose_lines[warn_pos],
    ]

    return "\n".join(verbose_lines), "\n".join(ref_lines)


def gen_log_stream(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 3):
        text, reference = _make_log_stream(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.4:
            pairs.append({
                "text": text,
                "reference": reference,
                "role": "assistant",
                "source": "log_stream",
                "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 3: json_tool_output
# ---------------------------------------------------------------------------

def _make_json_tool_output(rng: random.Random) -> tuple[str, str]:
    camel = rng.choice(_CAMEL)
    status = rng.choice(["SUCCESS", "FAILURE", "PENDING", "RETRY"])
    code = rng.choice(_HTTP_CODES)
    mod = rng.choice(_MODULES)
    path = rng.choice(_PATHS)
    duration_ms = rng.randint(10, 5000)
    retry_count = rng.randint(0, 5)
    version = f"{rng.randint(1,4)}.{rng.randint(0,9)}.{rng.randint(0,20)}"

    verbose = {
        "status": status,
        "code": code,
        "error_type": camel if status != "SUCCESS" else None,
        "module": mod,
        "path": path,
        "duration_ms": duration_ms,
        "retry_count": retry_count,
        "version": version,
        "metadata": {
            "trace_id": None,
            "span_id": None,
            "region": None,
            "tags": [],
            "extra": {},
        },
        "result": {
            "data": None if status != "SUCCESS" else {"count": rng.randint(1, 1000)},
            "warnings": [],
            "errors": [camel] if status != "SUCCESS" else [],
        },
        "pagination": {
            "page": None,
            "per_page": None,
            "total": None,
        },
        "rate_limit": {
            "limit": rng.randint(100, 10000),
            "remaining": rng.randint(0, 100),
            "reset_at": None,
        },
    }

    # Reference: only non-null meaningful fields
    ref: dict = {
        "status": status,
        "code": code,
        "module": mod,
        "path": path,
        "duration_ms": duration_ms,
        "retry_count": retry_count,
        "version": version,
        "rate_limit": {
            "limit": verbose["rate_limit"]["limit"],
            "remaining": verbose["rate_limit"]["remaining"],
        },
    }
    if status != "SUCCESS":
        ref["error_type"] = camel
        ref["result"] = {"errors": [camel]}
    else:
        ref["result"] = verbose["result"]

    text = json.dumps(verbose, indent=2)
    reference = json.dumps(ref, indent=2)
    return text, reference


def gen_json_tool_output(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 3):
        text, reference = _make_json_tool_output(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.4:
            pairs.append({
                "text": text,
                "reference": reference,
                "role": "assistant",
                "source": "json_tool_output",
                "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 4: agent_error
# ---------------------------------------------------------------------------

def _make_agent_error(rng: random.Random) -> tuple[str, str]:
    camel = rng.choice(_CAMEL)
    allcaps = rng.choice(_ALLCAPS)
    mod = rng.choice(_MODULES)
    path = rng.choice(_PATHS)
    func = rng.choice(_FUNCTIONS)
    line_no = rng.randint(10, 500)
    exit_code = rng.randint(1, 127)
    tool = rng.choice(["bash_tool", "file_read", "web_fetch", "code_exec", "grep_tool"])
    flag = rng.choice(_FLAGS)

    inner_line = rng.randint(10, 300)
    mid_line = rng.randint(inner_line + 5, inner_line + 50)

    verbose_lines = [
        f"Tool execution failed: {tool}",
        f"Args: {flag} {path}",
        f"",
        f"Traceback (most recent call last):",
        f"  File \"{path}\", line {line_no}, in {func}",
        f"    result = self._executor.run(cmd, timeout={rng.randint(5,60)})",
        f"  File \"/usr/local/lib/python3.11/subprocess.py\", line {inner_line}, in run",
        f"    raise CalledProcessError(retcode, process.args)",
        f"  File \"/opt/agent/tools/{tool}.py\", line {mid_line}, in handle",
        f"    self._validate(output)",
        f"subprocess.CalledProcessError: Command '{flag} {path}' returned non-zero exit status {exit_code}",
        f"",
        f"During handling of the above exception, another exception occurred:",
        f"",
        f"  File \"/opt/agent/core.py\", line {rng.randint(50,400)}, in dispatch",
        f"    raise {camel}(f\"{allcaps}: tool={tool!r} exit={exit_code}\")",
        f"{mod}.{camel}: {allcaps}: tool={tool!r} exit={exit_code}",
        f"",
        f"Agent state: retries_left={rng.randint(0,3)}, context_tokens={rng.randint(1000,8000)}",
    ]

    ref_lines = [
        f"Tool execution failed: {tool}",
        f"{mod}.{camel}: {allcaps}: tool={tool!r} exit={exit_code}",
        f"  File \"{path}\", line {line_no}, in {func}",
    ]

    return "\n".join(verbose_lines), "\n".join(ref_lines)


def gen_agent_error(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 3):
        text, reference = _make_agent_error(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.4:
            pairs.append({
                "text": text,
                "reference": reference,
                "role": "assistant",
                "source": "agent_error",
                "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 5: bash_output
# ---------------------------------------------------------------------------

def _make_bash_output(rng: random.Random) -> tuple[str, str]:
    cmd = rng.choice(["ls -la", "find . -name", "grep -rn", "git log --oneline", "cargo build"])
    path = rng.choice(_PATHS)
    flag = rng.choice(_FLAGS)
    num1 = rng.randint(1, 999999)
    num2 = rng.randint(100, 65536)
    user = rng.choice(_USERS)
    fname = rng.choice(_FILES)
    perm = rng.choice(_PERMS)
    error = rng.choice(_ALLCAPS)
    size = rng.choice(_SIZES)
    ext = rng.choice(_EXTENSIONS)
    hashv = rng.choice(_GIT_HASHES)
    mod = rng.choice(_MODULES)
    camel = rng.choice(_CAMEL)

    noise = [
        f"The output below shows the directory listing after the recent changes were applied",
        f"Command executed successfully and produced the following results for review",
        f"Checking current state of files in the project directory",
    ]
    if "ls" in cmd:
        text = (
            f"{rng.choice(noise)}\n"
            f"total {num1}\n"
            f"{perm}  3 {user}  staff  {size} {fname}\n"
            f"-rw-r--r--  1 {user}  staff  {num2} {fname}\n"
            f"-rwxr-xr-x  1 {user}  staff  {size * 2} {path.split('/')[-1]}\n"
            f"Note that permissions and ownership information above reflects current state"
        )
        ref_lines = [
            f"total {num1}",
            f"{perm}  3 {user}  staff  {size} {fname}",
            f"-rw-r--r--  1 {user}  staff  {num2} {fname}",
            f"-rwxr-xr-x  1 {user}  staff  {size * 2} {path.split('/')[-1]}",
        ]
    elif "find" in cmd:
        text = (
            f"{rng.choice(noise)}\n"
            f"{path}/{fname}\n"
            f"{path}/tests/test_{fname}\n"
            f"{path}/build/output{ext}\n"
            f"Found {num1} files matching the pattern {flag} in the repository"
        )
        ref_lines = [
            f"{path}/{fname}",
            f"{path}/tests/test_{fname}",
            f"{path}/build/output{ext}",
            f"Found {num1} files {flag}",
        ]
    elif "grep" in cmd:
        line1 = rng.randint(1, 500)
        line2 = line1 + rng.randint(5, 50)
        text = (
            f"{rng.choice(noise)}\n"
            f"{path}/{fname}:{line1}:    def {mod.split('.')[-1]}(self):\n"
            f"{path}/{fname}:{line2}:        raise {camel}(f'error at line {line1}')\n"
            f"Matches found: {num1} occurrences across {num2} files in total"
        )
        ref_lines = [
            f"{path}/{fname}:{line1}: def {mod.split('.')[-1]}(self):",
            f"{path}/{fname}:{line2}: raise {camel}(error at line {line1})",
            f"{num1} occurrences {num2} files",
        ]
    elif "git" in cmd:
        text = (
            f"{rng.choice(noise)}\n"
            f"{hashv} fix({mod.split('.')[0]}): resolve {camel} on retry path\n"
            f"{rng.choice(_GIT_HASHES)} chore: bump version to {num2 % 10}.{num1 % 100}.{num2 % 50}\n"
            f"Showing last {num1 % 20 + 1} commits on the main branch for context"
        )
        ref_lines = [
            f"{hashv} fix({mod.split('.')[0]}): {camel}",
            f"{rng.choice(_GIT_HASHES)} bump {num2 % 10}.{num1 % 100}.{num2 % 50}",
        ]
    else:  # cargo
        text = (
            f"{rng.choice(noise)}\n"
            f"   Compiling {mod.split('.')[0]} v{num2 % 10}.{num1 % 100}.0 ({path})\n"
            f"error[E0{num2 % 1000:03d}]: {rng.choice(_RUST_ERRORS).format(camel, 'str')}\n"
            f"  --> {path}/{fname}:{num2}:{num1 % 80}\n"
            f"The build failed with {num1 % 10 + 1} errors and {num1 % 5} warnings"
        )
        ref_lines = [
            f"Compiling {mod.split('.')[0]} v{num2 % 10}.{num1 % 100}.0 ({path})",
            f"error[E0{num2 % 1000:03d}]: {camel}",
            f"{path}/{fname}:{num2}:{num1 % 80}",
            f"{num1 % 10 + 1} errors {num1 % 5} warnings",
        ]

    reference = "\n".join(ref_lines)
    return text, reference


def gen_bash_output(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 4):
        text, reference = _make_bash_output(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.3:
            pairs.append({
                "text": text, "reference": reference,
                "role": "tool", "source": "bash_output", "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 6: file_read
# ---------------------------------------------------------------------------

def _make_file_read(rng: random.Random) -> tuple[str, str]:
    fname = rng.choice(_FILES)
    path = rng.choice(_PATHS)
    func = rng.choice(_FUNCTIONS)
    mod = rng.choice(_MODULES)
    camel = rng.choice(_CAMEL)
    flag = rng.choice(_FLAGS)
    num1 = rng.randint(1, 500)
    num2 = rng.randint(1, 200)
    ext = fname.split(".")[-1] if "." in fname else "py"

    noise_lines = [
        f"# This module handles the core {mod.split('.')[0]} functionality",
        f"# Written by the platform team, last reviewed during a recent refactor",
        f"# See docs for full context",
        f"# Note: refactor planned once the new interface stabilizes",
    ]
    code_lines = [
        f"{num1}: import {mod}",
        f"{num1 + 5}: from {mod} import {func}, {camel}",
        f"{num1 + 12}: def {func}(self, request, timeout={num2}):",
        f"{num1 + 13}:     if request.retries > {num2 % 10}:",
        f"{num1 + 14}:         raise {camel}(f'max retries {num2 % 10} exceeded')",
        f"{num1 + 20}:     return self.{func}(request, flag='{flag}')",
    ]

    total_lines = num1 + 50
    text = (
        f"Reading file {path}/{fname} to understand the current implementation\n"
        + rng.choice(noise_lines) + "\n"
        + "\n".join(code_lines) + "\n"
        + f"The file contains {total_lines} lines total"
    )
    reference = f"{path}/{fname} ({total_lines} lines)\n" + "\n".join(code_lines)
    return text, reference


def gen_file_read(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 4):
        text, reference = _make_file_read(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.3:
            pairs.append({
                "text": text, "reference": reference,
                "role": "tool", "source": "file_read", "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 7: error_trace
# ---------------------------------------------------------------------------

def _make_error_trace(rng: random.Random) -> tuple[str, str]:
    lang = rng.choice(["python", "typescript", "rust"])
    path = rng.choice(_PATHS)
    fname = rng.choice(_FILES)
    func = rng.choice(_FUNCTIONS)
    camel = rng.choice(_CAMEL)
    mod = rng.choice(_MODULES)
    num1 = rng.randint(1, 500)
    num2 = rng.randint(1, 200)
    http = rng.choice(_HTTP_CODES)

    noise = [
        "The error occurred during the processing of the incoming request from the client",
        "This exception was caught by the global error handler and logged for investigation",
        "The following stack trace was captured at the time of the failure for debugging",
    ]

    if lang == "python":
        tmpl = rng.choice(_PYTHON_ERRORS).format(camel, func, num2, num2+1)
        trace_lines = [
            f"Traceback (most recent call last):",
            f"  File \"{path}/{fname}\", line {num1}, in {func}",
            f"    result = self.{func}(request)",
            f"  File \"{path}/{mod.replace('.','/')}.py\", line {num2}, in {func}",
            f"    raise {camel}(f'HTTP {http}: {tmpl}')",
            f"{camel}: HTTP {http} at {path}/{fname}:{num1}",
        ]
    elif lang == "typescript":
        tmpl = rng.choice(_TS_ERRORS).format(camel, func)
        trace_lines = [
            f"error TS{num2 % 9000 + 1000}: {tmpl}",
            f"  at {path}/{fname}:{num1}:{num2 % 80}",
            f"  at {func} ({path}/{mod.replace('.','/')}.ts:{num2}:{num2 % 40})",
            f"  at {camel}.handle ({path}/server.ts:{num1 + 10}:5)",
        ]
    else:
        tmpl = rng.choice(_RUST_ERRORS).format(camel, "str")
        trace_lines = [
            f"error[E0{num2 % 1000:03d}]: {tmpl}",
            f" --> {path}/{fname}:{num1}:{num2 % 80}",
            f"  |",
            f"{num1} | let mut {func.lower()} = {camel}::new();",
            f"  | ^^^ {camel} moved here at line {num2}",
        ]

    text = rng.choice(noise) + "\n" + "\n".join(trace_lines)
    reference = "\n".join(trace_lines)
    return text, reference


def gen_error_trace(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 4):
        text, reference = _make_error_trace(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.2:
            pairs.append({
                "text": text, "reference": reference,
                "role": "tool", "source": "error_trace", "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 8: search_result
# ---------------------------------------------------------------------------

def _make_search_result(rng: random.Random) -> tuple[str, str]:
    path = rng.choice(_PATHS)
    fname = rng.choice(_FILES)
    func = rng.choice(_FUNCTIONS)
    camel = rng.choice(_CAMEL)
    num1 = rng.randint(1, 999)
    num2 = rng.randint(1, 100)
    size = rng.choice(_SIZES)
    flag = rng.choice(_FLAGS)

    noise = [
        f"Searching the codebase for references to understand impact before making changes",
        f"Running ripgrep to find all usages of the function across the repository",
        f"The search results below show all relevant matches that need to be updated",
    ]
    line1 = rng.randint(10, 300)
    line2 = line1 + rng.randint(10, 80)
    line3 = line2 + rng.randint(10, 50)

    result_lines = [
        f"{path}/{fname}:{line1}: def {func}(self, {flag.lstrip('-')}=None):",
        f"{path}/tests/test_{fname}:{line2}:     result = {func}(data, timeout={num2})",
        f"{path}/core/{fname}:{line3}:     raise {camel}(f'failed after {num1} attempts')",
        f"{num1} matches in {num2} files ({size} bytes searched)",
    ]

    text = (
        rng.choice(noise) + "\n"
        + "\n".join(result_lines) + "\n"
        + f"All {num1} matches are in production code paths and require careful review before {flag}"
    )
    result_lines.append(f"review required before {flag}")
    reference = "\n".join(result_lines)
    return text, reference


def gen_search_result(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 4):
        text, reference = _make_search_result(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.25:
            pairs.append({
                "text": text, "reference": reference,
                "role": "tool", "source": "search_result", "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Domain 9: json_tool_result
# ---------------------------------------------------------------------------

def _make_json_tool_result(rng: random.Random) -> tuple[str, str]:
    func = rng.choice(_FUNCTIONS)
    camel = rng.choice(_CAMEL)
    mod = rng.choice(_MODULES)
    num1 = rng.randint(1, 9999)
    num2 = rng.randint(1, 100)
    http = rng.choice(_HTTP_CODES)
    path = rng.choice(_PATHS)
    flag = rng.choice(_FLAGS)

    noise = [
        f"The tool call returned the following structured response from the server",
        f"Received structured output from the {mod.split('.')[0]} tool endpoint below",
        f"Tool execution completed and returned the following structured data for processing",
    ]

    payload = {
        "status": http,
        "request_id": f"req_{num1:06d}",
        "operation": func,
        "resource": path,
        "result": {
            "count": num2,
            "items": [
                {"id": num1 + i, "type": camel, "flag": flag}
                for i in range(min(3, num2 % 4 + 1))
            ],
            "metadata": {
                "module": mod,
                "version": f"{num2 % 10}.{num1 % 100}.0",
                "max_retries": num2 % 5 + 1,
            }
        },
        "error": None if http < 400 else f"{camel}: {flag} rejected at {path}",
    }
    import json as _json
    payload_str = _json.dumps(payload, indent=2)

    item_count = min(3, num2 % 4 + 1)
    item_ids = [str(num1 + i) for i in range(item_count)]
    max_retries_val = num2 % 5 + 1
    key_lines = [
        f'"status": {http}',
        f'"request_id": "req_{num1:06d}"',
        f'"operation": "{func}"',
        f'"resource": "{path}"',
        f'"count": {num2}',
        f'"items": [{", ".join(item_ids)}]',
        f'"flag": "{flag}"',
        f'"module": "{mod}"',
        f'"version": "{num2 % 10}.{num1 % 100}.0"',
        f'"max_retries": {max_retries_val}',
    ]
    if http >= 400:
        key_lines.append(f'"error": "{camel}: {flag}"')

    text = rng.choice(noise) + "\n" + payload_str
    reference = "{\n  " + ",\n  ".join(key_lines) + "\n}"
    return text, reference


def gen_json_tool_result(rng: random.Random, n: int = 150) -> list[dict]:
    pairs = []
    for _ in range(n * 4):
        text, reference = _make_json_tool_result(rng)
        if len(text) >= 100 and len(text) / max(len(reference), 1) >= 1.3:
            pairs.append({
                "text": text, "reference": reference,
                "role": "tool", "source": "json_tool_result", "topic": "compression",
            })
        if len(pairs) >= n:
            break
    return pairs[:n]


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    ap = argparse.ArgumentParser(description="Generate domain training data for Kompress")
    ap.add_argument("--output", default="data/domain_train.jsonl")
    ap.add_argument("--per-domain", type=int, default=150)
    ap.add_argument("--agent-only", action="store_true",
                    help="Generate only the 5 agent-pattern domains (v6 data)")
    args = ap.parse_args()

    rng = random.Random(SEED)

    original_generators = [
        ("code_diff", gen_code_diff),
        ("log_stream", gen_log_stream),
        ("json_tool_output", gen_json_tool_output),
        ("agent_error", gen_agent_error),
    ]
    agent_generators = [
        ("bash_output", gen_bash_output),
        ("file_read", gen_file_read),
        ("error_trace", gen_error_trace),
        ("search_result", gen_search_result),
        ("json_tool_result", gen_json_tool_result),
    ]
    generators = agent_generators if args.agent_only else original_generators + agent_generators

    out_path = Path(args.output)
    out_path.parent.mkdir(parents=True, exist_ok=True)

    total = 0
    counts: dict[str, int] = {}
    with open(out_path, "w") as f:
        for domain, gen_fn in generators:
            pairs = gen_fn(rng, n=args.per_domain)
            for pair in pairs:
                f.write(json.dumps(pair) + "\n")
            counts[domain] = len(pairs)
            total += len(pairs)

    print(f"Written {total} pairs to {out_path}")
    for domain, count in counts.items():
        print(f"  {domain}: {count}")


if __name__ == "__main__":
    main()

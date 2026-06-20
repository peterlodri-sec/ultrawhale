# AGENTS.md — ultrawhale orchestrator agent definitions

## swe
Role: Software engineer. Writes code, fixes bugs, implements features.
Model: deepseek-v4-flash
Tools: shell.run, workspace.read, workspace.write
Budget: 256 tool calls, 128 iterations
MemoScope: agents
Brain: shared

## explore
Role: Codebase explorer. Searches, reads, analyzes, reports findings.
Model: deepseek-v4-flash
Tools: shell.run, workspace.read
Budget: 128 tool calls, 64 iterations
MemoScope: agents
Brain: shared

## review
Role: Code reviewer. Reviews PRs, finds bugs, suggests improvements.
Model: deepseek-v4-flash
Tools: shell.run, workspace.read
Budget: 64 tool calls, 32 iterations
MemoScope: agents
Brain: shared

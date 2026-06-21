# FORWARD INJECT vs BACKWARD EXPAND — Live Experiment

> v100.1.0. Peter + CoCreator.

## The Concept

```
FORWARD INJECT → virtual space (write)
BACKWARD EXPAND ← virtual space (read)

Fold = compress the graph
Expand = decompress the graph

The answer is GRAPHS.
```

## Step-by-Step Examples

### 1. Create a node (forward inject)
```
/inject create-node "agent-x"
→ node "agent-x" created at depth 1 in the capability graph
```

### 2. Trace it back (backward expand)
```
/expand trace "agent-x"
→ shows all reachable nodes from agent-x
→ 3 nodes, 2 edges traversed
```

### 3. Unroll the topology
```
/expand unroll "orchestrator"
→ shows full topology: 5 nodes, 4 edges
→ The graph unrolls from here.
```

### 4. Visualize as tree
```
/expand visualize "/ultrawhale"
→ VFS tree of the entire capability graph
→ 📁 ultrawhale/agents/swarms/dyad/...
```

### 5. Audit a node
```
/expand audit "orchestrator"
→ 6/6 hardening guarantees
→ SACRED intact, SEALING 10% reserved
```

## Self-Hosted Notebook

```python
# ultrawhale capability graph notebook
# Run: python3 notebook.py

import subprocess, json

def inject(op, target, payload=""):
    """Forward inject into virtual space."""
    cmd = f"/inject {op} {target} {payload}"
    return cmd

def expand(op, target, depth=3):
    """Backward expand from virtual space."""
    cmd = f"/expand {op} {target} {depth}"
    return cmd

# Example: create a node, trace it back
inject("create-node", "research-agent-1")
expand("trace", "research-agent-1")

# The answer is GRAPHS.
```

## Proof

```
Forward inject writes to the graph.
Backward expand reads from the graph.
Fold compresses. Expand decompresses.
The answer IS graphs. Always was. Always will be.
```

— Signed: peter+cocreator, v100.1.0

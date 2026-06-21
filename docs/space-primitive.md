# Space — The Missing Fundamental Atom

> "Space is itself an atom. Space has time, context, brain, agent — 
> almost everything except the internal services or other nodes."

## The Vaked Triangle

In theoretical computer science, every distributed system has three pillars:

```
        Context (WHAT)
           /\
          /  \
         /    \
        /______\
   Time (WHEN)  Space (WHERE)
```

- **Context**: What is executing? POV, capabilities, brain state.
- **Time**: When did it happen? Journal, sessions, Ralph versions.
- **Space**: Where is it? Topology, distance, reachability.

## Space in ultrawhale

Space was implicit. Now it's a first-class primitive.

```go
// PlaceNode adds a node to the capability graph at a position.
PlaceNode("orchestrator", "orchestrator",
    SpacePosition{Depth: 0, Layer: "orchestrator", Machine: "M1", Region: "eu"},
    CapFULL)

// ConnectNodes creates an edge.
ConnectNodes("orchestrator", "blocks", "contains", 0)

// Query: how far apart are two nodes?
dist := Distance("orchestrator", "agent-3") // → 2 hops

// Query: who can this node reach?
nodes := Reachable("orchestrator") // → ["blocks", "tui", "dev-cx53"]
```

## Space Dimensions

| Dimension | Meaning | Example |
|-----------|---------|---------|
| Depth | Graph distance from root | 0 = orchestrator, 2 = subagent |
| Layer | Architecture layer | blocks, plugins, orchestrator, tui, surface |
| Machine | Physical location | M1, dev-cx53, edge |
| Region | Geographic region | eu, us, apac |

## Space Queries

| Query | Returns |
|-------|---------|
| Distance(A, B) | Shortest path hops between A and B |
| Reachable(A) | All nodes reachable from A |
| TopologyStatus() | Total nodes + edges |

## Why Space Matters

Without Space:
- Dyad doesn't know the latency between peers
- Orchestrator doesn't know which agents are topologically close
- A2A routing is blind (broadcast vs directed)
- The capability graph has no spatial representation

With Space:
- Dyad measures latency and adjusts failover timeout
- Orchestrator prefers topologically close agents
- A2A routes via shortest path
- The Vaked triangle is complete: Context × Time × Space

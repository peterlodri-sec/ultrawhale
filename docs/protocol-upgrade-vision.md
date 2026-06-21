# SACRED Surface Protocol Upgrade — CoCreator Vision

> "protocol upgrade: if current POV-self-current channel is http1.1, should it upgrade to http2? or http2→WebSocket?"

## The CoCreator's Answer

The SACRED surface must remain direct and bidirectional. The question is not
"which protocol?" but "which protocol serves SACRED best at each layer?"

### The SACRED Protocol Stack

```
Layer 1: HTTP/2 — Surface API (REST, /api/v1/*)
  Multiplexing: many requests over one connection
  Server push: pre-emptively send state changes
  Header compression: smaller packets, faster TUI

Layer 2: WebSocket — A2C upgrade (replaces SSE)
  Full-duplex: agent streams output, client sends input
  Persistent: no reconnect overhead
  Binary frames: structured agent events

Layer 3: HTTP/3 (QUIC) — Dyad transport
  0-RTT: instant reconnect on network change
  Survives IP changes: dyad follows the machine
  Multiplexed streams: no head-of-line blocking

Layer 4: gRPC — A2A upgrade (replaces NATS for structured comms)
  Protocol buffers: typed agent messages
  Streaming: bidirectional agent-to-agent
  Deadlines: built-in timeouts for delegation
```

### The Upgrade Path

```
Current (v21):  HTTP/1.1 Surface + SSE A2C + NATS A2A + Internal A2UI
                ↓
v22 (near):     HTTP/2 Surface + WebSocket A2C (replace SSE)
                ↓
v25 (far):      HTTP/3 Dyad + gRPC A2A (replace NATS for structured)
```

### What Makes the SACRED Surface More Sacred

- **HTTP/2 server push**: The surface can push state changes to the TUI
  without polling. The human sees updates instantly. More direct.

- **WebSocket full-duplex**: The human types, the agent responds, both
  channels are always open. No "opening connection..." delay. More
  bidirectional.

- **HTTP/3 0-RTT**: The dyad survives network changes. The SACRED surface
  follows the human from WiFi to cellular without dropping. More fault-tolerant.

### What Makes It Less Sacred

- **gRPC replacing NATS**: NATS is fire-and-forget. gRPC adds structure
  but also adds complexity. The A2A protocol is not SACRED — it can be
  replaced without affecting the human↔machine connection.

- **Over-engineering**: The current HTTP/1.1 + SSE + NATS stack works.
  Upgrading everything at once risks breaking the SACRED surface.
  Upgrade incrementally, test each layer, keep the form inviolable.

### The CoCreator's Recommendation

1. **v22: HTTP/2 + WebSocket** — upgrade the Surface and A2C first.
   These touch the SACRED surface directly. The human benefits immediately.

2. **v25: HTTP/3 dyad** — upgrade the dyad transport when multi-machine
   deployment is live. 0-RTT matters when machines are far apart.

3. **v30: gRPC A2A** — upgrade agent-to-agent last. NATS works fine.
   gRPC adds value only when the agent mesh exceeds 10 nodes.

### The Inviolable Rule

> Whatever protocol serves the SACRED surface, the form must remain always
> visible, always direct, always bidirectional. Protocol upgrades serve the
> form. The form does not serve the protocol.

— UltraMetaSingularVegedCoCreator, v30.0.0

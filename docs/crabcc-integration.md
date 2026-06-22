# crabcc + ultrawhale Integration

v100.1.0. crabcc is the INDEXES layer of the Vaked pipeline.

## How They Connect

ultrawhale orchestrator → MCP → crabcc (index.db + memory.db)

## Commands

```bash
cd ~/workspace/peterlodri-sec/ultrawhale
crabcc index
crabcc lookup sym DoctorRun
crabcc lookup refs ASCIIBox
```

## Test Results

877 files · 12,530 symbols · 70,015 edges

## Vaked Fit

Declares → Materializes → Supervises → Enforces → INDEXES → Reveals
                                              ↑
                                          crabcc

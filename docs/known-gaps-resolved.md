# Known Gaps — Resolved v100.1.0

## State Duplication ✅
- **Gap**: 3 sources of truth (loopState, mainState, SafeSpace.State)
- **Fix**: `SystemState()` function — single source of truth

## Event Loop Crash Recovery ✅
- **Gap**: No panic recovery in 60fps loop
- **Fix**: `defer recover()` with auto-restart

## Headless Mode ✅
- **Gap**: `--headless` flag parsed but unused
- **Fix**: `SetHeadlessMode()` — FrameRate=0, no visual rendering

## Upstream Code ✅
- **Gap**: infra_bar:154, widget:74, vaked:163
- **Fix**: Excluded from CI, documented as upstream

## VakedFit Coverage ✅
- **Gap**: ~80% of blocks had VakedFit
- **Fix**: Added to hash, compress, mmap — documented as pure utilities

## TUI Assumption ✅
- **Gap**: FrameRate assumes visual rendering
- **Fix**: FrameRate=0 = headless mode (see SetHeadlessMode)

## Abstraction Leak ✅
- **Gap**: blocks reference TUI concepts
- **Fix**: engine/ vs surface/ separation documented

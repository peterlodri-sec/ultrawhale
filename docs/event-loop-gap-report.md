# Event Loop Gap + Abstraction Leak Report — v82.0.0

## Gaps Found

| # | Gap | Severity | Detail |
|---|-----|----------|--------|
| 1 | **State triplication** | Medium | 3 places track state: `loopState` (event_loop.go), `mainState` (self_main_state.go), `SafeSpace.State` (safe_space.go). No single source of truth. |
| 2 | **No crash recovery** | High | If event loop goroutine panics, no recovery. No supervisor restart. |
| 3 | **TUI assumption leak** | Medium | `EventLoop` struct has `FrameRate` — assumes visual rendering. Headless mode can't use it. |
| 4 | **Unified status missing** | Low | No single `/status` command showing all 3 state views at once. |

## Abstraction Leaks

| Leak | Location | Fix |
|------|----------|-----|
| `FrameRate` assumes TUI refresh | `event_loop.go:30` | Make `FrameRate=0` mean "headless" (no visual tick) |
| `ActiveRecursion` duplicates `detectActiveRecursion()` | `event_loop.go:90` vs `recursion.go` | Use single source of truth from recursion module |
| `SelfLiveTick()` duplicates `AutoTransition()` logic | `self_live_webhook.go:95` vs `self_main_state.go:115` | Merge into one tick function |

## What's Perfect

- ✅ All 7 recursion types detected correctly
- ✅ Safe space check gates dyad existence
- ✅ Keyboard gate intact (one-way)
- ✅ SACRED surface health checked every frame
- ✅ 0 race conditions (all state is atomic or mutex-protected)

## Recommendations

1. **Unify state**: single `SystemState` struct combining loopState + mainState + safeSpace
2. **Crash recovery**: wrap event loop in `recover()` with auto-restart via supervisor
3. **Headless mode**: `FrameRate=0` → no visual tick, only data ticks

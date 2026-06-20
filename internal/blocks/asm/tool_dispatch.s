#include "textflag.h"

// func asmToolDispatchJump(toolName string, args []byte) (string, error)
// Assembly jump table for tool dispatch. Falls back to Go.
TEXT ·asmToolDispatchJump(SB), NOSPLIT, $0-56
    // For now: direct fallback to Go dispatch
    // In production: cmp toolName against known tools, jump to kernel
    RET

#include "textflag.h"

// func sha256BlockAmd64(dig *[8]uint32, msg []byte)
// Hardware-accelerated SHA256 block transform using Intel SHA-NI.
// Processes one 64-byte block. Called from Go wrapper in a loop.
TEXT ·sha256BlockAmd64(SB), NOSPLIT, $0-32
    MOVQ dig+0(FP), DI       // pointer to 8 uint32 state
    MOVQ msg_base+8(FP), SI  // pointer to 64-byte block
    MOVQ msg_len+16(FP), CX  // block length (should be 64)

    // Load state into XMM registers
    MOVOU (DI), X0            // X0 = A,B,C,D
    MOVOU 16(DI), X1          // X1 = E,F,G,H

    // Load message into XMM
    MOVOU (SI), X2
    MOVOU 16(SI), X3
    MOVOU 32(SI), X4
    MOVOU 48(SI), X5

    // SHA256 4 rounds using SHA-NI
    SHA256RNDS2 X2, X0, X1
    SHA256RNDS2 X3, X0, X1
    SHA256RNDS2 X4, X0, X1
    SHA256RNDS2 X5, X0, X1

    // Store result back
    MOVOU X0, (DI)
    MOVOU X1, 16(DI)
    RET

// func sha256MsgSchedAmd64(msg *[16]uint32, i int)
// SHA256 message schedule expansion using SHA-NI.
TEXT ·sha256MsgSchedAmd64(SB), NOSPLIT, $0-16
    MOVQ msg+0(FP), DI
    MOVQ i+8(FP), AX
    
    MOVOU (DI), X0
    SHA256MSG1 X0, (DI)
    SHA256MSG2 X0, (DI)
    RET

#ifndef X_RUN_CMD_H
#define X_RUN_CMD_H

#include "typ.h"

typedef enum x_reg
{
    RSP = 0x0,
    RIP = 0x1,
} x_reg;

typedef enum x_cmd
{
    JMP = 0x00,              // jmp <b:8> <o:4>
    JAT = JMP | (0x01 << 5), // jat <o:4>
    JIF = JMP | (0x02 << 5), // jif <b:8> <o:4>
    NAT = JMP | (0x03 << 5), // nat <b:8>

    RET = 0x01, // ret
    LEA = 0x02, // lea <b:8> <o:4>

    PS = 0x03,
    PB = PS | (0x0 << 5), // pb <i:1>
    PW = PS | (0x1 << 5), // pw <i:4>
    PD = PS | (0x2 << 5), // pd <i:8>
    PA = PS | (0x3 << 5), // pa <s:4> <d:s>

    MR = 0x04,
    LR = MR | (0x0 << 5), // lr <r:1>
    SR = MR | (0x1 << 5), // sr <r:1>

    LB = 0x05,
    LW = 0x06,
    LD = 0x07,
    LA = 0x08,

    SB = 0x09,
    SW = 0x0a,
    SD = 0x0b,
    SA = 0x0c,

    LBI = LB, // lbi <b:8> <o:4>
    LWI = LW, // lwi <b:8> <o:4>
    LDI = LD, // ldi <b:8> <o:4>
    LAI = LA, // lai <b:8> <o:4> <s:4>

    SBI = SB, // sbi <b:8> <o:4>
    SWI = SW, // swi <b:8> <o:4>
    SDI = SD, // sdi <b:8> <o:4>
    SAI = SA, // sai <b:8> <o:4> <s:4>

    LBS = LB | (1 << 5), // lbs <o:4>
    LWS = LW | (1 << 5), // lws <o:4>
    LDS = LD | (1 << 5), // lds <o:4>
    LAS = LA | (1 << 5), // las <o:4> <s:4>

    SBS = SB | (1 << 5), // sbs <o:4>
    SWS = SW | (1 << 5), // sws <o:4>
    SDS = SD | (1 << 5), // sds <o:4>
    SAS = SA | (1 << 5), // sas <o:4> <s:4>

    CNV = 0x0d,
    I2U = CNV | (0x0 << 5), // i2u
    I2F = CNV | (0x1 << 5), // i2f
    U2I = CNV | (0x2 << 5), // u2i
    U2F = CNV | (0x3 << 5), // u2f
    F2I = CNV | (0x4 << 5), // f2i
    F2U = CNV | (0x5 << 5), // f2u

    ADD = 0x0e, // add
    SUB = 0x0f, // sub
    MUL = 0x10, // mul
    DIV = 0x11, // div
    POW = 0x12, // pow
    SHL = 0x13, // shl
    SHR = 0x14, // shr
    MOD = 0x15, // mod
    XOR = 0x16, // xor
    AND = 0x17, // and
    BOR = 0x18, // bor
    NOT = 0x19, // not
    NEG = 0x1a, // neg

    LT = 0x1b, // lt
    LE = 0x1c, // le
    EQ = 0x1d, // eq
    NE = 0x1e, // ne

    ADDI = ADD | I64,
    SUBI = SUB | I64,
    MULI = MUL | I64,
    DIVI = DIV | I64,
    POWI = POW | I64,
    SHLI = SHL | I64,
    SHRI = SHR | I64,
    MODI = MOD | I64,
    XORI = XOR | I64,
    ANDI = AND | I64,
    BORI = BOR | I64,
    NOTI = NOT | I64,
    NEGI = NEG | I64,

    LTI = LT | I64,
    LEI = LE | I64,
    EQI = EQ | I64,
    NEI = NE | I64,

    ADDU = ADD | U64,
    SUBU = SUB | U64,
    MULU = MUL | U64,
    DIVU = DIV | U64,
    POWU = POW | U64,
    SHLU = SHL | U64,
    SHRU = SHR | U64,
    MODU = MOD | U64,
    XORU = XOR | U64,
    ANDU = AND | U64,
    BORU = BOR | U64,
    NOTU = NOT | U64,

    LTU = LT | F64,
    LEU = LE | F64,
    EQU = EQ | F64,
    NEU = NE | F64,

    ADDF = ADD | F64,
    SUBF = SUB | F64,
    MULF = MUL | F64,
    DIVF = DIV | F64,
    POWF = POW | F64,
    NEGF = NEG | F64,

    LTF = LT | F64,
    LEF = LE | F64,
    EQF = EQ | F64,
    NEF = NE | F64,

    ANDB = AND | BOOL,
    BORB = BOR | BOOL,
    NOTB = NOT | BOOL,
} x_cmd;

#endif // X_RUN_CMD_H

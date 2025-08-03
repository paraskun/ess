#ifndef O2_RUN_CMD_H
#define O2_RUN_CMD_H

#include <o2/typ.h>

typedef enum o2_reg
{
    O2_RSP = 0x0,
    O2_RIP = 0x1,
    O2_RBP = 0x2,
} o2_reg;

typedef enum o2_cmd
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

    ADDI = ADD | O2_I64,
    SUBI = SUB | O2_I64,
    MULI = MUL | O2_I64,
    DIVI = DIV | O2_I64,
    POWI = POW | O2_I64,
    SHLI = SHL | O2_I64,
    SHRI = SHR | O2_I64,
    MODI = MOD | O2_I64,
    XORI = XOR | O2_I64,
    ANDI = AND | O2_I64,
    BORI = BOR | O2_I64,
    NOTI = NOT | O2_I64,
    NEGI = NEG | O2_I64,

    LTI = LT | O2_I64,
    LEI = LE | O2_I64,
    EQI = EQ | O2_I64,
    NEI = NE | O2_I64,

    ADDU = ADD | O2_U64,
    SUBU = SUB | O2_U64,
    MULU = MUL | O2_U64,
    DIVU = DIV | O2_U64,
    POWU = POW | O2_U64,
    SHLU = SHL | O2_U64,
    SHRU = SHR | O2_U64,
    MODU = MOD | O2_U64,
    XORU = XOR | O2_U64,
    ANDU = AND | O2_U64,
    BORU = BOR | O2_U64,
    NOTU = NOT | O2_U64,

    LTU = LT | O2_F64,
    LEU = LE | O2_F64,
    EQU = EQ | O2_F64,
    NEU = NE | O2_F64,

    ADDF = ADD | O2_F64,
    SUBF = SUB | O2_F64,
    MULF = MUL | O2_F64,
    DIVF = DIV | O2_F64,
    POWF = POW | O2_F64,
    NEGF = NEG | O2_F64,

    LTF = LT | O2_F64,
    LEF = LE | O2_F64,
    EQF = EQ | O2_F64,
    NEF = NE | O2_F64,

    ANDB = AND | O2_LOG,
    BORB = BOR | O2_LOG,
    NOTB = NOT | O2_LOG,
} o2_cmd;

#endif // O2_RUN_CMD_H

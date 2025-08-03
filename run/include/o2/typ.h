#ifndef O2_RUN_TYP_H
#define O2_RUN_TYP_H

typedef enum o2_typ
{
    O2_LOG = (0x0 << 5),
    O2_I64 = (0x1 << 5),
    O2_U64 = (0x2 << 5),
    O2_F64 = (0x3 << 5),
    O2_STR = (0x4 << 5),
} o2_typ;

#endif // O2_RUN_TYP_H

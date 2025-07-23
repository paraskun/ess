#ifndef X_RUN_TYP_H
#define X_RUN_TYP_H

typedef enum x_typ
{
    BOOl = (0x0 << 5),
    I64 = (0x1 << 5),
    U64 = (0x2 << 5),
    F64 = (0x3 << 5),
    STR = (0x4 << 5),
} x_typ;

#endif // X_RUN_TYP_H

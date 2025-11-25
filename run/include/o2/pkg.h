#ifndef X_RUN_PKG_H
#define X_RUN_PKG_H

#include <stdint.h>

typedef enum o2_seg
{
    TEXT = 0x0,
    DATA = 0x1,
} o2_seg;

typedef struct [[gnu::packed]] o2_sym
{
    uint32_t idf;
    uint32_t off; // INT_MAX for native

    o2_seg seg;
} o2_sym;

typedef struct [[gnu::packed]] o2_rel
{
    uint32_t mod;
    uint32_t pkg;
    uint32_t idf;
    uint32_t off;

    o2_seg seg;
} o2_rel;

typedef struct [[gnu::packed]] o2_pkg
{
    uint32_t psz; // package size
    uint32_t tsz; // text size
    uint32_t dsz; // data size
    uint32_t ssz; // symbol table size
    uint32_t rsz; // relocation table size

    uint8_t *text;
    uint8_t *data;

    struct o2_sym *sym;
    struct o2_rel *rel;

    void *nat; // native handler, maybe null
} o2_pkg;

int o2_pkg_lup(const char *mod, const char *pkg, struct o2_pkg **ptr);
int o2_pkg_rip(struct o2_pkg *pkg);

#endif // X_RUN_PKG_H

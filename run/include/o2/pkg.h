#ifndef X_RUN_PKG_H
#define X_RUN_PKG_H

#include <stdint.h>

// o2_seg is a type of package segment.
typedef enum o2_seg
{
    TEXT = 0x0,
    DATA = 0x1,
} o2_seg;

// o2_sym is a defined symbol entry.
typedef struct [[gnu::packed]] o2_sym
{
    uint32_t idf;
    uint32_t off; // INT_MAX for native

    o2_seg seg;
} o2_sym;

// o2_rel is a relocation entry.
typedef struct [[gnu::packed]] o2_rel
{
    uint32_t mod;
    uint32_t pkg;
    uint32_t idf;
    uint32_t off;

    o2_seg seg;
} o2_rel;

// o2_pkg is a precompiled package.
typedef struct [[gnu::packed]] o2_pkg
{
    uint32_t psz;
    uint32_t tsz;
    uint32_t dsz;
    uint32_t ssz;
    uint32_t rsz;

    uint8_t *text;
    uint8_t *data;

    struct o2_sym *sym;
    struct o2_rel *rel;

    void *nat; // native handler, maybe null
} o2_pkg;

/**
 *  o2_pkg_lup used to lookup for a precompiled package by the given fully qualified name.
 */
int o2_pkg_lup(const char *mod, const char *pkg, struct o2_pkg **ptr);

/**
 *  pkg_rip used to unmount given package.
 */
int o2_pkg_rip(struct o2_pkg *pkg);

#endif // X_RUN_PKG_H

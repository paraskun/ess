#ifndef X_RUN_PKG_H
#define X_RUN_PKG_H

#include <stdint.h>

typedef enum o2_seg
{
    TEXT = 0x0,
    DATA = 0x1,
} o2_seg;

typedef struct o2_sym
{
    uint32_t idf;
    uint32_t off; // INT_MAX for native

    o2_seg seg;
} o2_sym;

typedef struct o2_rel
{
    uint32_t pkg;
    uint32_t idf;
    uint32_t off;

    o2_seg seg;
} o2_rel;

typedef struct o2_pkg
{
    uint32_t psz;
    uint32_t tsz;
    uint32_t dsz;
    uint32_t ssz;
    uint32_t rsz;

    uint8_t *name; // name = data (first string)
    uint8_t *text;
    uint8_t *data;

    struct o2_sym *sym; // symbol table
    struct o2_rel *rel; // relocation table

    void *nat; // native handler, maybe null
} o2_pkg;

/**
 *  pkg_srh used to search package at given endpoint.
 */
int o2_pkg_srh(const char *url, struct o2_pkg **pkg);

/**
 *  pkg_map used to setup segment pointers.
 */
int o2_pkg_map(struct o2_pkg *pkg);

/**
 *  pkg_rip used to unmount given package.
 */
int o2_pkg_rip(struct o2_pkg *pkg);

#endif // X_RUN_PKG_H

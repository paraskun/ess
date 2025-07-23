#ifndef X_RUN_PKG_H
#define X_RUN_PKG_H

#include <stdint.h>

typedef enum s_seg
{
    TEXT = 0x0,
    DATA = 0x1,
} s_seg;

typedef struct x_sym
{
    uint32_t idf;
    uint32_t off; // INT_MAX for native

    s_seg seg;
} x_sym;

typedef struct x_rel
{
    uint32_t pkg;
    uint32_t idf;
    uint32_t off;

    s_seg seg;
} x_rel;

typedef struct x_pkg
{
    uint32_t psz;
    uint32_t tsz;
    uint32_t dsz;
    uint32_t ssz;
    uint32_t rsz;

    uint8_t *name; // name = data (first string)
    uint8_t *text;
    uint8_t *data;

    struct x_sym *sym; // symbol table
    struct x_rel *rel; // relocation table

    void *nat; // native handler, maybe null
} x_pkg;

/**
 *  pkg_srh used to search package at given endpoint.
 */
int pkg_srh(const char *url, struct x_pkg **pkg);

/**
 *  pkg_map used to setup segment pointers.
 */
int pkg_map(struct x_pkg *pkg);

/**
 *  pkg_rip used to unmount given package.
 */
int pkg_rip(struct x_pkg *pkg);

#endif // X_RUN_PKG_H

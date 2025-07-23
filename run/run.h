#ifndef X_RUN_RUN_H
#define X_RUN_RUN_H

#include "pkg.h"

typedef struct x_ctx
{
    int psz;
    int rsz;

    struct x_pkg **pkg;
    struct x_run **run;
} x_ctx;

int ctx_add(struct x_ctx *ctx, struct x_pkg *pkg);

typedef struct x_run
{
    struct x_ctx *ctx;
    struct x_pkg *pkg;
} x_run;

int run_new(struct x_run *run, struct x_ctx *ctx, struct x_pkg *pkg, const char *ent);
int run_exe(struct x_run *run, uint8_t *arg, uint8_t **ret);
int run_rip(struct x_run *run);

#endif // X_RUN_RUN_H

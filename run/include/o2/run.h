#ifndef O2_RUN_RUN_H
#define O2_RUN_RUN_H

#include <o2/pkg.h>

typedef struct o2_ctx
{
    int psz;
    int rsz;

    struct o2_pkg **pkg;
    struct o2_run **run;
} o2_ctx;

int o2_ctx_add(struct o2_ctx *ctx, struct o2_pkg *pkg);

typedef struct o2_run
{
    struct o2_ctx *ctx;
    struct o2_pkg *pkg;
} o2_run;

int o2_run_new(struct o2_run *run, struct o2_ctx *ctx, struct o2_pkg *pkg, const char *ent);
int o2_run_exe(struct o2_run *run, uint8_t *arg, uint8_t **ret);
int o2_run_rip(struct o2_run *run);

#endif // O2_RUN_RUN_H

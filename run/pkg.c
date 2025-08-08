#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <o2/pkg.h>

static FILE *pkg_open(const char *mod, const char *pkg);
static void *nat_open(const char *path, uint32_t size, uint8_t *data);

int o2_pkg_lup(const char *mod, const char *pkg, struct o2_pkg **ptr)
{
    FILE *f = pkg_open(mod, pkg);

    if (f == NULL) {
        return -1;
    }

    uint32_t psz;

    if (fread(&psz, 4, 1, f) != 1) {
        fclose(f);
        return -1;
    }

    struct o2_pkg *p = malloc(psz);

    if (p == NULL) {
        fclose(f);
        return -1;
    }

    if (fread((void *)p, 1, psz, f) != psz) {
        fclose(f);
        free(p);
        return -1;
    }

    fclose(f);

    uint8_t *data = (void *)p + sizeof(struct o2_pkg);

    p->text = data;
    p->data = p->text + p->tsz;

    p->sym = (struct o2_sym *)(p->data + p->dsz);
    p->rel = (struct o2_rel *)(p->sym + p->ssz);

    uint32_t nsz = p->psz - p->tsz - p->dsz - p->ssz - p->rsz;

    if (nsz != 0) {
        char path[256] = "/tmp/o2/nat/";

        strcat(path, mod);
        strcat(path, "/");
        strcat(path, pkg);
        strcat(path, ".so");

        p->nat = nat_open(path, nsz, (uint8_t *)p->rel + p->rsz);

        if (p->nat == NULL) {
            free(p);
            return -1;
        }
    }

    *ptr = p;

    return 0;
}

static int mod_load(const char *mod);

static FILE *pkg_open(const char *mod, const char *pkg)
{
    char *home = getenv("HOME");

    if (home == NULL) {
        return NULL;
    }

    char path[256] = {0};

    strcat(path, home);
    strcat(path, "/o2/lib/");
    strcat(path, mod);
    strcat(path, "/");
    strcat(path, pkg);
    strcat(path, ".co2");

    FILE *f = fopen(path, "r");

    if (f == NULL) {
        if (mod_load(mod) == -1) {
            return NULL;
        }

        f = fopen(path, "r");
    }

    return f;
}

static int mod_load(const char *mod)
{
    char url[256] = "https://";

    const char *host = mod;
    const char *user = strchr(host, '/') + 1;
    const char *repo = strchr(user, '/') + 1;
    const char *path = strchr(repo, '/') + 1;
    const char *ver = strchr(user, '@') + 1;

    strncat(url, host, user - host);
    strncat(url, user, repo - user);

    if (path != NULL) {
        strncat(url, repo, path - repo);
    } else {
        strncat(url, repo, ver - repo);
    }

    strcat(url, "/archive/refs/tags/");

    if (path != NULL) {
        strncat(url, path, ver - path);
    }

    strcat(url, "-");
    strcat(url, ver);

    // download module

    return 0;
}

static void *nat_open(const char *path, uint32_t size, uint8_t *data)
{
    FILE *f = fopen(path, "wb");

    if (f == NULL) {
        return NULL;
    }

    if (fwrite(data, 1, size, f) != size) {
        fclose(f);
        return NULL;
    }

    fclose(f);

    return dlopen(path, RTLD_NOW);
}

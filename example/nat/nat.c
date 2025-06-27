#include <stdint.h>

typedef int64_t i64;
typedef uint64_t u64;

i64 add(i64 a, i64 b) { return a + b; }
i64 sub(i64 a, i64 b) { return a - b; }

struct [[gnu::packed]] ctx {
  u64 num;
  i64 sum;
};

i64 mean(void *ctx, i64 a) {
  struct ctx* c = (struct ctx*)ctx;

  c->num += 1;
  c->sum += a;

  return c->sum / c->num;
}

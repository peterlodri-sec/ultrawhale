#include <mimalloc.h>
#include <stdlib.h>
#include <string.h>

// Override standard allocator symbols — mimalloc takes over all allocations
void* malloc(size_t size)          { return mi_malloc(size); }
void* calloc(size_t n, size_t sz)  { return mi_calloc(n, sz); }
void* realloc(void* p, size_t sz)  { return mi_realloc(p, sz); }
void  free(void* p)                { mi_free(p); }
void* aligned_alloc(size_t a, size_t sz) { return mi_aligned_alloc(a, sz); }
int   posix_memalign(void** m, size_t a, size_t sz) {
    *m = mi_aligned_alloc(a, sz);
    return *m ? 0 : 12; // ENOMEM
}
char* strdup(const char* s) {
    size_t len = strlen(s) + 1;
    char* d = mi_malloc(len);
    return d ? memcpy(d, s, len) : NULL;
}

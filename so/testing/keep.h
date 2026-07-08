// so_keep forces the object at p to be materialized in a register or memory and
// marks memory as clobbered. This prevents the C optimizer from deleting
// benchmarked code whose result is otherwise unused, and makes the address of a
// stack local "escape" so operations on it cannot be proven dead. It emits no
// instructions. The T macro argument (the pointed-to C type) is unused; the
// pointer operand carries the type.
#define so_keep(T, p) \
    __asm__ volatile("" : : "r,m"(*(p)) : "memory")

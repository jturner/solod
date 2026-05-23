#include "main.h"

// -- Implementation --

int main(void) {
    pkg1_T1 t1 = (pkg1_T1){.Val = 42};
    pkg2_T2 t2 = (pkg2_T2){.Val = 42};
    if (t1.Val != t2.Val) {
        so_panic("t1 != t2");
    }
    return 0;
}

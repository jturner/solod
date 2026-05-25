#include "main.h"

// -- Forward declarations --
static void regularGoto(void);
static void labeledLoop(void);
static void labeledBreak(void);

// -- Implementation --

static void regularGoto(void) {
    so_int fails = 0;
    for (so_int i = 0; i < 10; i++) {
        if (i % 2 == 0) {
            goto next;
        }
        next:;
        fails++;
        if (fails > 2) {
            goto fallback;
        }
    }
    fallback:;
    if (fails != 3) {
        so_panic("fails != 3");
    }
}

static void labeledLoop(void) {
    so_int x = 0;
    outer:;
    for (so_int _i = 0; _i < 5; _i++) {
        x++;
    }
    outer_end:;
    if (x < 10) {
        goto outer;
    }
    if (x != 10) {
        so_panic("x != 10");
    }
}

static void labeledBreak(void) {
    so_int sum = 0;
    outer:;
    for (so_int i = 0; i < 5; i++) {
        for (so_int j = 0; j < 5; j++) {
            if (i + j > 3) {
                goto outer_end;
            }
            sum += i + j;
        }
    }
    outer_end:;
    if (sum != 6) {
        so_panic("sum != 6");
    }
}

int main(void) {
    regularGoto();
    labeledLoop();
    labeledBreak();
    return 0;
}

#include "value.h"
#include <stdio.h>

Tagged make_int(int x) {
    Tagged t;
    t.type = 'i';
    t.val.i = x;
    return t;
}

Tagged make_float(float x) {
    Tagged t;
    t.type = 'f';
    t.val.f = x;
    return t;
}

int tagged_as_int(Tagged t)     { return t.val.i; }
float tagged_as_float(Tagged t) { return t.val.f; }

void tagged_print(Tagged t) {
    if (t.type == 'i') {
        printf("[C] Tagged{type='i', val=%d}\n", t.val.i);
    } else {
        printf("[C] Tagged{type='f', val=%.2f}\n", t.val.f);
    }
    fflush(stdout);
}

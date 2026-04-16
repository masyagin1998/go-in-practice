#include <stdio.h>
#include "_cgo_export.h"

// add складывает два числа (чистая C-функция).
int add(int a, int b) {
    return a + b;
}

// orchestrate вызывает Go-функцию GoCompute из C.
// GoCompute, в свою очередь, вызывает C-функцию add.
void orchestrate(int x, int y) {
    printf("orchestrate(): calling GoCompute(%d, %d) from C...\n", x, y);
    fflush(stdout);
    GoCompute(x, y);
}

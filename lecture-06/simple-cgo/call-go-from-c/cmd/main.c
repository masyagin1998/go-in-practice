// main.c — точка входа на Си.
// Вызывает Go-функцию GoGreet, экспортированную через c-archive.
#include <stdio.h>
#include "libgogreet.h"

int main() {
    printf("main(): calling GoGreet from C...\n");
    fflush(stdout);
    GoGreet();
    printf("main(): done\n");
    return 0;
}

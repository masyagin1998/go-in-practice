#include <stdio.h>

// greet печатает приветствие в stdout.
void greet(const char* name) {
    printf("Hello, %s! (from C)\n", name);
    fflush(stdout);
}

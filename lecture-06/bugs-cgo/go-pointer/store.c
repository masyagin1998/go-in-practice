#include "store.h"
#include <stdio.h>

static void* g_ptr = NULL;

void store_ptr(void* ptr) {
    g_ptr = ptr;
    printf("[C] сохранили указатель: %p\n", ptr);
    fflush(stdout);
}

void print_ptr(void) {
    printf("[C] указатель: %p\n", g_ptr);
    fflush(stdout);
}

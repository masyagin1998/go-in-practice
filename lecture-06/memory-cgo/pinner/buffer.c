#include "buffer.h"
#include <stdio.h>

// Глобальный указатель на данные и их размер.
// Указывает на Go-память — обязательно закреплённую через runtime.Pinner.
static int*   g_data = NULL;
static size_t g_len  = 0;

// buffer_set сохраняет указатель в глобальной переменной.
void buffer_set(int* data, size_t len) {
    g_data = data;
    g_len  = len;
    printf("[C] buffer_set: сохранили указатель %p, len=%zu\n", (void*)data, len);
    fflush(stdout);
}

// buffer_print печатает содержимое буфера из глобального указателя.
void buffer_print(void) {
    printf("[C] buffer(%zu): [", g_len);
    for (size_t i = 0; i < g_len; i++) {
        if (i > 0) printf(", ");
        printf("%d", g_data[i]);
    }
    printf("]\n");
    fflush(stdout);
}

// buffer_scale умножает все элементы буфера на factor.
void buffer_scale(int factor) {
    for (size_t i = 0; i < g_len; i++) {
        g_data[i] *= factor;
    }
}

// Линейный аллокатор — самый простой: сдвигаем указатель вперёд, free нет.

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

#define TAG "[LINEAR] "

typedef struct {
    char  *buf;   // начало буфера
    size_t used;  // сколько занято
    size_t cap;   // общий размер
} LinearAlloc;

// init — создаём аллокатор с буфером заданного размера
void la_init(LinearAlloc *a, size_t size) {
    a->buf  = malloc(size);
    a->used = 0;
    a->cap  = size;
    printf(TAG "init: buf=%p, size=%zu\n", a->buf, size);
}

// alloc — выделяем size байт, сдвигая указатель вперёд
void *la_alloc(LinearAlloc *a, size_t size) {
    if (a->used + size > a->cap) {
        printf(TAG "alloc: FAIL, size=%zu, no space (used=%zu/%zu)\n",
               size, a->used, a->cap);
        return NULL;
    }
    void *ptr = a->buf + a->used;
    a->used += size;
    printf(TAG "alloc: ptr=%p, size=%zu, used=%zu/%zu\n",
           ptr, size, a->used, a->cap);
    return ptr;
}

// free — не поддерживается :(
// (как на слайде: "Free не будет")

// destroy — освобождаем весь буфер
void la_destroy(LinearAlloc *a) {
    printf(TAG "destroy: buf=%p, was used=%zu/%zu\n",
           a->buf, a->used, a->cap);
    free(a->buf);
    a->buf  = NULL;
    a->used = 0;
    a->cap  = 0;
}

int main(void) {
    LinearAlloc a;
    la_init(&a, 256);

    void *p1 = la_alloc(&a, 64);
    void *p2 = la_alloc(&a, 100);
    la_alloc(&a, 50);
    la_alloc(&a, 100); // не поместится

    // Записываем данные для проверки
    if (p1) *(int *)p1 = 42;
    if (p2) *(int *)p2 = 99;
    printf(TAG "check: *p1=%d, *p2=%d\n", *(int *)p1, *(int *)p2);

    la_destroy(&a);
    return 0;
}

// Стековый аллокатор — перед каждым выделением кладём заголовок (header)
// с размером. Free работает только для последнего блока (LIFO).

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

#define TAG "[STACK]  "

// Заголовок перед каждым выделением
typedef struct {
    size_t size; // размер пользовательских данных (без header)
} Header;

typedef struct {
    char  *buf;
    size_t used;
    size_t cap;
} StackAlloc;

// init — создаём аллокатор
void sa_init(StackAlloc *a, size_t size) {
    a->buf  = malloc(size);
    a->used = 0;
    a->cap  = size;
    printf(TAG "init: buf=%p, size=%zu\n", a->buf, size);
}

// alloc — пишем header + выделяем size байт
void *sa_alloc(StackAlloc *a, size_t size) {
    size_t total = sizeof(Header) + size;
    if (a->used + total > a->cap) {
        printf(TAG "alloc: FAIL, size=%zu, no space (used=%zu/%zu)\n",
               size, a->used, a->cap);
        return NULL;
    }

    Header *hdr = (Header *)(a->buf + a->used);
    hdr->size = size;
    void *ptr = (char *)hdr + sizeof(Header);
    a->used += total;

    printf(TAG "alloc: ptr=%p, size=%zu, used=%zu/%zu (header at %p)\n",
           ptr, size, a->used, a->cap, (void *)hdr);
    return ptr;
}

// free — освобождаем последний выделенный блок (LIFO)
void sa_free(StackAlloc *a, void *ptr) {
    Header *hdr = (Header *)((char *)ptr - sizeof(Header));

    // Проверяем что это действительно последний блок
    size_t expected_used = (char *)ptr + hdr->size - a->buf;
    if (expected_used != a->used) {
        printf(TAG "free: FAIL, ptr=%p is not the top block!\n", ptr);
        return;
    }

    a->used -= sizeof(Header) + hdr->size;
    printf(TAG "free: ptr=%p, size=%zu, used=%zu/%zu\n",
           ptr, hdr->size, a->used, a->cap);
}

// destroy — освобождаем буфер
void sa_destroy(StackAlloc *a) {
    printf(TAG "destroy: buf=%p, was used=%zu/%zu\n",
           a->buf, a->used, a->cap);
    free(a->buf);
    a->buf  = NULL;
    a->used = 0;
    a->cap  = 0;
}

int main(void) {
    StackAlloc a;
    sa_init(&a, 256);

    void *p1 = sa_alloc(&a, 32);
    void *p2 = sa_alloc(&a, 64);
    void *p3 = sa_alloc(&a, 16);

    if (p1) *(int *)p1 = 111;
    if (p3) *(int *)p3 = 333;
    printf(TAG "check: *p1=%d, *p3=%d\n", *(int *)p1, *(int *)p3);

    // Освобождаем в обратном порядке (LIFO)
    sa_free(&a, p3);
    sa_free(&a, p2);

    // Попробуем освободить не-верхний блок
    void *p4 = sa_alloc(&a, 48);
    sa_free(&a, p1); // ошибка — p1 не на вершине

    sa_free(&a, p4); // ок
    sa_free(&a, p1); // теперь p1 на вершине

    sa_destroy(&a);
    return 0;
}

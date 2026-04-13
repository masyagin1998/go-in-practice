// Блочный аллокатор — делим буфер на блоки фиксированного размера,
// свободные блоки храним в двусвязном списке.

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

#define TAG "[BLOCK]  "

// Указатели prev/next лежат прямо внутри свободного блока
// (блок всё равно не используется — классический приём).
typedef struct FreeNode {
    struct FreeNode *prev;
    struct FreeNode *next;
} FreeNode;

typedef struct {
    char     *buf;        // начало буфера
    size_t    block_size; // размер одного блока
    size_t    count;      // общее кол-во блоков
    size_t    free_count; // кол-во свободных
    FreeNode *free_list;  // голова двусвязного списка
} BlockAlloc;

// init — создаём буфер и нарезаем на блоки
void ba_init(BlockAlloc *a, size_t block_size, size_t count) {
    // Блок должен вмещать хотя бы два указателя
    if (block_size < sizeof(FreeNode))
        block_size = sizeof(FreeNode);

    a->buf        = malloc(block_size * count);
    a->block_size = block_size;
    a->count      = count;
    a->free_count = count;
    a->free_list  = NULL;

    // Строим двусвязный список свободных блоков
    for (size_t i = 0; i < count; i++) {
        FreeNode *node = (FreeNode *)(a->buf + i * block_size);
        node->prev = NULL;
        node->next = a->free_list;
        if (a->free_list)
            a->free_list->prev = node;
        a->free_list = node;
    }

    printf(TAG "init: buf=%p, block_size=%zu, count=%zu\n",
           a->buf, block_size, count);
}

// alloc — снимаем голову свободного списка
void *ba_alloc(BlockAlloc *a) {
    if (!a->free_list) {
        printf(TAG "alloc: FAIL, no free blocks\n");
        return NULL;
    }

    FreeNode *node = a->free_list;
    a->free_list = node->next;
    if (a->free_list)
        a->free_list->prev = NULL;
    a->free_count--;

    printf(TAG "alloc: ptr=%p, free=%zu/%zu\n",
           (void *)node, a->free_count, a->count);
    return (void *)node;
}

// free — возвращаем блок в голову списка
void ba_free(BlockAlloc *a, void *ptr) {
    FreeNode *node = (FreeNode *)ptr;
    node->prev = NULL;
    node->next = a->free_list;
    if (a->free_list)
        a->free_list->prev = node;
    a->free_list = node;
    a->free_count++;

    printf(TAG "free: ptr=%p, free=%zu/%zu\n",
           ptr, a->free_count, a->count);
}

// destroy — освобождаем буфер
void ba_destroy(BlockAlloc *a) {
    printf(TAG "destroy: buf=%p, free=%zu/%zu\n",
           a->buf, a->free_count, a->count);
    free(a->buf);
    a->buf       = NULL;
    a->free_list = NULL;
}

int main(void) {
    BlockAlloc a;
    ba_init(&a, 64, 8);

    void *p1 = ba_alloc(&a);
    void *p2 = ba_alloc(&a);
    void *p3 = ba_alloc(&a);

    // Записываем данные
    if (p1) *(int *)p1 = 111;
    if (p2) *(int *)p2 = 222;
    printf(TAG "check: *p1=%d, *p2=%d\n", *(int *)p1, *(int *)p2);

    ba_free(&a, p2);
    void *p4 = ba_alloc(&a); // получим обратно блок p2
    printf(TAG "reuse: p2 was %p, p4 is %p\n", p2, p4);

    ba_free(&a, p1);
    ba_free(&a, p3);
    ba_free(&a, p4);

    ba_destroy(&a);
    return 0;
}

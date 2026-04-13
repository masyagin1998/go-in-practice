// Stupid malloc — комбинация стекового (header), блочного (free list)
// и линейного (рост вперёд) аллокаторов. Работает поверх brk или mmap.
// Может служить заменой стандартного malloc в single-threaded программе.

#define _DEFAULT_SOURCE
#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <sys/mman.h>

#define TAG "[STUPID] "
#define MAX_BLOCK (64 * 1024) // 64 КБ — размер глобального блока

// Выбор бэкенда: USE_BRK=1 — brk, USE_BRK=0 — mmap
#ifndef USE_BRK
#define USE_BRK 0
#endif

// ---------------------------------------------------------------
// Заголовок каждого блока (двусвязный список всех блоков)
// ---------------------------------------------------------------
typedef struct Block {
    size_t        size;    // размер пользовательских данных
    int           is_free; // 1 — свободен, 0 — занят
    struct Block *prev;
    struct Block *next;
} Block;

#define HEADER_SIZE (sizeof(Block))
#define BLOCK_DATA(b) ((void *)((char *)(b) + HEADER_SIZE))
#define DATA_BLOCK(p) ((Block *)((char *)(p) - HEADER_SIZE))

// Глобальное состояние аллокатора
static Block *head = NULL; // голова списка всех блоков
static Block *tail = NULL; // хвост — для быстрого расширения

// ---------------------------------------------------------------
// Запрос памяти у системы
// ---------------------------------------------------------------
static void *request_memory(size_t size) {
#if USE_BRK
    void *ptr = sbrk(size);
    if (ptr == (void *)-1) return NULL;
    printf(TAG "sbrk: requested %zu bytes, got %p\n", size, ptr);
    return ptr;
#else
    void *ptr = mmap(NULL, size, PROT_READ | PROT_WRITE,
                     MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (ptr == MAP_FAILED) return NULL;
    printf(TAG "mmap: requested %zu bytes, got %p\n", size, ptr);
    return ptr;
#endif
}

// ---------------------------------------------------------------
// Поиск свободного блока (first-fit)
// ---------------------------------------------------------------
static Block *find_free(size_t size) {
    for (Block *b = head; b; b = b->next) {
        if (b->is_free && b->size >= size)
            return b;
    }
    return NULL;
}

// ---------------------------------------------------------------
// Разделение блока: если свободный блок сильно больше запроса,
// отрезаем от него нужный кусок.
// ---------------------------------------------------------------
static void split_block(Block *b, size_t size) {
    // Разделяем только если остаток вмещает header + хотя бы 16 байт
    if (b->size < size + HEADER_SIZE + 16)
        return;

    Block *new_block = (Block *)((char *)b + HEADER_SIZE + size);
    new_block->size    = b->size - size - HEADER_SIZE;
    new_block->is_free = 1;
    new_block->prev    = b;
    new_block->next    = b->next;

    if (b->next)
        b->next->prev = new_block;
    else
        tail = new_block;

    b->next = new_block;
    b->size = size;

    printf(TAG "split: block %p (size=%zu) + new free block %p (size=%zu)\n",
           (void *)b, b->size, (void *)new_block, new_block->size);
}

// ---------------------------------------------------------------
// Выделение нового глобального блока
// ---------------------------------------------------------------
static Block *grow(size_t size) {
    // Запрашиваем не меньше MAX_BLOCK
    size_t total = HEADER_SIZE + size;
    if (total < MAX_BLOCK)
        total = MAX_BLOCK;

    void *mem = request_memory(total);
    if (!mem) return NULL;

    Block *b = (Block *)mem;
    b->size    = total - HEADER_SIZE;
    b->is_free = 0;
    b->prev    = tail;
    b->next    = NULL;

    if (tail)
        tail->next = b;
    else
        head = b;
    tail = b;

    // Если выделили больше чем нужно — разделяем
    split_block(b, size);
    return b;
}

// ---------------------------------------------------------------
// stupid_malloc
// ---------------------------------------------------------------
void *stupid_malloc(size_t size) {
    if (size == 0) return NULL;

    // Выравнивание до 8 байт
    size = (size + 7) & ~(size_t)7;

    // Ищем подходящий свободный блок
    Block *b = find_free(size);
    if (b) {
        b->is_free = 0;
        split_block(b, size);
        printf(TAG "alloc: ptr=%p, size=%zu (reused free block)\n",
               BLOCK_DATA(b), size);
        return BLOCK_DATA(b);
    }

    // Не нашли — растём
    b = grow(size);
    if (!b) {
        printf(TAG "alloc: FAIL, size=%zu\n", size);
        return NULL;
    }

    b->is_free = 0;
    printf(TAG "alloc: ptr=%p, size=%zu (new block)\n",
           BLOCK_DATA(b), size);
    return BLOCK_DATA(b);
}

// ---------------------------------------------------------------
// Дефрагментация (coalescing) — сливаем с соседями
// ---------------------------------------------------------------
// Проверка физической смежности двух блоков
static int adjacent(Block *a, Block *b) {
    return (char *)a + HEADER_SIZE + a->size == (char *)b;
}

static void coalesce(Block *b) {
    // Сливаем с правым соседом (только если физически смежны)
    if (b->next && b->next->is_free && adjacent(b, b->next)) {
        Block *right = b->next;
        b->size += HEADER_SIZE + right->size;
        b->next = right->next;
        if (right->next)
            right->next->prev = b;
        else
            tail = b;
        printf(TAG "coalesce: merged with next, new_size=%zu\n", b->size);
    }

    // Сливаем с левым соседом (только если физически смежны)
    if (b->prev && b->prev->is_free && adjacent(b->prev, b)) {
        Block *left = b->prev;
        left->size += HEADER_SIZE + b->size;
        left->next = b->next;
        if (b->next)
            b->next->prev = left;
        else
            tail = left;
        printf(TAG "coalesce: merged with prev, new_size=%zu\n", left->size);
    }
}

// ---------------------------------------------------------------
// stupid_free
// ---------------------------------------------------------------
void stupid_free(void *ptr) {
    if (!ptr) return;

    Block *b = DATA_BLOCK(ptr);
    b->is_free = 1;
    printf(TAG "free: ptr=%p, size=%zu\n", ptr, b->size);

    coalesce(b);
}

// ---------------------------------------------------------------
// stupid_destroy — освобождаем всё
// ---------------------------------------------------------------
void stupid_destroy(void) {
    printf(TAG "destroy: releasing all blocks\n");

#if !USE_BRK
    // Для mmap нужно отдать каждый регион отдельно.
    // Т.к. grow() делает отдельный mmap на каждый запрос,
    // нужно найти границы каждого mmap-региона.
    // Упрощённый вариант: проходим по списку и munmap от head каждого
    // глобального блока (блоки внутри одного mmap-региона смежны).
    Block *b = head;
    while (b) {
        // Ищем конец непрерывного региона
        Block *start = b;
        while (b->next &&
               (char *)b + HEADER_SIZE + b->size == (char *)b->next) {
            b = b->next;
        }
        Block *next = b->next;
        size_t region_size = (char *)b + HEADER_SIZE + b->size - (char *)start;
        printf(TAG "munmap: %p, size=%zu\n", (void *)start, region_size);
        munmap(start, region_size);
        b = next;
    }
#else
    // Для brk достаточно сбросить указатель на начало
    if (head) {
        printf(TAG "brk: resetting to %p\n", (void *)head);
        brk(head);
    }
#endif
    head = NULL;
    tail = NULL;
}

// ---------------------------------------------------------------
// Демо
// ---------------------------------------------------------------
int main(void) {
    printf("=== stupid_malloc demo (backend: %s) ===\n\n",
           USE_BRK ? "brk" : "mmap");

    void *p1 = stupid_malloc(128);
    void *p2 = stupid_malloc(256);
    void *p3 = stupid_malloc(64);

    // Записываем данные
    if (p1) memset(p1, 'A', 128);
    if (p2) memset(p2, 'B', 256);
    if (p3) *(int *)p3 = 42;
    printf(TAG "check: p3 value = %d\n\n", *(int *)p3);

    // Освобождаем средний блок — должен слиться с соседями
    printf("--- free p2 ---\n");
    stupid_free(p2);

    // Выделяем что-то, что поместится в освобождённый блок
    printf("\n--- alloc 100 (should reuse p2's block) ---\n");
    void *p4 = stupid_malloc(100);

    // Освобождаем всё по одному
    printf("\n--- free all ---\n");
    stupid_free(p1);
    stupid_free(p3);
    stupid_free(p4);

    // Большой запрос (> MAX_BLOCK)
    printf("\n--- large alloc (128KB) ---\n");
    void *big = stupid_malloc(128 * 1024);
    if (big) {
        memset(big, 'X', 128 * 1024);
        printf(TAG "big block ok, ptr=%p\n", big);
    }
    stupid_free(big);

    printf("\n--- destroy ---\n");
    stupid_destroy();

    return 0;
}

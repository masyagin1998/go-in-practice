// XOR-связный список + Boehm GC — сборщик мусора ломает список.
//
// Boehm GC — консервативный сборщик мусора для C/C++. Он сканирует
// стек и кучу в поисках значений, похожих на указатели. Но в XOR-списке
// ни один узел не хранит «настоящий» указатель на соседей — только
// prev XOR next. GC не распознаёт эти замаскированные указатели
// и считает узлы недостижимыми → собирает их → список разрушается.
//
// Сборка (нужен libgc-dev):
//   gcc -Wall -Wextra -g -o 02_boehm_gc 02_boehm_gc.c -lgc

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include <gc/gc.h>

// === XOR-связный список (аллокация через GC) ===

typedef struct Node {
    int value;
    uintptr_t both; // prev XOR next — «невидимый» для GC указатель.
} Node;

// Вставка в конец. Память выделяем через GC_MALLOC.
Node *xor_push_back(Node **head, Node **tail, int value) {
    Node *node = (Node *)GC_MALLOC(sizeof(Node));
    if (!node) {
        fprintf(stderr, "GC_MALLOC failed\n");
        exit(1);
    }
    node->value = value;
    node->both = (uintptr_t)(*tail);

    if (*tail) {
        (*tail)->both = (*tail)->both ^ (uintptr_t)node;
    } else {
        *head = node;
    }
    *tail = node;
    return node;
}

// Проход вперёд с печатью.
void xor_traverse_forward(Node *head) {
    Node *prev = NULL;
    Node *curr = head;
    int count = 0;
    printf("  Проход вперёд:");
    while (curr && count < 10) { // Лимит на случай повреждения списка.
        printf(" %d", curr->value);
        Node *next = (Node *)(curr->both ^ (uintptr_t)prev);
        prev = curr;
        curr = next;
        count++;
    }
    if (count >= 10 && curr) {
        printf(" ... (обрезано, возможно повреждение)");
    }
    printf("\n");
}

int main(void) {
    printf("=== XOR-список + Boehm GC ===\n\n");

    GC_INIT();

    Node *head = NULL;
    Node *tail = NULL;

    // Заполняем список.
    for (int i = 1; i <= 5; i++) {
        xor_push_back(&head, &tail, i * 10);
    }
    printf("  Список создан: 10 → 20 → 30 → 40 → 50\n\n");

    // До GC — список ещё цел (узлы не успели собраться).
    printf("  [до GC]\n");
    xor_traverse_forward(head);

    // Создаём давление на память, чтобы спровоцировать сборку.
    // Промежуточные узлы доступны только через XOR — GC их не видит.
    printf("\n  Вызываем GC_gcollect()...\n");
    GC_gcollect();

    // Выделяем ещё памяти, чтобы GC точно переиспользовал освобождённые блоки.
    for (int i = 0; i < 10000; i++) {
        volatile void *p = GC_MALLOC(sizeof(Node));
        (void)p;
    }
    GC_gcollect();

    printf("\n  [после GC — проход 1]\n");
    xor_traverse_forward(head);

    printf("\n  [после GC — проход 2]\n");
    xor_traverse_forward(head);

    printf("\n  GC собрал узлы, на которые нет «нормальных» указателей.\n");
    printf("  XOR-маскировка сделала их невидимыми для сборщика.\n");

    return 0;
}

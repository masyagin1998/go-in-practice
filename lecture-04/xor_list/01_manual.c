// XOR-связный список на чистом C (ручное управление памятью).
//
// В XOR-списке каждый узел хранит не два указателя (prev, next),
// а одно значение: prev XOR next. Это экономит память, но делает
// указатели «невидимыми» для любого сборщика мусора.
//
// При ручном управлении памятью (malloc/free) всё работает корректно,
// потому что мы сами контролируем время жизни каждого узла.

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

// === XOR-связный список ===

typedef struct Node {
    int value;
    uintptr_t both; // prev XOR next
} Node;

// Вставка в конец списка. Возвращает новый tail.
Node *xor_push_back(Node **head, Node **tail, int value) {
    Node *node = (Node *)malloc(sizeof(Node));
    if (!node) {
        perror("malloc");
        exit(1);
    }
    node->value = value;
    node->both = (uintptr_t)(*tail); // prev=tail, next=NULL → tail XOR 0

    if (*tail) {
        // Старый tail: both = prev XOR NULL → нужно заменить на prev XOR node.
        // prev = tail->both XOR NULL = tail->both
        (*tail)->both = (*tail)->both ^ (uintptr_t)node;
    } else {
        *head = node;
    }
    *tail = node;
    return node;
}

// Проход по списку от head до конца с печатью.
void xor_traverse_forward(Node *head) {
    Node *prev = NULL;
    Node *curr = head;
    printf("  Проход вперёд:");
    while (curr) {
        printf(" %d", curr->value);
        Node *next = (Node *)(curr->both ^ (uintptr_t)prev);
        prev = curr;
        curr = next;
    }
    printf("\n");
}

// Проход от tail к head.
void xor_traverse_backward(Node *tail) {
    Node *next = NULL;
    Node *curr = tail;
    printf("  Проход назад: ");
    while (curr) {
        printf(" %d", curr->value);
        Node *prev = (Node *)(curr->both ^ (uintptr_t)next);
        next = curr;
        curr = prev;
    }
    printf("\n");
}

// Освобождение всех узлов.
void xor_free(Node *head) {
    Node *prev = NULL;
    Node *curr = head;
    while (curr) {
        Node *next = (Node *)(curr->both ^ (uintptr_t)prev);
        prev = curr;
        free(curr);
        curr = next;
    }
}

int main(void) {
    printf("=== XOR-список: ручное управление памятью ===\n\n");

    Node *head = NULL;
    Node *tail = NULL;

    // Заполняем список значениями 10, 20, 30, 40, 50.
    for (int i = 1; i <= 5; i++) {
        xor_push_back(&head, &tail, i * 10);
    }
    printf("  Список создан: 10 → 20 → 30 → 40 → 50\n\n");

    xor_traverse_forward(head);
    xor_traverse_backward(tail);

    printf("\n  Никаких проблем: мы сами управляем памятью,\n");
    printf("  поэтому XOR-указатели остаются валидными.\n");

    xor_free(head);
    printf("\n  Память освобождена вручную.\n");
    return 0;
}

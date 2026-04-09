// Пример 3: Самодельные «умные указатели» на чистом C.
// В C нет деструкторов и шаблонов, но можно реализовать подсчёт ссылок
// вручную. Это основа многих реальных C-проектов (GLib, CPython, COM).

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// === Заголовок «умного указателя» с подсчётом ссылок ===

typedef struct {
    int ref_count;       // Счётчик ссылок.
    void (*destructor)(void *); // Пользовательский деструктор (может быть NULL).
} SmartHeader;

// Выделяем блок: заголовок + пользовательские данные.
// Возвращаем указатель на пользовательские данные (сразу после заголовка).
void *smart_alloc(size_t size, void (*destructor)(void *)) {
    SmartHeader *header = malloc(sizeof(SmartHeader) + size);
    if (!header) {
        return NULL;
    }
    header->ref_count = 1;
    header->destructor = destructor;
    void *data = header + 1; // Указатель на пользовательские данные.
    memset(data, 0, size);
    printf("  [smart_alloc] Выделено %zu байт, ref_count=1\n", size);
    return data;
}

// Получаем заголовок по указателю на данные.
static SmartHeader *get_header(void *ptr) {
    return ((SmartHeader *)ptr) - 1;
}

// Увеличиваем счётчик ссылок (аналог shared_ptr::operator=).
void *smart_retain(void *ptr) {
    if (!ptr) return NULL;
    SmartHeader *h = get_header(ptr);
    h->ref_count++;
    printf("  [smart_retain] ref_count=%d\n", h->ref_count);
    return ptr;
}

// Уменьшаем счётчик ссылок. Когда он достигает 0 — освобождаем.
void smart_release(void *ptr) {
    if (!ptr) return;
    SmartHeader *h = get_header(ptr);
    h->ref_count--;
    printf("  [smart_release] ref_count=%d\n", h->ref_count);
    if (h->ref_count == 0) {
        // Вызываем пользовательский деструктор, если задан.
        if (h->destructor) {
            h->destructor(ptr);
        }
        printf("  [smart_release] Память освобождена\n");
        free(h);
    }
}

// === Пример использования: строка с подсчётом ссылок ===

typedef struct {
    char text[128];
} MyString;

void my_string_destructor(void *ptr) {
    MyString *s = (MyString *)ptr;
    printf("  [деструктор] Уничтожение строки: \"%s\"\n", s->text);
}

void demo_basic() {
    printf("\n--- Базовый пример ---\n");

    MyString *s = smart_alloc(sizeof(MyString), my_string_destructor);
    strcpy(s->text, "Привет, умные указатели!");
    printf("  Строка: %s\n", s->text);

    // Создаём вторую ссылку на тот же объект.
    MyString *s2 = smart_retain(s);
    printf("  s2->text: %s\n", s2->text);

    // Отпускаем первую ссылку — объект ещё жив (ref_count=1).
    smart_release(s);

    // Отпускаем вторую ссылку — объект уничтожается (ref_count=0).
    smart_release(s2);
}

// === Пример: массив с подсчётом ссылок ===

typedef struct {
    int *items;
    int count;
} IntArray;

void int_array_destructor(void *ptr) {
    IntArray *arr = (IntArray *)ptr;
    printf("  [деструктор] Уничтожение массива из %d элементов\n", arr->count);
    free(arr->items); // Освобождаем вложенный ресурс.
}

void demo_nested_resources() {
    printf("\n--- Вложенные ресурсы ---\n");

    IntArray *arr = smart_alloc(sizeof(IntArray), int_array_destructor);
    arr->count = 5;
    arr->items = malloc(sizeof(int) * arr->count);
    for (int i = 0; i < arr->count; i++) {
        arr->items[i] = i * 10;
    }

    printf("  Элементы:");
    for (int i = 0; i < arr->count; i++) {
        printf(" %d", arr->items[i]);
    }
    printf("\n");

    // Один release — и деструктор освободит и items, и сам массив.
    smart_release(arr);
}

int main(void) {
    demo_basic();
    demo_nested_resources();

    printf("\nВсе ресурсы освобождены.\n");
    return 0;
}

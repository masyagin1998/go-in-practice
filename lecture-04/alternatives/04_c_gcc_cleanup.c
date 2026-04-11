// Пример 4: Псевдо-RAII на чистом C через расширение GCC __attribute__((cleanup)).
//
// GCC (и Clang) поддерживают атрибут cleanup: при выходе переменной из scope
// автоматически вызывается указанная функция. Это аналог деструктора в C++.
//
// Преимущества:
//   - Работает на чистом C (не нужен C++ компилятор).
//   - Ресурсы освобождаются автоматически, даже при раннем return или goto.
//   - Используется в реальных проектах: systemd, GLib (g_autofree, g_autoptr).
//
// Недостатки:
//   - Нестандартное расширение — не работает с MSVC.
//   - Cleanup-функция получает указатель на переменную (тип **), не на значение.
//   - Нет RAII для возвращаемых значений — только для локальных переменных.
//
// Сборка (Makefile собирает автоматически):
//   gcc -Wall -Wextra -g -o 04_c_gcc_cleanup 04_c_gcc_cleanup.c
//
// Запуск:
//   ./04_c_gcc_cleanup

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// === Cleanup-функции ===
// Каждая принимает УКАЗАТЕЛЬ на переменную (т.е. тип **).

// Автоматический free() для malloc-памяти.
static void autofree(void *ptr) {
    void **p = (void **)ptr;
    if (*p) {
        printf("  [autofree] Освобождаю %p\n", *p);
        free(*p);
        *p = NULL;
    }
}

// Автоматический fclose() для файлов.
static void autoclose(FILE **fp) {
    if (*fp) {
        printf("  [autoclose] Закрываю файл\n");
        fclose(*fp);
        *fp = NULL;
    }
}

// === Удобные макросы (как в systemd/GLib) ===

#define _cleanup_free_  __attribute__((cleanup(autofree)))
#define _cleanup_fclose_ __attribute__((cleanup(autoclose)))

// === Примеры использования ===

// Пример 1: автоматическое освобождение памяти.
void demo_autofree(void) {
    printf("\n--- autofree ---\n");

    // При выходе из scope вызовется autofree(&buf).
    _cleanup_free_ char *buf = malloc(256);
    if (!buf) return;

    snprintf(buf, 256, "Привет, cleanup!");
    printf("  buf = %s\n", buf);

    // Ранний return — buf всё равно освободится!
    if (strlen(buf) > 5) {
        printf("  Ранний return — buf будет освобождён автоматически.\n");
        return;
    }

    printf("  Эта строка не будет напечатана.\n");
    // autofree(&buf) вызывается здесь автоматически.
}

// Пример 2: автоматическое закрытие файла.
void demo_autoclose(void) {
    printf("\n--- autoclose ---\n");

    _cleanup_fclose_ FILE *fp = fopen("/tmp/cleanup_test.txt", "w");
    if (!fp) {
        printf("  Не удалось открыть файл.\n");
        return;
    }

    fprintf(fp, "Данные из demo_autoclose\n");
    printf("  Записали в файл.\n");

    // fclose(fp) вызовется автоматически при выходе.
}

// Пример 3: несколько ресурсов — все освободятся в обратном порядке.
void demo_multiple_resources(void) {
    printf("\n--- Несколько ресурсов (обратный порядок) ---\n");

    _cleanup_free_ char *first = malloc(100);
    _cleanup_free_ char *second = malloc(200);
    _cleanup_free_ char *third = malloc(300);

    if (!first || !second || !third) return;

    strcpy(first, "Первый");
    strcpy(second, "Второй");
    strcpy(third, "Третий");

    printf("  first=%s, second=%s, third=%s\n", first, second, third);

    // Освобождение в обратном порядке: third → second → first.
    // Как деструкторы в C++!
}

// Пример 4: cleanup для собственной структуры.
typedef struct {
    char *name;
    int *data;
    size_t size;
} MyResource;

static void my_resource_cleanup(MyResource **res) {
    if (*res) {
        printf("  [my_resource_cleanup] Уничтожаю ресурс \"%s\"\n", (*res)->name);
        free((*res)->name);
        free((*res)->data);
        free(*res);
        *res = NULL;
    }
}

#define _cleanup_myresource_ __attribute__((cleanup(my_resource_cleanup)))

MyResource *my_resource_new(const char *name, size_t size) {
    MyResource *r = malloc(sizeof(MyResource));
    if (!r) return NULL;

    r->name = strdup(name);
    r->data = calloc(size, sizeof(int));
    r->size = size;

    if (!r->name || !r->data) {
        free(r->name);
        free(r->data);
        free(r);
        return NULL;
    }

    printf("  [new] Создан ресурс \"%s\" (size=%zu)\n", r->name, r->size);
    return r;
}

void demo_custom_struct(void) {
    printf("\n--- Собственная структура ---\n");

    _cleanup_myresource_ MyResource *res = my_resource_new("Сенсор", 10);
    if (!res) return;

    res->data[0] = 42;
    printf("  data[0] = %d\n", res->data[0]);

    // my_resource_cleanup(&res) вызовется автоматически.
}

int main(void) {
    demo_autofree();
    demo_autoclose();
    demo_multiple_resources();
    demo_custom_struct();

    printf("\nВсе ресурсы освобождены автоматически.\n");
    return 0;
}

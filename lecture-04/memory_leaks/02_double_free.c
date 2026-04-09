// Пример 2: Двойной free — неопределённое поведение (undefined behavior).
// После первого free() указатель становится «висячим» (dangling pointer).
// Повторный free() может повредить структуры аллокатора, вызвать краш
// или, что ещё хуже, быть эксплуатирован злоумышленником.
//
// Запуск:
//   ./02_double_free
//
// Запуск с valgrind:
//   valgrind --leak-check=full --show-leak-kinds=all ./02_double_free

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {
    char *buf = malloc(64);
    if (!buf) {
        return 1;
    }
    strcpy(buf, "Hello");
    printf("buf = %s\n", buf);

    free(buf); // Первый free — корректный.
    free(buf); // Второй free — UNDEFINED BEHAVIOR!
    // После первого free() аллокатор пометил блок как свободный.
    // Повторный free() ломает внутренние списки свободных блоков.

    return 0;
}

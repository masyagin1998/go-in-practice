// Пример 1: Забытый free — утечка памяти.
// Память выделяется, но никогда не освобождается.
// В реальных программах такие утечки накапливаются и приводят к исчерпанию памяти.
//
// Запуск:
//   ./01_forgotten_free
//
// Запуск с valgrind:
//   valgrind --leak-check=full --show-leak-kinds=all ./01_forgotten_free

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char *create_greeting(const char *name) {
    // Выделяем память под строку приветствия.
    char *buf = malloc(256);
    if (!buf) {
        return NULL;
    }
    snprintf(buf, 256, "Привет, %s!", name);
    return buf; // Вызывающий код ДОЛЖЕН вызвать free(), но забудет...
}

int main(void) {
    // Каждую итерацию выделяется 256 байт, но free() нигде не вызывается.
    // За 1 000 000 итераций утечёт ~256 МБ.
    for (int i = 0; i < 1000000; i++) {
        char *msg = create_greeting("Мир");
        printf("%s\n", msg);
        // Забыли: free(msg);
    }

    return 0;
}

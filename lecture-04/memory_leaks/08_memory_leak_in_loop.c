// Пример 8: Утечка при перезаписи указателя.
// Если присвоить указателю новый адрес без free() старого —
// старый блок становится недоступен, но не освобождён.
//
// Запуск:
//   ./08_memory_leak_in_loop
//
// Запуск с valgrind:
//   valgrind --leak-check=full --show-leak-kinds=all ./08_memory_leak_in_loop

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {
    char *data = NULL;

    // Каждая итерация выделяет новый блок, но старый не освобождается.
    // Указатель data перезаписывается — доступ к старому блоку потерян.
    for (int i = 0; i < 100; i++) {
        data = malloc(1024); // Старый блок утёк!
        if (!data) {
            return 1;
        }
        snprintf(data, 1024, "Итерация %d", i);
    }

    printf("Последнее значение: %s\n", data);
    free(data); // Освобождаем только последний блок. 99 блоков утекли.

    return 0;
}

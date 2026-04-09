// Пример 7: Разыменование NULL-указателя.
// Если malloc() вернул NULL (не хватило памяти), а мы не проверили —
// обращение по адресу 0 приведёт к SIGSEGV (краш).
//
// Запуск:
//   ./07_null_dereference
//
// Запуск с valgrind:
//   valgrind --leak-check=full --show-leak-kinds=all ./07_null_dereference

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char *allocate_huge_buffer(void) {
    // Пытаемся выделить нереально большой блок — malloc вернёт NULL.
    char *buf = malloc((size_t)1024 * 1024 * 1024 * 1024); // 1 ТБ
    // Забыли проверить на NULL!
    printf("malloc вернул: %p\n", (void *)buf);
    return buf;
}

int main(void) {
    char *buf = allocate_huge_buffer();

    // buf == NULL, но мы этого не проверили.
    // strcpy попытается записать по адресу 0 → SIGSEGV.
    strcpy(buf, "Hello");

    printf("%s\n", buf);
    free(buf);

    return 0;
}

// Пример 3: Использование памяти после освобождения (use-after-free).
// Память освобождена, но указатель всё ещё используется.
// Данные могут быть перезаписаны другим malloc(), что приведёт
// к чтению мусора или уязвимости безопасности.
//
// Запуск:
//   ./03_use_after_free
//
// Запуск с valgrind:
//   valgrind --leak-check=full --show-leak-kinds=all ./03_use_after_free

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    char name[32];
    int age;
} Person;

int main(void) {
    Person *p = malloc(sizeof(Person));
    if (!p) {
        return 1;
    }
    strcpy(p->name, "Алиса");
    p->age = 30;
    printf("До free: %s, %d лет (адрес структуры: %p)\n", p->name, p->age, (void *)p);

    free(p);

    // Выделяем новый блок — аллокатор может переиспользовать тот же адрес.
    int *numbers = malloc(sizeof(Person));
    if (numbers) {
        memset(numbers, 0xAB, sizeof(Person));
    }
    printf("Адрес нового массива: %p\n", (void *)numbers);

    // Обращаемся к уже освобождённой памяти — читаем мусор!
    printf("После free: %s, %d лет (адрес структуры: %p)\n", p->name, p->age, (void *)p);

    free(numbers);
    return 0;
}

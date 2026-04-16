#ifndef PRODUCER_H
#define PRODUCER_H

#include <stddef.h>

// === Простые значения из C ===
int get_answer(void);
double get_pi(void);

// === Структура из C (возврат по значению — копия по стеку) ===
typedef struct {
    int r, g, b, a;
} Color;

Color make_color(int r, int g, int b, int a);
void  print_color(Color c);

// === C-массив (аллоцирован в C-heap) ===
// Вызывающий код должен освободить через free().
typedef struct {
    int*   data;
    size_t len;
} IntArray;

// make_fibonacci создаёт массив первых n чисел Фибоначчи.
IntArray make_fibonacci(size_t n);

#endif

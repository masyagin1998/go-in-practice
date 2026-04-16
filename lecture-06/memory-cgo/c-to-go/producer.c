#include "producer.h"
#include <stdio.h>
#include <stdlib.h>

// get_answer возвращает простое число.
int get_answer(void) {
    return 42;
}

// get_pi возвращает приближение числа пи.
double get_pi(void) {
    return 3.14159265358979323846;
}

// make_color создаёт структуру Color и возвращает по значению.
Color make_color(int r, int g, int b, int a) {
    Color c = { r, g, b, a };
    return c;
}

// print_color печатает цвет.
void print_color(Color c) {
    printf("[C] Color(r=%d, g=%d, b=%d, a=%d)\n", c.r, c.g, c.b, c.a);
    fflush(stdout);
}

// make_fibonacci создаёт массив в C-heap и заполняет числами Фибоначчи.
IntArray make_fibonacci(size_t n) {
    IntArray arr;
    arr.len = n;
    arr.data = (int*)malloc(n * sizeof(int));
    if (n > 0) arr.data[0] = 0;
    if (n > 1) arr.data[1] = 1;
    for (size_t i = 2; i < n; i++) {
        arr.data[i] = arr.data[i-1] + arr.data[i-2];
    }
    return arr;
}

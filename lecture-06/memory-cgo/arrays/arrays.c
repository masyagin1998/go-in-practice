#include "arrays.h"

// sum_ints считает сумму элементов массива.
long long sum_ints(const int* arr, size_t len) {
    long long sum = 0;
    for (size_t i = 0; i < len; i++) {
        sum += arr[i];
    }
    return sum;
}

// fill_squares заполняет массив квадратами индексов.
void fill_squares(int* arr, size_t len) {
    for (size_t i = 0; i < len; i++) {
        arr[i] = (int)(i * i);
    }
}

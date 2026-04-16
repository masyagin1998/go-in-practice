#ifndef ARRAYS_H
#define ARRAYS_H

#include <stddef.h>

// sum_ints считает сумму элементов массива.
long long sum_ints(const int* arr, size_t len);

// fill_squares заполняет массив квадратами индексов: arr[i] = i*i.
void fill_squares(int* arr, size_t len);

#endif

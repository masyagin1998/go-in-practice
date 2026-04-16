#ifndef STRUTIL_H
#define STRUTIL_H

#include <stddef.h>

// print_cstring печатает C-строку (нуль-терминированную).
void print_cstring(const char* s);

// to_upper_cstring переводит C-строку в верхний регистр in-place.
void to_upper_cstring(char* s);

// print_raw_gostring печатает строку по указателю и длине
// (эквивалент Go-строки без нуль-терминатора).
void print_raw_gostring(const char* p, size_t n);

#endif

#include "strutil.h"
#include <stdio.h>
#include <ctype.h>

// print_cstring печатает C-строку (нуль-терминированную).
void print_cstring(const char* s) {
    printf("[C] CString: \"%s\"\n", s);
    fflush(stdout);
}

// to_upper_cstring переводит C-строку в верхний регистр in-place.
void to_upper_cstring(char* s) {
    for (; *s; s++) {
        *s = toupper((unsigned char)*s);
    }
}

// print_raw_gostring печатает строку по указателю и длине.
void print_raw_gostring(const char* p, size_t n) {
    printf("[C] GoString (len=%zu): \"%.*s\"\n", n, (int)n, p);
    fflush(stdout);
}

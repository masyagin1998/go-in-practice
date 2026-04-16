#ifndef BUFFER_H
#define BUFFER_H

#include <stddef.h>

// buffer_set сохраняет указатель на массив и его длину в глобальных переменных C.
// После вызова C-код может обращаться к данным без повторной передачи указателя.
// ВАЖНО: Go-память должна быть закреплена (runtime.Pinner), иначе GC
// может переместить объект и глобальный указатель станет невалидным.
void buffer_set(int* data, size_t len);

// buffer_print печатает содержимое буфера из глобального указателя.
void buffer_print(void);

// buffer_scale умножает все элементы буфера на factor через глобальный указатель.
void buffer_scale(int factor);

#endif

// fileops.h — файловые операции, которые могут завершиться с ошибкой.
#ifndef FILEOPS_H
#define FILEOPS_H

#include <sys/types.h>

// try_open пытается открыть файл. Возвращает fd или -1 при ошибке (errno устанавливается).
int try_open(const char* path);

// try_chdir пытается сменить текущую директорию. Возвращает 0 или -1 (errno).
int try_chdir(const char* path);

// safe_div делит a на b. При b == 0 устанавливает errno = EINVAL, возвращает 0.
int safe_div(int a, int b);

#endif

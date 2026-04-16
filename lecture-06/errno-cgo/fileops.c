// fileops.c — реализация файловых операций с errno.
#include "fileops.h"
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>

int try_open(const char* path) {
    return open(path, O_RDONLY);
}

int try_chdir(const char* path) {
    return chdir(path);
}

int safe_div(int a, int b) {
    if (b == 0) {
        errno = EINVAL;
        return 0;
    }
    return a / b;
}

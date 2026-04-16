// platform.c — платформозависимая реализация.
//
// Использует макросы PLATFORM_LINUX / PLATFORM_DARWIN, которые передаются
// через #cgo CFLAGS: -DPLATFORM_LINUX (или -DPLATFORM_DARWIN).
// Без этих директив код не скомпилируется — platform_name() не будет определена.
#include "include/platform.h"
#include <unistd.h>

#if defined(PLATFORM_LINUX)

const char* platform_name(void) {
    return "Linux";
}

#elif defined(PLATFORM_DARWIN)

const char* platform_name(void) {
    return "macOS (Darwin)";
}

#else
#error "Не определена платформа: нужен -DPLATFORM_LINUX или -DPLATFORM_DARWIN"
#endif

long page_size(void) {
    return sysconf(_SC_PAGESIZE);
}

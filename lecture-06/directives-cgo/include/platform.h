// platform.h — платформозависимые функции.
#ifndef PLATFORM_H
#define PLATFORM_H

// platform_name возвращает строку с именем ОС.
const char* platform_name(void);

// page_size возвращает размер страницы памяти.
long page_size(void);

#endif

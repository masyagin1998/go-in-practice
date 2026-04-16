// slow.h — длительная C-функция для демонстрации блокировки OS-треда.
#ifndef SLOW_H
#define SLOW_H

// busy_work выполняет тяжёлую работу в течение ~ms миллисекунд.
void busy_work(int ms);

#endif

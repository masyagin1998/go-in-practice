# Результаты демонстраций

## cpu-atomic: атомарность на уровне процессора

| # | Демо                    | Структура-аналог в C                                       | offset | sizeof | Torn read?          |
|---|-------------------------|-----------------------------------------------------------|--------|--------|---------------------|
| 1 | cpu-expected-atomic     | `struct { int32_t shit; int64_t value; }`                  | 8      | 16     | Нет — внутри линии  |
| 2 | cpu-unexpected-atomic   | `__packed__ struct { int32_t shit; int64_t value; }`       | 4      | 12     | Нет — внутри линии  |
| 3 | cpu-expected-non-atomic | `__packed__ __aligned__(64) struct { char[60]; int64_t; }` | 60     | 68     | Да — разрыв на 64   |

### Выводы
- На x86-64 выровненный 8-байтовый load/store атомарен на уровне CPU (Intel SDM §9.1.1).
- **Невыровненный** доступ тоже атомарен, **пока не пересекает границу кэш-линии**.
  cpu-unexpected-atomic: offset 4, байты [4..11] — всё внутри одной линии → torn read не возникает.
- Как только int64 **пересекает границу** (offset 60, байты [60..67] через рубеж на 64) —
  CPU обращается к двум кэш-линиям неатомарно → torn read.

## Синхронизация: spinlock'и

| # | Демо            | Алгоритм              | Fairness | Поведение                                       |
|---|-----------------|-----------------------|----------|-------------------------------------------------|
| 1 | stupid-spinlock | Test-and-set (CAS)    | Нет      | 4 горутины × 100 lock/unlock → counter = 400    |
| 2 | smart-spinlock  | Ticket lock (FIFO)    | Да       | 4 горутины × 100 lock/unlock → counter = 400    |

### Выводы
- **stupid-spinlock**: tight CAS loop → все ядра молотят одну кэш-линию,
  постоянные cache invalidation. Корректен, но starvation возможен.
- **smart-spinlock** (ticket lock): каждый ждёт свой номер → FIFO-порядок,
  нет starvation. `runtime.Gosched()` в spin loop отдаёт CPU.

## Синхронизация: Peterson's algorithm

| # | Демо     | Atomic? | counter (expected 200) | Комментарий                                |
|---|----------|---------|------------------------|--------------------------------------------|
| 1 | peterson | Да      | 200                    | Atomic обеспечивает memory barrier          |
| 2 | peterson | Нет*    | < 200 или зависание    | Без барьеров CPU переупорядочивает операции  |

\* — закомментированная версия в коде.

### Выводы
- Peterson's algorithm **корректен** только при наличии memory barrier (atomic/fence).
- Без atomic: CPU и компилятор переупорядочивают store/load → нарушение mutual exclusion.

## Проблемы конкурентности

| # | Демо         | Проблема              | Что происходит                                        |
|---|--------------|-----------------------|-------------------------------------------------------|
| 1 | deadlock     | Deadlock              | Две горутины берут два мьютекса в разном порядке → Go runtime паникует: `fatal error: all goroutines are asleep - deadlock!` |
| 2 | livelock     | Livelock              | Горутины активно уступают друг другу, но ни одна не продвигается. CPU загружен, но прогресса нет. Завершение по таймауту (5с) |
| 3 | out-of-order | Store-load reordering | T1: x=1, r1=y; T2: y=1, r2=x. Если r1==0 && r2==0 — CPU переупорядочил store после load. На x86-64 ловится редко (strong ordering), на ARM — часто |

### Выводы
- **Deadlock** vs **livelock**: deadlock — потоки спят, livelock — потоки работают,
  но оба — отсутствие прогресса. Go runtime умеет детектировать только deadlock.
- **Out-of-order**: даже на x86-64 (strong memory model) возможно store-load reordering.
  Это единственный тип переупорядочивания, который x86 допускает (TSO — Total Store Order).

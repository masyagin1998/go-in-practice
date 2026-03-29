# Результаты бенчмарков

## Режим fib: 500 000 параллельных запросов, fib(30)

| # | Сервер                              | Время      | Комментарий |
|---|-------------------------------------|------------|-------------|
| 1 | c_single_thread                     | ~30м (ест) | Все запросы последовательно, один за другим |
| 2 | c_coroutine                         | ~14м (ест) | Корутины не помогают — fib() блокирует единственный поток |
| 3 | c_single_thread_epoll               | 7м 39.9с   | epoll не помогает — fib() блокирует единственный поток event loop |
| 4 | c_multi_thread                      | 2м 12.3с   | Создание 500K OS-потоков — ощутимый overhead (~50мкс × 500K) |
| 5 | go_goroutine_pool (14, lock-free)   | 2м 07.2с   | 14 горутин, Go's per-call overhead на рекурсии |
| 6 | go_multi_goroutine (GOMAXPROCS=14)  | 2м 06.7с   | 500K горутин дешевы, но fib() медленнее чем в C |
| 7 | c_multi_thread_epoll (SO_REUSEPORT) | 1м 57.6с   | 14 epoll-потоков, каждый со своим accept — нет очереди |
| 8 | c_thread_pool (14, lock-free)       | 1м 42.2с   | Фиксированные 14 потоков, нет overhead на создание |
| 9 | c_thread_pool_epoll (14, eventfd)   | 1м 41.7с   | I/O в epoll-потоке, воркеры только считают — чуть чище |

## Режим sleep: 500 000 параллельных запросов, sleep 100мс

| # | Сервер                              | Время      | Комментарий |
|---|-------------------------------------|------------|-------------|
| 1 | c_single_thread                     | ~14ч (ест) | 500K × 100мс последовательно |
| 2 | c_thread_pool (14, lock-free)       | ~60м (ест) | 14 воркеров с nanosleep — 500K/14 × 100мс |
| 3 | c_thread_pool_epoll (14, eventfd)   | ~60м (ест) | То же — воркеры блокируются на nanosleep |
| 4 | go_goroutine_pool (14, lock-free)   | ~60м (ест) | 14 горутин с time.Sleep — то же ограничение |
| 5 | go_multi_goroutine (GOMAXPROCS=14)  | 1м 47.6с   | 500K горутин спят одновременно, runtime жонглирует таймерами |
| 6 | c_multi_thread_epoll (SO_REUSEPORT) | 1м 47.4с   | 14 потоков с timerfd — все таймеры неблокирующие |
| 7 | c_multi_thread                      | 1м 47.1с   | 500K OS-потоков, каждый спит независимо |
| 8 | c_single_thread_epoll               | 1м 47.0с   | ОДИН поток, 500K timerfd — аналог Go runtime вручную |
| 9 | c_coroutine                         | 1м 46.4с   | ОДИН поток, корутины + timerfd — код как в Go, скорость как в C |

## Выводы

### CPU-bound (fib)
- **Thread pool побеждает** при большом количестве запросов: 14 фиксированных потоков
  без overhead на создание/уничтожение. Lock-free очередь почти бесплатна.
- **Go на ~20% медленнее C** из-за overhead на каждый вызов функции:
  Go вставляет preemption check (проверку "не пора ли переключить горутину")
  в каждый call. При fib(30) это ~1.3M лишних проверок на запрос.
  При fib(42) было ~268M — и Go отставал уже в 4 раза.
- **epoll и корутины в однопоточном режиме бесполезны** для CPU-bound — fib()
  блокирует единственный поток, все запросы всё равно последовательны.
  c_coroutine даже медленнее c_single_thread_epoll из-за overhead на context switch.

### I/O-bound (sleep)
- **Все concurrent решения упираются в ~1м47с** — bottleneck на стороне клиента
  (500K процессов nc через xargs), а не сервера.
- **c_coroutine (1 поток!) = c_single_thread_epoll = Go multi-goroutine = c_multi_thread**.
  Корутинное решение самое быстрое — и при этом код handle_client выглядит
  как обычный линейный код, в точности как в Go.
- **Pool-решения непрактичны для I/O-bound**: 14 воркеров с блокирующим sleep —
  искусственное ограничение параллелизма.

### Что такое Go runtime на самом деле
Go runtime ≈ **c_coroutine**, размноженный на GOMAXPROCS потоков:
- Корутины = горутины (cooperative green threads с собственным стеком)
- coro_read/coro_write = net.Conn.Read/Write (yield при EAGAIN)
- coro_sleep = time.Sleep (timerfd + yield)
- Scheduler = Go's M:N scheduler (epoll event loop + ready queue)

`go handleClient(conn)` — это синтаксический сахар над тем, что в нашем
c_coroutine занимает ~200 строк. Go просто прячет эту сложность за
ключевым словом `go` и добавляет многопоточный планировщик.

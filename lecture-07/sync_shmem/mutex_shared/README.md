# mutex_shared — pthread_mutex в shared memory

Между **процессами**, а не горутинами. Go создаёт shmem + инициализирует
мьютекс с `PTHREAD_PROCESS_SHARED`, N C-воркеров атакуют общий `counter`.

Итог: `counter == N × iterations_per_worker`. Без `PROCESS_SHARED` —
lost updates.

## Запуск

```bash
make peer
# терминал 1:
go run .
# терминалы 2-5:
./peer 0
./peer 1
./peer 2
./peer 3
# Ctrl+C в терминале 1 — увидите итоговый counter
```

или `make run` (автоматизирует 4 воркера).

## Почему нельзя просто горутины

Участники — разные процессы (Go + C). Общая только `MAP_SHARED`-область;
`sync.Mutex` живёт в Go-heap одного процесса и не виден другому.

## Ключевые моменты

- **PTHREAD_PROCESS_SHARED**: без него mutex лочит только внутри одного
  процесса.
- **Мьютекс должен физически лежать в `MAP_SHARED`**. Копирование
  инициализированного mutex'а в другой mmap — UB (futex-адрес).
- **Layout**: Go видит `C.shared_data_t` через cgo, совпадение гарантировано.
- **Cleanup**: при аварийном выходе `pthread_mutex_destroy` не вызывается.
  Для таких случаев — robust mutex, см. `../mutex_death/`.

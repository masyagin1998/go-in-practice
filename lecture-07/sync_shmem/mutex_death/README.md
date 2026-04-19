# mutex_death — robust mutex и EOWNERDEAD

Обычный `pthread_mutex` в shmem — мина: процесс упал внутри `lock()` →
все остальные зависли навсегда. `PTHREAD_MUTEX_ROBUST` даёт ядру
заметить смерть владельца и отдать mutex следующему с флагом
`EOWNERDEAD`; тот обязан починить инварианты данных и сделать
`pthread_mutex_consistent`.

## Сценарий

1. Go создаёт shmem + robust mutex (`PROCESS_SHARED + ROBUST`).
2. Go в цикле делает `trylock`:
   - `0` — никто не взял, отпускаем и ждём;
   - `EBUSY` — peer взял, переходим к blocking lock;
   - `EOWNERDEAD` — peer уже успел умереть, восстанавливаем.
3. C-peer берёт mutex, пишет `value=42`, `_exit(0)` без unlock.
4. Go получает EOWNERDEAD, чинит `value=99`, отпускает.

## Запуск

```bash
make peer
# терминал 1:
go run .
# терминал 2:
./peer
```

или `make run`.

## Когда это нужно

- Долгоживущие in-memory БД/кеши в shmem: воркеры падают, остальные
  должны продолжать.
- Multi-process с горячим релоадом: нельзя терять lock при рестарте.

## Ограничения

- Linux+glibc поддерживают robust-mutex полностью; macOS/BSD — частично.
- "Восстановление инвариантов" — ответственность приложения. Ядро лишь
  говорит "владелец умер".
- Для простых rendezvous без критической секции — POSIX sem
  (см. `../../ipc/07_semaphore/`).

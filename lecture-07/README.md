# lecture-07: IPC на Go

Межпроцессное взаимодействие на практике. Порт `hp-systems-course/lecture-03`
(POSIX C) в Go-редакцию: Go на одной стороне, C — на другой, оба
запускаются в **отдельных терминалах** (Go НЕ форкает C).

## Карта лекции

```
fork/                        # создание процессов в Go
  exec-spawn/                # os/exec.Command — идиоматический spawn
  syscall-exec/              # syscall.Exec — self-replace (10 поколений)

ipc/                         # Go ↔ C, по возрастанию скорости/сложности:
  01_pipes/                  # shell pipe `./peer | go run .`
  02_pipes_zc/               #   + vmsplice(SPLICE_F_GIFT) zero-copy на TX
  03_mkfifo/                 # named FIFO
  04_mkfifo_zc/              #   + vmsplice через FIFO
  05_sockets/                # TCP loopback
  06_sockets_zc/             #   + SO_ZEROCOPY + send(MSG_ZEROCOPY)
  07_unix_sockets/           # AF_UNIX stream
  08_unix_sockets_zc/        #   + SCM_RIGHTS + memfd — полный zero-copy
  09_queue/                  # POSIX message queue
  10_shmem/                  # shared memory + busy-wait
  11_semaphore/              # shmem + POSIX sem — самый быстрый

sync_shmem/                  # синхронизация в shared memory
  mutex_shared/              # pthread_mutex + PTHREAD_PROCESS_SHARED
  mutex_death/               # robust mutex, EOWNERDEAD recovery

signals/                     # сигналы: lifecycle + IPC с payload
  graceful-shutdown/         # SIGINT/SIGTERM → ctx.Cancel
  sighup-reload/             # SIGHUP → перечитать конфиг
  signalfd/                  # signals как fd, payload через sigqueue

benchmark/                   # ping-pong round-trip: msgs/sec, ns/op
```

## Как запускать

Каждый пример — отдельный Go-модуль, внутри — `main.go` и (обычно)
`c/peer.c`. По умолчанию Go и C — **в разных терминалах**:

```bash
make peer          # собрать C-peer
# терминал 1:
go run .
# терминал 2:
./peer
```

Для CI / быстрых прогонов у каждого `Makefile` есть цель `run`, которая
автоматизирует: запускает Go в фоне, выполняет peer, возвращает управление.
На верхнем уровне:

```bash
make run          # прогнать все демо
make bench        # ping-pong бенчмарк (~10 минут)
make bench-quick  # короткий бенч (~10 секунд)
make clean        # убрать артефакты
```

Требования: `gcc`, `make`, Go 1.24+. Cgo-примеры линкуются с `-lrt`/`-lpthread`.

## Стиль

- Комментарии на русском.
- Код компактный, в духе `hp-systems-course/lecture-03`.
- Внешних Go-зависимостей нет — только stdlib и cgo.

## Что не вошло (и почему)

- **`shmem_anon`** — между неродственными процессами `MAP_ANONYMOUS|MAP_SHARED`
  бесполезен. Аналог — `memfd_create` + `SCM_RIGHTS` (следующая лекция).
- **System V IPC** (`msgget`, `semget`, `shmget`) — legacy, POSIX-аналоги мощнее.
- **Двойной `fork()` для демонизации** — антипаттерн в Go:
  используйте `systemd Type=simple`. См. `fork/README.md`.
- **`vfork`** — в Go не экспонируется; `syscall.ForkExec` внутри использует
  `clone(CLONE_VM|CLONE_VFORK)` на Linux.

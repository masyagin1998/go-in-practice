# 01_pipes — shell pipe

Самый простой IPC. Оболочка через `|` создаёт `pipe(2)` и соединяет
stdout одного процесса со stdin другого.

## Запуск

```bash
make peer             # собрать C-пир
./peer | go run .     # C пишет в stdout, Go читает из stdin
```

или сразу `make run`.

## Реальные кейсы

- **Shell-пайплайны**: `ps aux | grep go`.
- **Build-тулчейны**: `go build` внутри дёргает `cc1`/`as`/`ld` через пайпы.
- **ffmpeg-обёртки**: `ffmpeg -i - -f mp4 -` — чистое stdin/stdout API.
- **git helpers**: smart-HTTP transport → `git-http-backend` по пайпам.

## Ограничения

- **Однонаправленный поток**. Для двусторонних — две `pipe()` с общим родителем
  (fork+exec из Go) или FIFO. См. `../03_mkfifo/`.
- **Буфер ядра ~64K** (Linux). Переполнение → writer блокируется.
- **Кончается вместе с процессами**: долгоживущий обмен — FIFO/socket.

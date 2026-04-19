# 02_mkfifo — named pipe

То же, что anonymous pipe, но с именем в ФС. Не требует общего родителя.

## Запуск в двух терминалах

```bash
make peer
# терминал 1:
go run .
# терминал 2 (после "Ждём peer'а..."):
./peer
```

или сразу `make run` (Go на фоне + peer).

## Реальные кейсы

- Rendezvous между демонами на одной машине без TCP.
- Hot-reload триггеры: `touch /var/run/myapp.fifo` → демон перечитывает конфиг.
- Legacy syslog: `syslog-ng`/`rsyslog` собирает через FIFO из скриптов.
- Обёртки вокруг CLI, которые хотят только путь (`tar xf /tmp/fifo`).

## Подводные камни

- `open(O_RDONLY)`/`open(O_WRONLY)` **блокируются**, пока не откроется
  вторая сторона. В Go — параллельные goroutines или `O_NONBLOCK`.
- Stream-формат: нет границ сообщений, нужен framing (длина/разделитель).
- Атомарность `write` ≤ `PIPE_BUF` (4K на Linux) при нескольких писателях.

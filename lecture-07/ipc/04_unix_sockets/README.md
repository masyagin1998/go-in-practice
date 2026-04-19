# 04_unix_sockets — AF_UNIX stream

Тот же API, что у TCP, но без сетевого стека. Быстрее, изолирован правами
файловой системы.

## Запуск в двух терминалах

```bash
make peer
# терминал 1:
go run .
# терминал 2:
./peer
```

или `make run`.

## Реальные кейсы

- **Docker**: `/var/run/docker.sock` — права через `chmod`, без открытого порта.
- **systemd socket activation**: `ListenStream=/run/app.sock` → приложение получает fd #3.
- **Postgres local**: `/var/run/postgresql/.s.PGSQL.5432` — на 20–40% быстрее TCP.
- **gRPC over UDS** — sidecar ↔ main process, когда сеть не нужна.
- **SCM_RIGHTS** — передача fd между процессами (следующая лекция).

## Особенности

- **Abstract namespace** (Linux): путь с `\0` в начале → сокет не в ФС,
  только в ядре: `net.Listen("unix", "\x00myapp")`.
- **Datagram-вариант** (`SOCK_DGRAM`) сохраняет границы сообщений — им пользуются
  `syslog`/`journald`.
- **Permissions**: `chmod 600` на сокет и он недоступен другим пользователям.

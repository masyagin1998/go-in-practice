# 05_queue — POSIX message queue

`mq_open`/`mq_send`/`mq_receive` — сообщения с приоритетами. Framing
не нужен: одно `mq_send` = одно атомарное сообщение.

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

- **Real-time embedded** (automotive DO-178C, медтех): предсказуемый latency,
  никаких динамических аллокаций в kernel path.
- **Inter-service сигналы с приоритетами**: "shutdown" идёт prio=31,
  метрики — prio=0.
- **Bounded producer/consumer**: `mq_maxmsg` ограничивает очередь, `mq_send`
  блокируется — backpressure бесплатно.
- **`mq_notify`**: ядро разбудит процесс сигналом при появлении сообщения —
  альтернатива активному ожиданию.

## Отличие от System V `msgget`

- POSIX mq видна в `/dev/mqueue` (можно `ls`, `cat`), System V — `ipcs`.
- POSIX интегрируется с `epoll`; API проще.

## Подводные камни

- **Лимиты**: `/proc/sys/fs/mqueue/msg_max`, `msgsize_max` — на Ubuntu 10/8192.
- **Cleanup**: `mq_unlink` удаляет имя; пока fd открыт — очередь жива.
- **Блокировки**: без `O_NONBLOCK` `mq_send`/`mq_receive` блокируются.

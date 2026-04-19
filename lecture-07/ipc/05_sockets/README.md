# 05_sockets — TCP loopback

Самый универсальный транспорт: тот же API и для 127.0.0.1, и между машинами.

## Запуск в двух терминалах

```bash
make peer
# терминал 1:
go run .
# терминал 2 (после "Слушаем ..."):
./peer
```

или `make run`.

## Реальные кейсы

- Всё сетевое: HTTP, gRPC, Postgres/Redis/Kafka.
- Sidecar/proxy (Envoy, linkerd) слушает localhost, app ходит через loopback.
- Кросс-языковая отладка через фиксированный `127.0.0.1:PORT`.

## Важно

- Loopback идёт через полный сетевой стек (netfilter, qdisc, TCP FSM).
  UDS избегает значительной части — см. `../07_unix_sockets/`.
- Go включает `SO_REUSEADDR` по умолчанию — без него `TIME_WAIT` блокировал бы
  повторный bind.
- `SO_REUSEPORT` раздаёт accept'ы между процессами — основа prefork-серверов.

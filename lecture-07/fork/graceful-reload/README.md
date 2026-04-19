# graceful-reload

Zero-downtime релоад HTTP-сервера через наследование слушающего сокета.

## Сценарий

```bash
# терминал 1:
make run
# [PID 12345] первичный, слушаем 127.0.0.1:8080

# терминал 2:
curl localhost:8080/           # hi from PID 12345
kill -HUP 12345
curl localhost:8080/           # hi from PID 12346 — новый процесс отвечает
```

При релоаде `curl` не получает ни `connection refused`, ни reset —
accept-очередь ядра копится у общего listening-сокета, новый процесс
её разбирает.

## Ключевые моменты

- `(*net.TCPListener).File()` возвращает **дубликат** fd — оригинальный
  listener у родителя продолжает жить, дочке уходит копия.
- `cmd.ExtraFiles = []*os.File{f}` → fd №3 в child. 0/1/2 зарезервированы
  под stdin/stdout/stderr.
- Дочь видит `GRACEFUL_INHERIT=1`, забирает fd 3 через `os.NewFile` +
  `net.FileListener`.
- `srv.Shutdown(ctx)` ждёт in-flight хендлеры; без таймаута зависнет
  на долгих запросах.

# signals/ — сигналы как IPC и lifecycle-инструмент

Сигнал — асинхронное уведомление ядра процессу. Номер + опциональный
payload (см. `sigqueue`). Минималистичный IPC: никаких данных, только
"событие пришло".

В Go сигналы в двух ролях:

| Роль | Сигналы | Пример |
|---|---|---|
| **lifecycle** — управлять своим процессом | SIGINT/SIGTERM/SIGHUP | `graceful-shutdown/`, `sighup-reload/` |
| **IPC** — события от другого процесса | SIGUSR1/SIGUSR2, custom | `signalfd/` |

## Карта

```
graceful-shutdown/   # SIGINT/SIGTERM → ctx.Cancel → goroutines завершаются
sighup-reload/       # SIGHUP → перечитать config (atomic.Pointer)
signalfd/            # signals как fd, получение payload (ssi_int) от C-peer'а
```

## Что нужно знать про сигналы в Go

- **`signal.Notify(ch, ...)`** — подписка на сигнал. Канал ДОЛЖЕН быть
  буферизован (ёмкость ≥1): если никто не читает, сигнал теряется.
  Ядро не копит — второй такой же, пришедший подряд, сольётся с первым.
- **`signal.Notify` не даёт payload**. `sigqueue(2)` несёт `ssi_int`/`ssi_ptr`,
  но Go показывает только номер. Payload → `signalfd(2)` через cgo
  (пример в `signalfd/`).
- **Маска — per-thread**. `kill(pid, sig)` доставляется ЛЮБОМУ треду
  с разблокированным сигналом. Если нужно, чтобы сигнал забрал не
  Go-шный обработчик, а `signalfd` / `sigwait` — блокируйте его на
  всех тредах (пример в `signalfd/`).
- **Сигналы и Go-runtime**. Runtime сам слушает SIGURG (preemption),
  SIGPIPE, SIGPROF. Не трогайте их. Безопасны для приложения:
  SIGINT/TERM/HUP, SIGUSR1/2, SIGCHLD.
- **SIGKILL и SIGSTOP не ловятся** — обрабатываются ядром напрямую.
  Идемпотентность важнее чистого shutdown.

## Что не вошло

- **Двойной fork + setsid** (демонизация) — антипаттерн в Go,
  см. `../fork/README.md`.
- **Signal-safe code внутри обработчика** — в Go это не проблема:
  сигнал приходит в канал, прикладной код — в обычной goroutine.
- **SIGCHLD + waitpid** — `exec.Cmd.Wait()` делает всё правильно.

# 02_trace — runtime/trace, планировщик и GC

`pprof` говорит "где тормозит". `runtime/trace` показывает **когда** и
**почему**: таймлайн каждой goroutine, GC-фазы, syscall-блокировки,
задержки планировщика.

## Запуск

```bash
make run
go tool trace trace.out    # откроет UI в браузере
```

## Что полезного в UI

| Раздел | Для чего |
|---|---|
| **View trace** | основной таймлайн: P0..PN × время. Видно GC-stop-the-world как узкую полосу. |
| **Goroutine analysis** | сколько каждая goroutine ждала, бежала, блокировалась. |
| **Scheduler latency profile** | сколько времени goroutine "стояла в очереди" до того, как её запустили. |
| **Network blocking profile** | ожидание `netpoller`. |
| **Synchronization blocking profile** | блокировки на mutex/chan. |
| **Syscall blocking profile** | сколько goroutine просидела в syscall. |
| **User-defined regions** | `trace.WithRegion(...)` в коде — свои метки. |
| **MMU** | minimum mutator utilization — сколько процента времени приложение реально работало (а не стояло под GC). |

## GODEBUG=gctrace=1 — quick look

Для "срочно, сейчас" без трейса:

```bash
GODEBUG=gctrace=1 go run .
```

Каждая строчка — отчёт GC:
```
gc 5 @0.012s 0%: 0.018+0.45+0.009 ms clock, ...
```
- `5` — номер цикла GC
- `0.45 ms clock` — время stop-the-world
- `0%` — GC overhead

Если GC забирает > 25% CPU — время смотреть heap-профиль (`../01_heap/`)
и понимать, откуда аллокации.

## GOMEMLIMIT и GOGC

Два регулятора GC:

- `GOGC=100` (default) — GC запускается, когда heap вырос в 2 раза от
  живого.
- `GOMEMLIMIT=512MiB` — жёсткий потолок. GC будет бежать чаще, чтобы
  не перевалить через лимит. Полезно в k8s с ограничением памяти.

Trace сразу покажет, если `GOMEMLIMIT` слишком маленький: GC забьёт
весь таймлайн.

## Trace vs pprof — когда что

- **pprof cpu** — "какая функция ест CPU".
- **pprof heap** — "кто аллоцирует, кто держит".
- **runtime/trace** — "почему мой хоттайм-лимит не выполняется":
  GC-паузы, блокировки каналов, syscall-хвосты.

Trace — последняя инстанция. Он дорогой (overhead 3-5%), файлы большие,
но даёт картинку, которую не даст ни один другой инструмент.

## Артефакт

- `trace-summary.txt` — выжимка scheduler-profile из trace (чтобы
  увидеть суть без UI).

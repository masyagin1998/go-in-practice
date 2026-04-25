# lecture-08: Debugging, profiling

Инструменты исследования уже написанных Go-программ: как искать баги в
рантайме, мерить производительность, находить узкие места, ловить гонки
и утечки горутин.

## Карта лекции

```
debugging/                   # дебагер и post-mortem анализ
  01_delve_basic/            # dlv debug: break, next, print
  02_delve_goroutines/       # dlv: goroutines, switch, backtrace
  03_runtime_stack/          # runtime.Stack по SIGUSR1 + GOTRACEBACK
  04_core_dump/              # gcore PID → dlv core (post-mortem)

benchmarking/                # testing.B в Go
  01_basic/                  # b.N, sub-benchmarks, -benchmem, sink
  02_benchstat/              # two impls → benchstat old.txt new.txt

profiling/                   # pprof и runtime/trace, по темам
  01_heap/                   # inuse_space vs alloc_space, утечка
  02_trace/                  # runtime/trace: GC-паузы, scheduler
  03_cpu/                    # CPU profile, flamegraph
  04_block_mutex/            # block + mutex profile
  05_real_program/           # HTTP-сервер + нагрузка: before → fix → after

race/                        # data race detector
  01_counter_race/           # read-modify-write на общем int
  02_flag_race/              # "безобидный" bool done
  03_atomic_fix/             # sync/atomic + таблица atomic/mutex/channel

goleak/                      # go.uber.org/goleak
  01_leak_basic/             # runtime.NumGoroutine — утечка видна цифрой
  02_goleak_test/            # goleak.VerifyTestMain в TestMain
  03_context_leak/           # worker игнорирует ctx.Done → фикс
```

## Как запускать

Каждое демо — отдельный Go-модуль с собственным `Makefile`.

```bash
make run          # прогнать все демо
make clean        # убрать артефакты
```

Для отдельной секции: `make -C profiling run`, `make -C benchmarking bench`
и так далее.

## Предустановка

- **Go 1.24+**
- **delve** — `go install github.com/go-delve/delve/cmd/dlv@latest`
- **benchstat** — `go install golang.org/x/perf/cmd/benchstat@latest`
- **graphviz** — `apt install graphviz` (для flamegraph в `pprof -http`)
- **gdb** — `apt install gdb` (нужен `gcore` для секции `04_core_dump`)
- `sysctl -w kernel.yama.ptrace_scope=0` либо `sudo` — для `gcore` /
  `dlv attach`

## Стиль

- Комментарии на русском.
- Код компактный, в духе lecture-06/07.
- Внешних Go-зависимостей нет. **Исключение**: `goleak/` содержит один
  общий go-модуль с `go.uber.org/goleak`. `benchstat` используется как
  CLI-утилита, а не как import.
- Артефакты (вывод `pprof -top`, `benchstat`, `race`, `goleak` и т.п.)
  закоммичены в репозиторий — по образцу `lecture-06/speedup-cgo/go_o2.txt`.
  Студент может запустить сам и сравнить.

## Что не вошло (и почему)

- **Continuous profiling** (Pyroscope, Parca) — production-инфра,
  за скоупом курса.
- **eBPF / bpftrace / bcc** — отдельная большая тема; в Go-мире чаще
  используется для observability ядра, а не самого приложения.
- **IDE-дебагеры** (Goland, VSCode-delve) — тонкие обёртки над `dlv`,
  принципы ровно те же.
- **`go test -fuzz`** — относится к лекции о тестировании, не сюда.
- **`gops`** (`github.com/google/gops`) — удобная обёртка, но дублирует
  `pprof` + `runtime`.
- **`sasha-s/go-deadlock`** — частный случай race + goleak.

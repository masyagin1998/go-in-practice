# 03_cpu — CPU-профиль

Классический пример: склейка 50 000 строк через `+=` против
`strings.Builder`. Снимаем cpu.pprof и смотрим, где горит CPU.

## Запуск

```bash
make run
```

## Как снимают CPU-профиль

Три варианта.

**1. Из кода (runtime/pprof):**
```go
f, _ := os.Create("cpu.pprof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
// работа
```

**2. Из теста/бенчмарка:**
```bash
go test -bench=. -cpuprofile=cpu.pprof
```

**3. С живого сервера (см. `05_real_program/`):**
```bash
go tool pprof http://host:6060/debug/pprof/profile?seconds=30
```

## Анализ

```bash
go tool pprof -top cpu.pprof
go tool pprof -top -cum cpu.pprof        # по cumulative вместо flat
go tool pprof -http=: cpu.pprof           # UI с flamegraph (нужен graphviz)

go tool pprof cpu.pprof
(pprof) top
(pprof) list main.joinNaive               # исходник с heat-map
(pprof) disasm main.joinNaive             # ассемблер
(pprof) web                               # SVG-граф
```

## Как читать top

```
      flat  flat%   sum%        cum   cum%
   5.20s  80.0% 80.0%      5.30s  81.5%  main.joinNaive
```

- **flat** — время **в самой функции**, без вызываемых.
- **cum** — время в функции **и всём, что она вызывает**.
- Для горячих листовых функций flat ≈ cum. Для "родителей" — flat
  маленький, cum большой.

Наивный `joinNaive` выдаст flat = почти всё время, потому что внутри
он вызывает `runtime.concatstrings`, который копирует память.

## Sampling profiler и его ограничения

Go-профилер — **sampling**, 100 Hz (SIGPROF каждые 10 мс). Из этого:

- Функции быстрее 10 мс в сумме **не попадают в профиль**.
- Нужен достаточно длинный прогон (секунды) — иначе мало сэмплов и
  картинка шумная.
- Overhead ~1-2%, в прод можно включать.

Что CPU-профиль **не покажет**:
- IO-wait (goroutine спит — не тратит CPU);
- syscall latency (см. `02_trace/`);
- ожидание блокировок (см. `04_block_mutex/`).

## Артефакт

- `cpu-top.txt` — top-10 по flat. Наглядно видно, что `joinNaive`
  доминирует.

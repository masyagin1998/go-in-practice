# 05_real_program — профилирование живого сервера

Capstone всей секции: HTTP-сервер, нагрузка, поиск узкого места и
его устранение.

## Что внутри

- `main.go` — сервер с `_ "net/http/pprof"` на `:6060`:
  - `/fast` — быстрый ответ;
  - `/search?q=...` — case-insensitive поиск подстроки в in-memory
    датасете из 4000 строк;
  - `/debug/pprof/*` — стандартный набор от `net/http/pprof`.
- `load/main.go` — простой генератор нагрузки (N workers, GET в цикле).
- Флажок `USE_REGEXP` в `main.go` — переключатель между "удобной"
  реализацией через regexp и "правильной" через strings-функции.

## Сценарий

### Терминал 1 — сервер

```bash
go run .
```

Или через Makefile — он собирает бинарь и запускает его:

```bash
make run
```

Подсказки сервера в логах покажут все нужные URL.

### Терминал 2 — нагрузка

```bash
cd load
DURATION=20s WORKERS=20 go run .
```

### Терминал 3 — снимаем профили

```bash
# CPU-профиль за 10 секунд:
go tool pprof -top 'http://localhost:6060/debug/pprof/profile?seconds=10'

# Heap сейчас:
go tool pprof -top 'http://localhost:6060/debug/pprof/heap'

# Все goroutine со стеками:
curl 'http://localhost:6060/debug/pprof/goroutine?debug=2' | head -40

# Allocs (что аллоцировали, включая уже собранное):
go tool pprof -top 'http://localhost:6060/debug/pprof/allocs'

# Flamegraph в браузере:
go tool pprof -http=: 'http://localhost:6060/debug/pprof/profile?seconds=10'
```

## Что видно в профиле (до фикса)

Топ CPU-cum с `USE_REGEXP=true` (реальный вывод, см. `before-top.txt`):

```
     cum   cum%
   96.70s 98.67%  main.search
   96.52s 98.49%  regexp.(*Regexp).MatchString
   95.91s 97.87%  regexp.(*Regexp).backtrack
   70.86s 72.31%  regexp.(*Regexp).tryBacktrack
```

99% CPU уходит в `regexp.backtrack`. Для обычного substring-поиска это
дикий оверкилл: regex движок гоняет backtracking machine, когда всё,
что нужно, — сравнить байты.

## Фикс и сравнение

Меняем `USE_REGEXP` → `false` (используем `strings.ToLower` +
`strings.Contains`), перезапускаем сервер, снимаем профиль ещё раз:

```bash
# before:
go tool pprof -top -cum 'http://.../profile?seconds=5' > before-top.txt
# ... правим код, перезапускаем сервер ...
# after:
go tool pprof -top -cum 'http://.../profile?seconds=5' > after-top.txt
diff before-top.txt after-top.txt > delta.txt
```

После фикса (реальные цифры, см. `after-top.txt`):

```
     cum   cum%
   31.49s 74.25%  main.search
   22.30s 52.58%  strings.ToLower
    8.40s 19.81%  strings.Contains
```

Total samples: **98s → 42s** на том же 5-секундном окне под той же
нагрузкой. CPU-загрузка сервера упала почти вдвое.

## /debug/pprof/* — что можно дёргать

| Эндпоинт | Что отдаёт |
|---|---|
| `/debug/pprof/` | HTML-страница со ссылками |
| `/debug/pprof/profile?seconds=N` | CPU-профиль за N секунд |
| `/debug/pprof/heap` | текущий снапшот heap |
| `/debug/pprof/allocs` | все аллокации с момента старта |
| `/debug/pprof/goroutine` | все goroutine (добавь `?debug=2` для стеков) |
| `/debug/pprof/block` | block profile (нужно SetBlockProfileRate) |
| `/debug/pprof/mutex` | mutex profile (нужно SetMutexProfileFraction) |
| `/debug/pprof/threadcreate` | когда создавались OS-потоки |
| `/debug/pprof/trace?seconds=N` | runtime/trace (см. `../02_trace/`) |

## Реальные кейсы (и предостережения)

- **В проде НЕ выставляйте `/debug/pprof/*` наружу**. Это утечка
  внутренних структур + вектор DoS (профилирование не бесплатно).
  Типичные подходы:
  - Отдельный listener на loopback (`127.0.0.1:6060`) + порт-форвард.
  - Unix socket: `net.Listen("unix", "/var/run/app.sock")`.
  - Basic-auth или mTLS middleware перед `pprof.Index`.
- **Серверы за балансировщиком**. Нагрузку распределяет LB, профиль
  снимаете с одного инстанса — возможно, как раз с лёгкого. Репрезентативную
  картину даёт continuous profiling (Pyroscope/Parca), но это вне скоупа
  этой лекции.
- **Golden path**: `pprof -base before.pprof after.pprof` — diff.
  Видно, где стало лучше/хуже относительно baseline.

## Артефакты

- `before-top.txt` — CPU-top с наивным регекспом в hot path.
- `after-top.txt` — CPU-top после фикса (создаётся вручную: правите
  код, перезапускаете, снимаете новый профиль).
- `delta.txt` — diff между ними.

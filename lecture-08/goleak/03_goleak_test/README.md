# 03_goleak_test — ловим утечки тестами

`go.uber.org/goleak` сравнивает goroutine "до теста" и "после". Если
появились новые — печатает их стеки и валит тест.

## Запуск

```bash
go test -v .
```

`TestClean` — PASS, `TestLeaky` — FAIL с дампом.

## Два способа интеграции

```go
// 1. Один раз на пакет — короче, но любая утечка красит весь пакет.
func TestMain(m *testing.M) { goleak.VerifyTestMain(m) }

// 2. Per-test через defer — красным становится только виновный тест.
func TestX(t *testing.T) {
    defer goleak.VerifyNone(t)
    ...
}
```

В коде используется (2).

## Опции фильтрации

```go
goleak.VerifyNone(t,
    goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
    goleak.IgnoreCurrent(),         // baseline "что уже висит сейчас"
    goleak.IgnoreAnyFunction("svc.bg"),
)
```

Полезно для библиотек с "вечными" фоновыми G (см. `../02_counter_false_positive/`).

## Как выглядит падение

```
--- FAIL: TestLeaky (0.26s)
    leaks.go:78: found unexpected goroutines:
        [Goroutine N in state chan receive, with worker.Leaky.func1 on top of the stack:
        worker.Leaky.func1()
            worker.go:34 +0x...
        created by worker.Leaky
            worker.go:32 +0x...
        ]
```

Точная строка старта и точка зависания — этого не даёт `NumGoroutine()`
из `01_leak_basic/`.

## Где goleak используют

etcd, grpc-go, kubernetes/client-go, uber/zap, uber/fx — практически
стандарт для сетевых сервисов, где утечка goroutine = утечка сокета.

## Артефакт

- `goleak-output.txt` — `go test -v` с PASS/FAIL и стеком утечки.

# 06_atomic_fix — фиксы к 04 и 05

`atomic.Int64` для счётчика, `atomic.Bool` для флага. Под `-race` чисто.

## Запуск

```bash
go run -race .
# counter = 2000000
# воркер вышел после ... спинов
```

## API (Go 1.19+, типизированные обёртки)

```go
var c atomic.Int64
c.Add(1); c.Load(); c.Store(42); c.CompareAndSwap(old, new)

var b atomic.Bool
b.Load(); b.Store(true); b.CompareAndSwap(false, true)

var p atomic.Pointer[T]
p.Load(); p.Store(&x)
```

Старый стиль (`atomic.AddInt64(&x, 1)`) работает, но обёртки безопаснее:
компилятор не даст смешать atomic/не-atomic доступ к одной переменной.

## Atomic / Mutex / Channel — когда что

| Ситуация | Инструмент |
|---|---|
| Один скаляр (счётчик, флаг, указатель) | `sync/atomic` |
| Несколько полей с инвариантом между ними | `sync.Mutex` / `sync.RWMutex` |
| Передача владения данными | канал |
| Однократная инициализация | `sync.Once` |
| Много чтений, редкие записи | `sync.RWMutex` или `atomic.Pointer` + COW |

Правило: `atomic` — это **один** LOAD/STORE/ADD. "Прочитать A, проверить,
изменить B" — уже не атомарно, нужен Mutex.

## Цена

- x86: ~10–20 нс на atomic-add (lock prefix).
- ARM: дороже из-за fence-инструкций.
- Под высокой contention atomic-add на одном счётчике инвалидирует
  cache line у всех CPU — sharded counters могут быть в разы быстрее.

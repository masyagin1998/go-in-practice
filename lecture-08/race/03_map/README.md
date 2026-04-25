# 03_map — встроенный детектор гонок на map

Только у `map` (не у slice/struct/int) есть встроенная защёлка от
конкурентных записей. На каждом `m[k] = v`, `delete(m, k)` и `m[k]`
рантайм проверяет бит `hashWriting` в `hmap.flags`; при коллизии
вызывает `runtime.throw` → `fatal error: concurrent map writes`.
Работает БЕЗ `-race`, не ловится `recover` (это `throw`, а не `panic`).

## Запуск

```bash
go run .          # fatal error: concurrent map writes
go run -race .    # сначала WARNING: DATA RACE, потом всё равно fatal
```

## Как это устроено

```go
// runtime/map.go (упрощённо)
const hashWriting = 4

func mapassign(...) ... {
    if h.flags & hashWriting != 0 {
        fatal("concurrent map writes")
    }
    h.flags ^= hashWriting
    // ... insert ...
    h.flags &^= hashWriting
}
```

Один `AND` + `JNZ` на каждом доступе — стоит почти ноль, поэтому
включено всегда и отключить нельзя.

## Какие сообщения бывают

| Что произошло | Что напечатает |
|---|---|
| Два concurrent write | `fatal error: concurrent map writes` |
| Read + write | `fatal error: concurrent map read and map write` |
| Два concurrent read без write | безопасно, ничего |

## Важные оговорки

- **Best-effort**, а не TSan. `h.flags` модифицируется без `lock`-
  префикса — какие-то гонки можно не поймать. Для гарантий нужен `-race`.
- **Только для map.** Срезы / структуры / числа всё ещё гонят молча;
  их ловит только `-race` (см. 04, 05).
- **`recover` не помогает.** `throw` обходит механизм паник: после
  concurrent-write структура map уже сломана (повисший флаг, частично
  вставленный элемент) — продолжать с ней нельзя, проще убить процесс.

## Чем заменить

- `sync.Map` — под "много read / редкие write" или "каждая G пишет в
  свои ключи". В общем случае медленнее обычной `map+mutex`.
- `map + sync.RWMutex` / `sync.Mutex` — универсальный, обычно самый
  быстрый вариант.
- Sharded map (массив `[N]struct{m map[K]V; mu sync.Mutex}` + ключ →
  `hash(key) % N`) — под высокую contention.

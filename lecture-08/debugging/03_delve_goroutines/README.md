# 03_delve_goroutines — goroutine-aware отладка

5 goroutine: четыре воркера отрабатывают и выходят, пятая висит на
чтении из канала, в который никто не пишет. Учимся находить
зависшую goroutine в `dlv`.

## Запуск

```bash
dlv debug
(dlv) break main.hang
(dlv) continue
(dlv) bt
(dlv) goroutines
(dlv) goroutine 6   # ID из предыдущего вывода
(dlv) bt
```

## Ключевые команды для goroutine'ов

| Команда | Что делает |
|---|---|
| `goroutines` | список всех goroutine с их текущей позицией |
| `goroutine` | показать ID и статус текущей |
| `goroutine <N>` | переключиться в контекст goroutine N |
| `goroutine <N> bt` | backtrace указанной goroutine (без переключения) |
| `goroutine <N> frame <F>` | стек-фрейм F в goroutine N |

## Что полезно увидеть в выводе

- Каждая goroutine показывает свой статус: `running`, `runnable`,
  `waiting` (и причину: `chan receive`, `select`, `IO wait`).
- Зависшая `hang` будет в статусе `chan receive` с local'ами, где видно,
  что `ch` — это наш `stuck`.
- Worker'ы после завершения пропадают из списка (если убили до `Sleep`) —
  делать `break` лучше до `wg.Wait()`.


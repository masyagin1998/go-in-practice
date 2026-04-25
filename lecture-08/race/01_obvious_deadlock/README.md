# 01_obvious_deadlock — рантайм ловит сам, `-race` мешает

AB-BA на двух мьютексах. Все goroutine (main, g1, g2) паркуются →
шедулер видит, что нет ни одной runnable, и печатает:

```
fatal error: all goroutines are asleep - deadlock!
```

Это `runtime.checkdead()` (см. `runtime/proc.go`) — лёгкая проверка
шедулера, всегда включена.

## Запуск

```bash
go run .          # fatal error за миллисекунды
go run -race .    # ВИСНЕТ — режем по таймауту
```

## Сюрприз: под `-race` детектор дедлоков сломан

Race-детектор (TSan) держит свои фоновые goroutine живыми, поэтому
`checkdead()` думает "есть кому работать" и не срабатывает. Программа
просто молча висит. Известный артефакт: [golang/go#13098][1].

То есть `-race` тут **не помогает** (он ловит data race, а не deadlock)
**и заодно отключает встроенный детектор дедлоков**. Поэтому `-race`
не заменяет нагрузочные/end-to-end запуски без него.

[1]: https://github.com/golang/go/issues/13098

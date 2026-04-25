# 04_counter_race — RMW на общем int

Две goroutine по 1M раз делают `counter++`. Ожидаем 2_000_000,
получаем меньше — каждый раз разное.

## Запуск

```bash
go run .          # неверный результат (< 2M)
go run -race .    # WARNING: DATA RACE
```

## Почему `counter++` не атомарно

```
mov rax, [counter]   ; read
add rax, 1           ; modify
mov [counter], rax   ; write
```

Между read и write вторая goroutine может прочитать ту же старую
`counter` или перетереть наш write — оба варианта = потерянный инкремент.

## Что печатает -race

```
==================
WARNING: DATA RACE
Read at 0x... by goroutine 7:  main.main.func1() main.go:N
Previous write at 0x... by goroutine 6: main.main.func1() main.go:N
==================
```

Детектор показывает обе стороны гонки со стеками. Фикс — в `06_atomic_fix/`.

## Артефакт

- `race.txt` — полный трейс детектора (после `make run`).

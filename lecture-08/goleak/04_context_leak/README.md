# 04_context_leak — забытый ctx.Done

Самая частая утечка в реальных сервисах: consumer читает из канала через
`range` / `<-in`, не слушает контекст. Producer уходит, не закрыв
канал — consumer висит вечно.

## Запуск

```bash
go test -v .
```

`TestWorkerLeaky` — FAIL с дампом, `TestWorkerFixed` — PASS.

`TestWorkerFixed` использует `goleak.IgnoreCurrent()`: goroutine из
предыдущего теста всё ещё жива в процессе и засорила бы baseline.

## Два паттерна consumer'а

### "Producer закрывает канал"
```go
for v := range in {
    ...
}
```
Работает, если producer **гарантированно** делает `close(in)` при
отмене. Типичный промах: producer и consumer оба слушают ctx, и ни
одна сторона не знает, кто закрывает канал.

### "Consumer слушает контекст" (канонический)
```go
for {
    select {
    case <-ctx.Done():
        return
    case v, ok := <-in:
        if !ok { return }
        ...
    }
}
```
Всегда работает. Это шаблон по умолчанию для worker-pool / pipeline.

## Вложенный select на записи

Если consumer пишет в out, эту запись тоже надо обернуть в select:

```go
select {
case out <- v:
case <-ctx.Done():
    return
}
```

Иначе тот же баг, но на стороне записи: ждём, пока кто-то заберёт,
а никто не забирает.

## Реальный сценарий

Сервис принимает задачи из Kafka → на каждую стартует
`WorkerLeaky(ctx, ...)` → клиент отменяет запрос (`cancel()`) →
goroutine продолжает читать → через неделю NumGoroutine = 50 000 → OOM.

goleak в юнит-тесте обработчика поймал бы это ещё в PR.

## Артефакт

- `goleak-output.txt` — `TestWorkerLeaky FAIL` со стеком, `TestWorkerFixed PASS`.

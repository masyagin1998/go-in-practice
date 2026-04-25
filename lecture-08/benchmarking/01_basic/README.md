# 01_basic — testing.B на пальцах

`sum.go` / `sum_test.go` — три реализации суммирования (`index`, `range`,
`unroll`) + тесты на корректность + бенчмарк-матрица 3×3 (реализация ×
размер входа).

`bufreuse_test.go` — пара бенчей про переиспользование буфера vs
`make` на каждой итерации.

## Запуск

```bash
# тесты
go test -v ./...

# только бенчмарки, тесты пропускаем через -run='^$'
go test -run='^$' -bench=. -benchmem -benchtime=1s

# конкретный бенч
go test -run='^$' -bench=BenchmarkSum -benchmem

# через make
make test
make bench
```

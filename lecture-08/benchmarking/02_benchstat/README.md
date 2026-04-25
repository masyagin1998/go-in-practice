# 02_benchstat — статистически значимое сравнение

`format.go` / `format_test.go` — две реализации `int → string`
(`FormatFmt` через `fmt.Sprintf`, `FormatStrconv` через `strconv.Itoa`)
+ тесты на корректность + бенчи под `benchstat`.

## Установка benchstat

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Запуск

```bash
# тесты
go test -v ./...

# бенчи + сравнение (через make: old.txt, new.txt, delta.txt)
make bench

# вручную
go test -run='^$' -bench=BenchmarkFormatFmt     -benchmem -count=10 > old.txt
go test -run='^$' -bench=BenchmarkFormatStrconv -benchmem -count=10 > new.txt
benchstat old.txt new.txt
```

В `Makefile` имена бенчей унифицируются через `sed` (`BenchmarkFormatFmt`
и `BenchmarkFormatStrconv` → `BenchmarkFormat`), чтобы `benchstat` свёл
их в одну строку. В реальном workflow вы бы просто правили код между
двумя замерами.

## Что показывает benchstat

```
          │   old.txt   │               new.txt                │
          │   sec/op    │    sec/op     vs base                │
Format-28   52.65n ± 4%   20.29n ± 13%  -61.45% (p=0.000 n=10)
```

- `±` — median absolute deviation.
- `p=...` — p-value t-теста; `> 0.05` → шум, benchstat напишет `~`.
- `n=10` — число замеров (`-count=10`).

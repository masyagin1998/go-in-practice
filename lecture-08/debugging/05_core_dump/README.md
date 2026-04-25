# 05_core_dump — post-mortem через gcore + dlv core

Снимаем core dump с живого процесса и анализируем его в `dlv`. Полезно,
когда:
- процесс висит, а ребутать нельзя (прод);
- баг воспроизводится раз в неделю, и нужно зафиксировать момент;
- dlv/gdb attach к реальному процессу слишком рискованно.

## Предусловия

```bash
# Разрешить ptrace без sudo:
sudo sysctl -w kernel.yama.ptrace_scope=0
# (или запускать gcore под sudo)
```

Пакет `gdb` должен быть установлен — в нём лежит `gcore`.

## Сценарий

Терминал 1:
```bash
go build -o gobin .
./gobin
```

Терминал 2 (пока программа "висит" на time.Sleep):
```bash
gcore $(pgrep gobin)               # создаст core.<PID> в cwd
dlv core ./gobin core.<PID>
(dlv) goroutines
(dlv) goroutine <N>
(dlv) bt
(dlv) frame <F>
(dlv) locals
```

Полный лог — в `core-session.txt`.

## Особенности

- **Живой процесс не трогается**: gcore делает ptrace-attach, читает
  память, detach. Пауза на секунды, процесс продолжает работу.
- **Размер core** — весь VSZ процесса, у Go-бинарей это легко сотни
  мегабайт. `ulimit -c unlimited` и достаточно свободного места.
- **Срез момента**: goroutine'ы, локальные, стек — всё как в живой
  отладке. Но корректно "шагать" `next`/`step` нельзя — процесс мёртв
  (в core). `print`, `bt`, `goroutines` работают.
- **Автоматические core при падении**: `GOTRACEBACK=crash` + системный
  `core_pattern` (см. `/proc/sys/kernel/core_pattern`) — ядро сохранит
  core автоматически при panic/SIGSEGV.


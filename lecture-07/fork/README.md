# fork/ — создание процессов в Go

В C-курсе здесь `fork()`, `vfork()`, `posix_spawn()`. В Go эта тройка
трансформируется:

- **Чистый `fork()` недоступен** пользовательскому коду. Go-runtime
  держит несколько OS-тредов (GC, scavenger, sysmon); дублирование
  адресного пространства без немедленного `exec` ломает их фон. В Go
  всегда **fork+exec**.
- **`vfork` не экспонируется**. `syscall.ForkExec` на Linux уже
  использует `clone(CLONE_VM|CLONE_VFORK)` — быстрее CoW.
- **`posix_spawn` аналог** — `os/exec.Command`, stdlib выбирает лучший
  под платформу.

Два реальных кейса:

| Подкаталог | Когда это нужно в проде |
|---|---|
| `exec-spawn/` | запустить внешний инструмент (git, ffmpeg, cc1, migrate) и прочитать вывод |
| `syscall-exec/` | upgrade-in-place, privilege drop, re-exec из supervisor'а |

## А что с демонизацией (двойной fork)?

В Go — **не надо**. Runtime многотредовый, fork без exec = UB. Современные
системы (systemd / launchd) делают detach сами: пишите `Type=simple`.
Без systemd — `nohup ./app &` или `setsid`.

## Запуск

```bash
make run         # оба примера
make -C exec-spawn run
```

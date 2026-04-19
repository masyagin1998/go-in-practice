# signalfd — сигналы как файловый дескриптор

Linux-specific механизм: `signalfd(2)` возвращает fd, из которого можно
читать приходящие сигналы как структуры `signalfd_siginfo` (128 байт
фиксированного layout). Полезно:

- **Получить payload**. `sigqueue(2)` умеет класть `sival_int`/`sival_ptr`,
  Go-овский `signal.Notify` этого не показывает — только номер.
- **Интегрировать сигналы в `epoll`/event loop**. Сигналы перестают быть
  "async magic", становятся ещё одним fd наравне с сокетами.
- **Знать, кто прислал**: `ssi_pid`/`ssi_uid` — PID и UID отправителя.

## Схема демо

```
Go-parent                                            C-peer (./peer)
─────────────                                        ─────────────────
sigprocmask BLOCK {USR1,USR2}  ──(inherited)──►      (наследует маску
(в C-constructor'е, до runtime)                       только если был бы fork,
signalfd(-1, &set, 0)  → fd                           здесь exec — не важно)
exec ./peer <pid>                                    sigqueue(pid, USR1, 100)
read fd → signalfd_siginfo                           sigqueue(pid, USR2, 101)
read fd → …                                          …
```

## Почему сигналы должны быть **заблокированы**

`signalfd` отдаёт только pending-сигналы. Если SIGUSR1 не заблокирован в
каком-то треде процесса, ядро доставит его этому треду (вызовет
Go-обработчик), и в `signalfd` он не попадёт.

Значит, нужно заблокировать SIGUSR1/SIGUSR2 **во всех Go-тредах**. Go
runtime создаёт M-треды по ходу работы программы, и каждый наследует
маску создающего треда.

**Трюк** в `main.go`: блокировать сигналы из
`__attribute__((constructor))`. Это выполняется на этапе загрузки
бинаря, **до** `runtime.rt0_go`. Главный тред стартует уже с
заблокированными SIGUSR1/SIGUSR2 — все последующие Go-M-треды
наследуют маску. Go runtime при создании M не разблокирует "blockable"
сигналы (а SIGUSR1/SIGUSR2 именно такие), поэтому маска сохраняется.

## Реальные кейсы

- **High-perf серверы**: собственный event loop на `epoll`, в нём же
  слушаем signalfd — без `signal.Notify` и лишней goroutine.
- **Супервизоры/orchestrator'ы**: отслеживать SIGCHLD через signalfd,
  получать `ssi_pid`/`ssi_status` упавших детей — не надо никаких
  `wait(-1)` лотерей.
- **Передача команд с payload**: другой процесс шлёт нам `sigqueue(pid,
  SIGUSR1, command_code)` — мы в одном месте разбираем `ssi_int`. Грубо,
  но работает без сокетов и без общей памяти.

## Ограничения

- **Linux-only**. На macOS/BSD нет signalfd — там `kqueue(EVFILT_SIGNAL)`
  с похожим API.
- **Хрупко с runtime'ом**. Конструкторная установка маски — паттерн на
  грани. Если Go runtime когда-нибудь начнёт разблокировать SIGUSR*
  в `minitSignals`, пример сломается без предупреждения. Для
  production-кода лучше `golang.org/x/sys/unix.Signalfd` +
  явное управление тредами через `runtime.LockOSThread` — но это
  нарастающая сложность.
- **128 байт на сигнал**. Ограниченный payload. Нужен реальный обмен —
  используйте UDS/shmem рядом, а сигналом только "пинайте".

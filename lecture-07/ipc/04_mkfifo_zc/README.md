# 04_mkfifo_zc — FIFO + vmsplice(SPLICE_F_GIFT)

То же, что `02_pipes_zc`, только pipe именованный: peer и reader могут
стартовать независимо (не нужен общий родитель — достаточно пути в ФС).

Демо одностороннее, как `01_pipes`: peer отправляет 5 сообщений через
`vmsplice(SPLICE_F_GIFT)`, Go читает обычным `read(2)`.

## Почему zero-copy только на TX-стороне

`splice(2)` может перенести данные из одного pipe'а в другой без копии
в user-space, но чтобы zero-copy работал на reader'е, Go должен читать
через `splice(pipe_fd, other_fd, ...)` — а не в свой буфер. У нас Go
только хочет увидеть данные → копия неизбежна.

Для end-to-end zero-copy нужно, чтобы оба конца были fd'ами ядра
(pipe→pipe, pipe→socket, file→socket). Пример такого — `sendfile(2)`
в `06_sockets_zc`.

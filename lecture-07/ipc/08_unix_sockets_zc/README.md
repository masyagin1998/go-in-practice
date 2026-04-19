# 08_unix_sockets_zc — AF_UNIX + SCM_RIGHTS + memfd

Настоящий zero-copy: данные кладутся в анонимный **memfd** (страницы в
page cache, без файла на диске), потом **сам fd** пересылается через
Unix-socket. Получатель `mmap`'ит его и видит те же физические страницы.

```
peer:                                      go:
  mfd = memfd_create(...)                   recvmsg + SCM_RIGHTS → fd
  write(mfd, data)                          mmap(fd) ─┐
  sendmsg(SCM_RIGHTS, mfd) ──────────►    ←─ оба указывают на ОДНИ
                                           │  физические страницы
  close(mfd)  // Go держит свою копию fd
```

Ни одной копии данных — ни user→kernel, ни kernel→user.

## Отличие от shmem (10_shmem)

| | `shm_open` | `memfd + SCM_RIGHTS` |
|---|---|---|
| namespace | `/dev/shm/name` — viewable, persistent | ephemeral, исчезает при закрытии fd |
| discovery | обе стороны знают path | передача fd через socket |
| кто может присоединиться | любой с правами на path | только тот, кому явно передали fd |
| чистка после краха | нужен `shm_unlink` | сам собой — fd умирает вместе с процессом |

## Когда использовать

- Высокочастотный обмен большими буферами между parent-child / cooperating процессами.
- Когда не хочется светить имя объекта в ФС (безопасность, изоляция контейнеров).
- Когда нужен **capability-based** контроль: "вот тебе fd, больше ни у кого его нет".

## Ограничение

Передать fd через сеть нельзя — только через Unix-socket на одной
машине. Для межмашинного zero-copy обычно используют RDMA или отдельные
транспорты (напр. `io_uring` + SQE с registered buffers).

// shared.h — layout shared memory, общий между Go и C.
#ifndef SHARED_H
#define SHARED_H

#include <signal.h>

#define SHM_NAME   "/ipc_shmem_demo"
#define SHM_BUF    1024
#define ITERATIONS 5

typedef struct {
    volatile sig_atomic_t client_ready;
    volatile sig_atomic_t server_ready;
    char buffer[SHM_BUF];
} shared_mem_t;

#endif

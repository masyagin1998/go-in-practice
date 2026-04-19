#ifndef SHARED_H
#define SHARED_H

#define SHM_NAME   "/ipc_sem_demo"
#define SEM_CLIENT "/ipc_sem_client"
#define SEM_SERVER "/ipc_sem_server"
#define SHM_BUF    1024
#define ITERATIONS 5

typedef struct {
    char buffer[SHM_BUF];
} shared_mem_t;

#endif

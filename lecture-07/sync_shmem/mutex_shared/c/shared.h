#ifndef SHARED_H
#define SHARED_H

#include <pthread.h>

#define SHM_NAME   "/ipc_mutex_shared"
#define WORKERS    4
#define ITERATIONS 100

typedef struct {
    pthread_mutex_t mutex;
    int counter;
} shared_data_t;

#endif

#ifndef SHARED_H
#define SHARED_H

#include <pthread.h>

#define SHM_NAME "/ipc_mutex_death"

typedef struct {
    pthread_mutex_t mutex;
    int value;
} shared_t;

#endif

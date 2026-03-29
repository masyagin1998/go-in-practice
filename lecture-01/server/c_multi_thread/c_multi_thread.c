/* Multi-threaded TCP server (thread-per-request). Usage: ./c_multi_thread <fib|sleep> */

#include "../c_lib/server_lib.h"
#include <pthread.h>

struct client_args { int fd; struct sockaddr_in addr; };

static void *client_thread(void *arg) {
    struct client_args *ca = arg;
    handle_client(ca->fd, &ca->addr);
    free(ca);
    return NULL;
}

int main(int argc, char **argv) {
    parse_mode(argc, argv);
    int srv = server_listen();
    log_msg("listening on :%d (multi-thread, mode=%s)", PORT, argv[1]);

    for (;;) {
        struct client_args *ca = malloc(sizeof(*ca));
        if (!ca) continue;
        socklen_t len = sizeof(ca->addr);
        ca->fd = accept(srv, (struct sockaddr *)&ca->addr, &len);
        if (ca->fd < 0) { free(ca); continue; }

        pthread_t tid;
        if (pthread_create(&tid, NULL, client_thread, ca) != 0) {
            close(ca->fd); free(ca); continue;
        }
        pthread_detach(tid);
    }
}

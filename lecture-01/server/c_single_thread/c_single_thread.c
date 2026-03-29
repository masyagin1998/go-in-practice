/* Single-threaded TCP server. Usage: ./c_single_thread <fib|sleep> */

#include "../c_lib/server_lib.h"

int main(int argc, char **argv) {
    parse_mode(argc, argv);
    int srv = server_listen();
    log_msg("listening on :%d (single-thread, mode=%s)", PORT, argv[1]);

    for (;;) {
        struct sockaddr_in addr;
        socklen_t len = sizeof(addr);
        int fd = accept(srv, (struct sockaddr *)&addr, &len);
        if (fd < 0) continue;
        handle_client(fd, &addr);
    }
}

#include "server_lib.h"

enum server_mode g_mode;

void log_msg(const char *fmt, ...) {
    time_t now = time(NULL);
    struct tm tm;
    localtime_r(&now, &tm);
    fprintf(stderr, "%04d/%02d/%02d %02d:%02d:%02d ",
            tm.tm_year + 1900, tm.tm_mon + 1, tm.tm_mday,
            tm.tm_hour, tm.tm_min, tm.tm_sec);
    va_list ap;
    va_start(ap, fmt);
    vfprintf(stderr, fmt, ap);
    va_end(ap);
    fprintf(stderr, "\n");
}

uint64_t fib(uint64_t n) {
    if (n <= 1) return n;
    return fib(n - 1) + fib(n - 2);
}

void sleep_ms(void) {
    struct timespec ts = { .tv_sec = 0, .tv_nsec = SLEEP_MS * 1000000L };
    nanosleep(&ts, NULL);
}

void parse_mode(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "Usage: %s <fib|sleep>\n", argv[0]);
        exit(1);
    }
    if (strcmp(argv[1], "fib") == 0) {
        g_mode = MODE_FIB;
    } else if (strcmp(argv[1], "sleep") == 0) {
        g_mode = MODE_SLEEP;
    } else {
        fprintf(stderr, "Unknown mode: %s (use 'fib' or 'sleep')\n", argv[1]);
        exit(1);
    }
}

void handle_client(int fd, struct sockaddr_in *addr) {
    char peer[INET_ADDRSTRLEN];
    inet_ntop(AF_INET, &addr->sin_addr, peer, sizeof(peer));
    int peer_port = ntohs(addr->sin_port);

    char buf[BUF_SIZE];
    ssize_t n = read(fd, buf, sizeof(buf) - 1);
    if (n <= 0) {
        close(fd);
        return;
    }
    buf[n] = '\0';

    char resp[32];
    int len;

    if (g_mode == MODE_FIB) {
        uint64_t val;
        if (sscanf(buf, "%" SCNu64, &val) != 1) {
            const char *err = "error: expected integer N\n";
            write(fd, err, strlen(err));
            close(fd);
            return;
        }
        uint64_t result = fib(val);
        len = snprintf(resp, sizeof(resp), "%" PRIu64 "\n", result);
    } else {
        sleep_ms();
        len = snprintf(resp, sizeof(resp), "42\n");
    }

    write(fd, resp, len);
    close(fd);
}

int server_listen(void) {
    signal(SIGPIPE, SIG_IGN);

    int srv = socket(AF_INET, SOCK_STREAM, 0);
    if (srv < 0) { perror("socket"); exit(1); }

    int opt = 1;
    setsockopt(srv, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    struct sockaddr_in addr = {
        .sin_family = AF_INET,
        .sin_port = htons(PORT),
        .sin_addr.s_addr = INADDR_ANY,
    };

    if (bind(srv, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("bind"); exit(1);
    }
    if (listen(srv, 128) < 0) {
        perror("listen"); exit(1);
    }
    return srv;
}

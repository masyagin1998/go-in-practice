#ifndef SERVER_LIB_H
#define SERVER_LIB_H

#define _POSIX_C_SOURCE 200809L

#include <arpa/inet.h>
#include <errno.h>
#include <inttypes.h>
#include <netinet/in.h>
#include <signal.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <time.h>
#include <unistd.h>

#define PORT 9001
#define BUF_SIZE 256
#define SLEEP_MS 100

enum server_mode { MODE_FIB, MODE_SLEEP };

extern enum server_mode g_mode;

void log_msg(const char *fmt, ...);
uint64_t fib(uint64_t n);
void sleep_ms(void);

/* Parse mode from argv: "fib" or "sleep". Exits on bad input. */
void parse_mode(int argc, char **argv);

/* Blocking handle: read request, compute/sleep, write response, close. */
void handle_client(int fd, struct sockaddr_in *addr);

/* Create, bind, listen on TCP socket. Returns fd or exits on error. */
int server_listen(void);

#endif

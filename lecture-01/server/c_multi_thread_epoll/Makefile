CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c11
LDFLAGS = -pthread
LIB_SRC = ../c_lib/server_lib.c
TARGET = c_multi_thread_epoll

all: $(TARGET)

$(TARGET): $(TARGET).c $(LIB_SRC)
	$(CC) $(CFLAGS) -o $@ $(TARGET).c $(LIB_SRC) $(LDFLAGS)

clean:
	rm -f $(TARGET)

.PHONY: all clean

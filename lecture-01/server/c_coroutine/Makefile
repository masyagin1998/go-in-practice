CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=gnu11
LIB_SRC = ../c_lib/server_lib.c
TARGET = c_coroutine

all: $(TARGET)

$(TARGET): $(TARGET).c $(LIB_SRC)
	$(CC) $(CFLAGS) -o $@ $(TARGET).c $(LIB_SRC)

clean:
	rm -f $(TARGET)

.PHONY: all clean

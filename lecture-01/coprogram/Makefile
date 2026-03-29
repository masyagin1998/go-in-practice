CC = gcc
CFLAGS = -std=gnu11 -O2 -Wall -Wextra
TARGET = coro

$(TARGET): coprogram.c
	$(CC) $(CFLAGS) -o $(TARGET) coprogram.c

clean:
	rm -f $(TARGET)

.PHONY: clean

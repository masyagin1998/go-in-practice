CC = gcc
CFLAGS = -Wall -Wextra -g

.PHONY: all clean

all: 01_manual 02_boehm_gc

01_manual: 01_manual.c
	$(CC) $(CFLAGS) -o $@ $<

# Требует libgc-dev: sudo apt install libgc-dev
02_boehm_gc: 02_boehm_gc.c
	$(CC) $(CFLAGS) -o $@ $< -lgc

clean:
	rm -f 01_manual 02_boehm_gc

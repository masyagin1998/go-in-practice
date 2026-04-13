CC      = gcc
CFLAGS  = -Wall -Wextra -g -O0
TARGETS = linear_allocator block_allocator stack_allocator stupid_malloc

all: $(TARGETS)

linear_allocator: linear_allocator.c
	$(CC) $(CFLAGS) -o $@ $<

block_allocator: block_allocator.c
	$(CC) $(CFLAGS) -o $@ $<

stack_allocator: stack_allocator.c
	$(CC) $(CFLAGS) -o $@ $<

stupid_malloc: stupid_malloc.c
	$(CC) $(CFLAGS) -o $@ $<

# Собрать с brk вместо mmap
stupid_malloc_brk: stupid_malloc.c
	$(CC) $(CFLAGS) -DUSE_BRK=1 -o $@ $<

clean:
	rm -f $(TARGETS) stupid_malloc_brk

.PHONY: all clean

CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra -D_GNU_SOURCE

.PHONY: run clean

run: peer
	./peer | go run .

peer: c/peer.c
	$(CC) $(CFLAGS) -o peer c/peer.c

clean:
	rm -f peer
	go clean

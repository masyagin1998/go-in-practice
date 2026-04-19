CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Ручной запуск:
#   ./peer | go run .
# make run делает то же самое.
run: peer
	./peer | go run .

peer: c/peer.c
	$(CC) $(CFLAGS) -o peer c/peer.c

clean:
	rm -f peer
	go clean

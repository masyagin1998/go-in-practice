CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Два терминала:
#   терминал 1:  go run .    (создаёт shmem и busy-waits серверного флага)
#   терминал 2:  ./peer
run: peer
	go run . & GO=$$!; \
	sleep 0.3; \
	./peer; \
	wait $$GO

peer: c/peer.c c/shared.h
	$(CC) $(CFLAGS) -o peer c/peer.c -lrt

clean:
	rm -f peer /dev/shm/ipc_shmem_demo
	go clean

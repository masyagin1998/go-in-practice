CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Два терминала:
#   терминал 1:  go run .     (создаёт shmem+robust mutex, ждёт на lock)
#   терминал 2:  ./peer       (берёт mutex и падает без unlock)
run: peer gobin
	./gobin & GO=$$!; \
	sleep 0.3; \
	./peer; \
	wait $$GO

gobin: main.go go.mod
	go build -o gobin .

peer: c/peer.c c/shared.h
	$(CC) $(CFLAGS) -pthread -o peer c/peer.c -lrt

clean:
	rm -f peer gobin /dev/shm/ipc_mutex_death
	go clean

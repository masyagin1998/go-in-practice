CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Два терминала:
#   терминал 1:  go run .    (создаёт очереди и шлёт ping'и)
#   терминал 2:  ./peer      (читает ping'и, отвечает echo)
run: peer
	go run . & GO=$$!; \
	sleep 0.5; \
	./peer; \
	wait $$GO

peer: c/peer.c c/shared.h
	$(CC) $(CFLAGS) -o peer c/peer.c -lrt

clean:
	rm -f peer /dev/mqueue/ipc_mq_c2s /dev/mqueue/ipc_mq_s2c
	go clean

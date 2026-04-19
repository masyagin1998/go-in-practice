CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Ручной запуск в двух терминалах:
#   терминал 1:  go run .
#   терминал 2:  ./peer
# make run автоматизирует: Go на фоне, peer в foreground.
run: peer
	go run . & GO=$$!; \
	sleep 0.5; \
	./peer; \
	wait $$GO

peer: c/peer.c
	$(CC) $(CFLAGS) -o peer c/peer.c

clean:
	rm -f peer /tmp/ipc_fifo_c2s /tmp/ipc_fifo_s2c
	go clean

CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Несколько терминалов:
#   терминал 1:  go run .
#   терминалы 2-5:  ./peer 0, ./peer 1, ./peer 2, ./peer 3
# make run автоматизирует: Go на фоне, 4 peer'а, затем SIGINT в Go.
# Используем go build, чтобы kill -INT точно попал в процесс (а не в go run).
run: peer gobin
	./gobin & GO=$$!; \
	sleep 0.3; \
	./peer 0 & P1=$$!; \
	./peer 1 & P2=$$!; \
	./peer 2 & P3=$$!; \
	./peer 3 & P4=$$!; \
	wait $$P1 $$P2 $$P3 $$P4; \
	kill -INT $$GO; \
	wait $$GO

gobin: main.go go.mod
	go build -o gobin .

peer: c/peer.c c/shared.h
	$(CC) $(CFLAGS) -pthread -o peer c/peer.c -lrt

clean:
	rm -f peer gobin /dev/shm/ipc_mutex_shared
	go clean

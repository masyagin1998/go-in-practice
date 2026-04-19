CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Два терминала:
#   терминал 1:  go run .
#   терминал 2:  ./peer <PID> 100
# make run автоматизирует.
run: peer gobin
	./gobin & GO=$$!; \
	sleep 0.3; \
	./peer $$GO 100; \
	sleep 0.3; \
	kill -INT $$GO; \
	wait $$GO

gobin: main.go go.mod
	go build -o gobin .

peer: c/peer.c
	$(CC) $(CFLAGS) -o peer c/peer.c

clean:
	rm -f peer gobin
	go clean

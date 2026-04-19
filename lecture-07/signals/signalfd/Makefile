CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Два терминала:
#   терминал 1:  go run .         (печатает свой PID, ждёт 5 сигналов)
#   терминал 2:  ./peer <PID>     (шлёт 5 sigqueue'ов)
# make run автоматизирует (с built-бинарём, чтобы было проще найти PID).
run: peer gobin
	./gobin & GO=$$!; \
	sleep 0.3; \
	./peer $$GO; \
	wait $$GO

gobin: main.go go.mod
	go build -o gobin .

peer: c/peer.c
	$(CC) $(CFLAGS) -o peer c/peer.c

clean:
	rm -f peer gobin
	go clean

CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

.PHONY: run clean

# Два терминала:
#   терминал 1:  go run .
#   терминал 2:  ./peer
# make run: Go на фоне, peer в foreground.
run: peer
	go run . & GO=$$!; \
	sleep 0.5; \
	./peer; \
	wait $$GO

peer: c/peer.c
	$(CC) $(CFLAGS) -o peer c/peer.c

clean:
	rm -f peer
	go clean

CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra -D_GNU_SOURCE

PEERS = peer_pipe peer_pipe_zc peer_fifo peer_fifo_zc \
        peer_tcp peer_tcp_zc peer_uds peer_uds_zc \
        peer_mq peer_shmsem

.PHONY: bench bench-quick run clean

# Полный прогон: 10 транспортов × 2 размера сообщения × REPEATS. ~30 мин.
bench: $(PEERS)
	go run .

bench-quick: $(PEERS)
	BENCH_REPEATS=3 BENCH_ITERS=50000 go run .

run: bench-quick

peer_pipe:    peers/pipe.c
	$(CC) $(CFLAGS) -o $@ $<
peer_pipe_zc: peers/pipe_zc.c
	$(CC) $(CFLAGS) -o $@ $<
peer_fifo:    peers/fifo.c
	$(CC) $(CFLAGS) -o $@ $<
peer_fifo_zc: peers/fifo_zc.c
	$(CC) $(CFLAGS) -o $@ $<
peer_tcp:     peers/tcp.c
	$(CC) $(CFLAGS) -o $@ $<
peer_tcp_zc:  peers/tcp_zc.c
	$(CC) $(CFLAGS) -o $@ $<
peer_uds:     peers/uds.c
	$(CC) $(CFLAGS) -o $@ $<
peer_uds_zc:  peers/uds_zc.c
	$(CC) $(CFLAGS) -o $@ $<
peer_mq:      peers/mq.c
	$(CC) $(CFLAGS) -o $@ $< -lrt
peer_shmsem:  peers/shm_sem.c
	$(CC) $(CFLAGS) -o $@ $< -lrt -lpthread

clean:
	rm -f $(PEERS)
	rm -f /tmp/ipc_bench_fifo_c2s /tmp/ipc_bench_fifo_s2c \
	      /tmp/ipc_bench_fifo_zc_c2s /tmp/ipc_bench_fifo_zc_s2c \
	      /tmp/ipc_bench_uds /tmp/ipc_bench_uds_zc
	rm -f /dev/shm/ipc_bench_shm
	rm -f /dev/shm/sem.ipc_bench_sem_c /dev/shm/sem.ipc_bench_sem_s
	rm -f /dev/mqueue/ipc_bench_mq_c2s /dev/mqueue/ipc_bench_mq_s2c
	go clean

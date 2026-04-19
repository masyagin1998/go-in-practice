CC     = gcc
CFLAGS = -std=c11 -O2 -Wall -Wextra

PEERS = peer_pipe peer_fifo peer_mq peer_shmsem peer_tcp peer_uds

.PHONY: bench bench-quick run clean

# Полный прогон: по умолчанию ~10 минут (REPEATS=30, ITERS=500000).
# Можно варьировать через env: BENCH_REPEATS=50 BENCH_ITERS=1000000 make bench
bench: $(PEERS)
	go run .

# Быстрый прогон для CI / разработки: несколько секунд.
bench-quick: $(PEERS)
	BENCH_REPEATS=3 BENCH_ITERS=100000 go run .

# В root-Makefile'е цель `run` обходит RUN_DIRS; benchmark запускается через
# bench/bench-quick. Здесь `run` даём как алиас на bench-quick, чтобы локально
# быстро прогнать без ожидания.
run: bench-quick

peer_pipe: peers/pipe.c
	$(CC) $(CFLAGS) -o $@ $<

peer_fifo: peers/fifo.c
	$(CC) $(CFLAGS) -o $@ $<

peer_mq: peers/mq.c
	$(CC) $(CFLAGS) -o $@ $< -lrt

peer_shmsem: peers/shm_sem.c
	$(CC) $(CFLAGS) -o $@ $< -lrt -lpthread

peer_tcp: peers/tcp.c
	$(CC) $(CFLAGS) -o $@ $<

peer_uds: peers/uds.c
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm -f $(PEERS)
	rm -f /tmp/ipc_bench_fifo_c2s /tmp/ipc_bench_fifo_s2c /tmp/ipc_bench_uds
	rm -f /dev/shm/ipc_bench_shm
	rm -f /dev/shm/sem.ipc_bench_sem_c /dev/shm/sem.ipc_bench_sem_s
	rm -f /dev/mqueue/ipc_bench_mq_c2s /dev/mqueue/ipc_bench_mq_s2c
	go clean

# Запускает все примеры lecture-07 (IPC).
#
# make run          — прогнать все демо (fork, ipc, sync_shmem, signals)
# make bench        — полный ping-pong бенчмарк (~10 минут)
# make bench-quick  — быстрый бенчмарк для разработки/CI (~10 секунд)
# make clean        — очистить артефакты
# make all          — run + bench

RUN_DIRS   = fork ipc sync_shmem signals
BENCH_DIRS = benchmark

.PHONY: all run bench bench-quick clean

all: run bench

run:
	@for d in $(RUN_DIRS); do \
		echo "###### $$d ######"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

bench:
	@for d in $(BENCH_DIRS); do \
		echo "###### $$d ######"; \
		$(MAKE) -C $$d bench; \
		echo; \
	done

bench-quick:
	@for d in $(BENCH_DIRS); do \
		echo "###### $$d (quick) ######"; \
		$(MAKE) -C $$d bench-quick; \
		echo; \
	done

clean:
	@for d in $(RUN_DIRS) $(BENCH_DIRS); do $(MAKE) -C $$d clean; done

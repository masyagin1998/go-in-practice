# Запускает все примеры lecture-06 (CGo).
#
# make run   — запустить все примеры
# make bench — запустить бенчмарки
# make clean — очистить артефакты
# make all   — run + bench

RUN_DIRS    = simple-cgo memory-cgo errno-cgo directives-cgo scheduler-cgo semaphore-cgo bugs-cgo
BENCH_DIRS  = bench-empty-cgo speedup-cgo

.PHONY: all run bench clean

all: run bench

run:
	@for d in $(RUN_DIRS); do \
		echo "====== $$d ======"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

bench:
	@for d in $(BENCH_DIRS); do \
		echo "====== $$d ======"; \
		$(MAKE) -C $$d bench; \
		echo; \
	done

clean:
	@for d in $(RUN_DIRS) $(BENCH_DIRS); do $(MAKE) -C $$d clean; done

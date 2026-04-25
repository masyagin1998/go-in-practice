# Запускает все демо lecture-08.
#
# make run    — прогнать debugging, benchmarking, profiling, race, goleak
# make clean  — убрать артефакты
# make all    — то же, что run

RUN_DIRS = debugging benchmarking profiling race goleak

.PHONY: all run clean

all: run

run:
	@for d in $(RUN_DIRS); do \
		echo "###### $$d ######"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

clean:
	@for d in $(RUN_DIRS); do $(MAKE) -C $$d clean; done

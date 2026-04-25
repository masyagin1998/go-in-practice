SUBDIRS = 01_heap 02_trace 03_cpu 04_block_mutex 05_real_program

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "=== profiling/$$d ==="; \
		$(MAKE) -C $$d run; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

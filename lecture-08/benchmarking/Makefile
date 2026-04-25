SUBDIRS = 01_basic 02_benchstat

.PHONY: run bench clean $(SUBDIRS)

run: bench

bench:
	@for d in $(SUBDIRS); do \
		echo "=== benchmarking/$$d ==="; \
		$(MAKE) -C $$d bench; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

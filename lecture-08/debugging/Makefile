SUBDIRS = 01_delve_basic 02_fibonacci 03_delve_goroutines 04_runtime_stack 05_core_dump

.PHONY: run clean $(SUBDIRS)

run:
	@for d in $(SUBDIRS); do \
		echo "=== debugging/$$d ==="; \
		$(MAKE) -C $$d run; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

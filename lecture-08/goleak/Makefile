SUBDIRS = 01_leak_basic 02_counter_false_positive 03_goleak_test 04_context_leak

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "=== goleak/$$d ==="; \
		$(MAKE) -C $$d run; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

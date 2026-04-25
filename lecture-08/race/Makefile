SUBDIRS = 01_obvious_deadlock 02_subtle_deadlock 03_map 04_counter_race 05_flag_race 06_atomic_fix

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "=== race/$$d ==="; \
		$(MAKE) -C $$d run || true; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

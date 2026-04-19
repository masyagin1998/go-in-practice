SUBDIRS = mutex_shared mutex_death

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "====== sync_shmem/$$d ======"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

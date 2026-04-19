SUBDIRS = 01_pipes 02_mkfifo 03_sockets 04_unix_sockets 05_queue 06_shmem 07_semaphore

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "====== ipc/$$d ======"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

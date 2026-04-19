SUBDIRS = 01_pipes 02_pipes_zc 03_mkfifo 04_mkfifo_zc 05_sockets 06_sockets_zc 07_unix_sockets 08_unix_sockets_zc 09_queue 10_shmem 11_semaphore

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "====== ipc/$$d ======"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

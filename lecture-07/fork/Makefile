SUBDIRS = exec-spawn syscall-exec graceful-reload

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "====== fork/$$d ======"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

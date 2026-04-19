SUBDIRS = graceful-shutdown sighup-reload signalfd thread-masking

.PHONY: run clean

run:
	@for d in $(SUBDIRS); do \
		echo "====== signals/$$d ======"; \
		$(MAKE) -C $$d run; \
		echo; \
	done

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

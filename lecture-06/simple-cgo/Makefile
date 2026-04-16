SUBDIRS = call-c-from-go call-go-from-c combo

.PHONY: run clean $(SUBDIRS)

run: $(SUBDIRS)

$(SUBDIRS):
	@echo "=== simple-cgo/$@ ==="
	$(MAKE) -C $@ run
	@echo

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

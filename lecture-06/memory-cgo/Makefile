SUBDIRS = simple-values strings structs struct-fields arrays c-to-go pinner

.PHONY: run clean $(SUBDIRS)

run: $(SUBDIRS)

$(SUBDIRS):
	@echo "=== memory-cgo/$@ ==="
	$(MAKE) -C $@ run
	@echo

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

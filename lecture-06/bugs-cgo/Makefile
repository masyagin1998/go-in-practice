SUBDIRS = variadic go-pointer preamble-gap complex-align union

.PHONY: run clean $(SUBDIRS)

run: $(SUBDIRS)

$(SUBDIRS):
	@echo "=== bugs-cgo/$@ ==="
	$(MAKE) -C $@ run
	@echo

clean:
	@for d in $(SUBDIRS); do $(MAKE) -C $$d clean; done

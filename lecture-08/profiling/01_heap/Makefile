.PHONY: run flame clean

run:
	go run .
	@echo
	@echo "=== inuse_space (живёт прямо сейчас) ==="
	@go tool pprof -top -sample_index=inuse_space -nodecount=10 heap.pprof \
		| tee heap-inuse.txt
	@echo
	@echo "=== alloc_space (выделено за всё время) ==="
	@go tool pprof -top -sample_index=alloc_space -nodecount=10 heap.pprof \
		| tee heap-alloc.txt
	@echo
	@echo "=== inuse_objects ==="
	@go tool pprof -top -sample_index=inuse_objects -nodecount=10 heap.pprof
	@echo
	@echo "=== alloc_objects ==="
	@go tool pprof -top -sample_index=alloc_objects -nodecount=10 heap.pprof

# Интерактивный flamegraph в браузере.
flame:
	go tool pprof -http=: heap.pprof

clean:
	rm -f heap.pprof heap-inuse.txt heap-alloc.txt
	go clean

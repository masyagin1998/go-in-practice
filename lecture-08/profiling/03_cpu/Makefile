.PHONY: run clean

run:
	go run .
	@echo "--- cpu profile top-10 (cumulative) ---"
	@go tool pprof -top -cum -nodecount=10 cpu.pprof | tee cpu-top.txt

clean:
	rm -f cpu.pprof cpu-top.txt
	go clean

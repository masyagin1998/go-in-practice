.PHONY: run clean

run:
	go run .
	@echo "--- trace summary (sched) ---"
	@go tool trace -pprof=sched trace.out > trace-sched.pprof
	@go tool pprof -top -nodecount=10 trace-sched.pprof | tee trace-summary.txt
	@rm -f trace-sched.pprof

clean:
	rm -f trace.out trace-summary.txt trace-sched.pprof
	go clean

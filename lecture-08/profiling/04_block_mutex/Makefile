.PHONY: run clean

run:
	go run .
	@echo "--- block profile ---"
	@go tool pprof -top -nodecount=5 block.pprof | tee block.txt
	@echo "--- mutex profile ---"
	@go tool pprof -top -nodecount=5 mutex.pprof | tee mutex.txt

clean:
	rm -f block.pprof mutex.pprof block.txt mutex.txt
	go clean

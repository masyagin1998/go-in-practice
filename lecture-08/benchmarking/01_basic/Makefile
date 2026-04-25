.PHONY: run test bench clean

run: bench

test:
	go test -v ./...

bench:
	go test -run='^$$' -bench=. -benchmem -benchtime=1s | tee bench.txt

clean:
	rm -f bench.txt
	go clean

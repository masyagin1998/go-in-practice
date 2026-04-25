.PHONY: run test bench clean

run: bench

test:
	go test -v ./...

bench:
	go test -run='^$$' -bench=BenchmarkFormatFmt -benchmem -count=10 \
		| sed 's/BenchmarkFormatFmt/BenchmarkFormat/' > old.txt
	go test -run='^$$' -bench=BenchmarkFormatStrconv -benchmem -count=10 \
		| sed 's/BenchmarkFormatStrconv/BenchmarkFormat/' > new.txt
	benchstat old.txt new.txt | tee delta.txt

clean:
	rm -f old.txt new.txt delta.txt
	go clean

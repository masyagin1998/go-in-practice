.PHONY: run clean

run:
	go test -v . | tee goleak-output.txt ; true

clean:
	rm -f goleak-output.txt
	go clean -testcache

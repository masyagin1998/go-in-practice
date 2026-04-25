.PHONY: run clean

# TestClean проходит, TestLeaky ожидаемо падает — это и есть демо.
# go test -v покажет стектрейс утекшей goroutine в выводе TestLeaky.
run:
	go test -v . | tee goleak-output.txt ; true

clean:
	rm -f goleak-output.txt
	go clean -testcache

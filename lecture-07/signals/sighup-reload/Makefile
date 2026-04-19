.PHONY: run clean

# В ручном запуске жмём Ctrl+C; под `make run` — SIGTERM после 3 сек.
run: gobin
	timeout --signal=TERM 3 ./gobin || true

gobin: main.go go.mod
	go build -o gobin .

clean:
	rm -f gobin /tmp/sighup_demo.conf
	go clean

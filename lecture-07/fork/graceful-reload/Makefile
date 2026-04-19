.PHONY: run clean

# Ручной сценарий:
#   make run             # сервер на 127.0.0.1:8080 (отработает ~5 сек под make)
#   curl localhost:8080/ # в другом терминале
#   kill -HUP <PID>      # релоад
# make run в CI: Go + таймаут (SIGTERM → srv.Shutdown).
run: gobin
	timeout --signal=TERM 5 ./gobin || true

gobin: main.go go.mod
	go build -o gobin .

clean:
	rm -f gobin
	go clean

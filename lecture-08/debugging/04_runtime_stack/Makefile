.PHONY: run clean

# Фоновая программа + автоматический SIGUSR1 через 1 сек.
# Останавливаем через 2 сек SIGTERM'ом.
run:
	@go build -o gobin .
	@./gobin & \
	 PID=$$!; \
	 sleep 1; \
	 kill -USR1 $$PID; \
	 sleep 1; \
	 kill -TERM $$PID; \
	 wait $$PID || true

clean:
	rm -f gobin stacks.txt
	go clean

.PHONY: run clean

# run — просто стартует сервер и нагрузку на 5 секунд.
# Для before/after-сравнения переключайте USE_REGEXP в main.go руками
# (см. README.md). Готовые артефакты before-top.txt / after-top.txt /
# delta.txt уже лежат в репо.
run:
	@go build -o server .
	@./server & \
	 PID=$$!; \
	 sleep 1; \
	 echo ">>> нагрузка 5 сек"; \
	 (cd load && DURATION=5s go run . ) & \
	 LOAD=$$!; \
	 sleep 1; \
	 echo ">>> CPU-профиль на 3 секунды"; \
	 go tool pprof -top -cum -nodecount=10 \
	   'http://localhost:6060/debug/pprof/profile?seconds=3' 2>/dev/null \
	   | tee /tmp/realprog-last.txt; \
	 wait $$LOAD 2>/dev/null || true; \
	 kill -TERM $$PID 2>/dev/null || true; \
	 wait $$PID 2>/dev/null || true

clean:
	rm -f server
	go clean
	-cd load && go clean

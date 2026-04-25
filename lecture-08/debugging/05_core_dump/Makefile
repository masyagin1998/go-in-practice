.PHONY: run clean

# run — только проверяем, что программа собирается и стартует.
# Полный сценарий с gcore/dlv описан в README.md и core-session.txt;
# его стоит прогонять вручную (нужен sudo или ptrace_scope=0).
run:
	@go build -o gobin .
	@./gobin & \
	 PID=$$!; \
	 sleep 1; \
	 kill -TERM $$PID; \
	 wait $$PID || true

clean:
	rm -f gobin core.*
	go clean

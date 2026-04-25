.PHONY: run clean

# Без -race: видим неверный результат (недосчёт).
# С -race:    детектор печатает трейс в stderr, exit-code != 0.
run:
	@echo "--- без -race ---"
	@go run .
	@echo "--- с -race ---"
	@go run -race . 2>&1 | tee race.txt ; true

clean:
	rm -f race.txt
	go clean

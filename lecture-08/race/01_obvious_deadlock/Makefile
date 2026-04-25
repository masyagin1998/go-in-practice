.PHONY: run clean

# Без -race: рантайм ловит сам, fatal error за миллисекунды.
# С -race:  TSan ломает checkdead() — программа висит, режем по таймауту.
run:
	@echo "--- без -race (рантайм ловит сам) ---"
	@go run . ; true
	@echo "--- с -race (детектор checkdead сломан, режем по таймауту) ---"
	@timeout 3 go run -race . ; true

clean:
	go clean

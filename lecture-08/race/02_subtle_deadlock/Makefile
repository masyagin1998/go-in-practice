.PHONY: run clean

# Программа выходит чисто. Ни рантайм, ни -race не сообщают, что две
# goroutine висят на AB-BA — это и есть весь смысл примера.
run:
	@echo "--- без -race ---"
	@go run .
	@echo "--- с -race ---"
	@go run -race .

clean:
	go clean

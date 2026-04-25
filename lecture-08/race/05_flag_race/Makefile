.PHONY: run clean

run:
	@echo "--- без -race (на современном Go обычно не виснет) ---"
	@go run .
	@echo "--- с -race ---"
	@go run -race . 2>&1 | tee race.txt ; true

clean:
	rm -f race.txt
	go clean

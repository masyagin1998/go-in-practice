.PHONY: run clean

# Падает с fatal error и без -race, и под -race — детектор map встроенный.
run:
	@echo "--- без -race ---"
	@go run . 2>&1 | head -20 ; true
	@echo "--- с -race ---"
	@go run -race . 2>&1 | head -30 ; true

clean:
	go clean

.PHONY: all clean run-go run-c

all: go/stress_test_go c/stress_test_c

go/stress_test_go: go/main.go
	cd go && go build -o stress_test_go .

c/stress_test_c: c/stress_test.c
	gcc -O2 -Wall -pthread -o c/stress_test_c c/stress_test.c

run-go: go/stress_test_go
	./go/stress_test_go

run-c: c/stress_test_c
	./c/stress_test_c

clean:
	rm -f go/stress_test_go c/stress_test_c

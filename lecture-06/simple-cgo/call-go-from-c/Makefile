# Сборка Go-кода как C-архива, затем компиляция C-точки входа.
#
# 1. go build -buildmode=c-archive → libgogreet.a + libgogreet.h
# 2. gcc cmd/main.c + libgogreet.a → исполняемый файл main

.PHONY: run clean

main: libgogreet.a cmd/main.c
	gcc -I. cmd/main.c libgogreet.a -lpthread -o main

libgogreet.a: lib.go
	go build -buildmode=c-archive -o libgogreet.a .

run: main
	./main

clean:
	rm -f main libgogreet.a libgogreet.h

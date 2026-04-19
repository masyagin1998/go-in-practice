package main

// exec-spawn — `os/exec.Command` = fork+exec под капотом.
// Реальные кейсы: git, ffmpeg, cc1, migrate, kubectl, build-тулчейны.

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

func main() {
	// 1. Простая команда, захват вывода целиком.
	out, err := exec.Command("uname", "-a").Output()
	if err != nil {
		log.Fatalf("uname: %v", err)
	}
	fmt.Printf("uname -a: %s", out)

	// 2. Стрим stdout построчно — для долгих команд (tail -f, сборка).
	cmd := exec.Command("sh", "-c", "for i in 1 2 3; do echo строка $i; done")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("start: %v", err)
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Printf("[child] %s\n", scanner.Text())
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("wait: %v", err)
	}

	// 3. Двусторонний обмен — пишем в stdin, читаем stdout.
	cmd = exec.Command("tr", "a-z", "A-Z")
	stdin, _ := cmd.StdinPipe()
	stdout, _ = cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		log.Fatalf("start tr: %v", err)
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "hello from go\n")
	}()
	data, _ := io.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatalf("wait tr: %v", err)
	}
	fmt.Printf("tr вернул: %s\n", strings.TrimRight(string(data), "\n"))
}

package main

// exec-spawn — `os/exec.Command` = fork+exec под капотом.
// Реальные кейсы: git, ffmpeg, cc1, migrate, kubectl, build-тулчейны.

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func main() {
	// 1. Простая команда, захват вывода целиком.
	out, _ := exec.Command("uname", "-a").Output()
	fmt.Printf("uname -a: %s", out)

	// 2. Стрим stdout построчно — для долгих команд (tail -f, сборка).
	cmd := exec.Command("sh", "-c", "for i in 1 2 3; do echo строка $i; done")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Printf("[child] %s\n", scanner.Text())
	}
	cmd.Wait()

	// 3. Двусторонний обмен — пишем в stdin, читаем stdout.
	cmd = exec.Command("tr", "a-z", "A-Z")
	stdin, _ := cmd.StdinPipe()
	stdout, _ = cmd.StdoutPipe()
	cmd.Start()
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "hello from go\n")
	}()
	data, _ := io.ReadAll(stdout)
	cmd.Wait()
	fmt.Printf("tr вернул: %s\n", strings.TrimRight(string(data), "\n"))
}

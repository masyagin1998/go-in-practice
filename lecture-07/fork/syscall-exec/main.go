package main

// syscall-exec — замена текущего процесса новым бинарём через execve(2).
// PID сохраняется, память полностью заменяется. Реальные кейсы:
//   * upgrade-in-place у supervisor'ов,
//   * privilege drop после setup'а,
//   * re-exec из init-системы.
//
// Здесь программа 10 раз перезапускает сама себя, инкрементируя поколение.

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
)

const maxGen = 10

func main() {
	gen, _ := strconv.Atoi(os.Getenv("GENERATION")) // пусто → 0
	fmt.Printf("PID=%d generation=%d argv=%v\n", os.Getpid(), gen, os.Args)

	if gen >= maxGen {
		fmt.Printf("дошли до generation=%d, выходим\n", gen)
		return
	}

	self, err := os.Executable()
	if err != nil {
		log.Fatalf("executable: %v", err)
	}

	next := strconv.Itoa(gen + 1)
	os.Setenv("GENERATION", next)

	fmt.Printf("exec'имся в поколение %s (тот же PID)...\n", next)
	if err := syscall.Exec(self, []string{self, "--re-exec"}, os.Environ()); err != nil {
		log.Fatalf("exec: %v", err)
	}
}

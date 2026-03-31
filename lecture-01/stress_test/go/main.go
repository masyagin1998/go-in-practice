// Stress test: запускаем 1 000 000 горутин, каждая спит 10 секунд.
// Демонстрирует, что горутины дешёвые (~2-8 KB стека).
//
// Следить: watch -n1 'ps -o pid,vsz,rss,nlwp -p $(pgrep -f stress_test_go)'
package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	numGoroutines = 1_000_000
	sleepDuration = 10 * time.Second
)

func main() {
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	fmt.Printf("Запускаем %d горутин (sleep %s)...\n", numGoroutines, sleepDuration)
	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(sleepDuration)
		}()
	}

	launched := time.Since(start)
	fmt.Printf("Все горутины запущены за %s\n", launched.Round(time.Millisecond))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Горутин: %d, OS-тредов: -, Alloc: %d MB, Sys: %d MB\n",
		runtime.NumGoroutine(), m.Alloc/1024/1024, m.Sys/1024/1024)

	// Читаем RSS из /proc/self/status
	if data, err := os.ReadFile("/proc/self/status"); err == nil {
		for _, line := range splitLines(data) {
			s := string(line)
			if len(s) > 7 && s[:7] == "Threads" {
				fmt.Println(s)
			} else if len(s) > 6 && (s[:5] == "VmRSS" || s[:6] == "VmSize") {
				var name string
				var kb int64
				if n, _ := fmt.Sscanf(s, "%s %d", &name, &kb); n == 2 {
					fmt.Printf("%s\t%d MB\n", name, kb/1024)
				}
			}
		}
	}

	fmt.Println("Ждём завершения...")
	wg.Wait()
	fmt.Printf("Готово за %s\n", time.Since(start).Round(time.Millisecond))
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	for len(data) > 0 {
		i := 0
		for i < len(data) && data[i] != '\n' {
			i++
		}
		lines = append(lines, data[:i])
		if i < len(data) {
			i++
		}
		data = data[i:]
	}
	return lines
}

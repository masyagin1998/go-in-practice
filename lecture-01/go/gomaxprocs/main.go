// Демонстрация влияния GOMAXPROCS на параллельное выполнение.
//
// Запускаем CPU-bound задачу на нескольких горутинах и замеряем
// время при разных значениях GOMAXPROCS. С увеличением GOMAXPROCS
// (до числа ядер) время выполнения падает — горутины реально
// исполняются параллельно на разных OS-тредах.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	workers    = 4
	iterations = 300_000_000
)

// cpuWork — чистая CPU-нагрузка без аллокаций.
func cpuWork(wg *sync.WaitGroup) {
	defer wg.Done()
	x := 0
	for i := 0; i < iterations; i++ {
		x += i * i
	}
	_ = x
}

func bench(procs int) time.Duration {
	runtime.GOMAXPROCS(procs)

	var wg sync.WaitGroup
	wg.Add(workers)

	start := time.Now()
	for i := 0; i < workers; i++ {
		go cpuWork(&wg)
	}
	wg.Wait()

	return time.Since(start)
}

func main() {
	numCPU := runtime.NumCPU()
	fmt.Printf("CPU ядер: %d\n", numCPU)
	fmt.Printf("Горутин:  %d, итераций на горутину: %d\n\n", workers, iterations)

	procsToTest := []int{1, 2}
	if numCPU >= 4 {
		procsToTest = append(procsToTest, 4)
	}
	if numCPU >= 8 {
		procsToTest = append(procsToTest, 8)
	}
	if numCPU > 8 {
		procsToTest = append(procsToTest, numCPU)
	}

	baseline := time.Duration(0)
	for _, p := range procsToTest {
		d := bench(p)
		if baseline == 0 {
			baseline = d
		}
		speedup := float64(baseline) / float64(d)
		fmt.Printf("GOMAXPROCS=%-3d  время: %-12s  ускорение: %.2fx\n", p, d.Round(time.Millisecond), speedup)
	}
}

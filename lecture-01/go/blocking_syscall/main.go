// Демонстрация: при блокирующем сисколле горутина «паркуется»,
// а её OS-тред отсоединяется от P — и P передаётся другой горутине.
//
// Запускаем несколько горутин, каждая делает блокирующий ввод-вывод
// (запись во временный файл с fsync). Во время блокировки рантайм
// создаёт дополнительные OS-треды, чтобы остальные горутины
// продолжали работу. Это видно по количеству тредов через
// runtime/pprof.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"
)

func threadCount() int {
	p := pprof.Lookup("threadcreate")
	if p == nil {
		return -1
	}
	return p.Count()
}

// blockingIO выполняет блокирующую запись + fsync.
func blockingIO(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.CreateTemp("", "block-*")
	if err != nil {
		fmt.Printf("goroutine %d: ошибка создания файла: %v\n", id, err)
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()

	for i := 0; i < 5; i++ {
		data := make([]byte, 4*1024*1024) // 4 MiB
		_, _ = f.Write(data)
		syscall.Fsync(int(f.Fd())) // блокирующий сисколл
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("  goroutine %d: write %d done (threads: %d)\n", id, i+1, threadCount())
	}
}

func main() {
	runtime.GOMAXPROCS(2)
	fmt.Printf("GOMAXPROCS = %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Начальное кол-во тредов: %d\n\n", threadCount())

	const N = 6
	var wg sync.WaitGroup
	wg.Add(N)

	for i := 1; i <= N; i++ {
		go blockingIO(i, &wg)
	}

	// Параллельно наблюдаем за количеством тредов.
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("  [observer] goroutines: %d, OS threads: %d\n",
					runtime.NumGoroutine(), threadCount())
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	close(done)

	fmt.Printf("\nИтоговое кол-во тредов: %d\n", threadCount())
	fmt.Println("Все горутины завершились, несмотря на блокирующие сисколлы.")
}

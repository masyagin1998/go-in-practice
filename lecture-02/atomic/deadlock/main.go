// go run main.go
//
// Классический deadlock: две горутины захватывают два мьютекса в разном порядке.
// Go runtime обнаружит deadlock и завершит программу с паникой:
// "fatal error: all goroutines are asleep - deadlock!"

package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	mutex1 sync.Mutex
	mutex2 sync.Mutex
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		mutex1.Lock()
		fmt.Println("Горутина 1 захватила mutex1")
		time.Sleep(1 * time.Second)
		mutex2.Lock()
		fmt.Println("Горутина 1 захватила mutex2")
		mutex2.Unlock()
		mutex1.Unlock()
	}()

	go func() {
		defer wg.Done()
		mutex2.Lock()
		fmt.Println("Горутина 2 захватила mutex2")
		time.Sleep(1 * time.Second)
		mutex1.Lock()
		fmt.Println("Горутина 2 захватила mutex1")
		mutex1.Unlock()
		mutex2.Unlock()
	}()

	wg.Wait()
}

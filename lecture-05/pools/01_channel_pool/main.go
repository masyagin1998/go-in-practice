// Пример 1: «Колхозный» пул на buffered channel.
//
// Идея: buffered канал фиксированного размера хранит заранее
// выделенные слайсы. Берём из канала — используем — возвращаем.
// Канал ограничивает максимальное число живых буферов.
//
// Плюсы:
//   - Просто, понятно, потокобезопасно «из коробки»
//   - Жёсткий лимит на число буферов (размер канала)
//   - Если канал пуст — блокируемся (backpressure) или делаем select
//
// Минусы:
//   - Нет автоматической подстройки под нагрузку
//   - Буферы живут вечно (не освобождаются при простое)
//   - Нужно самим следить за размерами возвращаемых слайсов
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

const (
	poolSize = 4   // максимум буферов в пуле
	bufSize  = 1024 // начальный размер каждого буфера
)

// ChannelPool — пул []byte на основе buffered канала.
type ChannelPool struct {
	ch chan []byte
}

// NewChannelPool — создаёт пул и заполняет его буферами.
func NewChannelPool(size, bufCap int) *ChannelPool {
	p := &ChannelPool{ch: make(chan []byte, size)}
	for range size {
		p.ch <- make([]byte, 0, bufCap)
	}
	return p
}

// Get — берём буфер из пула. Если пул пуст — блокируемся.
func (p *ChannelPool) Get() []byte {
	return <-p.ch
}

// TryGet — неблокирующий вариант: nil если пул пуст.
func (p *ChannelPool) TryGet() []byte {
	select {
	case buf := <-p.ch:
		return buf
	default:
		return nil
	}
}

// Put — возвращаем буфер в пул (сбрасываем длину, сохраняем capacity).
func (p *ChannelPool) Put(buf []byte) {
	p.ch <- buf[:0]
}

func main() {
	pool := NewChannelPool(poolSize, bufSize)

	fmt.Printf("=== Channel Pool (размер=%d, буфер=%d байт) ===\n\n", poolSize, bufSize)

	// Демонстрация: 8 горутин конкурируют за 4 буфера.
	var wg sync.WaitGroup
	for i := range 8 {
		id := i
		wg.Go(func() {
			// Берём буфер (блокируемся, если все заняты).
			buf := pool.Get()
			fmt.Printf("  [горутина %d] получила буфер: len=%d, cap=%d\n",
				id, len(buf), cap(buf))

			// «Работаем» — заполняем случайными данными.
			n := rand.IntN(512)
			buf = buf[:n]
			for j := range buf {
				buf[j] = byte(j)
			}
			fmt.Printf("  [горутина %d] использовала %d байт\n", id, n)

			// Возвращаем.
			pool.Put(buf)
			fmt.Printf("  [горутина %d] вернула буфер\n", id)
		})
	}
	wg.Wait()

	// Проверяем: все буферы вернулись.
	fmt.Printf("\n  Буферов в пуле: %d (ожидается %d)\n", len(pool.ch), poolSize)

	// Демонстрация TryGet.
	fmt.Println("\n=== Неблокирующий TryGet ===")
	for range poolSize {
		pool.Get() // забираем все
	}
	buf := pool.TryGet()
	fmt.Printf("  TryGet из пустого пула: %v (nil = пул пуст)\n", buf)
}

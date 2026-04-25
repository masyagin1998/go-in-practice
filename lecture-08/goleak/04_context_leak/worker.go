package worker

import "context"

// WorkerLeaky — читает из канала без оглядки на контекст. Если producer
// ушёл, не закрыв in, goroutine висит вечно.
func WorkerLeaky(_ context.Context, in <-chan int, out chan<- int) {
	go func() {
		for v := range in {
			out <- v * 2
		}
	}()
}

// WorkerFixed — то же самое, но с select на ctx.Done. И на чтении, и
// на записи: иначе можно повиснуть на out <- ... если получателя нет.
func WorkerFixed(ctx context.Context, in <-chan int, out chan<- int) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- v * 2:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
}

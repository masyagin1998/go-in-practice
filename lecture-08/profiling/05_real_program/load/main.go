package main

// Простой нагрузочный генератор для соседнего сервера.
// Параметры через env: URL, WORKERS, DURATION.

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := envOr("URL", "http://localhost:6060/search?q=fox")
	workers := envOrInt("WORKERS", 20)
	dur := envOrDur("DURATION", 10*time.Second)

	log.Printf("load: %s, %d workers, %s", url, workers, dur)

	var done atomic.Bool
	var reqs atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := &http.Client{Timeout: 5 * time.Second}
			for !done.Load() {
				resp, err := c.Get(url)
				if err != nil {
					log.Println("get:", err)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				reqs.Add(1)
			}
		}()
	}

	time.Sleep(dur)
	done.Store(true)
	wg.Wait()

	total := reqs.Load()
	rps := float64(total) / dur.Seconds()
	fmt.Printf("total=%d rps=%.1f\n", total, rps)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func envOrDur(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

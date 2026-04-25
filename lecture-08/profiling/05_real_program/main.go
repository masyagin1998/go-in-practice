package main

// HTTP-сервер с намеренным узким местом.
//
// Эндпоинт /search принимает ?q=... и ищет совпадения в in-memory
// датасете. Для "удобства" case-insensitive поиск реализован через
// regexp: `regexp.MustCompile("(?i)" + QuoteMeta(q))`, затем
// MatchString на каждой строке. Под нагрузкой pprof покажет, что
// 99% CPU горит в regexp.backtrack — это дикий оверкилл для простой
// подстроки.
//
// Сценарий лекции:
//   1. go run .  (в одном терминале)
//   2. go run ./load  (в другом, нагружает /search)
//   3. go tool pprof -top -cum http://localhost:6060/debug/pprof/profile?seconds=5
//   4. Виден regexp в топе по CUM.
//   5. Фикс: ICASE=false — перейти на strings.ToLower + strings.Contains.
//   6. Снова pprof — узкое место ушло.

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // подключает /debug/pprof/*
	"os"
	"regexp"
	"runtime"
	"strings"
)

// dataset — "база" текстов, по которой идёт поиск.
var dataset []string

func init() {
	// Несколько тысяч строк разной длины.
	base := []string{
		"the quick brown fox jumps over the lazy dog",
		"go is a statically typed compiled language",
		"profiling real programs with pprof",
		"goroutine leaks are hard to find without tools",
		"benchmarks lie if you do not use benchstat",
	}
	for i := 0; i < 4000; i++ {
		dataset = append(dataset, base[i%len(base)]+fmt.Sprintf(" #%d", i))
	}

	// Включаем block/mutex-профили заодно — пригодится в упражнении.
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
}

// USE_REGEXP = true — медленная реализация через regexp.
// Поменяйте на false и перезапустите, чтобы увидеть "фикс":
// strings.ToLower + strings.Contains. Для простой подстрочной проверки
// разница в CPU — на порядок.
const USE_REGEXP = true

func search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing q", http.StatusBadRequest)
		return
	}

	hits := 0
	if USE_REGEXP {
		// Антипаттерн: regexp для простого case-insensitive substring-поиска.
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(q))
		for _, line := range dataset {
			if re.MatchString(line) {
				hits++
			}
		}
	} else {
		// Фикс: strings-функции, никакого регекспа.
		lq := strings.ToLower(q)
		for _, line := range dataset {
			if strings.Contains(strings.ToLower(line), lq) {
				hits++
			}
		}
	}

	fmt.Fprintf(w, "q=%q hits=%d\n", q, hits)
}

func fast(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "ok\n")
}

func main() {
	http.HandleFunc("/search", search)
	http.HandleFunc("/fast", fast)
	// /debug/pprof/* зарегистрированы через import _ "net/http/pprof"

	addr := ":6060"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Printf("listening on %s; pprof: http://localhost%s/debug/pprof/",
		addr, addr)
	log.Printf("dataset size: %d lines", len(dataset))
	log.Printf("USE_REGEXP=%v", USE_REGEXP)

	// Для удобства: показать, как curl'ить.
	log.Println(strings.TrimSpace(`
подсказки:
  curl 'http://localhost:6060/search?q=fox'
  curl 'http://localhost:6060/debug/pprof/goroutine?debug=2'
  go tool pprof 'http://localhost:6060/debug/pprof/profile?seconds=10'
`))

	log.Fatal(http.ListenAndServe(addr, nil))
}

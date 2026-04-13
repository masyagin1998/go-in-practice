// Пример 5: sync.Pool и вложенные структуры — грязные данные.
//
// Когда мы кладём в sync.Pool структуру с вложенными полями,
// при повторном Get() мы получаем объект с ДАННЫМИ ОТ ПРОШЛОГО ПОЛЬЗОВАТЕЛЯ.
//
// Три подхода:
//   1. Не обнулять — получим грязные данные (баг)
//   2. Обнулять вручную при Get() — корректно, но многословно
//   3. Pool оптимизирует только аллокацию «корневой» структуры —
//      вложенные слайсы/мапы всё равно аллоцируются заново
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"strings"
	"sync"
)

// Request — типичная структура с вложенными данными.
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
	Tags    []string
}

func main() {
	// =============================================================
	// Кейс 1: Грязные данные — забыли обнулить
	// =============================================================
	fmt.Println("=== Кейс 1: Грязные данные (без Reset) ===")
	fmt.Println()

	dirtyPool := sync.Pool{
		New: func() any {
			return &Request{
				Headers: make(map[string]string),
			}
		},
	}

	// «Первый запрос» — заполняем данными.
	req1 := dirtyPool.Get().(*Request)
	req1.Method = "POST"
	req1.Path = "/api/admin/delete-user"
	req1.Headers["Authorization"] = "Bearer secret-token-12345"
	req1.Headers["X-User-ID"] = "admin"
	req1.Body = []byte(`{"user": "victim"}`)
	req1.Tags = append(req1.Tags, "admin", "dangerous")

	fmt.Printf("  Запрос 1: %s %s\n", req1.Method, req1.Path)
	fmt.Printf("  Headers:  %v\n", req1.Headers)
	fmt.Printf("  Body:     %s\n", req1.Body)
	fmt.Printf("  Tags:     %v\n", req1.Tags)

	// Возвращаем в пул БЕЗ очистки.
	dirtyPool.Put(req1)

	// «Второй запрос» — получаем тот же объект.
	req2 := dirtyPool.Get().(*Request)
	req2.Method = "GET"
	req2.Path = "/api/public/health"
	// Забыли очистить Headers, Body, Tags!

	fmt.Printf("\n  Запрос 2: %s %s\n", req2.Method, req2.Path)
	fmt.Printf("  Headers:  %v  ← УТЕЧКА ДАННЫХ!\n", req2.Headers)
	fmt.Printf("  Body:     %s  ← ГРЯЗНЫЕ ДАННЫЕ!\n", req2.Body)
	fmt.Printf("  Tags:     %v  ← ЧУЖИЕ ТЕГИ!\n", req2.Tags)

	// =============================================================
	// Кейс 2: Правильный подход — Reset при Get
	// =============================================================
	fmt.Println()
	fmt.Println("=== Кейс 2: Корректный Reset при Get ===")
	fmt.Println()

	cleanPool := sync.Pool{
		New: func() any {
			return &Request{}
		},
	}

	// resetRequest — обнуляет все поля, переиспользуя capacity где можно.
	resetRequest := func(r *Request) {
		r.Method = ""
		r.Path = ""
		// Мапу дешевле пересоздать, чем чистить (clear() с Go 1.21).
		// Но если ключей мало, можно clear(r.Headers).
		r.Headers = nil
		r.Body = r.Body[:0] // сохраняем capacity
		r.Tags = r.Tags[:0] // сохраняем capacity
	}

	// Первый запрос.
	r1 := cleanPool.Get().(*Request)
	resetRequest(r1)
	r1.Method = "POST"
	r1.Path = "/api/secret"
	r1.Headers = map[string]string{"Authorization": "Bearer xyz"}
	r1.Body = append(r1.Body, []byte("secret body")...)
	r1.Tags = append(r1.Tags, "confidential")
	cleanPool.Put(r1)

	// Второй запрос.
	r2 := cleanPool.Get().(*Request)
	resetRequest(r2)
	r2.Method = "GET"
	r2.Path = "/api/public"

	fmt.Printf("  Запрос: %s %s\n", r2.Method, r2.Path)
	fmt.Printf("  Headers: %v  ← чисто\n", r2.Headers)
	fmt.Printf("  Body:    \"%s\"  ← чисто (cap=%d — переиспользован)\n",
		r2.Body, cap(r2.Body))
	fmt.Printf("  Tags:    %v  ← чисто (cap=%d — переиспользован)\n",
		r2.Tags, cap(r2.Tags))
	cleanPool.Put(r2)

	// =============================================================
	// Кейс 3: Pool оптимизирует только «корень» структуры
	// =============================================================
	fmt.Println()
	fmt.Println("=== Кейс 3: Pool оптимизирует только корневую аллокацию ===")
	fmt.Println()

	type Response struct {
		Status  int
		Headers map[string]string   // аллокация хипа
		Body    []byte              // аллокация хипа
		Errors  []string            // аллокация хипа
	}

	respPool := sync.Pool{
		New: func() any {
			return &Response{}
		},
	}

	// Без пула: каждый раз аллоцируем Response + 3 вложенных объекта = 4 аллокации.
	// С пулом: Response переиспользуется = 1 аллокация сэкономлена,
	// но вложенные map/slice всё равно новые (или нужен reset).

	fmt.Println("  Что экономит пул:")
	fmt.Println("  ┌──────────────────────────────────────────┐")
	fmt.Println("  │ Response (корень)    — из пула ✓         │")
	fmt.Println("  │ ├── Headers (map)    — заново / clear()  │")
	fmt.Println("  │ ├── Body ([]byte)    — reset len→0       │")
	fmt.Println("  │ └── Errors ([]string)— reset len→0       │")
	fmt.Println("  └──────────────────────────────────────────┘")

	// Демонстрация: что реально экономится.
	resp := respPool.Get().(*Response)
	resp.Status = 200
	resp.Headers = map[string]string{"Content-Type": "application/json"}
	resp.Body = []byte(`{"ok":true}`)
	resp.Errors = nil

	// При возврате: Body и Errors сохраняют capacity,
	// но Headers пересоздаётся (или clear).
	resp.Status = 0
	resp.Headers = nil
	resp.Body = resp.Body[:0]
	resp.Errors = resp.Errors[:0]
	respPool.Put(resp)

	resp2 := respPool.Get().(*Response)
	fmt.Printf("\n  Повторный Get():\n")
	fmt.Printf("    Status:  %d (нужно задать)\n", resp2.Status)
	fmt.Printf("    Headers: %v (nil — нужно создать)\n", resp2.Headers)
	fmt.Printf("    Body:    len=%d, cap=%d (capacity сохранён!)\n",
		len(resp2.Body), cap(resp2.Body))
	fmt.Printf("    Errors:  len=%d, cap=%d (capacity сохранён!)\n",
		len(resp2.Errors), cap(resp2.Errors))
	respPool.Put(resp2)

	fmt.Println()
	fmt.Println("  → Вывод: sync.Pool экономит аллокацию корневой структуры.")
	fmt.Println("    Вложенные данные надо либо обнулять (reset), либо")
	fmt.Println("    пересоздавать. Без reset — утечка данных между запросами.")
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("  ИТОГО: выбирайте стратегию по ситуации:")
	fmt.Println("    • Маленькая структура без вложений → Pool идеален")
	fmt.Println("    • Структура со слайсами → Reset(len→0), capacity живёт")
	fmt.Println("    • Структура с map → clear() или пересоздание")
	fmt.Println("    • Много вложенных ссылок → Pool экономит мало")
}

// go run main.go
//
// sync.Once — гарантирует, что функция выполнится ровно один раз,
// даже если Do() вызывается из множества горутин одновременно.

package main

import (
	"fmt"
	"sync"
)

// --- 1. Базовый вариант: однократная инициализация ---

var (
	instance *DB
	once     sync.Once
)

type DB struct{ dsn string }

func GetDB() *DB {
	once.Do(func() {
		fmt.Println("Подключаемся к БД (выполнится один раз)...")
		instance = &DB{dsn: "postgres://localhost/mydb"}
	})
	return instance
}

// --- 2. sync.OnceValue — возвращает значение (Go 1.21+) ---

var getConfig = sync.OnceValue(func() map[string]string {
	fmt.Println("Читаем конфиг (выполнится один раз)...")
	return map[string]string{
		"host": "localhost",
		"port": "8080",
	}
})

// --- 3. sync.OnceValues — возвращает значение + ошибку (Go 1.21+) ---

var loadCert = sync.OnceValues(func() ([]byte, error) {
	fmt.Println("Загружаем сертификат (выполнится один раз)...")
	// В реальности: os.ReadFile("cert.pem")
	return []byte("CERTIFICATE DATA"), nil
})

func main() {
	// 1. Базовый Once — 10 горутин, инициализация один раз.
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db := GetDB()
			fmt.Printf("  got db: %s\n", db.dsn)
		}()
	}
	wg.Wait()

	fmt.Println()

	// 2. OnceValue — конфиг вычисляется один раз, возвращается многим.
	for range 3 {
		cfg := getConfig()
		fmt.Printf("  config: %v\n", cfg)
	}

	fmt.Println()

	// 3. OnceValues — с обработкой ошибки.
	for range 3 {
		cert, err := loadCert()
		fmt.Printf("  cert: %s, err: %v\n", cert, err)
	}
}

// Пример 4: Настройка GOGC и GOMEMLIMIT.
//
// GOGC контролирует частоту сборки мусора.
//   GOGC=100 (по умолчанию) — GC запускается, когда куча вырастает на 100%
//                              относительно «живых» данных после прошлой сборки.
//   GOGC=50  — агрессивнее: GC чаще, меньше памяти, больше CPU.
//   GOGC=200 — ленивее: GC реже, больше памяти, меньше CPU.
//   GOGC=off — GC отключён (опасно: память будет расти бесконечно).
//
// GOMEMLIMIT (с Go 1.19) — мягкий (soft) лимит памяти.
//   Рантайм старается не превышать лимит, запуская GC агрессивнее.
//   В отличие от GOGC, который задаёт соотношение, GOMEMLIMIT — абсолютный порог.
//
//   Это именно SOFT-лимит: рантайм гарантирует, что GC не будет тратить
//   больше ~50% CPU. Если для соблюдения лимита нужно больше — рантайм
//   позволит куче вырасти за лимит, но не уронит приложение в GC thrashing.
//   Последствия: при слишком низком лимите процесс может превысить его
//   и быть убит OOM killer'ом, хотя лимит формально задан.
//
// Запуск с разными значениями:
//   GOGC=50      go run main.go
//   GOGC=200     go run main.go
//   GOGC=off     go run main.go
//   GOMEMLIMIT=64MiB go run main.go

package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

// allocateGarbage выделяет N объектов, часть из которых становится мусором.
func allocateGarbage(live []*[1024]byte) []*[1024]byte {
	for range 10_000 {
		obj := new([1024]byte)
		obj[0] = 1

		// ~10% объектов сохраняем как «живые», остальные — мусор.
		if len(live) < 1000 {
			live = append(live, obj)
		}
	}
	return live
}

func main() {
	// Читаем текущие настройки GC.
	gogc := os.Getenv("GOGC")
	if gogc == "" {
		gogc = "100 (по умолчанию)"
	}
	memlimit := os.Getenv("GOMEMLIMIT")
	if memlimit == "" {
		memlimit = "не задан"
	}

	fmt.Printf("GOGC=%s, GOMEMLIMIT=%s\n", gogc, memlimit)
	fmt.Printf("GOMAXPROCS=%d\n\n", runtime.GOMAXPROCS(0))

	var live []*[1024]byte
	var stats runtime.MemStats

	start := time.Now()

	// 50 раундов аллокаций — наблюдаем поведение GC.
	for round := range 50 {
		live = allocateGarbage(live)

		if round%10 == 9 {
			runtime.ReadMemStats(&stats)
			fmt.Printf("Раунд %2d: куча=%4d МБ, живых=%4d КБ, циклов GC=%d, пауза последняя=%v\n",
				round+1,
				stats.HeapAlloc/1024/1024,
				stats.HeapInuse/1024,
				stats.NumGC,
				time.Duration(stats.PauseNs[(stats.NumGC+255)%256]))
		}
	}

	elapsed := time.Since(start)
	runtime.ReadMemStats(&stats)

	fmt.Printf("\n=== Итого ===\n")
	fmt.Printf("Время:            %v\n", elapsed)
	fmt.Printf("Аллоцировано:     %d МБ\n", stats.TotalAlloc/1024/1024)
	fmt.Printf("Циклов GC:        %d\n", stats.NumGC)
	fmt.Printf("Суммарная пауза:  %v\n", time.Duration(stats.PauseTotalNs))
	fmt.Printf("Куча сейчас:      %d МБ\n", stats.HeapAlloc/1024/1024)

	// Программный способ настройки (вместо переменных окружения).
	_ = debug.SetGCPercent      // Аналог GOGC.
	_ = debug.SetMemoryLimit    // Аналог GOMEMLIMIT.
}

// Пример 6: Наблюдение за работой GC в реальном времени.
//
// Переменная окружения GODEBUG=gctrace=1 заставляет рантайм печатать
// строку в stderr на каждый цикл GC. Формат:
//
//   gc # @#s #%: #+#+# ms clock, #+#/#/#+# ms cpu, #->#-># MB, # MB goal, # P
//
//   gc #        — номер цикла GC
//   @#s         — время с запуска программы
//   #%          — процент времени, потраченный на GC
//   #+#+# ms    — wall-clock время (sweep, mark, term)
//   #+#/#/#+#   — CPU время (assist, bg mark, idle mark)
//   #->#-># MB  — размер кучи: до GC → после GC → живые данные
//   # MB goal   — целевой размер кучи
//   # P         — число процессоров
//
// Запуск:
//   GODEBUG=gctrace=1 go run main.go
//
// Также полезно:
//   GODEBUG=gctrace=1,gcpacertrace=1 go run main.go

package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	fmt.Println("Запустите с GODEBUG=gctrace=1 для трассировки GC.")
	fmt.Println("Начинаю аллокации...")

	var live [][]byte
	var stats runtime.MemStats

	for round := range 20 {
		// Выделяем 50 000 объектов по 1 КБ — ~50 МБ за раунд.
		for range 50_000 {
			buf := make([]byte, 1024)
			buf[0] = 1

			// 1% сохраняем как «живые» данные, остальное — мусор.
			if len(live) < 500 {
				live = append(live, buf)
			}
		}

		runtime.ReadMemStats(&stats)
		fmt.Printf("Раунд %2d: куча=%3d МБ, GC=%d, пауза=%v\n",
			round+1,
			stats.HeapAlloc/1024/1024,
			stats.NumGC,
			time.Duration(stats.PauseNs[(stats.NumGC+255)%256]))
	}

	runtime.ReadMemStats(&stats)
	fmt.Printf("\nИтого: аллоцировано %d МБ, %d циклов GC, суммарная пауза %v\n",
		stats.TotalAlloc/1024/1024,
		stats.NumGC,
		time.Duration(stats.PauseTotalNs))
}

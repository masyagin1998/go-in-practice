package main

// Комбо-пример: Go → C (orchestrate) → Go (GoCompute) → C (add).

/*
extern void orchestrate(int x, int y);
extern int add(int a, int b);
*/
import "C"

import "fmt"

//export GoCompute
func GoCompute(x, y C.int) {
	// Из Go вызываем C-функцию add.
	result := C.add(x, y)
	fmt.Printf("GoCompute(): add(%d, %d) = %d (C вернул результат в Go)\n", x, y, result)
}

func main() {
	// Go вызывает C-функцию orchestrate,
	// которая вызывает Go-функцию GoCompute,
	// которая вызывает C-функцию add.
	C.orchestrate(3, 4)

	fmt.Println("Done (from Go)")
}

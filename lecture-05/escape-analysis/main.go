// Escape Analysis в Go — исчерпывающий набор примеров.
//
// Escape analysis — статический анализ компилятора, определяющий,
// где разместить переменную: на стеке (дёшево) или в куче (дорого).
// Переменная «убегает» (escapes) в кучу, если компилятор не может
// доказать, что её время жизни ограничено рамками функции.
//
// Правило простое: если компилятор НЕ УВЕРЕН — он кладёт в кучу.
// Лучше лишняя аллокация, чем dangling pointer.
//
// Сборка с выводом решений escape analysis:
//   go build -gcflags="-m -l" main.go
//
// Подробный вывод (причины побега):
//   go build -gcflags="-m -m -l" main.go
//
// Флаг -l отключает инлайнинг, чтобы видеть побег на уровне
// каждой функции, а не после встраивания.
//
// ПРИМЕЧАНИЕ: в этом файле используется //go:noinline вместо -l,
// чтобы каждая функция контролировала инлайнинг независимо.
// При проверке через go build -gcflags="-m -l" результат тот же:
// -l и //go:noinline дают одинаковый эффект для помеченных функций.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// ============================================================================
// 1. Возврат указателя из функции — КЛАССИКА побега.
// ============================================================================

type Point struct {
	X, Y int
}

// returnPointer — возвращает указатель на локальную переменную.
// Компилятор видит: p используется после выхода → куча.
//
// go build -gcflags="-m -l":
//   moved to heap: p
//
//go:noinline
func returnPointer() *Point {
	p := Point{X: 1, Y: 2} // moved to heap: p
	return &p
}

// returnValue — возвращает копию. Ничего не убегает → стек.
//
// go build -gcflags="-m -l":
//   (нет вывода — ничего не убегает)
//
//go:noinline
func returnValue() Point {
	p := Point{X: 1, Y: 2} // остаётся на стеке
	return p
}

// ============================================================================
// 2. Передача в interface{} / any — ЧАСТАЯ причина побега.
// ============================================================================

// printViaInterface — fmt.Println принимает ...any.
// Компилятор не может гарантировать, что значение не будет
// сохранено внутри fmt — поэтому кладёт в кучу.
//
// go build -gcflags="-m -l":
//   42 escapes to heap
//
//go:noinline
func printViaInterface() {
	x := 42
	fmt.Println(x) // x escapes to heap (через интерфейс)
}

// printDirect — fmt.Fprintf с форматированием: аргументы тоже
// передаются через ...any → всё равно побег.
//
// go build -gcflags="-m -l":
//   "Go" escapes to heap
//
//go:noinline
func printDirect() {
	name := "Go"
	fmt.Fprintf(os.Stdout, "Hello, %s\n", name) // name escapes
}

// noInterface — без интерфейсов, всё на стеке.
//
// go build -gcflags="-m -l":
//   (нет вывода — ничего не убегает)
//
//go:noinline
func noInterface() int {
	x := 42
	y := 58
	return x + y // никакого побега
}

// ============================================================================
// 3. new() и &T{} — НЕ ВСЕГДА куча!
// ============================================================================

// newStaysOnStack — new(Point) выделяет память, но если результат
// не убегает из функции, компилятор размещает его на стеке.
// new() — это подсказка, а не приказ. Компилятор умнее.
//
// go build -gcflags="-m -l":
//   new(Point) does not escape
//
//go:noinline
func newStaysOnStack() int {
	p := new(Point) // несмотря на new — остаётся на стеке!
	p.X = 10
	p.Y = 20
	return p.X + p.Y
}

// newEscapes — тот же new, но результат возвращается → куча.
//
// go build -gcflags="-m -l":
//   new(Point) escapes to heap
//
//go:noinline
func newEscapes() *Point {
	p := new(Point) // moved to heap
	p.X = 10
	p.Y = 20
	return p
}

// compositeStaysOnStack — &Point{} тоже может жить на стеке,
// если указатель не покидает функцию.
//
// go build -gcflags="-m -l":
//   &Point{...} does not escape
//
//go:noinline
func compositeStaysOnStack() int {
	p := &Point{X: 5, Y: 10} // стек!
	return p.X + p.Y
}

// ============================================================================
// 4. Структура с полем-указателем — классическая ловушка.
// ============================================================================

type Config struct {
	Name    string
	Logger  *Logger
}

type Logger struct {
	Prefix string
}

// configAllOnStack — структура и её поле-указатель инициализируются
// вместе, ничего не убегает → всё на стеке.
//
// go build -gcflags="-m -l":
//   cfg.Logger.Prefix + ": " + cfg.Name escapes to heap
//   (сама структура на стеке, убегает только результат конкатенации)
//
//go:noinline
func configAllOnStack() string {
	logger := Logger{Prefix: "[app]"} // стек
	cfg := Config{                     // стек
		Name:   "myapp",
		Logger: &logger, // указатель на стековую переменную — ОК,
		// потому что cfg тоже на стеке и не убегает.
	}
	return cfg.Logger.Prefix + ": " + cfg.Name
}

// configFieldEscapes — Logger создаётся отдельно и возвращается
// через Config → Logger убегает в кучу.
//
// go build -gcflags="-m -l":
//   moved to heap: logger
//   moved to heap: cfg
//
//go:noinline
func configFieldEscapes() *Config {
	logger := Logger{Prefix: "[app]"} // moved to heap (убегает через cfg)
	cfg := Config{                     // moved to heap (возвращается указатель)
		Name:   "myapp",
		Logger: &logger,
	}
	return &cfg
}

// configOnlyInnerEscapes — Config на стеке, но Logger убегает,
// потому что возвращается отдельно.
//
// go build -gcflags="-m -l":
//   moved to heap: logger
//
//go:noinline
func configOnlyInnerEscapes() *Logger {
	logger := Logger{Prefix: "[app]"} // moved to heap
	cfg := Config{                     // стек (не убегает сам)
		Name:   "myapp",
		Logger: &logger,
	}
	_ = cfg.Name
	return cfg.Logger // Logger убегает через возврат
}

// ============================================================================
// 5. Замыкания (closures) — захваченные переменные убегают.
// ============================================================================

// closureEscapes — переменная count захвачена замыканием.
// Замыкание живёт дольше функции → count убегает в кучу.
//
// go build -gcflags="-m -l":
//   moved to heap: count
//   func literal escapes to heap
//
//go:noinline
func closureEscapes() func() int {
	count := 0 // moved to heap (захвачена замыканием)
	return func() int {
		count++
		return count
	}
}

// closureNoEscape — замыкание используется и умирает внутри функции.
// Захваченная переменная НЕ убегает.
//
// go build -gcflags="-m -l":
//   func literal does not escape
//
//go:noinline
func closureNoEscape() int {
	sum := 0
	add := func(x int) { // замыкание не убегает
		sum += x
	}
	add(10)
	add(20)
	return sum // 30, всё на стеке
}

// ============================================================================
// 6. Слайсы — размер и возврат определяют судьбу.
// ============================================================================

// smallSliceOnStack — маленький слайс с известным размером,
// не возвращается → стек.
//
// go build -gcflags="-m -l":
//   make([]int, 4) does not escape
//
//go:noinline
func smallSliceOnStack() int {
	s := make([]int, 4) // стек (маленький, не убегает)
	s[0] = 1
	s[1] = 2
	s[2] = 3
	s[3] = 4
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

// sliceEscapesReturn — слайс возвращается → куча.
//
// go build -gcflags="-m -l":
//   make([]int, 4) escapes to heap
//
//go:noinline
func sliceEscapesReturn() []int {
	s := make([]int, 4) // moved to heap (возвращается)
	s[0] = 42
	return s
}

// sliceDynamicSize — размер известен только в рантайме.
// Несмотря на это, если слайс не убегает из функции,
// современный Go (1.20+) может разместить его на стеке.
// Компилятор использует runtime.stackalloc для динамических размеров.
//
// go build -gcflags="-m -l":
//   make([]int, n) does not escape
//
//go:noinline
func sliceDynamicSize(n int) int {
	s := make([]int, n) // does not escape (стек, даже при динамическом размере!)
	for i := range s {
		s[i] = i
	}
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

// sliceTooBig — слайс слишком большой для стека.
// Даже если не убегает — компилятор перемещает в кучу.
//
// go build -gcflags="-m -l":
//   make([]int, 100000) escapes to heap
//
//go:noinline
func sliceTooBig() int {
	s := make([]int, 100_000) // moved to heap (слишком большой)
	s[0] = 1
	return s[0]
}

// ============================================================================
// 7. Map — ВСЕГДА куча.
// ============================================================================

// mapInternalsOnHeap — map внутри использует кучу для хеш-таблицы,
// даже если сама переменная m «не убегает» с точки зрения escape analysis.
// Заголовок map (hmap) может остаться на стеке, но бакеты (buckets) —
// всегда в куче, потому что их размер и количество динамические.
// Итого: map = скрытые аллокации, даже если компилятор говорит "does not escape".
//
// go build -gcflags="-m -l":
//   map[string]int{...} does not escape
//   (но бакеты всё равно аллоцируются в куче — это скрыто от -m)
//
//go:noinline
func mapInternalsOnHeap() int {
	m := map[string]int{"a": 1, "b": 2} // hmap на стеке, но бакеты в куче
	return m["a"] + m["b"]
}

// ============================================================================
// 8. Отправка в канал — побег.
// ============================================================================

// sendToChannel — данные, отправленные в канал, убегают.
// Канал — разделяемая структура между горутинами;
// компилятор не может гарантировать, когда данные будут прочитаны.
//
// go build -gcflags="-m -l":
//   moved to heap: p
//
//go:noinline
func sendToChannel() int {
	ch := make(chan *Point, 1)
	p := Point{X: 1, Y: 2} // moved to heap (отправляется в канал)
	ch <- &p
	result := <-ch
	return result.X + result.Y
}

// ============================================================================
// 9. Горутины — захваченные переменные убегают.
// ============================================================================

// goroutineEscape — переменная, захваченная горутиной, убегает.
// Горутина может пережить функцию-родителя → куча.
//
// go build -gcflags="-m -l":
//   moved to heap: result
//   moved to heap: wg
//   func literal escapes to heap
//
//go:noinline
func goroutineEscape() int {
	result := 0 // moved to heap (захвачена горутиной)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result = 42
	}()
	wg.Wait()
	return result
}

// ============================================================================
// 10. Присвоение в глобальную переменную / поле интерфейса — побег.
// ============================================================================

var globalPoint *Point // глобальная переменная

// assignToGlobal — присвоение адреса в глобальную переменную.
// Глобальная переменная живёт вечно → данные убегают.
//
// go build -gcflags="-m -l":
//   moved to heap: p
//
//go:noinline
func assignToGlobal() {
	p := Point{X: 100, Y: 200} // moved to heap
	globalPoint = &p
}

// ============================================================================
// 11. Интерфейсные вызовы — метод через интерфейс.
// ============================================================================

type Writer interface {
	Write(data string)
}

type ConsoleWriter struct {
	Prefix string
}

func (w *ConsoleWriter) Write(data string) {
	fmt.Fprintf(os.Stdout, "%s: %s\n", w.Prefix, data)
}

// interfaceMethodEscape — объект хранится в интерфейсной переменной.
// Компилятор девиртуализирует вызов Write → видит конкретный тип
// *ConsoleWriter и знает, что w не сохраняется. Сама w остаётся
// на стеке. Убегают только аргументы fmt.Fprintf внутри Write:
// w.Prefix и "hello" — через ...any.
//
// go build -gcflags="-m -l":
//   devirtualizing iw.Write to *ConsoleWriter
//   (w не убегает — leaking param content: w означает утечку содержимого,
//    а не самой структуры; w.Prefix и data убегают внутри Write через fmt.Fprintf)
//
//go:noinline
func interfaceMethodEscape() {
	w := ConsoleWriter{Prefix: "LOG"} // стек (девиртуализация)
	var iw Writer = &w                // девиртуализируется в *ConsoleWriter
	iw.Write("hello")                 // w.Prefix и "hello" убегают через fmt.Fprintf
}

// ============================================================================
// 12. io.Writer и передача в стандартную библиотеку.
// ============================================================================

// writeToWriter — передача []byte в io.Writer.
// Сам слайс может убежать, потому что Writer — интерфейс.
//
// go build -gcflags="-m -l":
//   leaking param: w
//   ([]byte)("escape analysis demo\n") escapes to heap
//
//go:noinline
func writeToWriter(w io.Writer) {
	data := []byte("escape analysis demo\n") // убегает через w.Write
	_, _ = w.Write(data)
}

// ============================================================================
// 13. Массивы vs слайсы — массив фиксированного размера на стеке.
// ============================================================================

// arrayOnStack — массив фиксированного размера. В отличие от слайса,
// массив — value type, его размер известен на этапе компиляции → стек.
//
// go build -gcflags="-m -l":
//   (нет вывода — ничего не убегает)
//
//go:noinline
func arrayOnStack() int {
	var arr [100]int // стек (фиксированный размер, не убегает)
	for i := range arr {
		arr[i] = i
	}
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
}

// ============================================================================
// 14. Строки и конкатенация.
// ============================================================================

// stringConcatEscape — конкатенация строк создаёт новую строку.
// Если результат передаётся в интерфейс (fmt.Println) → куча.
//
// go build -gcflags="-m -l":
//   a + b escapes to heap
//   c escapes to heap
//
//go:noinline
func stringConcatEscape() {
	a := "hello"
	b := " world"
	c := a + b      // новая строка → может убежать
	fmt.Println(c) // через интерфейс → куча
}

// stringConcatStack — конкатенация без побега → может быть на стеке.
//
// go build -gcflags="-m -l":
//   a + b does not escape
//
//go:noinline
func stringConcatStack() int {
	a := "hello"
	b := " world"
	c := a + b
	return len(c) // не убегает → стек (если компилятор достаточно умён)
}

// ============================================================================
// 15. Указатель как аргумент — НЕ убегает, если не сохраняется.
// ============================================================================

// processPoint — принимает указатель, но не сохраняет его.
// Аргумент p «не убегает» с точки зрения вызываемой функции.
// Это значит, что вызывающая сторона может передать адрес
// стековой переменной — и это будет безопасно.
//
// go build -gcflags="-m -l":
//   p does not escape
//
//go:noinline
func processPoint(p *Point) int {
	return p.X*p.X + p.Y*p.Y // p does not escape
}

// callerCanUseStack — Point на стеке, даже при передаче по указателю.
//
// go build -gcflags="-m -l":
//   (нет вывода — Point остаётся на стеке)
//
//go:noinline
func callerCanUseStack() int {
	p := Point{X: 3, Y: 4}    // стек
	return processPoint(&p)    // &p не убегает — безопасно
}

// ============================================================================
// 16. Рекурсия и указатели — каждый фрейм стека независим.
// ============================================================================

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// buildTree — рекурсивное создание дерева.
// Каждый узел возвращается по указателю → все узлы в куче.
//
// go build -gcflags="-m -l":
//   moved to heap: node
//
//go:noinline
func buildTree(depth int) *TreeNode {
	if depth == 0 {
		return nil
	}
	node := TreeNode{Val: depth} // moved to heap
	node.Left = buildTree(depth - 1)
	node.Right = buildTree(depth - 1)
	return &node
}

// treeSum — обход дерева. Ничего не создаёт → всё на стеке.
//
// go build -gcflags="-m -l":
//   node does not escape
//
//go:noinline
func treeSum(node *TreeNode) int {
	if node == nil {
		return 0
	}
	return node.Val + treeSum(node.Left) + treeSum(node.Right) // node does not escape
}

// ============================================================================
// 17. Множественный возврат — указатель в кортеже.
// ============================================================================

// returnMultiple — возвращает указатель как часть кортежа → побег.
//
// go build -gcflags="-m -l":
//   moved to heap: p
//
//go:noinline
func returnMultiple() (*Point, error) {
	p := Point{X: 1, Y: 2} // moved to heap
	return &p, nil
}

// returnMultipleValue — возвращает значение, не указатель → стек.
//
// go build -gcflags="-m -l":
//   (нет вывода — ничего не убегает)
//
//go:noinline
func returnMultipleValue() (Point, error) {
	p := Point{X: 1, Y: 2} // стек
	return p, nil
}

// ============================================================================
// 18. Взятие адреса поля структуры.
// ============================================================================

type Pair struct {
	A int
	B int
}

// fieldAddressNoEscape — берём адрес поля, но не выносим наружу.
//
// go build -gcflags="-m -l":
//   (нет вывода — ничего не убегает)
//
//go:noinline
func fieldAddressNoEscape() int {
	pair := Pair{A: 10, B: 20}
	pa := &pair.A // указатель на поле — ОК, если не убегает
	*pa = 100
	return pair.A + pair.B // 120, всё на стеке
}

// fieldAddressEscapes — адрес поля возвращается → вся структура в куче.
//
// go build -gcflags="-m -l":
//   moved to heap: pair
//
//go:noinline
func fieldAddressEscapes() *int {
	pair := Pair{A: 10, B: 20} // moved to heap (берём адрес поля)
	return &pair.A               // указатель на поле → pair убегает
}

// ============================================================================
// main — запуск всех примеров.
// ============================================================================

func main() {
	fmt.Println("=== Escape Analysis: примеры ===")
	fmt.Println()

	// 1. Возврат указателя vs значения.
	fmt.Println("--- 1. Возврат указателя vs значения ---")
	p1 := returnPointer()
	p2 := returnValue()
	fmt.Printf("  returnPointer (куча):  %+v, адрес: %p\n", *p1, p1)
	fmt.Printf("  returnValue  (стек):   %+v\n", p2)
	fmt.Println()

	// 2. Интерфейсы.
	fmt.Println("--- 2. Передача в interface{} ---")
	printViaInterface()
	printDirect()
	fmt.Printf("  noInterface (стек): %d\n", noInterface())
	fmt.Println()

	// 3. new() не обязательно = куча.
	fmt.Println("--- 3. new() — не всегда куча ---")
	fmt.Printf("  newStaysOnStack: %d (стек!)\n", newStaysOnStack())
	p3 := newEscapes()
	fmt.Printf("  newEscapes:      %+v, адрес: %p (куча)\n", *p3, p3)
	fmt.Printf("  compositeStaysOnStack: %d (стек!)\n", compositeStaysOnStack())
	fmt.Println()

	// 4. Структура с полем-указателем.
	fmt.Println("--- 4. Структура с полем-указателем ---")
	fmt.Printf("  configAllOnStack:       %q (всё на стеке)\n", configAllOnStack())
	cfg := configFieldEscapes()
	fmt.Printf("  configFieldEscapes:     %+v (всё в куче)\n", cfg)
	logger := configOnlyInnerEscapes()
	fmt.Printf("  configOnlyInnerEscapes: %+v (Logger в куче)\n", logger)
	fmt.Println()

	// 5. Замыкания.
	fmt.Println("--- 5. Замыкания ---")
	counter := closureEscapes()
	fmt.Printf("  closureEscapes:  %d, %d, %d\n", counter(), counter(), counter())
	fmt.Printf("  closureNoEscape: %d (всё на стеке)\n", closureNoEscape())
	fmt.Println()

	// 6. Слайсы.
	fmt.Println("--- 6. Слайсы ---")
	fmt.Printf("  smallSliceOnStack: %d (стек)\n", smallSliceOnStack())
	s := sliceEscapesReturn()
	fmt.Printf("  sliceEscapesReturn: %v (куча)\n", s)
	fmt.Printf("  sliceDynamicSize(5): %d (стек — динамический размер, но не убегает)\n", sliceDynamicSize(5))
	fmt.Printf("  sliceTooBig: %d (куча — слишком большой)\n", sliceTooBig())
	fmt.Println()

	// 7. Map.
	fmt.Println("--- 7. Map — скрытые аллокации ---")
	fmt.Printf("  mapInternalsOnHeap: %d\n", mapInternalsOnHeap())
	fmt.Println()

	// 8. Каналы.
	fmt.Println("--- 8. Каналы ---")
	fmt.Printf("  sendToChannel: %d (побег через канал)\n", sendToChannel())
	fmt.Println()

	// 9. Горутины.
	fmt.Println("--- 9. Горутины ---")
	fmt.Printf("  goroutineEscape: %d (побег в горутину)\n", goroutineEscape())
	fmt.Println()

	// 10. Глобальные переменные.
	fmt.Println("--- 10. Глобальные переменные ---")
	assignToGlobal()
	fmt.Printf("  assignToGlobal: %+v (побег в глобальную)\n", *globalPoint)
	fmt.Println()

	// 11. Интерфейсные вызовы.
	fmt.Println("--- 11. Метод через интерфейс ---")
	interfaceMethodEscape()
	fmt.Println()

	// 12. io.Writer.
	fmt.Println("--- 12. io.Writer ---")
	writeToWriter(os.Stdout)

	// 13. Массивы vs слайсы.
	fmt.Println("--- 13. Массив vs слайс ---")
	fmt.Printf("  arrayOnStack: %d (массив на стеке)\n", arrayOnStack())
	fmt.Println()

	// 14. Строки.
	fmt.Println("--- 14. Строки ---")
	stringConcatEscape()
	fmt.Printf("  stringConcatStack: %d\n", stringConcatStack())
	fmt.Println()

	// 15. Указатель как аргумент без побега.
	fmt.Println("--- 15. Указатель без побега ---")
	fmt.Printf("  callerCanUseStack: %d (Point на стеке!)\n", callerCanUseStack())
	fmt.Println()

	// 16. Рекурсия.
	fmt.Println("--- 16. Рекурсия ---")
	tree := buildTree(3)
	fmt.Printf("  buildTree(3) sum: %d (все узлы в куче)\n", treeSum(tree))
	fmt.Println()

	// 17. Множественный возврат.
	fmt.Println("--- 17. Множественный возврат ---")
	pm, _ := returnMultiple()
	pv, _ := returnMultipleValue()
	fmt.Printf("  returnMultiple:      %+v (куча)\n", *pm)
	fmt.Printf("  returnMultipleValue: %+v (стек)\n", pv)
	fmt.Println()

	// 18. Адрес поля структуры.
	fmt.Println("--- 18. Адрес поля структуры ---")
	fmt.Printf("  fieldAddressNoEscape: %d (стек)\n", fieldAddressNoEscape())
	fa := fieldAddressEscapes()
	fmt.Printf("  fieldAddressEscapes:  %d (куча — адрес поля убежал)\n", *fa)
	fmt.Println()

	fmt.Println("=== Проверьте вывод escape analysis ===")
	fmt.Println("  go build -gcflags=\"-m -l\" .")
	fmt.Println("  go build -gcflags=\"-m -m -l\" .")
}

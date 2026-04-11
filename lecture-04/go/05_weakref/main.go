// Пример 5: Слабые ссылки (weak references) в Go.
//
// Пакет weak (с Go 1.24) позволяет хранить ссылку на объект,
// не препятствуя его сборке мусором. Когда GC собирает объект —
// weak.Pointer.Value() возвращает nil.
//
// Типичное применение:
//   - Кэши: храним объект, пока он кому-то нужен; GC сам вычищает неиспользуемое.
//   - Канонизация (interning): unique.Handle использует weak refs внутри.
//   - Наблюдатели (observers): подписка без удержания объекта.
//
// Запуск:
//   go run main.go

package main

import (
	"fmt"
	"runtime"
	"weak"
)

type BigObject struct {
	Name string
	Data [1024 * 1024]byte // ~1 МБ
}

// demoBasicWeakRef — базовый пример: слабая ссылка обнуляется после GC.
func demoBasicWeakRef() {
	fmt.Println("=== Базовый weak ref ===")

	// Создаём объект и слабую ссылку на него.
	obj := &BigObject{Name: "Тяжёлый объект"}
	obj.Data[0] = 42

	w := weak.Make(obj)

	// Пока obj жив — слабая ссылка работает.
	if val := w.Value(); val != nil {
		fmt.Printf("  До GC: %s (Data[0]=%d)\n", val.Name, val.Data[0])
	}

	// Убираем единственную сильную ссылку.
	obj = nil

	// Запускаем GC — объект может быть собран.
	runtime.GC()

	// Проверяем слабую ссылку.
	if val := w.Value(); val != nil {
		fmt.Printf("  После GC: %s (объект выжил)\n", val.Name)
	} else {
		fmt.Println("  После GC: объект собран, weak ref вернул nil")
	}
}

// WeakCache — простой кэш на слабых ссылках.
// Объекты живут в кэше, пока на них есть сильные ссылки снаружи.
// Когда внешний код перестаёт держать объект — GC его собирает,
// и при следующем обращении к кэшу получаем nil (промах).
type WeakCache struct {
	entries map[string]weak.Pointer[BigObject]
}

func NewWeakCache() *WeakCache {
	return &WeakCache{entries: make(map[string]weak.Pointer[BigObject])}
}

// Put — кладём объект в кэш. Кэш хранит только слабую ссылку.
func (c *WeakCache) Put(key string, obj *BigObject) {
	c.entries[key] = weak.Make(obj)
}

// Get — достаём объект. Если GC его собрал — возвращаем nil, ok=false.
func (c *WeakCache) Get(key string) (*BigObject, bool) {
	wp, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	val := wp.Value()
	if val == nil {
		// Объект собран GC — чистим запись.
		delete(c.entries, key)
		return nil, false
	}
	return val, true
}

// demoWeakCache — пример использования кэша на слабых ссылках.
func demoWeakCache() {
	fmt.Println("\n=== Кэш на weak refs ===")
	cache := NewWeakCache()

	// Создаём объекты и кладём в кэш.
	// Сохраняем сильные ссылки в слайс — пока слайс жив, объекты не соберутся.
	holders := make([]*BigObject, 0, 3)
	for _, name := range []string{"Альфа", "Бета", "Гамма"} {
		obj := &BigObject{Name: name}
		cache.Put(name, obj)
		holders = append(holders, obj)
		fmt.Printf("  Добавлен: %s\n", name)
	}

	// Все объекты живы — кэш работает.
	fmt.Println("  --- Все ссылки живы ---")
	for _, name := range []string{"Альфа", "Бета", "Гамма"} {
		if obj, ok := cache.Get(name); ok {
			fmt.Printf("  Кэш[%s]: попадание (%s)\n", name, obj.Name)
		} else {
			fmt.Printf("  Кэш[%s]: промах\n", name)
		}
	}

	// Убираем сильные ссылки на «Альфа» и «Бета» — оставляем только «Гамма».
	gamma := holders[2]
	holders = nil
	runtime.GC()

	fmt.Println("  --- После GC (держим только Гамму) ---")
	for _, name := range []string{"Альфа", "Бета", "Гамма"} {
		if obj, ok := cache.Get(name); ok {
			fmt.Printf("  Кэш[%s]: попадание (%s)\n", name, obj.Name)
		} else {
			fmt.Printf("  Кэш[%s]: промах (собран GC)\n", name)
		}
	}

	// Убираем последнюю ссылку.
	runtime.KeepAlive(gamma) // Гарантируем, что gamma жива до этой точки.
	gamma = nil
	runtime.GC()

	fmt.Println("  --- После GC (все ссылки убраны) ---")
	for _, name := range []string{"Альфа", "Бета", "Гамма"} {
		if obj, ok := cache.Get(name); ok {
			fmt.Printf("  Кэш[%s]: попадание (%s)\n", name, obj.Name)
		} else {
			fmt.Printf("  Кэш[%s]: промах (собран GC)\n", name)
		}
	}
}

func main() {
	demoBasicWeakRef()
	demoWeakCache()
}

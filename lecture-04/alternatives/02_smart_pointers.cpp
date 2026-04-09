// Пример 2: Умные указатели в C++ (unique_ptr, shared_ptr, weak_ptr).
// Стандартная библиотека C++ предоставляет RAII-обёртки над указателями,
// которые автоматически вызывают delete при уничтожении.

#include <cstdio>
#include <memory>
#include <string>
#include <vector>

struct Sensor {
    std::string name;
    double value;

    Sensor(std::string n, double v) : name(std::move(n)), value(v) {
        std::printf("  [Sensor] Создан: %s\n", name.c_str());
    }
    ~Sensor() {
        std::printf("  [Sensor] Уничтожен: %s\n", name.c_str());
    }
};

// === unique_ptr: единоличное владение ===
// Только один unique_ptr может владеть объектом.
// При уничтожении unique_ptr объект автоматически удаляется.
void demo_unique_ptr() {
    std::printf("\n--- unique_ptr ---\n");

    // make_unique — безопасное создание (нет утечки при исключении).
    auto sensor = std::make_unique<Sensor>("Температура", 36.6);
    std::printf("  Показание: %s = %.1f\n", sensor->name.c_str(), sensor->value);

    // Передача владения через std::move.
    auto transferred = std::move(sensor);
    // sensor теперь nullptr — владение передано.
    std::printf("  sensor == nullptr: %s\n", sensor == nullptr ? "да" : "нет");
    std::printf("  transferred: %s = %.1f\n",
                transferred->name.c_str(), transferred->value);

    // При выходе из функции transferred уничтожится → Sensor удалится.
}

// === shared_ptr: разделяемое владение ===
// Несколько shared_ptr могут указывать на один объект.
// Объект удаляется, когда последний shared_ptr уничтожен (счётчик ссылок = 0).
void demo_shared_ptr() {
    std::printf("\n--- shared_ptr ---\n");

    auto sensor = std::make_shared<Sensor>("Давление", 760.0);
    std::printf("  Счётчик ссылок: %ld\n", sensor.use_count()); // 1

    {
        // Создаём ещё один shared_ptr на тот же объект.
        auto copy = sensor;
        std::printf("  Счётчик ссылок: %ld\n", sensor.use_count()); // 2

        copy->value = 755.0;
        // copy уничтожается при выходе из блока, счётчик → 1.
    }

    std::printf("  Счётчик ссылок: %ld\n", sensor.use_count()); // 1
    std::printf("  Значение: %.1f\n", sensor->value);            // 755.0

    // sensor уничтожается → Sensor удаляется.
}

// === weak_ptr: слабая ссылка (не продлевает жизнь объекта) ===
// Используется для разрыва циклических зависимостей.
void demo_weak_ptr() {
    std::printf("\n--- weak_ptr ---\n");

    std::weak_ptr<Sensor> weak;

    {
        auto sensor = std::make_shared<Sensor>("Влажность", 45.0);
        weak = sensor; // weak_ptr не увеличивает счётчик.
        std::printf("  Счётчик ссылок: %ld\n", sensor.use_count()); // 1

        // Чтобы использовать weak_ptr, нужно «продвинуть» его до shared_ptr.
        if (auto locked = weak.lock()) {
            std::printf("  Объект жив: %s = %.1f\n",
                        locked->name.c_str(), locked->value);
        }
    }

    // sensor уничтожен — weak_ptr это знает.
    std::printf("  Объект жив: %s\n", weak.expired() ? "нет" : "да");
    if (auto locked = weak.lock()) {
        std::printf("  Этого не будет напечатано.\n");
    } else {
        std::printf("  weak.lock() вернул nullptr — объект удалён.\n");
    }
}

int main() {
    demo_unique_ptr();
    demo_shared_ptr();
    demo_weak_ptr();

    std::printf("\nВсе ресурсы автоматически освобождены.\n");
    return 0;
}

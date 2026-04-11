// Пример 5: Утечка памяти из-за циклических ссылок shared_ptr.
//
// shared_ptr использует счётчик ссылок. Если два объекта ссылаются
// друг на друга через shared_ptr, счётчик никогда не достигнет нуля —
// память утекает, деструкторы не вызываются.
//
// Решение — заменить одну из ссылок на weak_ptr (показано ниже).

#include <cstdio>
#include <memory>
#include <string>

// ============================================================
// Часть 1: Утечка — оба поля shared_ptr
// ============================================================

struct BadNode {
    std::string name;
    std::shared_ptr<BadNode> peer; // сильная ссылка → цикл!

    BadNode(std::string n) : name(std::move(n)) {
        std::printf("  [BadNode]  Создан: %s\n", name.c_str());
    }
    ~BadNode() {
        std::printf("  [BadNode]  Уничтожен: %s\n", name.c_str());
    }
};

// Демонстрация утечки: A → B → A, счётчик никогда не станет 0.
void demo_leak() {
    std::printf("\n=== Циклическая ссылка (shared_ptr) — УТЕЧКА ===\n");

    auto a = std::make_shared<BadNode>("A");
    auto b = std::make_shared<BadNode>("B");

    // Создаём цикл: A ↔ B.
    a->peer = b;
    b->peer = a;

    std::printf("  a.use_count = %ld (a сам + b->peer)\n", a.use_count()); // 2
    std::printf("  b.use_count = %ld (b сам + a->peer)\n", b.use_count()); // 2

    // При выходе из функции локальные a и b уничтожаются:
    //   a.use_count: 2 → 1  (b->peer всё ещё жив)
    //   b.use_count: 2 → 1  (a->peer всё ещё жив)
    // Счётчик не дошёл до 0 → деструкторы НЕ вызываются → утечка!
    std::printf("  Выход из функции — ждём деструкторы...\n");
}

// ============================================================
// Часть 2: Исправление — weak_ptr разрывает цикл
// ============================================================

struct GoodNode {
    std::string name;
    std::weak_ptr<GoodNode> peer; // слабая ссылка → цикла нет

    GoodNode(std::string n) : name(std::move(n)) {
        std::printf("  [GoodNode] Создан: %s\n", name.c_str());
    }
    ~GoodNode() {
        std::printf("  [GoodNode] Уничтожен: %s\n", name.c_str());
    }

    // Чтобы использовать peer, нужно «продвинуть» weak_ptr до shared_ptr.
    void greet_peer() {
        if (auto p = peer.lock()) {
            std::printf("  %s → привет, %s!\n", name.c_str(), p->name.c_str());
        } else {
            std::printf("  %s → peer уже удалён\n", name.c_str());
        }
    }
};

// Демонстрация: weak_ptr не увеличивает счётчик — цикла нет.
void demo_fixed() {
    std::printf("\n=== weak_ptr разрывает цикл — БЕЗ УТЕЧКИ ===\n");

    auto a = std::make_shared<GoodNode>("A");
    auto b = std::make_shared<GoodNode>("B");

    a->peer = b;
    b->peer = a;

    // weak_ptr не увеличивает use_count.
    std::printf("  a.use_count = %ld\n", a.use_count()); // 1
    std::printf("  b.use_count = %ld\n", b.use_count()); // 1

    a->greet_peer();
    b->greet_peer();

    std::printf("  Выход из функции — ждём деструкторы...\n");
    // use_count: 1 → 0 → деструкторы вызваны!
}

// ============================================================

int main() {
    demo_leak();
    std::printf("  ^ Деструкторы BadNode НЕ были вызваны — утечка!\n");

    demo_fixed();
    std::printf("  ^ Деструкторы GoodNode вызваны — всё чисто.\n");

    return 0;
}

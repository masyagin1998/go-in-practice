// Пример 1: RAII (Resource Acquisition Is Initialization) в C++.
// Ресурс захватывается в конструкторе и освобождается в деструкторе.
// При выходе из области видимости деструктор вызывается автоматически —
// утечки невозможны, даже при исключениях.

#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <stdexcept>

// RAII-обёртка над динамическим буфером.
class Buffer {
public:
    // Конструктор: захват ресурса.
    explicit Buffer(size_t size) : size_(size) {
        data_ = new char[size];
        std::printf("  [Buffer] Выделено %zu байт\n", size);
    }

    // Деструктор: освобождение ресурса. Вызывается автоматически!
    ~Buffer() {
        delete[] data_;
        std::printf("  [Buffer] Освобождено %zu байт\n", size_);
    }

    // Запрещаем копирование (чтобы не было двойного delete).
    Buffer(const Buffer &) = delete;
    Buffer &operator=(const Buffer &) = delete;

    // Разрешаем перемещение.
    Buffer(Buffer &&other) noexcept : data_(other.data_), size_(other.size_) {
        other.data_ = nullptr;
        other.size_ = 0;
    }

    char *data() { return data_; }
    size_t size() const { return size_; }

private:
    char *data_;
    size_t size_;
};

// RAII-обёртка над файлом.
class File {
public:
    explicit File(const char *path, const char *mode) {
        fp_ = std::fopen(path, mode);
        if (!fp_) {
            throw std::runtime_error("Не удалось открыть файл");
        }
        std::printf("  [File] Файл открыт: %s\n", path);
    }

    ~File() {
        if (fp_) {
            std::fclose(fp_);
            std::printf("  [File] Файл закрыт\n");
        }
    }

    File(const File &) = delete;
    File &operator=(const File &) = delete;

    FILE *get() { return fp_; }

private:
    FILE *fp_;
};

void process() {
    std::printf("Начало process():\n");

    Buffer buf(256);                         // Память выделена.
    std::snprintf(buf.data(), buf.size(), "Привет, RAII!");
    std::printf("  Содержимое: %s\n", buf.data());

    File f("/tmp/raii_example.txt", "w");    // Файл открыт.
    std::fprintf(f.get(), "%s\n", buf.data());

    std::printf("Конец process():\n");
    // Здесь автоматически вызываются деструкторы: ~File(), затем ~Buffer().
    // Даже если бы выше было throw — деструкторы всё равно вызвались бы!
}

int main() {
    process();
    // К этому моменту все ресурсы гарантированно освобождены.
    std::printf("Все ресурсы освобождены.\n");
    return 0;
}

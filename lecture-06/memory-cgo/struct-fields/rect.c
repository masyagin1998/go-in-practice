#include "rect.h"

// area вычисляет площадь прямоугольника.
double area(Rect r) {
    return r.size.x * r.size.y;
}

// scale_width масштабирует ширину прямоугольника.
void scale_width(Rect* r, double factor) {
    r->size.x *= factor;
}

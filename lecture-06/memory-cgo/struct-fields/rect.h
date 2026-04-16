#ifndef RECT_H
#define RECT_H

// Rect — прямоугольник с двумя точками (без указателей).
typedef struct {
    double x;
    double y;
} Vec2;

typedef struct {
    Vec2 origin;
    Vec2 size;
} Rect;

// area вычисляет площадь прямоугольника.
double area(Rect r);

// scale_width масштабирует ширину прямоугольника.
void scale_width(Rect* r, double factor);

#endif

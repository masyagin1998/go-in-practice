#include "point.h"
#include <math.h>

// distance вычисляет расстояние от точки до начала координат.
double distance(Point p) {
    return sqrt(p.x * p.x + p.y * p.y);
}

// translate сдвигает точку на (dx, dy) и возвращает новую.
Point translate(Point p, double dx, double dy) {
    Point result = { p.x + dx, p.y + dy };
    return result;
}

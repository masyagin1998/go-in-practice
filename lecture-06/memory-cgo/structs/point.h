#ifndef POINT_H
#define POINT_H

// Point — точка в 2D без указателей внутри.
typedef struct {
    double x;
    double y;
} Point;

// distance вычисляет расстояние от точки до начала координат.
double distance(Point p);

// translate сдвигает точку на (dx, dy) и возвращает новую.
Point translate(Point p, double dx, double dy);

#endif

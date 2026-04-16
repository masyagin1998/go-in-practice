// mathutil.c — реализация с использованием libm.
#include "include/mathutil.h"
#include <math.h>

double hypotenuse(double a, double b) {
    return sqrt(a * a + b * b);
}

#ifndef VALUE_H
#define VALUE_H

// Union — все поля разделяют одну и ту же память.
// CGo представляет union как [N]byte — доступа к полям нет.
typedef union {
    int   i;
    float f;
    char  bytes[4];
} Value;

// Tagged union — типичный паттерн в C для «полиморфных» значений.
typedef struct {
    char  type; // 'i' = int, 'f' = float
    Value val;
} Tagged;

Tagged make_int(int x);
Tagged make_float(float x);

int   tagged_as_int(Tagged t);
float tagged_as_float(Tagged t);
void  tagged_print(Tagged t);

#endif

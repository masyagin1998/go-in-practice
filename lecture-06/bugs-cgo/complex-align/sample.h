#ifndef SAMPLE_H
#define SAMPLE_H

#include <complex.h>

// Packed-структура с complex float.
// CGo не может представить поля с нарушенным выравниванием —
// val и count сворачиваются в непрозрачный массив байт (_).
typedef struct __attribute__((packed)) {
    char          tag;    // offset 0, size 1
    float complex val;    // offset 1, size 8 (выравнивание нарушено!)
    int           count;  // offset 9, size 4
} Sample;                 // sizeof = 13

// C-функции-аксессоры — единственный способ работать с такими полями из Go.
Sample make_sample(char tag, float re, float im, int count);

char          sample_tag(Sample s);
float         sample_re(Sample s);
float         sample_im(Sample s);
int           sample_count(Sample s);
void          sample_print(Sample s);

#endif

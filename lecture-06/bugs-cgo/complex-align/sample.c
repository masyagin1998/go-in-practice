#include "sample.h"
#include <stdio.h>

Sample make_sample(char tag, float re, float im, int count) {
    Sample s;
    s.tag   = tag;
    s.val   = re + im * I;
    s.count = count;
    return s;
}

char sample_tag(Sample s)   { return s.tag; }
float sample_re(Sample s)   { return crealf(s.val); }
float sample_im(Sample s)   { return cimagf(s.val); }
int sample_count(Sample s)  { return s.count; }

void sample_print(Sample s) {
    printf("[C] Sample{tag='%c', val=%.1f+%.1fi, count=%d}\n",
           s.tag, crealf(s.val), cimagf(s.val), s.count);
    fflush(stdout);
}

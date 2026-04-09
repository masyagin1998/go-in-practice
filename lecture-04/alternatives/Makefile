CC = gcc
CXX = g++
CFLAGS = -Wall -Wextra -g
CXXFLAGS = -Wall -Wextra -g -std=c++17

C_SOURCES = $(wildcard *.c)
CXX_SOURCES = $(wildcard *.cpp)
C_TARGETS = $(C_SOURCES:.c=)
CXX_TARGETS = $(CXX_SOURCES:.cpp=)
TARGETS = $(C_TARGETS) $(CXX_TARGETS)

.PHONY: all clean

all: $(TARGETS)

%: %.c
	$(CC) $(CFLAGS) -o $@ $<

%: %.cpp
	$(CXX) $(CXXFLAGS) -o $@ $<

clean:
	rm -f $(TARGETS)

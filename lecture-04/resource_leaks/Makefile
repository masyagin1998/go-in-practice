VALGRIND_DIRS = 00_no_leak
DIRS = 01_goroutine_leak 02_fd_leak 03_http_body_leak 04_ticker_leak 05_context_leak

ALL_DIRS = $(VALGRIND_DIRS) $(DIRS)

.PHONY: all clean $(ALL_DIRS)

all: $(ALL_DIRS)

$(VALGRIND_DIRS):
	cd $@ && go build -tags valgrind -o $@ .

$(DIRS):
	cd $@ && go build -o $@ .

clean:
	@for d in $(ALL_DIRS); do rm -f $$d/$$d; done

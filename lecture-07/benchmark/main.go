package main

// benchmark — ping-pong round-trip между Go-клиентом и C-peer'ом через
// разные транспорты. Каждый транспорт гоняется REPEATS раз, итоговая
// таблица сортируется по медиане времени round-trip'а (быстрые сверху).
//
// Параметры через env:
//   BENCH_REPEATS (default 30)     — сколько раз повторить каждый транспорт
//   BENCH_ITERS   (default 500000) — ping-pong'ов в одном прогоне
//
// Короткий прогон (≈5 секунд):
//   BENCH_REPEATS=3 BENCH_ITERS=100000 make bench
// Длинный дефолт — около 10 минут суммарно.
//
// Реальные числа чувствительны к CPU governor, NUMA, nodelay, нагрузке —
// это порядок величин, не продакшн-бенч.

/*
#cgo LDFLAGS: -lrt -lpthread
#include <errno.h>
#include <fcntl.h>
#include <mqueue.h>
#include <semaphore.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

static int shm_create(const char *name, size_t size) {
    int fd = shm_open(name, O_CREAT | O_RDWR, 0600);
    if (fd < 0) return -1;
    if (ftruncate(fd, (off_t)size) < 0) { close(fd); shm_unlink(name); return -1; }
    return fd;
}

static sem_t *sem_create(const char *name) {
    sem_unlink(name);
    return sem_open(name, O_CREAT | O_EXCL, 0600, 0);
}

static mqd_t mq_create(const char *name, long maxmsg, long msgsize, int readonly) {
    struct mq_attr a = {0};
    a.mq_maxmsg = maxmsg;
    a.mq_msgsize = msgsize;
    int flags = (readonly ? O_RDONLY : O_WRONLY) | O_CREAT | O_EXCL;
    return mq_open(name, flags, 0600, &a);
}

static int mq_is_err(mqd_t q) { return q == (mqd_t)-1; }

static int sem_wait_w(sem_t *s) { return sem_wait(s); }
static int sem_post_w(sem_t *s) { return sem_post(s); }
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

const msgSize = 256

// Дефолты подогнаны под ~10 минут полного прогона (6 транспортов ×
// 50 запусков × 500k round-trip'ов + warmup). На TCP loopback это самая
// длинная часть; на shm+sem — самая короткая.
var (
	repeats = envInt("BENCH_REPEATS", 50)
	iters   = envInt("BENCH_ITERS", 500_000)
)

type transport struct {
	name string
	fn   func(iters int) (time.Duration, error)
}

type result struct {
	name              string
	ok                int
	best, median, max time.Duration
}

func main() {
	// Имена совпадают с ipc/01_..07_. 06_shm (raw busy-wait) не включён —
	// см. ipc/06_shmem/ как standalone-демо.
	transports := []transport{
		{"01_pipe", benchPipe},
		{"02_fifo", benchFifo},
		{"03_tcp", benchTCP},
		{"04_uds", benchUDS},
		{"05_mq", benchMQ},
		{"07_shm+sem", benchShmSem},
	}

	fmt.Printf("MSG_SIZE=%d B, ITERS=%d, REPEATS=%d (+1 warmup на транспорт)\n",
		msgSize, iters, repeats)
	fmt.Printf("Грубая оценка полного прогона: ~%d сек\n\n",
		estimateSeconds(repeats, iters, len(transports)))

	results := make([]result, 0, len(transports))
	totalStart := time.Now()
	for _, t := range transports {
		results = append(results, runTransport(t))
	}
	fmt.Printf("\nВсего: %v\n\n", time.Since(totalStart).Round(time.Second))

	// Быстрые — сверху (по медиане round-trip'а).
	slices.SortFunc(results, func(a, b result) int {
		return int(a.median - b.median)
	})

	printTable(results)
}

func runTransport(t transport) result {
	fmt.Printf("[%-8s] warmup... ", t.name)
	if _, err := t.fn(iters); err != nil {
		fmt.Printf("warmup failed: %v\n", err)
		return result{name: t.name}
	}

	durs := make([]time.Duration, 0, repeats)
	for i := range repeats {
		d, err := t.fn(iters)
		if err != nil {
			fmt.Printf("\n  run %d: %v", i, err)
			continue
		}
		durs = append(durs, d)
		fmt.Printf("\r[%-8s] %d/%d  last=%v ",
			t.name, i+1, repeats, d.Round(time.Millisecond))
	}
	fmt.Println()
	return summarize(t.name, durs)
}

func summarize(name string, runs []time.Duration) result {
	if len(runs) == 0 {
		return result{name: name}
	}
	sorted := slices.Clone(runs)
	slices.Sort(sorted)
	return result{
		name:   name,
		ok:     len(sorted),
		best:   sorted[0],
		median: sorted[len(sorted)/2],
		max:    sorted[len(sorted)-1],
	}
}

func printTable(results []result) {
	fmt.Println("Результаты (сортировка по медиане ns/round-trip, быстрые сверху):")
	fmt.Println()
	fmt.Printf("%-10s %5s %12s %12s %12s %12s\n",
		"transport", "runs", "msgs/sec*", "min ns/rt", "median ns/rt", "max ns/rt")
	fmt.Println("------------------------------------------------------------------------")
	for _, r := range results {
		if r.ok == 0 {
			fmt.Printf("%-10s %5s %12s %12s %12s %12s\n", r.name, "—", "—", "—", "—", "—")
			continue
		}
		msgsPerS := float64(iters) / r.median.Seconds()
		fmt.Printf("%-10s %5d %12.0f %12d %12d %12d\n",
			r.name, r.ok, msgsPerS,
			int64(r.best)/int64(iters),
			int64(r.median)/int64(iters),
			int64(r.max)/int64(iters))
	}
	fmt.Println()
	fmt.Println("* msgs/sec посчитаны от медианы round-trip'а.")
}

func estimateSeconds(repeats, iters, nTransports int) int {
	// Грубо: ~4µs/rt в среднем по транспортам × iters × (repeats+1 warmup) × N.
	const avgNsPerRt = int64(4000)
	total := avgNsPerRt * int64(iters) * int64(repeats+1) * int64(nTransports)
	return int(total / 1_000_000_000)
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// ---------------------- pipe ----------------------

func benchPipe(iters int) (time.Duration, error) {
	cmd := exec.Command("./peer_pipe", strconv.Itoa(iters))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	msg := make([]byte, msgSize)
	reply := make([]byte, msgSize)
	start := time.Now()
	for range iters {
		if _, err := stdin.Write(msg); err != nil {
			return 0, err
		}
		if _, err := io.ReadFull(stdout, reply); err != nil {
			return 0, err
		}
	}
	elapsed := time.Since(start)
	stdin.Close()
	return elapsed, cmd.Wait()
}

// ---------------------- fifo ----------------------

func benchFifo(iters int) (time.Duration, error) {
	const (
		c2s = "/tmp/ipc_bench_fifo_c2s"
		s2c = "/tmp/ipc_bench_fifo_s2c"
	)
	_ = os.Remove(c2s)
	_ = os.Remove(s2c)
	if err := syscall.Mkfifo(c2s, 0600); err != nil {
		return 0, err
	}
	if err := syscall.Mkfifo(s2c, 0600); err != nil {
		return 0, err
	}
	defer os.Remove(c2s)
	defer os.Remove(s2c)

	cmd := exec.Command("./peer_fifo", strconv.Itoa(iters))
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	type res struct {
		f   *os.File
		err error
	}
	wCh := make(chan res, 1)
	rCh := make(chan res, 1)
	go func() { f, e := os.OpenFile(c2s, os.O_WRONLY, 0); wCh <- res{f, e} }()
	go func() { f, e := os.OpenFile(s2c, os.O_RDONLY, 0); rCh <- res{f, e} }()
	wr := <-wCh
	rd := <-rCh
	if wr.f != nil {
		defer wr.f.Close()
	}
	if rd.f != nil {
		defer rd.f.Close()
	}
	if wr.err != nil {
		return 0, wr.err
	}
	if rd.err != nil {
		return 0, rd.err
	}

	msg := make([]byte, msgSize)
	reply := make([]byte, msgSize)
	start := time.Now()
	for range iters {
		if _, err := wr.f.Write(msg); err != nil {
			return 0, err
		}
		if _, err := io.ReadFull(rd.f, reply); err != nil {
			return 0, err
		}
	}
	elapsed := time.Since(start)
	return elapsed, cmd.Wait()
}

// ---------------------- mq ----------------------

func benchMQ(iters int) (time.Duration, error) {
	const (
		c2s = "/ipc_bench_mq_c2s"
		s2c = "/ipc_bench_mq_s2c"
	)
	cc := C.CString(c2s)
	cs := C.CString(s2c)
	defer C.free(unsafe.Pointer(cc))
	defer C.free(unsafe.Pointer(cs))
	C.mq_unlink(cc)
	C.mq_unlink(cs)

	qC2S := C.mq_create(cc, 10, msgSize, 0)
	qS2C := C.mq_create(cs, 10, msgSize, 1)
	if C.mq_is_err(qC2S) != 0 || C.mq_is_err(qS2C) != 0 {
		return 0, errors.New("mq_create failed")
	}
	defer C.mq_unlink(cc)
	defer C.mq_unlink(cs)
	defer C.mq_close(qC2S)
	defer C.mq_close(qS2C)

	cmd := exec.Command("./peer_mq", strconv.Itoa(iters))
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	msg := make([]byte, msgSize)
	reply := make([]byte, msgSize)
	start := time.Now()
	for range iters {
		if rc := C.mq_send(qC2S,
			(*C.char)(unsafe.Pointer(&msg[0])),
			C.size_t(msgSize), 0); rc != 0 {
			return 0, errors.New("mq_send")
		}
		if n := C.mq_receive(qS2C,
			(*C.char)(unsafe.Pointer(&reply[0])),
			C.size_t(msgSize), nil); n < 0 {
			return 0, errors.New("mq_receive")
		}
	}
	elapsed := time.Since(start)
	return elapsed, cmd.Wait()
}

// ---------------------- shm+sem ----------------------

func benchShmSem(iters int) (time.Duration, error) {
	const (
		shmName   = "/ipc_bench_shm"
		semClient = "/ipc_bench_sem_c"
		semServer = "/ipc_bench_sem_s"
	)
	cShm := C.CString(shmName)
	cSc := C.CString(semClient)
	cSs := C.CString(semServer)
	defer C.free(unsafe.Pointer(cShm))
	defer C.free(unsafe.Pointer(cSc))
	defer C.free(unsafe.Pointer(cSs))

	C.shm_unlink(cShm)
	fd := C.shm_create(cShm, C.size_t(msgSize))
	if fd < 0 {
		return 0, errors.New("shm_create")
	}
	defer C.shm_unlink(cShm)

	sc := C.sem_create(cSc)
	ss := C.sem_create(cSs)
	if sc == nil || ss == nil {
		return 0, errors.New("sem_create")
	}
	defer C.sem_unlink(cSc)
	defer C.sem_unlink(cSs)
	defer C.sem_close(sc)
	defer C.sem_close(ss)

	region, err := syscall.Mmap(int(fd), 0, msgSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return 0, err
	}
	defer syscall.Munmap(region)
	syscall.Close(int(fd))

	cmd := exec.Command("./peer_shmsem", strconv.Itoa(iters))
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	msg := make([]byte, msgSize)
	reply := make([]byte, msgSize)
	start := time.Now()
	for range iters {
		copy(region, msg)
		C.sem_post_w(sc)
		C.sem_wait_w(ss)
		copy(reply, region)
	}
	elapsed := time.Since(start)
	return elapsed, cmd.Wait()
}

// ---------------------- tcp ----------------------

func benchTCP(iters int) (time.Duration, error) {
	// port=0 — ядро выбирает свободный, чтобы избежать TIME_WAIT между runs.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	cmd := exec.Command("./peer_tcp", strconv.Itoa(iters), strconv.Itoa(port))
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	conn, err := ln.Accept()
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetNoDelay(true)
	}
	return pingPong(conn, cmd, iters)
}

// ---------------------- uds ----------------------

func benchUDS(iters int) (time.Duration, error) {
	const path = "/tmp/ipc_bench_uds"
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	defer os.Remove(path)
	cmd := exec.Command("./peer_uds", strconv.Itoa(iters))
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	conn, err := ln.Accept()
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	return pingPong(conn, cmd, iters)
}

// ---------------------- helpers ----------------------

func pingPong(conn net.Conn, cmd *exec.Cmd, iters int) (time.Duration, error) {
	msg := make([]byte, msgSize)
	reply := make([]byte, msgSize)
	start := time.Now()
	for range iters {
		if _, err := conn.Write(msg); err != nil {
			return 0, err
		}
		if _, err := io.ReadFull(conn, reply); err != nil {
			return 0, err
		}
	}
	elapsed := time.Since(start)
	return elapsed, cmd.Wait()
}

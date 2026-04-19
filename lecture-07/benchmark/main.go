package main

// benchmark — ping-pong round-trip между Go-клиентом и C-peer'ом через
// разные транспорты. Для каждой пары (транспорт, размер сообщения)
// делается warmup + REPEATS прогонов, после чего по всем размерам
// печатается отдельная таблица.
//
// Параметры через env:
//   BENCH_REPEATS (default 20)     — сколько раз повторить каждый прогон
//   BENCH_ITERS   (default 100000) — ping-pong'ов в одном прогоне
//
// Короткий прогон (~10 секунд):
//   BENCH_REPEATS=3 BENCH_ITERS=50000 make bench
// Длинный дефолт — 10 транспортов × 2 размера × (20+1) прогонов.
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

// Размеры сообщений (байт). 256 — маленький, где zero-copy обычно
// проигрывает из-за оверхеда setup'а. 65536 — большой, где zc может
// реально выигрывать за счёт устранения копий.
var msgSizes = []int{256, 65536}

var (
	repeats = envInt("BENCH_REPEATS", 20)
	iters   = envInt("BENCH_ITERS", 100_000)
)

type benchFn func(iters, msgSize int) (time.Duration, error)

type transport struct {
	name string
	fn   benchFn
	// maxSize — если msgSize > maxSize, транспорт пропускается для
	// этого размера. 0 = без ограничений.
	maxSize int
}

type result struct {
	name              string
	msgSize           int
	ok                int
	best, median, max time.Duration
}

func main() {
	transports := []transport{
		{"01_pipe", benchPipe, 0},
		{"02_pipe_zc", benchPipeZC, 0},
		{"03_fifo", benchFifo, 0},
		{"04_fifo_zc", benchFifoZC, 0},
		{"05_tcp", benchTCP, 0},
		{"06_tcp_zc", benchTCPZC, 0},
		{"07_uds", benchUDS, 0},
		{"08_uds_zc", benchUDSZC, 0},
		// POSIX mqueue: системный лимит msgsize_max по умолчанию 8192.
		// На большем размере mq_open вернёт EINVAL — пропускаем.
		{"09_mq", benchMQ, 8192},
		{"10_shmsem", benchShmSem, 0},
	}

	fmt.Printf("MSG_SIZES=%v, ITERS=%d, REPEATS=%d (+1 warmup на конфигурацию)\n",
		msgSizes, iters, repeats)
	totalStart := time.Now()

	// results[size] = []result
	allResults := make(map[int][]result)

	for _, size := range msgSizes {
		fmt.Printf("\n=== msg_size=%d B ===\n", size)
		for _, t := range transports {
			if t.maxSize > 0 && size > t.maxSize {
				fmt.Printf("[%-10s] пропущен (размер > %d)\n", t.name, t.maxSize)
				continue
			}
			r := runTransport(t, size)
			allResults[size] = append(allResults[size], r)
		}
	}

	fmt.Printf("\nВсего: %v\n", time.Since(totalStart).Round(time.Second))

	for _, size := range msgSizes {
		rs := allResults[size]
		slices.SortFunc(rs, func(a, b result) int { return int(a.median - b.median) })
		printTable(rs, size)
	}
}

func runTransport(t transport, msgSize int) result {
	fmt.Printf("[%-10s size=%-5d] warmup... ", t.name, msgSize)
	if _, err := t.fn(iters, msgSize); err != nil {
		fmt.Printf("warmup failed: %v\n", err)
		return result{name: t.name, msgSize: msgSize}
	}

	durs := make([]time.Duration, 0, repeats)
	for i := range repeats {
		d, err := t.fn(iters, msgSize)
		if err != nil {
			fmt.Printf("\n  run %d: %v", i, err)
			continue
		}
		durs = append(durs, d)
		fmt.Printf("\r[%-10s size=%-5d] %d/%d  last=%v ",
			t.name, msgSize, i+1, repeats, d.Round(time.Millisecond))
	}
	fmt.Println()
	return summarize(t.name, msgSize, durs)
}

func summarize(name string, msgSize int, runs []time.Duration) result {
	if len(runs) == 0 {
		return result{name: name, msgSize: msgSize}
	}
	sorted := slices.Clone(runs)
	slices.Sort(sorted)
	return result{
		name:    name,
		msgSize: msgSize,
		ok:      len(sorted),
		best:    sorted[0],
		median:  sorted[len(sorted)/2],
		max:     sorted[len(sorted)-1],
	}
}

func printTable(results []result, msgSize int) {
	fmt.Println()
	fmt.Printf("msg_size=%d B, сортировка по медиане ns/round-trip:\n\n", msgSize)
	fmt.Printf("%-12s %5s %12s %14s %12s %12s %12s\n",
		"transport", "runs", "msgs/sec*", "MB/sec*", "min ns/rt", "median ns/rt", "max ns/rt")
	fmt.Println("------------------------------------------------------------------------------------------")
	for _, r := range results {
		if r.ok == 0 {
			fmt.Printf("%-12s %5s %12s %14s %12s %12s %12s\n",
				r.name, "—", "—", "—", "—", "—", "—")
			continue
		}
		msgsPerS := float64(iters) / r.median.Seconds()
		mbPerS := msgsPerS * float64(r.msgSize) / (1024 * 1024)
		fmt.Printf("%-12s %5d %12.0f %14.1f %12d %12d %12d\n",
			r.name, r.ok, msgsPerS, mbPerS,
			int64(r.best)/int64(iters),
			int64(r.median)/int64(iters),
			int64(r.max)/int64(iters))
	}
	fmt.Println()
	fmt.Println("* msgs/sec и MB/sec посчитаны от медианы round-trip'а (двусторонний обмен).")
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// ---------------------- pipe / pipe_zc ----------------------

func benchPipe(iters, msgSize int) (time.Duration, error) {
	return benchStdinStdout("./peer_pipe", iters, msgSize)
}

func benchPipeZC(iters, msgSize int) (time.Duration, error) {
	return benchStdinStdout("./peer_pipe_zc", iters, msgSize)
}

func benchStdinStdout(bin string, iters, msgSize int) (time.Duration, error) {
	cmd := exec.Command(bin, strconv.Itoa(iters), strconv.Itoa(msgSize))
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

// ---------------------- fifo / fifo_zc ----------------------

func benchFifo(iters, msgSize int) (time.Duration, error) {
	return benchFifoImpl("./peer_fifo",
		"/tmp/ipc_bench_fifo_c2s", "/tmp/ipc_bench_fifo_s2c",
		iters, msgSize)
}

func benchFifoZC(iters, msgSize int) (time.Duration, error) {
	return benchFifoImpl("./peer_fifo_zc",
		"/tmp/ipc_bench_fifo_zc_c2s", "/tmp/ipc_bench_fifo_zc_s2c",
		iters, msgSize)
}

func benchFifoImpl(bin, c2s, s2c string, iters, msgSize int) (time.Duration, error) {
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

	cmd := exec.Command(bin, strconv.Itoa(iters), strconv.Itoa(msgSize))
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

// ---------------------- tcp / tcp_zc ----------------------

func benchTCP(iters, msgSize int) (time.Duration, error) {
	return benchTCPImpl("./peer_tcp", iters, msgSize)
}

func benchTCPZC(iters, msgSize int) (time.Duration, error) {
	return benchTCPImpl("./peer_tcp_zc", iters, msgSize)
}

func benchTCPImpl(bin string, iters, msgSize int) (time.Duration, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	cmd := exec.Command(bin, strconv.Itoa(iters), strconv.Itoa(port), strconv.Itoa(msgSize))
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
	return pingPong(conn, cmd, iters, msgSize)
}

// ---------------------- uds / uds_zc ----------------------

func benchUDS(iters, msgSize int) (time.Duration, error) {
	const path = "/tmp/ipc_bench_uds"
	_ = os.Remove(path)
	ln, err := net.Listen("unix", path)
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	defer os.Remove(path)
	cmd := exec.Command("./peer_uds", strconv.Itoa(iters), strconv.Itoa(msgSize))
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	conn, err := ln.Accept()
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	return pingPong(conn, cmd, iters, msgSize)
}

func benchUDSZC(iters, msgSize int) (time.Duration, error) {
	const path = "/tmp/ipc_bench_uds_zc"
	_ = os.Remove(path)
	ln, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	defer os.Remove(path)

	cmd := exec.Command("./peer_uds_zc", strconv.Itoa(iters), strconv.Itoa(msgSize))
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	conn, err := ln.AcceptUnix()
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Первый recvmsg — получаем fd memfd'а от peer'а.
	buf := make([]byte, 1)
	oob := make([]byte, syscall.CmsgSpace(4))
	_, oobn, _, _, err := conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return 0, err
	}
	scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return 0, err
	}
	fds, err := syscall.ParseUnixRights(&scms[0])
	if err != nil {
		return 0, err
	}
	fd := fds[0]
	defer syscall.Close(fd)

	total := msgSize * 2
	shared, err := syscall.Mmap(fd, 0, total,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return 0, err
	}
	defer syscall.Munmap(shared)

	out := shared[:msgSize]       // go → peer
	in := shared[msgSize : 2*msgSize] // peer → go
	msg := make([]byte, msgSize)
	sig := []byte{0}

	start := time.Now()
	for range iters {
		copy(out, msg)
		if _, err := conn.Write(sig); err != nil {
			return 0, err
		}
		if _, err := io.ReadFull(conn, sig); err != nil {
			return 0, err
		}
		copy(msg, in) // считать из shared (noop для бенча, но честно)
	}
	elapsed := time.Since(start)
	return elapsed, cmd.Wait()
}

// ---------------------- mq ----------------------

func benchMQ(iters, msgSize int) (time.Duration, error) {
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

	qC2S := C.mq_create(cc, 10, C.long(msgSize), 0)
	qS2C := C.mq_create(cs, 10, C.long(msgSize), 1)
	if C.mq_is_err(qC2S) != 0 || C.mq_is_err(qS2C) != 0 {
		return 0, errors.New("mq_create failed")
	}
	defer C.mq_unlink(cc)
	defer C.mq_unlink(cs)
	defer C.mq_close(qC2S)
	defer C.mq_close(qS2C)

	cmd := exec.Command("./peer_mq", strconv.Itoa(iters), strconv.Itoa(msgSize))
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

func benchShmSem(iters, msgSize int) (time.Duration, error) {
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

	cmd := exec.Command("./peer_shmsem", strconv.Itoa(iters), strconv.Itoa(msgSize))
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

// ---------------------- helpers ----------------------

func pingPong(conn net.Conn, cmd *exec.Cmd, iters, msgSize int) (time.Duration, error) {
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

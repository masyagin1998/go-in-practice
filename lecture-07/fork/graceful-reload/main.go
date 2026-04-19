package main

// graceful-reload — zero-downtime HTTP-релоад.
// По SIGHUP форкаем нового себя, передаём ему слушающий сокет через
// ExtraFiles (fd #3), старый доделывает in-flight и выходит.
// Паттерн из tableflip/caddy/nginx/envoy hot-restart.

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const addr = "127.0.0.1:8080"

func main() {
	ln, inherited, err := obtainListener()
	if err != nil {
		log.Fatalf("listener: %v", err)
	}
	role := "первичный"
	if inherited {
		role = "child после SIGHUP"
	}
	log.Printf("[PID %d] %s, слушаем %s", os.Getpid(), role, addr)
	log.Printf("Запрос: curl %s/", addr)
	log.Printf("Релоад:  kill -HUP %d", os.Getpid())

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hi from PID %d\n", os.Getpid())
		}),
	}

	var child atomic.Pointer[exec.Cmd]
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		<-hup
		log.Printf("[PID %d] SIGHUP: spawning child", os.Getpid())
		c, err := spawnChild(ln)
		if err != nil {
			log.Printf("spawn: %v", err)
			return
		}
		child.Store(c)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Printf("serve: %v", err)
	}
	if c := child.Load(); c != nil {
		_ = c.Wait()
	}
	log.Printf("[PID %d] завершение", os.Getpid())
}

func obtainListener() (net.Listener, bool, error) {
	if os.Getenv("GRACEFUL_INHERIT") == "1" {
		f := os.NewFile(3, "inherited-listener")
		ln, err := net.FileListener(f)
		f.Close()
		return ln, true, err
	}
	ln, err := net.Listen("tcp", addr)
	return ln, false, err
}

func spawnChild(ln net.Listener) (*exec.Cmd, error) {
	tcp, ok := ln.(*net.TCPListener)
	if !ok {
		return nil, fmt.Errorf("не *net.TCPListener: %T", ln)
	}
	f, err := tcp.File()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(self)
	cmd.Env = append(os.Environ(), "GRACEFUL_INHERIT=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{f} // fd #3 в дочке
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

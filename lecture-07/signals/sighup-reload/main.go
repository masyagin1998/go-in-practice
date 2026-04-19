package main

// sighup-reload — перечитать конфиг по SIGHUP.
// Паттерн nginx/postgres/sshd/haproxy: атомарная подмена указателя на
// конфиг, никаких локов на hot-path читателей.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const configPath = "/tmp/sighup_demo.conf"

type config struct {
	LogLevel string `json:"log_level"`
	Workers  int    `json:"workers"`
}

var current atomic.Pointer[config]

func main() {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		writeConfig(config{LogLevel: "info", Workers: 2})
	}
	if err := load(); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-hup:
				if err := load(); err != nil {
					log.Printf("reload: %v", err)
					continue
				}
				fmt.Printf("[reload] %+v\n", *current.Load())
			case <-ctx.Done():
				return
			}
		}
	}()

	pid := os.Getpid()
	fmt.Printf("PID=%d, конфиг %s\n", pid, configPath)
	fmt.Printf("Правьте файл и шлите: kill -HUP %d\n", pid)
	fmt.Printf("Выход: Ctrl+C\n\n")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c := current.Load()
			fmt.Printf("[tick] log_level=%s workers=%d\n", c.LogLevel, c.Workers)
		case <-ctx.Done():
			fmt.Println("\n[main] завершаемся")
			return
		}
	}
}

func load() error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var c config
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	current.Store(&c)
	return nil
}

func writeConfig(c config) {
	data, _ := json.Marshal(c)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		log.Fatal(err)
	}
}

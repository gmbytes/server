package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	robot "server/internal/service/robot"
)

func main() {
	wsURL := flag.String("url", "ws://127.0.0.1:8080/ws", "Gate WebSocket URL")
	count := flag.Int("n", 10, "Number of robots to launch")
	interval := flag.Duration("interval", 100*time.Millisecond, "Interval between robot launches")
	verbose := flag.Bool("v", false, "Verbose per-robot logging")
	autoStart := flag.Bool("auto", true, "Auto-start robots on launch")
	flag.Parse()

	mgr := robot.NewRobotMgr(*wsURL, *verbose)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n[Main] shutting down...")
		mgr.StopAll()
		os.Exit(0)
	}()

	if *autoStart {
		mgr.Start(*count, *interval)
	}

	go func() {
		for {
			time.Sleep(15 * time.Second)
			if mgr.AliveCount() > 0 {
				mgr.PrintStats()
			}
		}
	}()

	printHelp()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "add":
			n := 10
			if len(parts) > 1 {
				if v, err := strconv.Atoi(parts[1]); err == nil && v > 0 {
					n = v
				}
			}
			dur := *interval
			if len(parts) > 2 {
				if d, err := time.ParseDuration(parts[2]); err == nil {
					dur = d
				}
			}
			mgr.Start(n, dur)

		case "stop":
			if len(parts) > 1 && parts[1] != "all" {
				if v, err := strconv.Atoi(parts[1]); err == nil && v > 0 {
					mgr.StopN(v)
				}
			} else {
				mgr.StopAll()
			}

		case "stats", "s":
			mgr.PrintStats()

		case "help", "h":
			printHelp()

		case "quit", "q", "exit":
			fmt.Println("[Main] shutting down...")
			mgr.StopAll()
			time.Sleep(500 * time.Millisecond)
			return

		default:
			fmt.Printf("unknown command: %s (type 'help' for usage)\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Print(`
Commands:
  add [N] [interval]  - Add N robots (default 10, interval e.g. 50ms)
  stop [N|all]        - Stop N robots or all
  stats / s           - Print current statistics
  help / h            - Show this help
  quit / q / exit     - Shutdown all robots and exit
`)
}

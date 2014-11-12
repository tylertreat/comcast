package throttler

import (
	"fmt"
	"os"
	"runtime"
)

const (
	Start = "start"
	stop  = "stop"
	any   = "any"
)

type config struct {
	latency    int
	bandwidth  int
	packetLoss float64
}

type throttler interface {
	setup(*config) error
	teardown() error
	exists() bool
	check() string
}

func setup(throttler throttler, config *config) {
	if throttler.exists() {
		fmt.Println("It looks like the packet rules are already setup")
		os.Exit(1)
	}

	if err := throttler.setup(config); err != nil {
		fmt.Println("I couldn't setup the packet rules")
		os.Exit(1)
	}

	fmt.Println("Packet rules setup...")
	fmt.Printf("Run `%s` to double check\n", throttler.check())
	fmt.Printf("Run `%s --mode %s` to reset\n", os.Args[0], stop)
}

func teardown(throttler throttler) {
	if !throttler.exists() {
		fmt.Println("It looks like the packet rules aren't setup")
		os.Exit(1)
	}

	if err := throttler.teardown(); err != nil {
		fmt.Println("Failed to stop packet controls")
		os.Exit(1)
	}

	fmt.Println("Packet rules stopped...")
	fmt.Printf("Run `%s` to double check\n", throttler.check())
	fmt.Printf("Run `%s --mode %s` to start\n", os.Args[0], Start)
}

func Run(mode string, latency, bandwidth int, packetLoss float64) {
	config := &config{
		latency:    latency,
		bandwidth:  bandwidth,
		packetLoss: packetLoss,
	}

	var throttler throttler
	switch runtime.GOOS {
	case "darwin":
		throttler = &ipfwThrottler{}
	case "linux":
		throttler = &tcThrottler{}
	default:
		fmt.Printf("I don't support your OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	switch mode {
	case Start:
		setup(throttler, config)
	case stop:
		teardown(throttler)
	default:
		fmt.Printf("I don't know what this mode is: %s\n", mode)
		fmt.Printf("Try '%s' or '%s'\n", Start, stop)
		os.Exit(1)
	}
}

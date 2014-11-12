package wiretap

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

type Config struct {
	Latency    int
	Bandwidth  int
	PacketLoss float64
}

type Wiretap interface {
	Setup(*Config) error
	Teardown() error
	Exists() bool
	Check() string
}

func setup(tap Wiretap, config *Config) {
	if tap.Exists() {
		fmt.Println("It looks like the packet rules are already setup")
		os.Exit(1)
	}

	if err := tap.Setup(config); err != nil {
		fmt.Println("I couldn't setup the packet rules")
		os.Exit(1)
	}

	fmt.Println("Packet rules setup...")
	fmt.Printf("Run `%s` to double check\n", tap.Check())
	fmt.Printf("Run `%s --mode %s` to reset\n", os.Args[0], stop)
}

func teardown(tap Wiretap) {
	if !tap.Exists() {
		fmt.Println("It looks like the packet rules aren't setup")
		os.Exit(1)
	}

	if err := tap.Teardown(); err != nil {
		fmt.Println("Failed to stop packet controls")
		os.Exit(1)
	}

	fmt.Println("Packet rules stopped...")
	fmt.Printf("Run `%s` to double check\n", tap.Check())
	fmt.Printf("Run `%s --mode %s` to start\n", os.Args[0], Start)
}

func Run(mode string, latency, bandwidth int, packetLoss float64) {
	config := &Config{
		Latency:    latency,
		Bandwidth:  bandwidth,
		PacketLoss: packetLoss,
	}

	var tap Wiretap
	switch runtime.GOOS {
	case "darwin":
		//tap = &DarwinWiretap{}
		tap = &LinuxWiretap{}
	case "linux":
		tap = &LinuxWiretap{}
	default:
		fmt.Printf("I don't support your OS: %s", runtime.GOOS)
		os.Exit(1)
	}

	switch mode {
	case Start:
		setup(tap, config)
	case stop:
		teardown(tap)
	default:
		fmt.Println("I don't know what this mode is: %s", mode)
		os.Exit(1)
	}
}

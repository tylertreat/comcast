package throttler

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	Start           = "start"
	stop            = "stop"
	any             = "any"
	linux           = "linux"
	darwin          = "darwin"
	freebsd         = "freebsd"
	checkOSXVersion = "sw_vers -productVersion"
)

// Config specifies options for configuring packet filter rules.
type Config struct {
	Mode       string
	Latency    int
	Bandwidth  int
	PacketLoss float64
}

type throttler interface {
	setup(*Config) error
	teardown() error
	exists() bool
	check() string
}

func setup(throttler throttler, config *Config) {
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

// Run executes the packet filter operation, either setting it up or tearing
// it down.
func Run(config *Config) {

	var throttler throttler
	switch runtime.GOOS {
	case darwin, freebsd:
		if runtime.GOOS == darwin && !osxVersionSupported() {
			// ipfw was removed in OSX 10.10 in favor of pfctl.
			// TODO: add support for pfctl.
			fmt.Println("I don't support your version of OSX")
			os.Exit(1)
		}
		throttler = &ipfwThrottler{}
	case linux:
		throttler = &tcThrottler{}
	default:
		fmt.Printf("I don't support your OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	switch config.Mode {
	case Start:
		setup(throttler, config)
	case stop:
		teardown(throttler)
	default:
		fmt.Printf("I don't know what this mode is: %s\n", config.Mode)
		fmt.Printf("Try '%s' or '%s'\n", Start, stop)
		os.Exit(1)
	}
}

func osxVersionSupported() bool {
	v, err := exec.Command("/bin/sh", "-c", checkOSXVersion).Output()
	if err != nil {
		return false
	}
	return !strings.HasPrefix(string(v), "10.10")
}

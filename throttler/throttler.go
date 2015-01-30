package throttler

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// Start is the mode to setup packet filter rules.
	Start           = "start"
	stop            = "stop"
	any             = "any"
	linux           = "linux"
	darwin          = "darwin"
	freebsd         = "freebsd"
	windows         = "windows"
	checkOSXVersion = "sw_vers -productVersion"
)

// Config specifies options for configuring packet filter rules.
type Config struct {
	Device           string
	Mode             string
	Latency          int
	TargetBandwidth  int
	DefaultBandwidth int
	PacketLoss       float64
	TargetIps        []string
	TargetPorts      []string
	TargetProtos     []string
	DryRun           bool
}

type throttler interface {
	setup(*Config) error
	teardown(*Config) error
	exists(*Config) bool
	check() string
}

func setup(t throttler, c *Config) {
	if t.exists(c) {
		log.Fatalln("It looks like the packet rules are already setup")
	}

	if err := t.setup(c); err != nil {
		log.Fatalln("I couldn't setup the packet rules")
	}

	log.Println("Packet rules setup...")
	log.Printf("Run `%s` to double check\n", t.check())
	log.Printf("Run `%s --mode %s` to reset\n", os.Args[0], stop)
}

func teardown(t throttler, c *Config) {
	if !t.exists(c) {
		log.Fatalln("It looks like the packet rules aren't setup")
	}

	if err := t.teardown(c); err != nil {
		log.Fatalln("Failed to stop packet controls")
	}

	log.Println("Packet rules stopped...")
	log.Printf("Run `%s` to double check\n", t.check())
	log.Printf("Run `%s --mode %s` to start\n", os.Args[0], Start)
}

// Run executes the packet filter operation, either setting it up or tearing
// it down.
func Run(c *Config) {
	var t throttler
	switch runtime.GOOS {
	case freebsd:
		if c.Device == "" {
			log.Fatalln("Device not specified, unable to default to eth0 on FreeBSD.")
		}

		t = &ipfwThrottler{}
	case darwin:
		if runtime.GOOS == darwin && !osxVersionSupported() {
			// ipfw was removed in OSX 10.10 in favor of pfctl.
			log.Fatalln("I don't support your version of OSX")

			// TODO: add support for pfctl.
			//t = &pfctlThrottler{}
		}

		if c.Device == "" {
			c.Device = "eth0"
		}

		t = &ipfwThrottler{}
	case linux:
		if c.Device == "" {
			c.Device = "eth0"
		}

		t = &tcThrottler{}
	case windows:
		log.Fatalln("I don't support your OS: %s\n", runtime.GOOS)
		//log.Fatalln("If you want to use Comcast on Windows, please install wipfw.")
		//t = &wipfwThrottler{}
	default:
		log.Fatalln("I don't support your OS: %s\n", runtime.GOOS)
	}

	switch c.Mode {
	case Start:
		setup(t, c)
	case stop:
		teardown(t, c)
	default:
		log.Printf("I don't know what this mode is: %s\n", c.Mode)
		log.Fatalf("Try %q or %q\n", Start, stop)
	}
}

func osxVersionSupported() bool {
	v, err := exec.Command("/bin/sh", "-c", checkOSXVersion).Output()
	if err != nil {
		return false
	}
	return !strings.HasPrefix(string(v), "10.10")
}

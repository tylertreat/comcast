package throttler

import (
	"bufio"
	"errors"
	"fmt"
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
	exists() bool
	check() string
}

type commander interface {
	execute(string) error
	executeGetLines(string) ([]string, error)
}

type dryRunCommander struct {}

type shellCommander struct {}

var dry bool

func setup(t throttler, cfg *Config) {
	if t.exists() {
		log.Fatalln("It looks like the packet rules are already setup")
	}

	if err := t.setup(cfg); err != nil {
		log.Fatalln("I couldn't setup the packet rules")
	}

	log.Println("Packet rules setup...")
	log.Printf("Run `%s` to double check\n", t.check())
	log.Printf("Run `%s --mode %s` to reset\n", os.Args[0], stop)
}

func teardown(t throttler, cfg *Config) {
	if !t.exists() {
		log.Fatalln("It looks like the packet rules aren't setup")
	}

	if err := t.teardown(cfg); err != nil {
		log.Fatalln("Failed to stop packet controls")
	}

	log.Println("Packet rules stopped...")
	log.Printf("Run `%s` to double check\n", t.check())
	log.Printf("Run `%s --mode %s` to start\n", os.Args[0], Start)
}

// Run executes the packet filter operation, either setting it up or tearing
// it down.
func Run(cfg *Config) {
	dry = cfg.DryRun
	var t throttler
	var c commander

	if cfg.DryRun {
		c = &dryRunCommander{}
	} else {
		c = &shellCommander{}
	}

	switch runtime.GOOS {
	case freebsd:
		if cfg.Device == "" {
			log.Fatalln("Device not specified, unable to default to eth0 on FreeBSD.")
		}

		t = &ipfwThrottler{c}
	case darwin:
		if runtime.GOOS == darwin && !osxVersionSupported() {
			// ipfw was removed in OSX 10.10 in favor of pfctl.
			log.Fatalln("I don't support your version of OSX")

			// TODO: add support for pfctl.
			//t = &pfctlThrottler{}
		}

		if cfg.Device == "" {
			cfg.Device = "eth0"
		}

		t = &ipfwThrottler{c}
	case linux:
		if cfg.Device == "" {
			cfg.Device = "eth0"
		}

		t = &tcThrottler{c}
	case windows:
		log.Fatalln("I don't support your OS: %s\n", runtime.GOOS)
		//log.Fatalln("If you want to use Comcast on Windows, please install wipfw.")
		//t = &wipfwThrottler{}
	default:
		log.Fatalln("I don't support your OS: %s\n", runtime.GOOS)
	}

	switch cfg.Mode {
	case Start:
		setup(t, cfg)
	case stop:
		teardown(t, cfg)
	default:
		log.Printf("I don't know what this mode is: %s\n", cfg.Mode)
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

func (c *dryRunCommander) execute(cmd string) error {
	fmt.Println(cmd)
	return nil
}

func (c *dryRunCommander) executeGetLines(cmd string) ([]string, error) {
	fmt.Println(cmd)
	return []string{}, nil
}

func (c *shellCommander) execute(cmd string) error {
	fmt.Println(cmd)
	return exec.Command("/bin/sh", "-c", cmd).Run()
}

func (c *shellCommander) executeGetLines(cmd string) ([]string, error) {
	lines := []string{}
	child := exec.Command("/bin/sh", "-c", cmd)

	out, err := child.StdoutPipe()
	if err != nil {
		return []string{}, err
	}

	err = child.Start()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []string{}, errors.New(fmt.Sprint("Error reading standard input:", err))
	}

	err = child.Wait()
	if err != nil {
		return []string{}, err
	}

	return lines, nil
}

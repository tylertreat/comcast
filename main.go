package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/tylertreat/wiretap/wiretap"
)

const (
	start = "start"
	stop  = "stop"
	any   = "any"
)

func setup(cutter wiretap.Wiretap, config *wiretap.Config) {
	if cutter.Exists() {
		fmt.Println("It looks like the packet rules are already setup")
		os.Exit(1)
	}

	if err := cutter.Setup(config); err != nil {
		fmt.Println("I couldn't setup the packet rules")
		os.Exit(1)
	}

	fmt.Println("Packet rules setup...")
	fmt.Printf("Run `%s` to double check\n", cutter.Check())
	fmt.Printf("Run `%s --mode %s` to reset\n", os.Args[0], stop)
}

func teardown(cutter wiretap.Wiretap) {
	if !cutter.Exists() {
		fmt.Println("It looks like the packet rules aren't setup")
		os.Exit(1)
	}

	if err := cutter.Teardown(); err != nil {
		fmt.Println("Failed to stop packet controls")
		os.Exit(1)
	}

	fmt.Println("Packet rules stopped...")
	fmt.Printf("Run `%s` to double check\n", cutter.Check())
	fmt.Printf("Run `%s --mode %s` to start\n", os.Args[0], start)
}

func main() {
	mode := flag.String("mode", start, "start or stop packet controls")
	host := flag.String("host", any, "remote host to apply rules to")
	latency := flag.Int("latency", -1, "latency to add in ms")
	bandwidth := flag.Int("bandwidth", -1, "bandwidth limit in kb/s")
	packetLoss := flag.Float64("packet-loss", 0, "packet-loss rate")
	flag.Parse()

	config := &wiretap.Config{
		Host:       *host,
		Latency:    *latency,
		Bandwidth:  *bandwidth,
		PacketLoss: *packetLoss,
	}

	var cutter wiretap.Wiretap
	switch runtime.GOOS {
	case "darwin":
		cutter = &wiretap.DarwinWiretap{}
	default:
		fmt.Printf("I don't support your OS: %s", runtime.GOOS)
		os.Exit(1)
	}

	switch *mode {
	case start:
		setup(cutter, config)
	case stop:
		teardown(cutter)
	default:
		fmt.Println("I don't know what this mode is: %s", *mode)
		os.Exit(1)
	}
}

package main

import (
	"flag"

	"github.com/tylertreat/comcast/throttler"
)

func main() {
	// TODO: add support for specific host/port.
	// Also support for other options like packet reordering, duplication, etc.
	var (
		device     = flag.String("device", "eth0", "interface (device) to use")
		mode       = flag.String("mode", throttler.Start, "start or stop packet controls")
		latency    = flag.Int("latency", -1, "latency to add in ms")
		bandwidth  = flag.Int("bandwidth", -1, "bandwidth limit in kb/s")
		packetLoss = flag.Float64("packet-loss", 0, "packet-loss rate")
	)
	flag.Parse()

	throttler.Run(&throttler.Config{
		Device:     *device,
		Mode:       *mode,
		Latency:    *latency,
		Bandwidth:  *bandwidth,
		PacketLoss: *packetLoss,
	})
}

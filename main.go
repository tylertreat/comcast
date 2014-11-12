package main

import (
	"flag"

	"github.com/tylertreat/Comcast/throttler"
)

func main() {
	mode := flag.String("mode", throttler.Start, "start or stop packet controls")
	latency := flag.Int("latency", -1, "latency to add in ms")
	bandwidth := flag.Int("bandwidth", -1, "bandwidth limit in kb/s")
	packetLoss := flag.Float64("packet-loss", 0, "packet-loss rate")
	flag.Parse()

	throttler.Run(*mode, *latency, *bandwidth, *packetLoss)
}

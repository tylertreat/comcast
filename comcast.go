package main

import (
	"flag"
	"github.com/tylertreat/Comcast/throttler"
	"log"
	"net"
	"strconv"
	"strings"
)

func main() {
	// TODO: Add support for other options like packet reordering, duplication, etc.
	var (
		device      = flag.String("device", "", "interface (device) to use")
		mode        = flag.String("mode", throttler.Start, "start or stop packet controls")
		latency     = flag.Int("latency", -1, "latency to add in ms")
		targetbw    = flag.Int("target-bw", -1, "target bandwidth limit in kb/s (slow-lane)")
		defaultbw   = flag.Int("default-bw", -1, "default bandwidth limit in kb/s (fast-lane)")
		packetLoss  = flag.Float64("packet-loss", 0, "packet-loss rate")
		targetaddr  = flag.String("target-addr", "", "target addresses, (eg: 10.0.0.1 or 10.0.0.0/24 or 10.0.0.1,192.168.0.0/24)")
		targetport  = flag.String("target-port", "", "target port(s) (eg: 80 or 1:65535 or 22,80,443,1000:1010)")
		targetproto = flag.String("target-proto", "", "target protocol TCP/UDP (eg: tcp or tcp,udp or icmp)")
		dryrun      = flag.Bool("dry-run", false, "specifies whether or not to actually commit the rule changes")
		//icmptype    = flag.String("icmp-type", "", "icmp message type (eg: reply or reply,request)") //TODO: Maybe later :3
	)
	flag.Parse()

	throttler.Run(&throttler.Config{
		Device:           *device,
		Mode:             *mode,
		Latency:          *latency,
		TargetBandwidth:  *targetbw,
		DefaultBandwidth: *defaultbw,
		PacketLoss:       *packetLoss,
		TargetIps:        parseAddrs(*targetaddr),
		TargetPorts:      parsePorts(*targetport),
		TargetProtos:     parseProtos(*targetproto),
		DryRun:           *dryrun,
	})
}

func parseAddrs(addrs string) []string {
	adrs := strings.Split(addrs, ",")
	parsed := []string{}

	if addrs != "" {
		for _, adr := range adrs {
			ip := net.ParseIP(adr)
			if ip != nil {
				parsed = append(parsed, adr)
			} else { //Not a valid single IP, could it be a CIDR?
				_, net, err := net.ParseCIDR(adr)
				if err == nil {
					parsed = append(parsed, net.String())
				} else {
					log.Fatalln("Incorrectly specified target IP or CIDR", adr)
				}
			}
		}
	}

	return parsed
}

func parsePorts(ports string) []string {
	prts := strings.Split(ports, ",")
	parsed := []string{}

	if ports != "" {
		for _, prt := range prts {
			if strings.Contains(prt, ":") {
				if validRange(prt) {
					parsed = append(parsed, prt)
				} else {
					log.Fatalln("Incorrectly specified port range:", prt)
				}
			} else { //Isn't a range, check if just a single port
				if validPort(prt) {
					parsed = append(parsed, prt)
				} else {
					log.Fatalln("Incorrectly specified port:", prt)
				}
			}
		}
	}

	return parsed
}

func parsePort(port string) int {
	prt, err := strconv.Atoi(port)
	if err != nil {
		return 0
	}

	return prt
}

func validPort(port string) bool {
	prt := parsePort(port)
	return prt > 0 && prt < 65536
}

func validRange(ports string) bool {
	pr := strings.Split(ports, ":")

	if len(pr) == 2 {
		if !validPort(pr[0]) || !validPort(pr[1]) {
			return false
		}

		if portHigher(pr[0], pr[1]) {
			return false
		}
	} else {
		return false
	}

	return true
}

func portHigher(prt1, prt2 string) bool {
	p1 := parsePort(prt1)
	p2 := parsePort(prt2)

	return p1 > p2
}

func parseProtos(protos string) []string {
	ptcs := strings.Split(protos, ",")
	parsed := []string{}

	if protos != "" {
		for _, ptc := range ptcs {
			p := strings.ToLower(ptc)
			if p == "udp" ||
				p == "tcp" ||
				p == "icmp" {
				parsed = append(parsed, p)
			} else {
				log.Fatalln("Incorrectly specified protocol:", p)
			}
		}
	}

	return parsed
}

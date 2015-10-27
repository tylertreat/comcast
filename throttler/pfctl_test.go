package throttler

import (
	"testing"
)

func TestPfctlDefaultConfigCommand(t *testing.T) {

	r := newCmdRecorder()
	th := &pfctlThrottler{r}
	c := defaultTestConfig
	c.PacketLoss = 0
	c.TargetIps = []string{}
	c.TargetIps6 = []string{}
	c.TargetBandwidth = -1
	c.TargetPorts = []string{}
	c.TargetProtos = []string{"tcp,udp,icmp"}

	th.setup(&c)
	r.verifyCommands(t, []string{
		"sudo pfctl -E",
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,
		`sudo dnctl pipe 1 config mask  proto tcp,udp,icmp`,
	})
}

func TestPfctlThrottleOnlyConfigCommand(t *testing.T) {

	var c = Config{
		Device:           "eth0",
		Stop:             false,
		Latency:          -1,
		TargetBandwidth:  -1,
		DefaultBandwidth: 20000,
		PacketLoss:       0.1,
	}
	r := newCmdRecorder()
	th := &pfctlThrottler{r}

	th.setup(&c)
	r.verifyCommands(t, []string{
		"sudo pfctl -E",
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,
		`sudo dnctl pipe 1 config plr 0.0010`,
	})
}
func TestPfctlNoIPThrottleConfigCommand(t *testing.T) {

	var c = Config{
		Device:           "eth0",
		Stop:             false,
		Latency:          -1,
		TargetBandwidth:  -1,
		DefaultBandwidth: 20000,
		PacketLoss:       0.1,
		TargetProtos:     []string{"tcp"},
	}
	r := newCmdRecorder()
	th := &pfctlThrottler{r}

	th.setup(&c)
	r.verifyCommands(t, []string{
		"sudo pfctl -E",
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,
		`sudo dnctl pipe 1 config plr 0.0010 mask  proto tcp`,
	})
}

func TestPfctlPacketSetup(t *testing.T) {

	r := newCmdRecorder()
	th := &pfctlThrottler{r}
	c := defaultTestConfig
	c.PacketLoss = 0.5

	th.setup(&c)
	r.verifyCommands(t, []string{
		"sudo pfctl -E",
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 src-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 dst-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 src-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 dst-ip 10.10.10.10 proto tcp`,
	})
}

func TestPfctlProtoSetup(t *testing.T) {

	r := newCmdRecorder()
	th := &pfctlThrottler{r}
	c := defaultTestConfig
	c.PacketLoss = 0.5
	c.TargetProtos = []string{"tcp", "udp", "icmp"}

	th.setup(&c)
	r.verifyCommands(t, []string{
		"sudo pfctl -E",
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 src-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 src-ip 10.10.10.10 proto udp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 src-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 dst-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 dst-ip 10.10.10.10 proto udp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  dst-port 80 dst-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 src-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 src-ip 10.10.10.10 proto udp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 src-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 dst-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 dst-ip 10.10.10.10 proto udp`,
		`sudo dnctl pipe 1 config plr 0.0050 mask  src-port 80 dst-ip 10.10.10.10 proto icmp`,
	})
}

func TestPfctlMultiplePortsAndIps(t *testing.T) {
	r := newCmdRecorder()
	th := &pfctlThrottler{r}
	cfg := defaultTestConfig
	cfg.TargetIps = []string{"1.1.1.1", "2.2.2.2"}
	cfg.TargetPorts = []string{"80", "8080"}
	cfg.TargetProtos = []string{"tcp"}
	th.setup(&cfg)
	r.verifyCommands(t, []string{
		"sudo pfctl -E",
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 80 src-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 80 dst-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 80 src-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 80 dst-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 80 src-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 80 dst-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 80 src-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 80 dst-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 8080 src-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 8080 dst-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 8080 src-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  dst-port 8080 dst-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 8080 src-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 8080 dst-ip 1.1.1.1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 8080 src-ip 2.2.2.2 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0010 mask  src-port 8080 dst-ip 2.2.2.2 proto tcp`,
	})
}

func TestPfctlMixedIPv6Setup(t *testing.T) {
	r := newCmdRecorder()
	th := &pfctlThrottler{r}
	cfg := defaultTestConfig
	cfg.TargetProtos = []string{"icmp", "tcp"}
	cfg.PacketLoss = 0.2
	cfg.TargetIps6 = []string{"2001:db8::1"}
	th.setup(&cfg)
	r.verifyCommands(t, []string{
		`sudo pfctl -E`,
		`(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`,
		`echo $'dummynet in all pipe 1' | sudo pfctl -a mop -f - `,

		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 src-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 src-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 dst-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 dst-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 src-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 src-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 dst-ip 10.10.10.10 proto icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 dst-ip 10.10.10.10 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 src-ip6 2001:db8::1 proto ipv6-icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 src-ip6 2001:db8::1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 dst-ip6 2001:db8::1 proto ipv6-icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  dst-port 80 dst-ip6 2001:db8::1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 src-ip6 2001:db8::1 proto ipv6-icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 src-ip6 2001:db8::1 proto tcp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 dst-ip6 2001:db8::1 proto ipv6-icmp`,
		`sudo dnctl pipe 1 config plr 0.0020 mask  src-port 80 dst-ip6 2001:db8::1 proto tcp`,
	})
}

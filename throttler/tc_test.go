package throttler

import (
	"testing"
)

type cmdRecorder struct {
	commands     []string
	responses    map[string][]string
	cmdBlackList []string
}

func newCmdRecorder() *cmdRecorder {
	return &cmdRecorder{[]string{}, map[string][]string{}, []string{}}
}

func (r *cmdRecorder) execute(cmd string) error {
	r.commands = append(r.commands, cmd)
	return nil
}

func (r *cmdRecorder) executeGetLines(cmd string) ([]string, error) {
	r.execute(cmd)
	if responses, found := r.responses[cmd]; found {
		return responses, nil
	}
	return []string{}, nil
}

func (r *cmdRecorder) commandExists(cmd string) bool {
	for _, blackListed := range r.cmdBlackList {
		if blackListed == cmd {
			return false
		}
	}
	return true
}

func (r *cmdRecorder) verifyCommands(t *testing.T, expected []string) {
	if len(expected) != len(r.commands) {
		for i, cmd := range expected {
			t.Logf("Expected (%d): %s", i, cmd)
		}
		for i, cmd := range r.commands {
			t.Logf("Actual   (%d): %s", i, cmd)
		}

		t.Fatalf("Expected to see %d commands, got %d", len(expected), len(r.commands))
	}

	for i, cmd := range expected {
		if actual := r.commands[i]; actual != cmd {
			t.Fatalf("Expected to see command `%s`, got `%s`", i, cmd, actual)
		}
	}
}

var defaultTestConfig = Config{
	Device:           "eth0",
	Stop:             false,
	Latency:          -1,
	TargetBandwidth:  -1,
	DefaultBandwidth: 20000,
	PacketLoss:       0.1,
	TargetIps:        []string{"10.10.10.10"},
	TargetPorts:      []string{"80"},
	TargetProtos:     []string{"tcp"},
	DryRun:           false,
}

func TestTcPacketLossSetup(t *testing.T) {
	r := newCmdRecorder()
	th := &tcThrottler{r}
	cfg := defaultTestConfig
	cfg.Device = "eth1"
	cfg.PacketLoss = 0.2
	th.setup(&cfg)
	r.verifyCommands(t, []string{
		"sudo tc qdisc add dev eth1 handle 10: root htb",
		"sudo tc class add dev eth1 parent 10: classid 10:1 htb rate 20000kbit",
		"sudo tc class add dev eth1 parent 10:1 classid 10:10 htb rate 20000kbit",
		"sudo tc qdisc add dev eth1 parent 10:10 handle 100: netem loss 0.20%",
		"sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p tcp --dport 80 -d 10.10.10.10",
	})
}

func TestTcMultiplePortsAndIps(t *testing.T) {
	r := newCmdRecorder()
	th := &tcThrottler{r}
	cfg := defaultTestConfig
	cfg.TargetIps = []string{"1.1.1.1", "2.2.2.2"}
	cfg.TargetPorts = []string{"80", "8080"}
	cfg.TargetProtos = []string{"tcp", "udp"}
	th.setup(&cfg)
	r.verifyCommands(t, []string{
		"sudo tc qdisc add dev eth0 handle 10: root htb",
		"sudo tc class add dev eth0 parent 10: classid 10:1 htb rate 20000kbit",
		"sudo tc class add dev eth0 parent 10:1 classid 10:10 htb rate 20000kbit",
		"sudo tc qdisc add dev eth0 parent 10:10 handle 100: netem loss 0.10%",
		"sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p tcp --match multiport --dports 80,8080 -d 1.1.1.1",
		"sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p udp --match multiport --dports 80,8080 -d 1.1.1.1",
		"sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p tcp --match multiport --dports 80,8080 -d 2.2.2.2",
		"sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p udp --match multiport --dports 80,8080 -d 2.2.2.2",
	})
}

func TestTcMixedIPv6Setup(t *testing.T) {
	r := newCmdRecorder()
	th := &tcThrottler{r}
	cfg := defaultTestConfig
	cfg.Device = "eth1"
	cfg.PacketLoss = 0.2
	cfg.TargetIps6 = []string{"2001:db8::1"}
	th.setup(&cfg)
	r.verifyCommands(t, []string{
		"sudo tc qdisc add dev eth1 handle 10: root htb",
		"sudo tc class add dev eth1 parent 10: classid 10:1 htb rate 20000kbit",
		"sudo tc class add dev eth1 parent 10:1 classid 10:10 htb rate 20000kbit",
		"sudo tc qdisc add dev eth1 parent 10:10 handle 100: netem loss 0.20%",
		"sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p tcp --dport 80 -d 10.10.10.10",
		"sudo ip6tables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10 -p tcp --dport 80 -d 2001:db8::1",
	})
}

func TestTcTeardown(t *testing.T) {
	r := newCmdRecorder()
	th := &tcThrottler{r}
	r.responses = map[string][]string{
		"sudo iptables -S -t mangle": {
			"-P PREROUTING ACCEPT",
			"-P INPUT ACCEPT",
			"-P FORWARD ACCEPT",
			"-P OUTPUT ACCEPT",
			"-P POSTROUTING ACCEPT",
			"-A POSTROUTING -d 10.10.10.10 -p tcp -m tcp --dport 80 -j CLASSIFY --set-class 0010:0010",
		},
		"sudo ip6tables -S -t mangle": {
			"-P PREROUTING ACCEPT",
			"-P INPUT ACCEPT",
			"-P FORWARD ACCEPT",
			"-P OUTPUT ACCEPT",
			"-P POSTROUTING ACCEPT",
		},
	}
	th.teardown(&defaultTestConfig)
	r.verifyCommands(t, []string{
		"sudo iptables -S -t mangle",
		"sudo iptables -t mangle -D POSTROUTING -d 10.10.10.10 -p tcp -m tcp --dport 80 -j CLASSIFY --set-class 0010:0010",
		"sudo ip6tables -S -t mangle",
		"sudo tc qdisc del dev eth0 handle 10: root",
	})
}

func TestTcTeardownNoIpTables(t *testing.T) {
	r := newCmdRecorder()
	th := &tcThrottler{r}
	th.teardown(&defaultTestConfig)
	r.verifyCommands(t, []string{
		"sudo iptables -S -t mangle",
		"sudo ip6tables -S -t mangle",
		"sudo tc qdisc del dev eth0 handle 10: root",
	})
}

func TestTcIPv6Teardown(t *testing.T) {
	r := newCmdRecorder()
	th := &tcThrottler{r}
	r.responses = map[string][]string{
		"sudo iptables -S -t mangle": {},
		"sudo ip6tables -S -t mangle": {
			"-P PREROUTING ACCEPT",
			"-P INPUT ACCEPT",
			"-P FORWARD ACCEPT",
			"-P OUTPUT ACCEPT",
			"-P POSTROUTING ACCEPT",
			"-A POSTROUTING -d 2001:db8::1 -p tcp -m tcp --dport 80 -j CLASSIFY --set-class 0010:0010",
		},
	}
	config := defaultTestConfig

	th.teardown(&config)
	r.verifyCommands(t, []string{
		"sudo iptables -S -t mangle",
		"sudo ip6tables -S -t mangle",
		"sudo ip6tables -t mangle -D POSTROUTING -d 2001:db8::1 -p tcp -m tcp --dport 80 -j CLASSIFY --set-class 0010:0010",
		"sudo tc qdisc del dev eth0 handle 10: root",
	})
}

func TestTcTeardownNoIPv6(t *testing.T) {
	r := newCmdRecorder()
	r.cmdBlackList = []string{"ip6tables"}
	th := &tcThrottler{r}
	r.responses = map[string][]string{
		"sudo iptables -S -t mangle": {
			"-P PREROUTING ACCEPT",
			"-P INPUT ACCEPT",
			"-P FORWARD ACCEPT",
			"-P OUTPUT ACCEPT",
			"-P POSTROUTING ACCEPT",
			"-A POSTROUTING -d 10.10.10.10 -p tcp -m tcp --dport 80 -j CLASSIFY --set-class 0010:0010",
		},
	}

	th.teardown(&defaultTestConfig)
	r.verifyCommands(t, []string{
		"sudo iptables -S -t mangle",
		"sudo iptables -t mangle -D POSTROUTING -d 10.10.10.10 -p tcp -m tcp --dport 80 -j CLASSIFY --set-class 0010:0010",
		"sudo tc qdisc del dev eth0 handle 10: root",
	})
}

package throttler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// TODO: use printf in favour of echo due to shell portability issues
	pfctlCreateAnchor    = `(cat /etc/pf.conf && echo "dummynet-anchor \"mop\"" && echo "anchor \"mop\"") | sudo pfctl -f -`
	pfctlTeardown        = `sudo pfctl -f /etc/pf.conf`
	dnctl                = `sudo dnctl pipe 1 config`
	pfctlCreateDummynet  = `echo $'dummynet in all pipe 1'`
	pfctlExecuteInline   = `%s | sudo pfctl -a mop -f - `
	pfctlEnableFirewall  = `sudo pfctl -E`
	pfctlEnableFwRegex   = `pf enabled`
	pfctlDisableFirewall = `sudo pfctl -d`
	pfctlDisbleFwRegex   = `pf disabled`
	pfctlIsEnabled       = `sudo pfctl -sa | grep -i enabled`
	dnctlIsConfigured    = `sudo dnctl show`
	pfctlIsEnabledRegex  = `Enabled`
	dnctlTeardown        = `sudo dnctl -q flush`
)

type pfctlThrottler struct {
	c commander
}

// Execute a command and check that any matching line in the result contains 'match'
func (i *pfctlThrottler) executeAndParse(cmd string, match string) bool {
	lines, err := i.c.executeGetLines(cmd)

	if err != nil {
		return false
	}

	for _, line := range lines {
		if strings.Contains(line, match) {
			return true
		}
	}
	return false
}

func (i *pfctlThrottler) setup(c *Config) error {
	// Enable firewall
	err := i.c.execute(pfctlEnableFirewall)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not enable firewall using: `%s`. Error: %s", pfctlEnableFirewall, err.Error()))
	}

	// Add the dummynet and anchor
	err = i.c.execute(pfctlCreateAnchor)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create anchor rule for dummynet using: `%s`. Error: %s", pfctlCreateAnchor, err.Error()))
	}

	// Add 'execute' portion of the command
	cmd := fmt.Sprintf(pfctlExecuteInline, pfctlCreateDummynet)

	err = i.c.execute(cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create dummynet using: `%s`. Error: %s", pfctlCreateDummynet, err.Error()))
	}

	// Apply the shaping etc.
	for _, cmd := range i.buildConfigCommand(c) {
		err = i.c.execute(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *pfctlThrottler) teardown(_ *Config) error {

	// Reset firewall rules, leave it running
	err := i.c.execute(pfctlTeardown)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not remove firewall rules using: `%s`. Error: %s", pfctlTeardown, err.Error()))
	}

	// Turn off the firewall, discarding any rules
	err = i.c.execute(pfctlDisableFirewall)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not disable firewall using: `%s`. Error: %s", pfctlDisableFirewall, err.Error()))
	}

	// Disable dnctl rules
	err = i.c.execute(dnctlTeardown)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not disable dnctl rules using: `%s`. Error: %s", dnctlTeardown, err.Error()))
	}

	return nil
}

func (i *pfctlThrottler) isFirewallRunning() bool {
	return i.executeAndParse(pfctlIsEnabled, pfctlIsEnabledRegex)
}
func (i *pfctlThrottler) exists() bool {
	if dry {
		return false
	}
	return i.executeAndParse(dnctlIsConfigured, "port") || i.isFirewallRunning()
}

func (i *pfctlThrottler) check() string {
	return pfctlIsEnabled
}

func addProtosToCommands(cmds []string, protos []string) []string {
	commands := make([]string, 0)

	for _, cmd := range cmds {
		for _, proto := range protos {
			commands = append(commands, fmt.Sprintf("%s proto %s", cmd, proto))
		}
	}

	return commands
}
func addPortsToCommand(cmd string, ports []string) []string {
	commands := make([]string, 0)

	for _, port := range ports {
		commands = append(commands, fmt.Sprintf("%s dst-port %s", cmd, port))
		commands = append(commands, fmt.Sprintf("%s src-port %s", cmd, port))
	}

	return commands
}

// Takes care of the annoying differences between ipv4 and ipv6
func addIpsAndProtoToCommands(ipVersion int, cmds []string, ips []string, protos []string) []string {

	commands := make([]string, 0)

	for _, cmd := range cmds {
		for _, ip := range ips {
			srcIpFlag := "src-ip"
			dstIpFlag := "dst-ip"
			if ipVersion == 6 {
				srcIpFlag = "src-ip6"
				dstIpFlag = "dst-ip6"
			}

			commands = append(commands, addProtoToCommands(ipVersion, fmt.Sprintf("%s %s %s", cmd, srcIpFlag, ip), protos)...)
			commands = append(commands, addProtoToCommands(ipVersion, fmt.Sprintf("%s %s %s", cmd, dstIpFlag, ip), protos)...)
		}
	}

	if len(ips) == 0 {

	}

	return commands
}

func addProtoToCommands(ipVersion int, cmd string, protos []string) []string {
	commands := make([]string, 0)
	for _, proto := range protos {
		if ipVersion == 6 {
			if proto == "icmp" {
				proto = "ipv6-icmp"
			}
		}
		commands = append(commands, fmt.Sprintf("%s proto %s", cmd, proto))
	}
	return commands
}

func (i *pfctlThrottler) buildConfigCommand(c *Config) []string {

	cmd := dnctl

	// Add all non tcp version dependent stuff first...
	if c.Latency > 0 {
		cmd = cmd + " delay " + strconv.Itoa(c.Latency) + "ms"
	}

	if c.TargetBandwidth > 0 {
		cmd = cmd + " bw " + strconv.Itoa(c.TargetBandwidth) + "Kbit/s"
	}

	if c.PacketLoss > 0 {
		cmd = cmd + " plr " + strconv.FormatFloat(c.PacketLoss/100, 'f', 4, 64)
	}

	// Add Mask keyword if we have pipe qualifiers
	if len(c.TargetPorts) > 0 || len(c.TargetProtos) > 0 || len(c.TargetIps) > 0 || len(c.TargetIps6) > 0 {
		cmd = cmd + " mask "
	}

	// Expand commands with ports
	commands := []string{cmd}

	if len(c.TargetPorts) > 0 {
		commands = addPortsToCommand(cmd, c.TargetPorts)
	}

	if len(c.TargetIps) == 0 && len(c.TargetIps6) == 0 {
		if len(c.TargetProtos) > 0 {
			return addProtosToCommands(commands, c.TargetProtos)
		}
		return commands
	}

	// create and combine the ipv4 and ipv6 IPs with the protocol version specific keywords
	return append(addIpsAndProtoToCommands(4, commands, c.TargetIps, c.TargetProtos), addIpsAndProtoToCommands(6, commands, c.TargetIps6, c.TargetProtos)...)
}

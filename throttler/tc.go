package throttler

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	tcRootQDisc    = `dev %s handle 10: root`
	tcDefaultClass = `dev %s parent 10: classid 10:1`
	tcTargetClass  = `dev %s parent 10:1 classid 10:10`
	tcNetemRule    = `dev %s parent 10:10 handle 100:`
	tcRate         = `rate %vkbit`
	tcDelay        = `delay %vms`
	tcLoss         = `loss %v%%`
	tcAddClass     = `sudo tc class add`
	tcDelClass     = `sudo tc class del`
	tcAddQDisc     = `sudo tc qdisc add`
	tcDelQDisc     = `sudo tc qdisc del`
	iptAddTarget   = `sudo iptables -A POSTROUTING -t mangle -j CLASSIFY --set-class 10:10`
	iptDelTarget   = `sudo iptables -D POSTROUTING -t mangle -j CLASSIFY --set-class 10:10`
	iptDestIP      = `-d %s`
	iptProto       = `-p %s`
	iptDestPorts   = `--match multiport --dports %s`
	iptDestPort    = `--dport %s`
	iptDelSearch   = `class 0010:0010`
	iptList        = `sudo iptables -S -t mangle`
	iptDel         = `sudo iptables -t mangle -D`
	tcExists       = `sudo tc qdisc show | grep "netem"`
	tcCheck        = `sudo tc -s qdisc`
)

type tcThrottler struct{}

func (t *tcThrottler) setup(c *Config) error {
	err := addRootQDisc(c) //The root node to append the filters
	if err != nil {
		return err
	}

	err = addDefaultClass(c) //The default class for all traffic that isn't classified
	if err != nil {
		return err
	}

	err = addTargetClass(c) //The class that the network emulator rule is assigned
	if err != nil {
		return err
	}

	err = addNetemRule(c) //The network emulator rule that contains the desired behavior
	if err != nil {
		return err
	}

	return addIptablesRules(c) //The network emulator rule that contains the desired behavior
}

func addRootQDisc(c *Config) error {
	//Add the root QDisc
	root := fmt.Sprintf(tcRootQDisc, c.Device)
	strs := []string{tcAddQDisc, root, "htb"}
	cmd := strings.Join(strs, " ")

	return runCommand(cmd)
}

func addDefaultClass(c *Config) error {
	//Add the default Class
	def := fmt.Sprintf(tcDefaultClass, c.Device)
	rate := ""

	if c.DefaultBandwidth > 0 {
		rate = fmt.Sprintf(tcRate, c.DefaultBandwidth)
	} else {
		rate = fmt.Sprintf(tcRate, 1000000)
	}

	strs := []string{tcAddClass, def, "htb", rate}
	cmd := strings.Join(strs, " ")

	return runCommand(cmd)
}

func addTargetClass(c *Config) error {
	//Add the target Class
	tar := fmt.Sprintf(tcTargetClass, c.Device)
	rate := ""

	if c.DefaultBandwidth > -1 {
		rate = fmt.Sprintf(tcRate, c.DefaultBandwidth)
	} else {
		rate = fmt.Sprintf(tcRate, 1000000)
	}

	strs := []string{tcAddClass, tar, "htb", rate}
	cmd := strings.Join(strs, " ")

	return runCommand(cmd)
}

func addNetemRule(c *Config) error {
	//Add the Network Emulator rule
	net := fmt.Sprintf(tcNetemRule, c.Device)
	strs := []string{tcAddQDisc, net, "netem"}

	if c.Latency > 0 {
		strs = append(strs, fmt.Sprintf(tcDelay, c.Latency))
	}

	if c.TargetBandwidth > -1 {
		strs = append(strs, fmt.Sprintf(tcRate, c.TargetBandwidth))
	}

	if c.PacketLoss > 0 {
		strs = append(strs, fmt.Sprintf(tcLoss, strconv.FormatFloat(c.PacketLoss, 'f', 2, 64)))
	}

	cmd := strings.Join(strs, " ")

	return runCommand(cmd)
}

func addIptablesRules(c *Config) error {
	rules := []string{}
	ports := ""

	if len(c.TargetPorts) > 0 {
		if len(c.TargetPorts) > 1 {
			prts := strings.Join(c.TargetPorts, ",")
			ports = fmt.Sprintf(iptDestPorts, prts)
		} else {
			ports = fmt.Sprintf(iptDestPort, c.TargetPorts[0])
		}
	}

	if len(c.TargetProtos) > 0 {
		for _, ptc := range c.TargetProtos {
			proto := fmt.Sprintf(iptProto, ptc)
			rule := strings.Join([]string{iptAddTarget, proto}, " ")

			if ptc != "icmp" {
				if ports != "" {
					rule += " " + ports
				}
			}

			rules = append(rules, rule)
		}
	} else {
		rules = []string{iptAddTarget}
	}

	if len(c.TargetIps) > 0 {
		iprules := []string{}
		for _, ip := range c.TargetIps {
			dest := fmt.Sprintf(iptDestIP, ip)
			if len(rules) > 0 {
				for _, rule := range rules {
					r := rule + " " + dest
					iprules = append(iprules, r)
				}
			} else {
				iprules = append(iprules, dest)
			}
		}
		if len(iprules) > 0 {
			rules = iprules
		}
	}

	for _, rule := range rules {
		err := runCommand(rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tcThrottler) teardown(c *Config) error {
	err := delIptablesRules()
	if err != nil {
		return err
	}

	err = delRootQDisc(c) //The root node to append the filters
	if err != nil {
		return err
	}

	return nil
}

func delIptablesRules() error {
	lines, err := runCommandGetLines(iptList)
	if err != nil {
		return err
	}

	if len(lines) > 0 {
		for _, line := range lines {
			if strings.Contains(line, iptDelSearch) {
				cmd := strings.Replace(line, "-A", iptDel, 1)
				err = runCommand(cmd)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func delRootQDisc(c *Config) error {
	//Delete the root QDisc
	root := fmt.Sprintf(tcRootQDisc, c.Device)

	strs := []string{tcDelQDisc, root}
	cmd := strings.Join(strs, " ")

	return runCommand(cmd)
}

func (t *tcThrottler) exists() bool {
	if dry {
		return false
	}
	err := runCommand(tcExists)
	return err == nil
}

func (t *tcThrottler) check() string {
	return tcCheck
}

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

type tcThrottler struct {
	c commander
}

func (t *tcThrottler) setup(cfg *Config) error {
	err := addRootQDisc(cfg, t.c) //The root node to append the filters
	if err != nil {
		return err
	}

	err = addDefaultClass(cfg, t.c) //The default class for all traffic that isn't classified
	if err != nil {
		return err
	}

	err = addTargetClass(cfg, t.c) //The class that the network emulator rule is assigned
	if err != nil {
		return err
	}

	err = addNetemRule(cfg, t.c) //The network emulator rule that contains the desired behavior
	if err != nil {
		return err
	}

	return addIptablesRules(cfg, t.c) //The network emulator rule that contains the desired behavior
}

func addRootQDisc(cfg *Config, c commander) error {
	//Add the root QDisc
	root := fmt.Sprintf(tcRootQDisc, cfg.Device)
	strs := []string{tcAddQDisc, root, "htb"}
	cmd := strings.Join(strs, " ")

	return c.execute(cmd)
}

func addDefaultClass(cfg *Config, c commander) error {
	//Add the default Class
	def := fmt.Sprintf(tcDefaultClass, cfg.Device)
	rate := ""

	if cfg.DefaultBandwidth > 0 {
		rate = fmt.Sprintf(tcRate, cfg.DefaultBandwidth)
	} else {
		rate = fmt.Sprintf(tcRate, 1000000)
	}

	strs := []string{tcAddClass, def, "htb", rate}
	cmd := strings.Join(strs, " ")

	return c.execute(cmd)
}

func addTargetClass(cfg *Config, c commander) error {
	//Add the target Class
	tar := fmt.Sprintf(tcTargetClass, cfg.Device)
	rate := ""

	if cfg.DefaultBandwidth > -1 {
		rate = fmt.Sprintf(tcRate, cfg.DefaultBandwidth)
	} else {
		rate = fmt.Sprintf(tcRate, 1000000)
	}

	strs := []string{tcAddClass, tar, "htb", rate}
	cmd := strings.Join(strs, " ")

	return c.execute(cmd)
}

func addNetemRule(cfg *Config, c commander) error {
	//Add the Network Emulator rule
	net := fmt.Sprintf(tcNetemRule, cfg.Device)
	strs := []string{tcAddQDisc, net, "netem"}

	if cfg.Latency > 0 {
		strs = append(strs, fmt.Sprintf(tcDelay, cfg.Latency))
	}

	if cfg.TargetBandwidth > -1 {
		strs = append(strs, fmt.Sprintf(tcRate, cfg.TargetBandwidth))
	}

	if cfg.PacketLoss > 0 {
		strs = append(strs, fmt.Sprintf(tcLoss, strconv.FormatFloat(cfg.PacketLoss, 'f', 2, 64)))
	}

	cmd := strings.Join(strs, " ")

	return c.execute(cmd)
}

func addIptablesRules(cfg *Config, c commander) error {
	rules := []string{}
	ports := ""

	if len(cfg.TargetPorts) > 0 {
		if len(cfg.TargetPorts) > 1 {
			prts := strings.Join(cfg.TargetPorts, ",")
			ports = fmt.Sprintf(iptDestPorts, prts)
		} else {
			ports = fmt.Sprintf(iptDestPort, cfg.TargetPorts[0])
		}
	}

	if len(cfg.TargetProtos) > 0 {
		for _, ptc := range cfg.TargetProtos {
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

	if len(cfg.TargetIps) > 0 {
		iprules := []string{}
		for _, ip := range cfg.TargetIps {
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
		if err := c.execute(rule); err != nil {
			return err
		}
	}

	return nil
}

func (t *tcThrottler) teardown(cfg *Config) error {
	if err := delIptablesRules(t.c); err != nil {
		return err
	}

	// The root node to append the filters
	if err := delRootQDisc(cfg, t.c); err != nil {
		return err
	}
	return nil
}

func delIptablesRules(c commander) error {
	lines, err := c.executeGetLines(iptList)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if strings.Contains(line, iptDelSearch) {
			cmd := strings.Replace(line, "-A", iptDel, 1)
			err = c.execute(cmd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func delRootQDisc(cfg *Config, c commander) error {
	//Delete the root QDisc
	root := fmt.Sprintf(tcRootQDisc, cfg.Device)

	strs := []string{tcDelQDisc, root}
	cmd := strings.Join(strs, " ")

	return c.execute(cmd)
}

func (t *tcThrottler) exists() bool {
	if dry {
		return false
	}
	err := t.c.execute(tcExists)
	return err == nil
}

func (t *tcThrottler) check() string {
	return tcCheck
}

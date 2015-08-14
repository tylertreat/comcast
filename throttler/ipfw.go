package throttler

import (
	"strconv"
)

const (
	ipfwAddPipe  = `sudo ipfw add 1 pipe 1 ip from any to any via `
	ipfwTeardown = `sudo ipfw delete 1`
	ipfwConfig   = `sudo ipfw pipe 1 config`
	ipfwExists   = `sudo ipfw list | grep "pipe 1"`
	ipfwCheck    = `sudo ipfw list`
)

type ipfwThrottler struct {
	c commander
}

func (i *ipfwThrottler) setup(c *Config) error {
	cmd := ipfwAddPipe + c.Device
	err := i.c.execute(cmd)
	if err != nil {
		return err
	}

	configCmd := i.buildConfigCommand(c)
	err = i.c.execute(configCmd)
	return err
}

func (i *ipfwThrottler) teardown(_ *Config) error {
	err := i.c.execute(ipfwTeardown)
	return err
}

func (i *ipfwThrottler) exists() bool {
	if dry {
		return false
	}
	err := i.c.execute(ipfwExists)
	return err == nil
}

func (i *ipfwThrottler) check() string {
	return ipfwCheck
}

func (i *ipfwThrottler) buildConfigCommand(c *Config) string {
	cmd := ipfwConfig

	if c.Latency > 0 {
		cmd = cmd + " delay " + strconv.Itoa(c.Latency) + "ms"
	}

	if c.TargetBandwidth > 0 {
		cmd = cmd + " bw " + strconv.Itoa(c.TargetBandwidth) + "Kbit/s"
	}

	if c.PacketLoss > 0 {
		cmd = cmd + " plr " + strconv.FormatFloat(c.PacketLoss/100, 'f', 4, 64)
	}

	return cmd
}

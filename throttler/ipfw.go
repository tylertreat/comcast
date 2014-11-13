package throttler

import (
	"fmt"
	"os/exec"
	"strconv"
)

const (
	ipfwAddPipe  = `sudo ipfw add 1 pipe 1 ip from any to any`
	ipfwTeardown = `sudo ipfw delete 1`
	ipfwConfig   = `sudo ipfw pipe 1 config`
	ipfwExists   = `sudo ipfw list | grep "pipe 1"`
	ipfwCheck    = `sudo ipfw list`
)

type ipfwThrottler struct{}

func (i *ipfwThrottler) setup(c *Config) error {
	fmt.Println(ipfwAddPipe)
	if err := exec.Command("/bin/sh", "-c", ipfwAddPipe).Run(); err != nil {
		return err
	}

	configCmd := i.buildConfigCommand(c)
	fmt.Println(configCmd)
	return exec.Command("/bin/sh", "-c", configCmd).Run()
}

func (i *ipfwThrottler) teardown() error {
	fmt.Println(ipfwTeardown)
	return exec.Command("/bin/sh", "-c", ipfwTeardown).Run()
}

func (i *ipfwThrottler) exists() bool {
	fmt.Println(ipfwExists)
	return exec.Command("/bin/sh", "-c", ipfwExists).Run() == nil
}

func (i *ipfwThrottler) check() string {
	return ipfwCheck
}

func (i *ipfwThrottler) buildConfigCommand(c *Config) string {
	cmd := ipfwConfig

	if c.Latency > 0 {
		cmd = cmd + " delay " + strconv.Itoa(c.Latency) + "ms"
	}

	if c.Bandwidth > 0 {
		cmd = cmd + " bw " + strconv.Itoa(c.Bandwidth) + "Kbit/s"
	}

	if c.PacketLoss > 0 {
		cmd = cmd + " plr " + strconv.FormatFloat(c.PacketLoss, 'f', 2, 64)
	}

	return cmd
}

package throttler

import (
	"fmt"
	"os/exec"
	"strconv"
)

const (
	tcAdd      = `sudo tc qdisc add dev eth0 root netem`
	tcTeardown = `sudo tc qdisc del dev eth0 root netem`
	tcExists   = `sudo tc qdisc show | grep "netem"`
	tcCheck    = `sudo tc -s qdisc`
)

type tcThrottler struct{}

func (t *tcThrottler) setup(c *Config) error {
	cmd := t.buildConfigCommand(c)
	fmt.Println(cmd)
	return exec.Command("/bin/sh", "-c", cmd).Run()
}

func (t *tcThrottler) teardown() error {
	fmt.Println(tcTeardown)
	return exec.Command("/bin/sh", "-c", tcTeardown).Run()
}

func (t *tcThrottler) exists() bool {
	fmt.Println(tcExists)
	return exec.Command("/bin/sh", "-c", tcExists).Run() == nil
}

func (t *tcThrottler) check() string {
	return tcCheck
}

func (t *tcThrottler) buildConfigCommand(c *Config) string {
	cmd := tcAdd

	if c.Latency > 0 {
		cmd = cmd + " delay " + strconv.Itoa(c.Latency) + "ms"
	}

	if c.Bandwidth > 0 {
		cmd = cmd + " rate " + strconv.Itoa(c.Bandwidth) + "kbit"
	}

	if c.PacketLoss > 0 {
		cmd = cmd + " loss " + strconv.FormatFloat(c.PacketLoss*100, 'f', 0, 64) + "%"
	}

	return cmd
}

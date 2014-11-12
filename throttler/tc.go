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

func (t *tcThrottler) setup(config *Config) error {
	cmd := t.buildConfigCommand(config)
	fmt.Println(cmd)
	return exec.Command("/bin/sh", "-c", cmd).Run()
}

func (t *tcThrottler) teardown() error {
	fmt.Println(tcTeardown)
	return exec.Command("/bin/sh", "-c", tcTeardown).Run()
}

func (t *tcThrottler) exists() bool {
	fmt.Println(tcExists)
	err := exec.Command("/bin/sh", "-c", tcExists).Run()
	return err == nil
}

func (t *tcThrottler) check() string {
	return tcCheck
}

func (t *tcThrottler) buildConfigCommand(config *Config) string {
	cmd := tcAdd
	if config.Latency > 0 {
		latencyStr := strconv.Itoa(config.Latency)
		cmd = cmd + " delay " + latencyStr + "ms"
	}

	if config.Bandwidth > 0 {
		bwStr := strconv.Itoa(config.Bandwidth)
		cmd = cmd + " rate " + bwStr + "kbit"
	}

	if config.PacketLoss > 0 {
		lossStr := strconv.FormatFloat(config.PacketLoss*100, 'f', 0, 64)
		cmd = cmd + " loss " + lossStr + "%"
	}

	return cmd
}

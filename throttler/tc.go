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

func (t *tcThrottler) setup(config *config) error {
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

func (t *tcThrottler) buildConfigCommand(config *config) string {
	cmd := tcAdd
	if config.latency > 0 {
		latencyStr := strconv.Itoa(config.latency)
		cmd = cmd + " delay " + latencyStr + "ms"
	}

	if config.bandwidth > 0 {
		bwStr := strconv.Itoa(config.bandwidth)
		cmd = cmd + " rate " + bwStr + "kbit"
	}

	if config.packetLoss > 0 {
		lossStr := strconv.FormatFloat(config.packetLoss*100, 'f', 0, 64)
		cmd = cmd + " loss " + lossStr + "%"
	}

	return cmd
}

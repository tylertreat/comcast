package wiretap

import (
	"fmt"
	"os/exec"
	"strconv"
)

const (
	tcTeardown = `sudo tc qdisc del dev eth0 root netem`
	tcExists   = `sudo tc qdisc show | grep "netem"`
	tcCheck    = `sudo tc qdisc show`
)

type LinuxWiretap struct{}

func (l *LinuxWiretap) Setup(config *Config) error {
	cmd := l.buildConfigCommand(config)
	return exec.Command("/bin/sh", "-c", cmd).Run()
}

func (l *LinuxWiretap) Teardown() error {
	return exec.Command("/bin/sh", "-c", tcTeardown).Run()
}

func (l *LinuxWiretap) Exists() bool {
	err := exec.Command("/bin/sh", "-c", tcExists).Run()
	return err == nil
}

func (l *LinuxWiretap) Check() string {
	return tcCheck
}

func (l *LinuxWiretap) buildConfigCommand(config *Config) string {
	cmd := "sudo tc qdisc add dev eth0 root netem"
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

	fmt.Println(cmd)
	return cmd
}

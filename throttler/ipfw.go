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

func (i *ipfwThrottler) setup(config *Config) error {
	fmt.Println(ipfwAddPipe)
	if err := exec.Command("/bin/sh", "-c", ipfwAddPipe).Run(); err != nil {
		return err
	}

	configCmd := i.buildConfigCommand(config)
	fmt.Println(configCmd)
	return exec.Command("/bin/sh", "-c", configCmd).Run()
}

func (i *ipfwThrottler) teardown() error {
	fmt.Println(ipfwTeardown)
	return exec.Command("/bin/sh", "-c", ipfwTeardown).Run()
}

func (i *ipfwThrottler) exists() bool {
	fmt.Println(ipfwExists)
	err := exec.Command("/bin/sh", "-c", ipfwExists).Run()
	return err == nil
}

func (i *ipfwThrottler) check() string {
	return ipfwCheck
}

func (d *ipfwThrottler) buildConfigCommand(config *Config) string {
	cmd := ipfwConfig
	if config.Latency > 0 {
		latencyStr := strconv.Itoa(config.Latency)
		cmd = cmd + " delay " + latencyStr + "ms"
	}

	if config.Bandwidth > 0 {
		bwStr := strconv.Itoa(config.Bandwidth)
		cmd = cmd + " bw " + bwStr + "Kbit/s"
	}

	if config.PacketLoss > 0 {
		plrStr := strconv.FormatFloat(config.PacketLoss, 'f', 2, 64)
		cmd = cmd + " plr " + plrStr
	}

	return cmd
}

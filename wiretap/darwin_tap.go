package wiretap

import (
	"fmt"
	"os/exec"
	"strconv"
)

// TODO: Add support for pfctl.

const (
	createInboundPipe  = `sudo ipfw add 1 pipe 1 ip from any to me`
	createOutboundPipe = `sudo ipfw add 2 pipe 2 ip from me to any`
	deleteInboundPipe  = `sudo ipfw delete 1`
	deleteOutboundPipe = `sudo ipfw delete 2`
	checkInboundPipe   = `sudo ipfw list | grep "pipe 1"`
	check              = `sudo ipfw pipe show`
)

type DarwinWiretap struct{}

func (d *DarwinWiretap) Setup(config *Config) error {
	d.buildConfigCommand(config)
	cmd := exec.Command("/bin/sh", "-c", createInboundPipe)
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("/bin/sh", "-c", createOutboundPipe)
	return cmd.Run()
}

func (d *DarwinWiretap) Teardown() error {
	var err error
	cmd := exec.Command("/bin/sh", "-c", deleteInboundPipe)
	if e := cmd.Run(); e != nil {
		err = e
	}
	cmd = exec.Command("/bin/sh", "-c", deleteOutboundPipe)
	if e := cmd.Run(); e != nil {
		err = e
	}
	return err
}

func (d *DarwinWiretap) Exists() bool {
	err := exec.Command("/bin/sh", "-c", checkInboundPipe).Run()
	return err == nil
}

func (d *DarwinWiretap) Check() string {
	return check
}

func (d *DarwinWiretap) buildConfigCommand(config *Config) string {
	cmd := "sudo ipfw pipe 1 config"
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

	fmt.Println(cmd)
	return cmd
}

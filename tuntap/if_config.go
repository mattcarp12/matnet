package tuntap

import "os/exec"

const DefaultIPv4Addr = "10.88.45.1/24"

func IfaceConfig(name, ipAddr string) {
	exec.Command("ip", "addr", "add", ipAddr, "dev", name).Run()
	exec.Command("ip", "link", "set", name, "up").Run()
}
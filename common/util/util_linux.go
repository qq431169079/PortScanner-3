package util

import (
	"fmt"
	"os/exec"
)

func GetNetInfo(dstIp string) (localIP, interfaceName, routeIp string, err error) {
	routeCmd := exec.Command("ip", "route", "get", dstIp)
	output, err := routeCmd.CombinedOutput()
	if err != nil {
		fmt.Println("exec ip command failed")
	}
	return parseIpRoute(output)
}

package net

import (
	"net"

	"github.com/designinlife/slib/errors"
)

// GetLocalIPAddr 读取本机 IP 地址。
func GetLocalIPAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Wrap(err, "failed to read the IP address of the machine")
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("failed to read the IP address of the machine")
}

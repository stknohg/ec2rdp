package cmd

import (
	"net"
	"strconv"
	"time"
)

func isPortOpen(hostName string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostName, strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	if conn == nil {
		return false
	}
	return true
}

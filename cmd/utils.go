package cmd

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/term"
)

func readPrompt(prompt string) string {
	fmt.Print(prompt)
	val, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Printf("\n")
	return string(val)
}

func isPortOpen(hostName string, port int) bool {
	if port < 1 || port > 65535 {
		return false
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostName, strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	if conn == nil {
		return false
	}
	return true
}

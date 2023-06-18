package cmd

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
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

func getLocalRDPPort(localHost string, startPort int) (int, error) {
	for i := startPort; i <= 65535; i++ {
		listener, err := net.Listen("tcp", net.JoinHostPort(localHost, strconv.Itoa(i)))
		if err == nil {
			defer listener.Close()
			return i, nil
		}
	}
	return 65535, fmt.Errorf("failed to find local proxy port")
}

func invokeRegionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// generate from : aws ec2 describe-regions --all-regions --query "sort_by(Regions,&RegionName)[].RegionName" --output json
	regions := []string{
		"af-south-1",
		"ap-east-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-south-1",
		"ap-south-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-southeast-3",
		"ap-southeast-4",
		"ca-central-1",
		"eu-central-1",
		"eu-central-2",
		"eu-north-1",
		"eu-south-1",
		"eu-south-2",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"me-central-1",
		"me-south-1",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
	return regions, cobra.ShellCompDirectiveDefault
}

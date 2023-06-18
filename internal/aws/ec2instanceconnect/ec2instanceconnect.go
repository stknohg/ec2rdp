package ec2instanceconnect

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func OpenTunnel(cfg aws.Config, ctx context.Context, endpointId string, endpointDnsName string, privateIpAddress string, localPort int, remotePort int) (int, error) {
	// execute aws ec2-instance-connect open-tunnel command.
	args := []string{}
	args = append(args, "ec2-instance-connect")
	args = append(args, "open-tunnel")
	args = append(args, "--instance-connect-endpoint-id")
	args = append(args, endpointId)
	args = append(args, "--instance-connect-endpoint-dns-name")
	args = append(args, endpointDnsName)
	args = append(args, "--private-ip-address")
	args = append(args, privateIpAddress)
	args = append(args, "--local-port")
	args = append(args, strconv.Itoa(localPort))
	args = append(args, "--remote-port")
	args = append(args, strconv.Itoa(remotePort))
	cmd := exec.Command("aws", args...)
	// pass environment variables in case MFA is required
	cred, _ := cfg.Credentials.Retrieve(ctx)
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", cred.AccessKeyID))
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", cred.SecretAccessKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_SESSION_TOKEN=%v", cred.SessionToken))
	cmd.Env = append(cmd.Env, fmt.Sprintf("AWS_REGION=%v", cfg.Region))
	// start process
	err := cmd.Start()
	if err != nil {
		return -1, err
	}
	return cmd.Process.Pid, nil
}

func CloseTunnel(pid int) error {
	if pid < 0 {
		return fmt.Errorf("invalid pid specified")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

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
	args = append(args,
		"ec2-instance-connect", "open-tunnel",
		"--instance-connect-endpoint-id", endpointId,
		"--instance-connect-endpoint-dns-name", endpointDnsName,
		"--private-ip-address", privateIpAddress,
		"--local-port", strconv.Itoa(localPort),
		"--remote-port", strconv.Itoa(remotePort),
	)
	cmd := exec.Command("aws", args...)
	// pass environment variables in case MFA is required
	cred, _ := cfg.Credentials.Retrieve(ctx)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%v", cred.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%v", cred.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%v", cred.SessionToken),
		fmt.Sprintf("AWS_REGION=%v", cfg.Region),
	)
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

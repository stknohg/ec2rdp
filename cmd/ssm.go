package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/stknohg/ec2rdp/internal/aws"
	"github.com/stknohg/ec2rdp/internal/aws/ec2"
	"github.com/stknohg/ec2rdp/internal/aws/ssm"
	"github.com/stknohg/ec2rdp/internal/connector"
)

var ssmInstanceId string
var ssmPemFile string
var ssmPort int
var ssmUserName string
var ssmUserPassword string
var ssmProfileName string
var ssmRegionName string

// ssmCmd represents the ssm command
var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "Connect to EC2 Instance via SSM Session Manager",
	Long:  `Connect to EC2 Instance via SSM Session Manager`,
	Args: func(cmd *cobra.Command, args []string) error {
		if !isSessionManagerPluginInstalled() {
			return errors.New("session-manager-plugin is not installed")
		}
		if ssmPemFile == "" && ssmUserPassword == "" {
			return errors.New("--pemfile or --password flag is requied")
		}
		if ssmPemFile != "" {
			_, err := os.Stat(ssmPemFile)
			if err != nil {
				return errors.New(".pem file does not exist")
			}
		}
		if ssmPort < 0 || ssmPort > 65535 {
			return errors.New("set port number between 1 and 65535")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return invokeSSMCommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(ssmCmd)
	ssmCmd.Flags().StringVarP(&ssmInstanceId, "instance", "i", "", "EC2 Instance ID")
	ssmCmd.Flags().StringVarP(&ssmPemFile, "pemfile", "p", "", ".pem file path")
	ssmCmd.Flags().IntVar(&ssmPort, "port", 3389, "RDP port no")
	ssmCmd.Flags().StringVar(&ssmUserName, "user", "Administrator", "RDP username")
	ssmCmd.Flags().StringVarP(&ssmUserPassword, "password", "P", "", "RDP passowrd")
	ssmCmd.Flags().StringVar(&ssmProfileName, "profile", "", "AWS profile name")
	ssmCmd.Flags().StringVar(&ssmRegionName, "region", "", "AWS region name")
	//
	ssmCmd.MarkFlagRequired("instance")
	ssmCmd.MarkFlagFilename("pemfile", "pem")
	ssmCmd.MarkFlagsMutuallyExclusive("pemfile", "password")
}

func invokeSSMCommand(cmd *cobra.Command, args []string) error {
	// get aws config
	cfg := aws.GetConfig(ssmProfileName, ssmRegionName)

	// check instance exists
	_, err := ec2.IsInstanceExist(cfg, ssmInstanceId)
	if err != nil {
		return err
	}

	// check instance status
	_, err = ssm.IsInstanceOnline(cfg, ssmInstanceId)
	if err != nil {
		return err
	}

	// get administrator password
	var password string
	if ssmUserPassword == "" {
		password, err = ec2.GetAdministratorPassword(cfg, ssmInstanceId, ssmPemFile)
		if err != nil {
			return err
		}
		if password == "" {
			return fmt.Errorf("EC2 PasswordData is empty. Use --password flag instead")
		}
		fmt.Println("Administrator password acquisition completed")
	} else {
		password = ssmUserPassword
	}

	// get hostname and local port
	var localHostName = "localhost"
	localPort, err := getLocalRDPPort(localHostName, 33389)
	if err != nil {
		return err
	}

	// start port forwarding with SSM Session Manager Plugin
	result, err := ssm.StartSSMSessionPortForward(cfg, ssmInstanceId, ssmPort, localPort, ssmProfileName)
	if err != nil {
		return err
	}
	fmt.Printf("Starting session with SessionId: %v\n", result.SessionId)
	for i := 1; ; i++ {
		if isPortOpen(localHostName, localPort) {
			break
		}
		time.Sleep(500 * time.Millisecond)
		if i >= 10 {
			return fmt.Errorf("%v port %v is not open", localHostName, localPort)
		}
	}
	fmt.Printf("Start listening %v:%v\n", localHostName, localPort)

	// connect
	return connectSSMInstance(localHostName, localPort, ssmUserName, password, result)
}

func isSessionManagerPluginInstalled() bool {
	_, err := exec.LookPath("session-manager-plugin")
	return err == nil
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

func connectSSMInstance(hostName string, port int, userName string, plainPassword string, pluginResult *ssm.StartSSMSessionPluginResult) error {
	con := connector.DefaultConnector{
		HostName:      hostName,
		Port:          port,
		UserName:      userName,
		PlainPassword: plainPassword,
		WaitFor:       true, // always true
	}
	err := con.PreConnect()
	if err != nil {
		return err
	}
	err = con.Connect()
	if err != nil {
		return err
	}
	defer func() {
		con.PostConnect()
		fmt.Printf("Terminate SSM session%v\n", pluginResult.SessionId)
		ssm.TerminateSSMSession(pluginResult.Config, pluginResult.SessionId)
	}()
	return nil
}

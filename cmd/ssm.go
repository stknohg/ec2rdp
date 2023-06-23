package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/stknohg/ec2rdp/internal/aws"
	"github.com/stknohg/ec2rdp/internal/aws/ec2"
	"github.com/stknohg/ec2rdp/internal/aws/ssm"
	"github.com/stknohg/ec2rdp/internal/connector"
)

// ssmCmd represents the ssm command
var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "Connect to EC2 instance via SSM Session Manager",
	Long:  `Connect to EC2 instance via SSM Session Manager`,
	Args: func(cmd *cobra.Command, args []string) error {
		if installed, err := isSessionManagerPluginInstalled(); !installed {
			return err
		}
		if cpPemFile == "" && !cpUserPassword {
			return errors.New("--pemfile or --password flag is requied")
		}
		if cpPemFile != "" {
			err := validatePemFile(cpPemFile)
			if err != nil {
				return err
			}
		}
		err := validatePort(cpPort)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return invokeSSMCommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(ssmCmd)
	ssmCmd.Flags().StringVarP(&cpInstanceId, "instance", "i", "", "EC2 Instance ID")
	ssmCmd.Flags().StringVarP(&cpPemFile, "pemfile", "p", "", ".pem file path")
	ssmCmd.Flags().IntVar(&cpPort, "port", 3389, "RDP port no")
	ssmCmd.Flags().StringVar(&cpUserName, "user", "Administrator", "RDP username")
	ssmCmd.Flags().BoolVarP(&cpUserPassword, "password", "P", false, "RDP passowrd")
	ssmCmd.Flags().StringVar(&cpProfileName, "profile", "", "AWS profile name")
	ssmCmd.Flags().StringVar(&cpRegionName, "region", "", "AWS region name")
	//
	ssmCmd.MarkFlagRequired("instance")
	ssmCmd.MarkFlagFilename("pemfile", "pem")
	ssmCmd.MarkFlagsMutuallyExclusive("pemfile", "password")
	// custom completion
	ssmCmd.RegisterFlagCompletionFunc("region", invokeRegionCompletion)
}

func invokeSSMCommand(cmd *cobra.Command, args []string) error {
	// check if connector application installed
	connector := connector.DefaultConnector{}
	_, err := connector.IsInstalled()
	if err != nil {
		return err
	}

	// get aws config
	cfg := aws.GetConfig(cpProfileName, cpRegionName)
	ec2api := ec2.NewAPI(cfg)
	ssmapi := ssm.NewAPI(cfg)
	ctx := context.Background()

	// check instance exists
	_, err = ec2.IsInstanceExist(ec2api, ctx, cpInstanceId)
	if err != nil {
		return err
	}

	// check instance status
	_, err = ssm.IsInstanceOnline(ssmapi, ctx, cpInstanceId)
	if err != nil {
		return err
	}

	// get administrator password
	password, message, err := getAdministratorPasswordWithPrompt(ec2api, ctx, cpInstanceId, cpPemFile, cpUserPassword)
	if err != nil {
		return err
	}
	if message != "" {
		fmt.Println(message)
	}

	// get hostname and local port
	var localHostName = "localhost"
	localPort, err := getLocalRDPPort(localHostName, 33389)
	if err != nil {
		return err
	}

	// start port forwarding with SSM Session Manager Plugin
	var ssmRegion = cfg.Region
	var ssmProfile = getSSMProfileName(cpProfileName)
	ssmResult, err := ssm.StartSSMSessionPortForward(ssmapi, ctx, cpInstanceId, cpPort, localPort, "ec2rdp ssm", ssmRegion, ssmProfile)
	if err != nil {
		return err
	}
	fmt.Printf("Starting session with SessionId: %v\n", ssmResult.SessionId)
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
	connector.HostName = localHostName
	connector.Port = localPort
	connector.UserName = cpUserName
	connector.PlainPassword = password
	connector.WaitFor = true // always true
	return connectSSMInstance(&connector, ssmResult)
}

func getSSMProfileName(input string) string {
	if input != "" {
		return input
	}
	return os.Getenv("AWS_PROFILE")
}

func isSessionManagerPluginInstalled() (bool, error) {
	_, err := exec.LookPath("session-manager-plugin")
	if err != nil {
		// ref : https://github.com/aws/aws-cli/blob/2.11.16/awscli/customizations/sessionmanager.py#L23-L28
		return false, errors.New(`SessionManagerPlugin is not found.
Please refer to SessionManager Documentation here: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-troubleshooting.html#plugin-not-found`)
	}
	return true, nil
}

func connectSSMInstance(con connector.Connector, ret *ssm.StartSSMSessionPluginResult) error {
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
		fmt.Printf("Terminate SSM session%v\n", ret.SessionId)
		ssm.TerminateSSMSession(ret.API, context.Background(), ret.SessionId)
	}()
	return nil
}

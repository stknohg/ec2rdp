package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/stknohg/ec2rdp/internal/aws"
	"github.com/stknohg/ec2rdp/internal/aws/ec2"
	"github.com/stknohg/ec2rdp/internal/aws/ec2instanceconnect"
	"github.com/stknohg/ec2rdp/internal/connector"
)

var (
	eiceEndpointId string
)

// eiceCmd represents the ssm command
var eiceCmd = &cobra.Command{
	Use:   "eice",
	Short: "Connect to EC2 instance via EC2 Instance Connect Endpoint",
	Long:  `Connect to EC2 instance via EC2 Instance Connect Endpoint`,
	Args: func(cmd *cobra.Command, args []string) error {
		if installed, err := isAWSCLIInstalled(); !installed {
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
		return invokeEICECommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(eiceCmd)
	eiceCmd.Flags().StringVarP(&cpInstanceId, "instance", "i", "", "EC2 Instance ID")
	eiceCmd.Flags().StringVarP(&cpPemFile, "pemfile", "p", "", ".pem file path")
	eiceCmd.Flags().IntVar(&cpPort, "port", 3389, "RDP port no")
	eiceCmd.Flags().StringVar(&cpUserName, "user", "Administrator", "RDP username")
	eiceCmd.Flags().BoolVarP(&cpUserPassword, "password", "P", false, "RDP passowrd")
	eiceCmd.Flags().StringVar(&cpProfileName, "profile", "", "AWS profile name")
	eiceCmd.Flags().StringVar(&cpRegionName, "region", "", "AWS region name")
	eiceCmd.Flags().StringVarP(&eiceEndpointId, "endpointid", "e", "", "EC2 Instance Connect Endpoint ID")
	//
	eiceCmd.MarkFlagRequired("instance")
	eiceCmd.MarkFlagFilename("pemfile", "pem")
	eiceCmd.MarkFlagsMutuallyExclusive("pemfile", "password")
	// custom completion
	eiceCmd.RegisterFlagCompletionFunc("region", invokeRegionCompletion)
}

func invokeEICECommand(_ *cobra.Command, _ []string) error {
	// check if connector application installed
	connector := connector.DefaultConnector{}
	_, err := connector.IsInstalled()
	if err != nil {
		return err
	}

	// get aws config
	cfg := aws.GetConfig(cpProfileName, cpRegionName)
	ec2api := ec2.NewAPI(cfg)
	ctx := context.Background()

	// check instance exists
	_, err = ec2.IsInstanceExist(ec2api, ctx, cpInstanceId)
	if err != nil {
		return err
	}

	// get instance metadata information
	metadata, err := ec2.GetInstanceMetadataForEICE(ec2api, ctx, cpInstanceId)
	if err != nil {
		return err
	}
	if metadata.State.Name != types.InstanceStateNameRunning {
		return fmt.Errorf("instance %v is %v (status code=%d)", cpInstanceId, metadata.State.Name, *metadata.State.Code)
	}

	// get EC2 Insntance Connect Endpoint information
	var fetchResult *ec2.EICEndpointMetadata
	if eiceEndpointId != "" {
		fetchResult, err = ec2.FetchEICEndpointById(ec2api, ctx, eiceEndpointId)
		if err != nil {
			return err
		}
	} else {
		fetchResult, err = ec2.FetchEICEndpointByVpc(ec2api, ctx, metadata.VpcId)
		if err != nil {
			return err
		}
		fmt.Printf("Find EC2 Instance Connect Endpoint %v in the VPC\n", fetchResult.EndpointId)
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

	// Open WebSocket tunnel with AWS CLI
	wspid, err := ec2instanceconnect.OpenTunnel(cfg, ctx, fetchResult.EndpointId, fetchResult.DnsName, metadata.PrivateIpAddress, localPort, cpPort)
	if err != nil {
		return err
	}
	fmt.Printf("Opening WebSocket tunnel (pid=%v)\n", wspid)
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
	return connectEICEInstance(&connector, wspid)
}

func isAWSCLIInstalled() (bool, error) {
	_, err := exec.LookPath("aws")
	if err != nil {
		return false, errors.New("AWS CLI is not found")
	}
	output, err := exec.Command("aws", "--version").Output()
	if err != nil {
		return false, errors.New("failed to get AWS CLI version")
	}
	cliVersion, err := version.NewVersion(strings.Split(strings.Split(string(output), " ")[0], "/")[1])
	if err != nil {
		return false, errors.New("failed to get AWS CLI version")
	}
	constraint, _ := version.NewConstraint(">=2.12.0")
	result := constraint.Check(cliVersion)
	if !result {
		return false, fmt.Errorf("AWS CLI 2.12.0 later is required (current version=%s)", cliVersion)
	}
	return true, nil
}

func connectEICEInstance(con connector.Connector, wspid int) error {
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
		fmt.Printf("Close WebSocket tunnel (pid=%v)\n", wspid)
		ec2instanceconnect.CloseTunnel(wspid)
	}()
	return nil
}

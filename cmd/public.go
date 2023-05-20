package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stknohg/ec2rdp/internal/aws"
	"github.com/stknohg/ec2rdp/internal/aws/ec2"
	"github.com/stknohg/ec2rdp/internal/connector"
)

var publicNoWait bool

// publicCmd represents the public command
var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Connect to public EC2 instance",
	Long:  `Connect to public EC2 instance`,
	Args: func(cmd *cobra.Command, args []string) error {
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
		return invokePublicCommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)
	publicCmd.Flags().StringVarP(&cpInstanceId, "instance", "i", "", "EC2 instance ID")
	publicCmd.Flags().StringVarP(&cpPemFile, "pemfile", "p", "", ".pem file path")
	publicCmd.Flags().IntVar(&cpPort, "port", 3389, "RDP port no")
	publicCmd.Flags().StringVar(&cpUserName, "user", "Administrator", "RDP username")
	publicCmd.Flags().BoolVarP(&cpUserPassword, "password", "P", false, "RDP passowrd")
	publicCmd.Flags().StringVar(&cpProfileName, "profile", "", "AWS profile name")
	publicCmd.Flags().StringVar(&cpRegionName, "region", "", "AWS region name")
	// original parameters
	publicCmd.Flags().BoolVar(&publicNoWait, "nowait", false, "")
	//
	publicCmd.MarkFlagRequired("instance")
	publicCmd.MarkFlagFilename("pemfile", "pem")
	publicCmd.MarkFlagsMutuallyExclusive("pemfile", "password")
	// custom completion
	publicCmd.RegisterFlagCompletionFunc("region", invokeRegionCompletion)
}

func invokePublicCommand(cmd *cobra.Command, args []string) error {
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

	// get public hostname
	hostName, err := ec2.GetPublicHostName(ec2api, ctx, cpInstanceId)
	if err != nil {
		return err
	}

	// test port is open
	if !isPortOpen(hostName, cpPort) {
		return fmt.Errorf("failed to test TCP connection. (Port=%v)", cpPort)
	}
	fmt.Printf("Remote host %v port %v is open\n", hostName, cpPort)

	// get administrator password
	var password string
	if !cpUserPassword {
		password, err = ec2.GetAdministratorPassword(ec2api, ctx, cpInstanceId, cpPemFile)
		if err != nil {
			return err
		}
		if password == "" {
			return fmt.Errorf("EC2 PasswordData is empty. Use --password flag instead")
		}
		fmt.Println("Administrator password acquisition completed")
	} else {
		password = readPrompt("Enter password:")
	}

	// connect
	connector.HostName = hostName
	connector.Port = cpPort
	connector.UserName = cpUserName
	connector.PlainPassword = password
	connector.WaitFor = !publicNoWait
	return connectPublicInstance(&connector)
}

func connectPublicInstance(con connector.Connector) error {
	err := con.PreConnect()
	if err != nil {
		return err
	}
	err = con.Connect()
	if err != nil {
		return err
	}
	defer con.PostConnect()
	return nil
}

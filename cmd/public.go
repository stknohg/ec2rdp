package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stknohg/ec2rdp/internal/aws"
	"github.com/stknohg/ec2rdp/internal/aws/ec2"
	"github.com/stknohg/ec2rdp/internal/connector"
)

var pubilcInstanceId string
var publicPemFile string
var publicPort int
var publicUserName string
var publicUserPassword string
var publicNoWait bool
var publicProfileName string
var publicRegionName string

// publicCmd represents the public command
var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Connect to public EC2 instance",
	Long:  `Connect to public EC2 instance`,
	Args: func(cmd *cobra.Command, args []string) error {
		if publicPemFile == "" && publicUserPassword == "" {
			return errors.New("--pemfile or --password flag is requied")
		}
		if publicPemFile != "" {
			_, err := os.Stat(publicPemFile)
			if err != nil {
				return errors.New(".pem file does not exist")
			}
		}
		if publicPort < 0 || publicPort > 65535 {
			return errors.New("set port number between 1 and 65535")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return invokePublicCommand(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)
	publicCmd.Flags().StringVarP(&pubilcInstanceId, "instance", "i", "", "EC2 instance ID")
	publicCmd.Flags().StringVarP(&publicPemFile, "pemfile", "p", "", ".pem file path")
	publicCmd.Flags().IntVar(&publicPort, "port", 3389, "RDP port no")
	publicCmd.Flags().StringVar(&publicUserName, "user", "Administrator", "RDP username")
	publicCmd.Flags().StringVarP(&publicUserPassword, "password", "P", "", "RDP passowrd")
	publicCmd.Flags().BoolVar(&publicNoWait, "nowait", false, "")
	publicCmd.Flags().StringVar(&publicProfileName, "profile", "", "AWS profile name")
	publicCmd.Flags().StringVar(&publicRegionName, "region", "", "AWS region name")
	//
	publicCmd.MarkFlagRequired("instance")
	publicCmd.MarkFlagFilename("pemfile", "pem")
	publicCmd.MarkFlagsMutuallyExclusive("pemfile", "password")
}

func invokePublicCommand(cmd *cobra.Command, args []string) error {
	// get aws config
	cfg := aws.GetConfig(publicProfileName, publicRegionName)

	// check instance exists
	_, err := ec2.IsInstanceExist(cfg, pubilcInstanceId)
	if err != nil {
		return err
	}

	// get public hostname
	hostName, err := ec2.GetPublicHostName(cfg, pubilcInstanceId)
	if err != nil {
		return err
	}

	// test port is open
	if !isPortOpen(hostName, publicPort) {
		return fmt.Errorf("failed to test TCP connection. (Port=%v)", publicPort)
	}
	fmt.Printf("Remote host %v port %v is open\n", hostName, publicPort)

	// get administrator password
	var password string
	if publicUserPassword == "" {
		password, err = ec2.GetAdministratorPassword(cfg, pubilcInstanceId, publicPemFile)
		if err != nil {
			return err
		}
		if password == "" {
			return fmt.Errorf("EC2 PasswordData is empty. Use --password flag instead")
		}
		fmt.Println("Administrator password acquisition completed")
	} else {
		password = publicUserPassword
	}

	// connect
	return connectPublicInstance(hostName, publicPort, publicUserName, password, !publicNoWait)
}

func connectPublicInstance(hostName string, port int, userName string, plainPassword string, waitFor bool) error {
	con := connector.DefaultConnector{
		HostName:      hostName,
		Port:          port,
		UserName:      userName,
		PlainPassword: plainPassword,
		WaitFor:       waitFor,
	}
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

package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

// Common parameters
var (
	cpInstanceId   string
	cpPemFile      string
	cpPort         int
	cpUserName     string
	cpUserPassword bool
	cpProfileName  string
	cpRegionName   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "ec2rdp",
	Short:        "Remote Desktop utility for Amazon EC2",
	Long:         `Remote Desktop utility for Amazon EC2.`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// do nothing
}

// Common validations
func validatePemFile(filePath string) error {
	if filePath == "" {
		return errors.New(".pem file path is empty")
	}
	_, err := os.Stat(filePath)
	if err != nil {
		return errors.New(".pem file does not exist")
	}
	return nil
}

func validatePort(portNo int) error {
	if portNo < 1 || portNo > 65535 {
		return errors.New("set port number between 1 and 65535")
	}
	return nil
}

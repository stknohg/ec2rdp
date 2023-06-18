package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cmdVersion string = "0.0.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  `Show version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmdVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

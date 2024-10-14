package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy to an instance",
	Run: func(cmd *cobra.Command, args []string) {
		instanceIP, _ := cmd.Flags().GetString("ip")
		fmt.Printf("Deploying to instance at IP %s\n", instanceIP)
		// Add your deployment logic here
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().String("ip", "", "Instance IP")
}

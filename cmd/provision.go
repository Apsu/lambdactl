package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch new instances",
	Run: func(cmd *cobra.Command, args []string) {
		vmType, _ := cmd.Flags().GetString("type")
		region, _ := cmd.Flags().GetString("region")
		count, _ := cmd.Flags().GetInt("count")
		fmt.Printf("Launching %d instance(s) of type %s in region %s\n", count, vmType, region)
		// Add your provisioning logic here
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)

	launchCmd.Flags().String("type", "gpu_1x_h100_sxm5", "Instance type")
	launchCmd.Flags().String("region", "us-south-2", "Region")
	launchCmd.Flags().Int("count", 1, "Number of instances")
}

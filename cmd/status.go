package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of instances",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Getting status...")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

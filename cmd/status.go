package cmd

import (
	"fmt"

	"lambdactl/pkg/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of instances",
	RunE:  statusFunc,
}

func statusFunc(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")

	client := api.NewAPIClient(viper.GetString("api-url"), viper.GetString("api-key"))

	instances, err := client.ListInstances()
	if err != nil {
		return fmt.Errorf("error listing instance: %v", err)
	}

	var match api.Instance
	for _, instance := range instances {
		if instance.Name == name {
			match = instance
			break
		}
	}

	if match.Name == "" {
		return fmt.Errorf("no matching instance found")
	}

	output, err := yaml.Marshal(match)
	if err != nil {
		return fmt.Errorf("error marshalling instance details: %v", err)
	}
	fmt.Println(string(output))

	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().String("name", "", "Instance name")
	statusCmd.MarkFlagRequired("name")
}

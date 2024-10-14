package cmd

import (
	"fmt"

	"lambdactl/pkg/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	Run:   listFunc,
}

func listFunc(cmd *cobra.Command, args []string) {
	client := api.NewAPIClient(viper.GetString("apiUrl"), viper.GetString("apiKey"))

	instances, err := client.ListInstances()
	if err != nil {
		fmt.Printf("error listing instance: %v", err)
		return
	}

	output, err := yaml.Marshal(instances)
	if err != nil {
		fmt.Printf("error marshalling instance: %v", err)
		return
	}
	fmt.Println(string(output))
}

func init() {
	rootCmd.AddCommand(listCmd)
}

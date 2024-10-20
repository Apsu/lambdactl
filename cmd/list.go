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
	Short: "List all launched instances",
	Run:   listFunc,
}

func listFunc(cmd *cobra.Command, args []string) {
	client := api.NewAPIClient(viper.GetString("api-url"), viper.GetString("api-key"))

	instances, err := client.ListInstances()
	if err != nil {
		fmt.Printf("error listing instance: %v", err)
		return
	}

	// Update some specs from type
	// for index, instance := range instances {
	// 	instanceSpecs := api.ParseInstanceType(instance.InstanceType)
	// 	if err == nil {
	// 		instances[index].InstanceType.Specs = instanceSpecs
	// 	}
	// }

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

package cmd

import (
	"fmt"

	"lambdactl/pkg/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch instance types available to launch",
	Run:   fetchFunc,
}

func fetchFunc(cmd *cobra.Command, args []string) {
	client := api.NewAPIClient(viper.GetString("apiUrl"), viper.GetString("apiKey"))

	instanceOptions, err := client.FetchInstanceOptions()
	if err != nil {
		fmt.Printf("error fetching instance options: %v", err)
		return
	}

	output, err := yaml.Marshal(instanceOptions)
	if err != nil {
		fmt.Printf("error marshalling instance options: %v", err)
		return
	}
	fmt.Println(string(output))
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}

package cmd

import (
	"fmt"

	"lambdactl/pkg/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into an instance",
	RunE:   sshFunc,
}

func sshFunc(cmd *cobra.Command, args []string) error {
	client := api.NewAPIClient(viper.GetString("apiUrl"), viper.GetString("apiKey"))

	err := client.SSHIntoMachine(api.InstanceDetails{
		IP: viper.GetString("ip"),
	})
	if err != nil {
		return fmt.Errorf("error connecting to instance: %v", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.Flags().String("ip", "", "Instance IP")
	viper.BindPFlag("ip", sshCmd.Flags().Lookup("ip"))
	sshCmd.MarkFlagRequired("ip")
}

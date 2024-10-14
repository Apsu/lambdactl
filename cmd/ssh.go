package cmd

import (
	"lambdactl/pkg/sshlib"

	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into an instance",
	RunE:  sshFunc,
}

func sshFunc(cmd *cobra.Command, args []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	user, _ := cmd.Flags().GetString("user")
	keyName, _ := cmd.Flags().GetString("keyName")
	return sshlib.NewShell(host, port, user, keyName)
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.MarkFlagRequired("host")
	sshCmd.Flags().String("host", "", "Hostname or IP")
	sshCmd.Flags().Int("port", 22, "Remote port")
	sshCmd.Flags().String("user", "ubuntu", "Remote user")
	sshCmd.Flags().String("keyName", "id_rsa", "SSH Key Name")
}

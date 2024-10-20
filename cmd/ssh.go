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
	ip, _ := cmd.Flags().GetString("ip")
	port, _ := cmd.Flags().GetInt("port")
	user, _ := cmd.Flags().GetString("user")
	keyName, _ := cmd.Flags().GetString("keyName")
	return sshlib.NewShell(ip, port, user, keyName)
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.MarkFlagRequired("ip")
	sshCmd.Flags().String("ip", "", "Target IP")
	sshCmd.Flags().Int("port", 22, "Remote port")
	sshCmd.Flags().String("user", "ubuntu", "Remote user")
	sshCmd.Flags().String("keyName", "id_rsa", "SSH Key Name")
}

package device

import (
	"github.com/spf13/cobra"
)

//CredentialsGroupCmd provides grouping for Credentials commands
func CredentialsGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Commands to manage credentials of the devices",
	}
	cmd.AddCommand(CredentialsUpdateCommand)
	return cmd
}

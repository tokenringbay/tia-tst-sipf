package device

import (
	"github.com/spf13/cobra"
)

//NewGroupCmd provides grouping for device commands
func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device",
		Short: "Device commands",
	}
	cmd.AddCommand(CredentialsGroupCmd())
	return cmd
}

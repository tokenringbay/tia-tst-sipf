package settings

import (
	"github.com/spf13/cobra"
)

//NewGroupCmd provides grouping of Fabric Settings attributes
func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setting",
		Short: "IP Fabric setting commands",
	}
	cmd.AddCommand(ShowCommand)
	cmd.AddCommand(UpdateCommand)

	return cmd
}

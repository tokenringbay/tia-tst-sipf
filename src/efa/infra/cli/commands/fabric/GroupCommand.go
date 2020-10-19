package fabric

import (
	"efa/infra/cli/commands/fabric/settings"
	"github.com/spf13/cobra"
)

//NewGroupCmd provides grouping for configure/deconfigure commands
func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fabric",
		Short: "Fabric commands",
	}
	cmd.AddCommand(ConfigureSwitchCommand)
	cmd.AddCommand(DeconfigureSwitchCommand)
	cmd.AddCommand(settings.NewGroupCmd())
	cmd.AddCommand(ShowFabricConfigCommand)
	cmd.AddCommand(ShowFabricCommand)
	return cmd
}

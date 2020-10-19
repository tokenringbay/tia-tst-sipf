package execution

import (
	"github.com/spf13/cobra"
)

//NewGroupCmd for grouping Execution commands
func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execution",
		Short: "Execution commands",
	}
	cmd.AddCommand(ShowCommand)
	return cmd
}

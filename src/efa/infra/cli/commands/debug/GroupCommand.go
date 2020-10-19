package debug

import (
	"github.com/spf13/cobra"
)

//NewGroupCmd groups Debug Commands
func NewGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug commands",
	}
	cmd.AddCommand(ClearConfigCommand)
	return cmd
}

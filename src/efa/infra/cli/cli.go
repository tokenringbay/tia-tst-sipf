package cli

import (
	"efa/infra/cli/commands"
	"efa/infra/cli/commands/debug"
	"efa/infra/cli/commands/device"
	"efa/infra/cli/commands/execution"
	"efa/infra/cli/commands/fabric"
	"efa/infra/constants"
	"github.com/spf13/cobra"
)

//StartCLI Starts the CLI Environment
func StartCLI() {

	rootCmd := GetRootCommand()

	rootCmd.Execute()
}

//GetRootCommand provides access to all Root commands
func GetRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{Use: constants.ApplicationName}
	rootCmd.AddCommand(fabric.NewGroupCmd())
	rootCmd.AddCommand(execution.NewGroupCmd())
	rootCmd.AddCommand(debug.NewGroupCmd())
	rootCmd.AddCommand(commands.ShowVersionCommand)
	rootCmd.AddCommand(commands.SupportSaveCommand)
	rootCmd.AddCommand(device.NewGroupCmd())
	return rootCmd
}

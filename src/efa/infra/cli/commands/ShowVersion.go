package commands

import (
	"efa/infra/cli/utils"

	"fmt"
	"github.com/spf13/cobra"

	"efa/infra"
)

//ShowVersionCommand provides commands to list execution status
var ShowVersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Display version of the application",
	RunE:  utils.TimedRunE(runShowVersion),
}

func init() {
}

func runShowVersion(cmd *cobra.Command, args []string) error {
	fmt.Println("Version :", infra.Version)
	fmt.Println("Time Stamp:", infra.BuildStamp)
	return nil
}

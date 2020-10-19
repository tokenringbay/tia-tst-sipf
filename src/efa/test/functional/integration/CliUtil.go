package integration

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)

func emptyRun(*cobra.Command, []string) {}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func resetCommandLineFlagSet() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
}

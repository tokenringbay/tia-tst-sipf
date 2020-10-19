package integration

import (
	"efa/infra/cli"

	"testing"

	"efa/infra/constants"

	"github.com/stretchr/testify/assert"
)

func TestTopHelp(t *testing.T) {

	rootCmd := cli.GetRootCommand()
	output, err := executeCommand(rootCmd, "--help")

	assert.Nil(t, err)
	assert.Contains(t, output, "fabric      Fabric commands")
	assert.Contains(t, output, "execution   Execution commands")
	assert.Contains(t, output, "debug       Debug commands")
	assert.Contains(t, output, "help        Help about any command")

}

func TestFabricHelp(t *testing.T) {
	rootCmd := cli.GetRootCommand()
	output, err := executeCommand(rootCmd, "fabric", "--help")

	assert.Nil(t, err)
	assert.Contains(t, output, "configure   Configure IP Fabric on the device")
	assert.Contains(t, output, "deconfigure Deconfigure IP Fabric from the device")
	assert.Contains(t, output, "setting     IP Fabric setting commands")
	assert.Contains(t, output, "show-config Display IP Fabric config")
}

func TestExecutionHelp(t *testing.T) {
	rootCmd := cli.GetRootCommand()
	output, err := executeCommand(rootCmd, "execution", "--help")

	assert.Nil(t, err)
	assert.Contains(t, output, constants.ApplicationName+" execution [command] --help")

}
func TestFabricSettingsHelp(t *testing.T) {
	rootCmd := cli.GetRootCommand()
	output, err := executeCommand(rootCmd, "fabric", "setting", "--help")

	assert.Nil(t, err)
	assert.Contains(t, output, "show        Display IP Fabric setting")
	assert.Contains(t, output, "update      Update fabric settings.")
}

func TestShowHelp(t *testing.T) {
	rootCmd := cli.GetRootCommand()
	output, err := executeCommand(rootCmd, "fabric", "show-config", "--help")

	assert.Nil(t, err)
	assert.Contains(t, output, "Display IP Fabric config")
}

func TestDebugHelp(t *testing.T) {
	rootCmd := cli.GetRootCommand()
	output, err := executeCommand(rootCmd, "debug", "--help")

	assert.Nil(t, err)
	assert.Contains(t, output, "clear-config Clear configuration from device")
}

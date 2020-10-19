package integration

import (
	"fmt"
	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
	"testing"

	"efa-server/infra/constants"
	"efa-server/test/functional"
	_ "efa-server/test/functional"
	"efa-server/test/testutils"
	"strings"
)

var Password = functional.DeviceAdminPassword

var Spine1IP = functional.IntegrationTestSpine1IP
var Leaf1IP = functional.IntegrationTestLeaf1IP
var Leaf2IP = functional.IntegrationTestLeaf2IP

//var Leaf3IP = functional.IntegrationTestLeaf3IP
//var Leaf4IP = functional.IntegrationTestLeaf4IP
//var Leaf5IP = functional.IntegrationTestLeaf4IP
//var Leaf6IP = functional.IntegrationTestLeaf4IP

//Main Function
func TestMain(m *testing.M) {
	testutils.TestMutex.Lock()
	defer testutils.TestMutex.Unlock()
	func() {
		c1 := testcli.Command(constants.ApplicationName, "fabric", "setting", "update", "--fabric-type", "clos")
		c1.Run()
		output1 := c1.Stdout()
		fmt.Println(output1)

		c := testcli.Command(constants.ApplicationName, "debug", "clear-config", "--device",
			fmt.Sprintf("%s,%s,%s", Spine1IP, Leaf1IP, Leaf2IP), "--username", "admin", "--password", Password)
		c.Run()
		output := c.Stdout()
		fmt.Println(output)
	}()
	m.Run()

	//Clean all devices after completing testing
	func() {
		cleanup()
	}()

}
func cleanup() {
	func() {

		c := testcli.Command(constants.ApplicationName, "fabric", "show")
		c.Run()
		fabricShowoutput := c.Stdout()
		fmt.Println(fabricShowoutput)

		if strings.Contains(fabricShowoutput, Spine1IP) {
			c = testcli.Command(constants.ApplicationName, "fabric", "deconfigure", "--device", Spine1IP)
			c.Run()
			output := c.Stdout()
			fmt.Println(output)
		}
		if strings.Contains(fabricShowoutput, Leaf1IP) {
			c = testcli.Command(constants.ApplicationName, "fabric", "deconfigure", "--device", Leaf1IP)
			c.Run()
			output := c.Stdout()
			fmt.Println(output)
		}
		if strings.Contains(fabricShowoutput, Leaf2IP) {
			c = testcli.Command(constants.ApplicationName, "fabric", "deconfigure", "--device", Leaf2IP)
			c.Run()
			output := c.Stdout()
			fmt.Println(output)
		}

	}()
}

//TestCLOSCLICollection so that we can run in sequence
func TestCLOSCLICollection(t *testing.T) {

	t.Run("TA_TestConfigure_Delete_NonExisting_Spine", TConfigureDeleteNonExistingSpine)
	t.Run("TB_TestConfigure_Single_Spine_Wrong_Credentials", TConfigureSingleSpineWrongCredentials)
	t.Run("TC_TestConfigure_Single_Spine", TConfigureSingleSpine)
	t.Run("TD_TestConfigure_Single_Leaf", TConfigureSingleLeaf)
	t.Run("TE_TestConfigure", TConfigure)
	t.Run("TF_TestConfigure_Add_Leaves_Then_Spine", TConfigureAddLeavesThenSpine)
	t.Run("TG_TestConfigure_Add_Spine_Then_Leaves", TConfigureAddSpineThenLeaves)

}

//Delete only one Spine which doesn't exist
func TConfigureDeleteNonExistingSpine(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "deconfigure", "--device", Spine1IP)
	c.Run()
	output := c.Stdout()

	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Deletion of  device(s) with ip-address = %s [Failed]", Spine1IP))
	assert.Contains(t, output, fmt.Sprintf("Switch %s doesn't exist in the fabric", Spine1IP))
}

//Configure only one Spine with wrong credentials
func TConfigureSingleSpineWrongCredentials(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure", "--spine", Spine1IP, "--username", "admin", "--password", "pass1234")
	c.Run()
	output := c.Stdout()

	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Addition of Spine device with ip-address = %s [Failed]", Spine1IP))
	assert.Contains(t, output, fmt.Sprintf("Switch %s connection Failed", Spine1IP))
	assert.Contains(t, output, "Add Device(s) [Failed]")
}

//Configure only one Spine
func TConfigureSingleSpine(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure", "--spine", Spine1IP,
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	fmt.Println(output)
	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Addition of Spine device with ip-address = %s [Succeeded]", Spine1IP))
	assert.Contains(t, output, "Validate Fabric [Failed]")
	assert.Contains(t, output, "No Leaf Devices")
}

//Configure only one Leaf
func TConfigureSingleLeaf(t *testing.T) {
	cleanup()

	c := testcli.Command(constants.ApplicationName, "fabric", "configure", "--leaf", Leaf1IP,
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()

	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf1IP))
	assert.Contains(t, output, "Validate Fabric [Failed]")
	assert.Contains(t, output, "No Spine Devices")
}

//Configure Three nodes
func TConfigure(t *testing.T) {

	cleanup()

	c := testcli.Command(constants.ApplicationName, "fabric", "configure", "--spine", Spine1IP,
		"--leaf", fmt.Sprintf("%s,%s", Leaf1IP, Leaf2IP),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	fmt.Println(output)

	showconfig := testcli.Command(constants.ApplicationName, "fabric", "show-config")
	showconfig.Run()
	fmt.Println(showconfig.Stdout())

	assert.Nil(t, c.Error())
	assert.Contains(t, output, fmt.Sprintf("Addition of Spine device with ip-address = %s [Succeeded]", Spine1IP))
	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf1IP))
	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf2IP))
	assert.Contains(t, output, "Validate Fabric [Success]")
	assert.Contains(t, output, "Configure Fabric [Success]")

}

func TConfigureAddLeavesThenSpine(t *testing.T) {

	cleanup()

	//Add leaves first
	c := testcli.Command(constants.ApplicationName, "fabric", "configure", "--leaf",
		fmt.Sprintf("%s,%s", Leaf1IP, Leaf2IP),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()

	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf1IP))
	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf2IP))
	assert.Contains(t, output, "Validate Fabric [Failed]")
	assert.Contains(t, output, "No Spine Devices")

	//Add Spines Next
	c = testcli.Command(constants.ApplicationName, "fabric", "configure", "--spine", Spine1IP,
		"--username", "admin", "--password", Password)
	c.Run()
	output = c.Stdout()

	assert.Nil(t, c.Error())
	assert.Contains(t, output, fmt.Sprintf("Addition of Spine device with ip-address = %s [Succeeded]", Spine1IP))
	assert.Contains(t, output, "Validate Fabric [Success]")
	assert.Contains(t, output, "Configure Fabric [Success]")
}

func TConfigureAddSpineThenLeaves(t *testing.T) {

	cleanup()

	//Add Spines first
	c := testcli.Command(constants.ApplicationName, "fabric", "configure", "--spine", Spine1IP,
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Addition of Spine device with ip-address = %s [Succeeded]", Spine1IP))
	assert.Contains(t, output, "Validate Fabric [Failed]")
	assert.Contains(t, output, "No Leaf Devices")

	//Add leaves Next
	c = testcli.Command(constants.ApplicationName, "fabric", "configure", "--leaf",
		fmt.Sprintf("%s,%s", Leaf1IP, Leaf2IP),
		"--username", "admin", "--password", Password)
	c.Run()
	output = c.Stdout()

	assert.Nil(t, c.Error())
	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf1IP))
	assert.Contains(t, output, fmt.Sprintf("Addition of Leaf device with ip-address = %s [Succeeded]", Leaf2IP))
	assert.Contains(t, output, "Validate Fabric [Success]")
	assert.Contains(t, output, "Configure Fabric [Success]")
}

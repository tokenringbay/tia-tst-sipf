package nonclos

import (
	"fmt"
	"github.com/rendon/testcli"
	"github.com/stretchr/testify/assert"
	"testing"

	"efa-server/infra/constants"
	"efa-server/test/functional"
	_ "efa-server/test/functional"
	"efa-server/test/testutils"
)

var Password = functional.DeviceAdminPassword

var Rack1IP1 = functional.IntegrationNonClosTestRack1IP1
var Rack1IP2 = functional.IntegrationNonClosTestRack1IP2
var Rack2IP1 = functional.IntegrationNonClosTestRack2IP1
var Rack2IP2 = functional.IntegrationNonClosTestRack2IP2
var Rack3IP1 = functional.IntegrationNonClosTestRack3IP1
var Rack3IP2 = functional.IntegrationNonClosTestRack3IP2
var Rack4IP1 = functional.IntegrationNonClosTestRack4IP1
var Rack4IP2 = functional.IntegrationNonClosTestRack4IP2
var Rack5IP1 = "1.1.1.1"
var Rack5IP2 = "2.2.2.2"

//Main Function
func TestMain(m *testing.M) {
	testutils.TestMutex.Lock()
	defer testutils.TestMutex.Unlock()
	func() {

		c := testcli.Command(constants.ApplicationName, "debug", "clear-config", "--device",
			fmt.Sprintf("%s,%s,%s,%s", Rack3IP1, Rack3IP2, Rack4IP1, Rack4IP2), "--username", "admin", "--password", Password)
		c.Run()
		output := c.Stdout()
		fmt.Println(output)
	}()
	m.Run()
	//At the end re-set to CLOS
	func() {

		c := testcli.Command(constants.ApplicationName, "fabric", "setting", "update", "--fabric-type", "clos")
		c.Run()
		output := c.Stdout()
		fmt.Println(output)
	}()
}

//TestNonCLOSCLICollection so that we can run in sequence
func TestNonCLOSCLICollection(t *testing.T) {

	t.Run("TA_TestNonClosConfigureFabricType", TNonClosConfigureFabricType)
	t.Run("TB_TestNonClosConfigure5Racks", TNonClosConfigure5Racks)
	t.Run("TC_TestNonClosDeconfigure5Racks", TNonClosDeconfigure5Racks)
	t.Run("TD_TestNonClosConfigureRacks", TNonClosConfigureRacks)
	t.Run("TE_TestNonClosAdd3RacksInExistingFabric", TNonClosAdd3RacksInExistingFabric)
	t.Run("TF_TestNonClosReConfigureExistingRack", TNonClosReConfigureExistingRack)
	t.Run("TG_TestNonClosConfigureRackWithAlreadyExistingIP", TNonClosConfigureRackWithAlreadyExistingIP)

	t.Run("TI_TestNonClosDeconfigureInvalidRack", TNonClosDeconfigureInvalidRack)
	t.Run("TJ_TestNonClosDeconfigureOneRack", TNonClosDeconfigureOneRack)
	t.Run("TK_TestNonClosVerifyFabricShow", TNonClosVerifyFabricShow)
	t.Run("TL_TestNonClosConfigureOneRack", TNonClosConfigureOneRack)
	t.Run("TM_TestNonClosDeconfigureRacks", TNonClosDeconfigureRacks)

}

//Configure Fabric-type as non-clos
func TNonClosConfigureFabricType(t *testing.T) {
	FabricType := "non-clos"

	c := testcli.Command(constants.ApplicationName, "fabric", "setting", "update", "--fabric-type", FabricType)
	c.Run()
	assert.Nil(t, c.Error())

	sc := testcli.Command(constants.ApplicationName, "fabric", "setting", "show")
	sc.Run()
	showoutput := sc.Stdout()
	assert.Nil(t, sc.Error())

	assert.Contains(t, showoutput, fmt.Sprintf("Fabric Type               | %s", FabricType))
}

//Configure 5 Racks in a Non-Clos Fabric
func TNonClosConfigure5Racks(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure",
		"--rack", fmt.Sprintf("%s,%s", Rack1IP1, Rack1IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack2IP1, Rack2IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack4IP1, Rack4IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack5IP1, Rack5IP2),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stderr()

	assert.Nil(t, c.Error())
	assert.Contains(t, output, "Error: Only 4 Rack Pairs are supported")
}

//Deconfigure 5 Racks in a Non-Clos Fabric
func TNonClosDeconfigure5Racks(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "deconfigure",
		"--rack", fmt.Sprintf("%s,%s", Rack1IP1, Rack1IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack2IP1, Rack2IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack4IP1, Rack4IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack5IP1, Rack5IP2))
	c.Run()
	output := c.Stderr()

	assert.Nil(t, c.Error())
	assert.Contains(t, output, "Error: Only 4 Rack Pairs are supported")
}

//Configure more than one Rack in a Non-Clos Fabric
func TNonClosConfigureRacks(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure",
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack4IP1, Rack4IP2),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	Rack3 := fmt.Sprintln(Rack3IP1, ",", Rack3IP2)
	assert.Contains(t, output, fmt.Sprintf("Addition of Rack device with ip-address = %s [Succeeded]", Rack3))
	Rack4 := fmt.Sprintln(Rack4IP1, ",", Rack4IP2)
	assert.Contains(t, output, fmt.Sprintf("Addition of Rack device with ip-address = %s [Succeeded]", Rack4))
	assert.Contains(t, output, "Validate Fabric [Success]")
	assert.Contains(t, output, "Configure Fabric [Success]")
}

//Add 3 Racks to an existing Non-Clos Fabric
func TNonClosAdd3RacksInExistingFabric(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure",
		"--rack", fmt.Sprintf("%s,%s", Rack1IP1, Rack1IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack2IP1, Rack2IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack5IP1, Rack5IP2),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	assert.Contains(t, output, "Maximum Configurable Racks cannot be more than 4.")
}

//Reconfigure an existing Rack in a Non-Clos Fabric
func TNonClosReConfigureExistingRack(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure",
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	Rack3 := fmt.Sprintln(Rack3IP1, ",", Rack3IP2)
	assert.Contains(t, output, fmt.Sprintf("Addition of Rack device with ip-address = %s [Succeeded]", Rack3))
	assert.Contains(t, output, "Validate Fabric [Success]")
	assert.Contains(t, output, "Configure Fabric [Success]")
}

//Configure a Rack with an IP from existing Rack in a Non-Clos Fabric
func TNonClosConfigureRackWithAlreadyExistingIP(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure",
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack5IP2),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Addition of  device with ip-address = %s [Failed]", Rack3IP1))
	assert.Contains(t, output, fmt.Sprintf("%s already present in another Rack", Rack3IP1))
}

//Deonfigure a rack which is not to a Non-Clos Fabric
func TNonClosDeconfigureInvalidRack(t *testing.T) {

	Rack := fmt.Sprintf("%s,%s", Rack5IP1, Rack5IP2)
	c := testcli.Command(constants.ApplicationName, "fabric", "deconfigure",
		"--rack", Rack)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	RackStr := fmt.Sprintf("%s\n", Rack)
	assert.Contains(t, output, fmt.Sprintf("Deletion of Rack device(s) with ip-address =  %s [Failed]", RackStr))
	assert.Contains(t, output, fmt.Sprintf("Invalid Rack IP Pairs Found  %s", Rack))
}

//Deconfigure a Rack in a Non-Clos Fabric
func TNonClosDeconfigureOneRack(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "deconfigure",
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2))
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Deletion of  device with ip-address = %s [Succeeded]", Rack3IP1))
	assert.Contains(t, output, fmt.Sprintf("Deletion of  device with ip-address = %s [Succeeded]", Rack3IP2))
}

//Verify the deconfigured rack present in the Non-clos Fabric
func TNonClosVerifyFabricShow(t *testing.T) {
	c := testcli.Command(constants.ApplicationName, "fabric", "show")
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	assert.NotContains(t, output, Rack3IP1)
	assert.NotContains(t, output, Rack3IP2)
}

//Configure a Rack in a Non-Clos Fabric
func TNonClosConfigureOneRack(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "configure",
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2),
		"--username", "admin", "--password", Password)
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	Rack3 := fmt.Sprintln(Rack3IP1, ",", Rack3IP2)
	assert.Contains(t, output, fmt.Sprintf("Addition of Rack device with ip-address = %s [Succeeded]", Rack3))
	assert.Contains(t, output, "Validate Fabric [Success]")
	assert.Contains(t, output, "Configure Fabric [Success]")
}

//Deconfigure all the Racks in a Non-Clos Fabric
func TNonClosDeconfigureRacks(t *testing.T) {

	c := testcli.Command(constants.ApplicationName, "fabric", "deconfigure",
		"--rack", fmt.Sprintf("%s,%s", Rack3IP1, Rack3IP2),
		"--rack", fmt.Sprintf("%s,%s", Rack4IP1, Rack4IP2))
	c.Run()
	output := c.Stdout()
	assert.Nil(t, c.Error())

	assert.Contains(t, output, fmt.Sprintf("Deletion of  device with ip-address = %s [Succeeded]", Rack3IP1))
	assert.Contains(t, output, fmt.Sprintf("Deletion of  device with ip-address = %s [Succeeded]", Rack3IP2))
	assert.Contains(t, output, fmt.Sprintf("Deletion of  device with ip-address = %s [Succeeded]", Rack4IP1))
	assert.Contains(t, output, fmt.Sprintf("Deletion of  device with ip-address = %s [Succeeded]", Rack4IP2))
}

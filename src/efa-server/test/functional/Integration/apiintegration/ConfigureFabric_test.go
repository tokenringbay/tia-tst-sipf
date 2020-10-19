package apiintegration

import (
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	//ad "efa-server/infra/device/adapter"
	//"efa-server/infra/device/client"
	"efa-server/test/functional"
	"efa-server/usecase"
	"fmt"
	//"os"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"

	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	"github.com/rifflock/lfshook"
	"io/ioutil"
)

var (
	UserName   = "admin"
	FabricName = constants.DefaultFabric
	Password   = functional.DeviceAdminPassword

	Spine1IP = functional.IntegrationTestSpine1IP
	Leaf1IP  = functional.IntegrationTestLeaf1IP
	Leaf2IP  = functional.IntegrationTestLeaf2IP
)

func init() {
	pathMap := lfshook.PathMap{
		log.InfoLevel:  "/var/log/esa_action_test_info.log",
		log.ErrorLevel: "/var/log/esa_action_test_error.log",
	}

	log.AddHook(lfshook.NewHook(
		pathMap,
		&log.JSONFormatter{},
	))
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.TextFormatter{DisableColors: true})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	//log.SetOutput(os.Stdout)
	log.SetOutput(ioutil.Discard)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
func TestConfigure_Switches(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)

	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear

	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	//Configure
	resp1, err := devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	fmt.Println(err, resp1)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

	//TODO Validation of the switch to see if they are indeed deleted Similat to ValidateFabricFetchResponse

}

func TestConfigure_Switches_Force_Conflict_OverlayGateway(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)

	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear

	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Create OverlayGateway with to simulate Failure
	netconfClient := &client.NetconfClient{Host: Leaf1IP, User: UserName, Password: Password}
	netconfClient.Login()
	defer netconfClient.Close()

	//Create Overlay Gateway
	detail, _ := ad.GetDeviceDetail(netconfClient)
	adapter := ad.GetAdapter(detail.Model)
	msg, err := adapter.CreateOverlayGateway(netconfClient, "OVG_FAIL", "layer2-extension",
		"1", "true")
	fmt.Println(msg)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	netconfClient2 := &client.NetconfClient{Host: Leaf2IP, User: UserName, Password: Password}
	netconfClient2.Login()
	defer netconfClient2.Close()

	//Create Overlay Gateway
	detail2, _ := ad.GetDeviceDetail(netconfClient2)
	adapter2 := ad.GetAdapter(detail2.Model)
	msg2, err2 := adapter2.CreateOverlayGateway(netconfClient2, "OVG_FAIL", "layer2-extension",
		"1", "true")
	fmt.Println(msg2)
	assert.Equal(t, "<ok/>", msg2, "")
	assert.Nil(t, err2, "")

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	//Configure without force should fail
	resp1, err := devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	fmt.Println(err, resp1)
	assert.Error(t, err)

	//Configure with force should pass
	resp, err = devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, true)
	assert.NoError(t, err)

	resp1, err = devUC.ConfigureFabric(context.Background(), FabricName, true, false)
	fmt.Println(err, resp1)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

}

func TestConfigure_Switches_Force_Conflict_AnycastGateway(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)

	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear

	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Create Anycast Gateway  to simulate Failure
	netconfClient := &client.NetconfClient{Host: Leaf1IP, User: UserName, Password: Password}
	netconfClient.Login()
	defer netconfClient.Close()

	//Create Anycast Gateway
	detail, _ := ad.GetDeviceDetail(netconfClient)
	adapter := ad.GetAdapter(detail.Model)
	msg, err := adapter.ConfigureAnycastGateway(netconfClient, "0201.0101.0104", "0201.0101.0105")
	fmt.Println(msg)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	//Configure without force should fail
	resp1, err := devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	assert.Error(t, err)

	//Configure with force should pass
	resp, err = devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, true)
	assert.NoError(t, err)

	resp1, err = devUC.ConfigureFabric(context.Background(), FabricName, true, false)
	fmt.Println("err", err, "resp1", resp1)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

}
func TestConfigure_Switches_Force_Conflict_EVPN(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)

	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear

	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Create EVPN with to simulate Failure
	netconfClient := &client.NetconfClient{Host: Leaf1IP, User: UserName, Password: Password}
	netconfClient.Login()
	defer netconfClient.Close()

	//Create EVPN
	detail, _ := ad.GetDeviceDetail(netconfClient)
	adapter := ad.GetAdapter(detail.Model)
	msg, err := adapter.CreateEvpnInstance(netconfClient, "EVPN_FAIL", "5", "3")

	fmt.Println(msg)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	resp1, err := devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	fmt.Println(err, resp1)
	assert.Error(t, err)

	//Configure with force should pass
	resp, err = devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, true)
	fmt.Println(resp)
	assert.NoError(t, err)

	resp1, err = devUC.ConfigureFabric(context.Background(), FabricName, true, false)
	fmt.Println(err, resp1)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

}

func TestConfigure_Switches_Force_Conflict_LinkInterfaceNumbered(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)

	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear

	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Create EVPN with to simulate Failure
	netconfClient := &client.NetconfClient{Host: Leaf1IP, User: UserName, Password: Password}
	netconfClient.Login()
	defer netconfClient.Close()

	//Create numbered IP
	detail, _ := ad.GetDeviceDetail(netconfClient)
	adapter := ad.GetAdapter(detail.Model)
	ifType := "ethernet"
	ifName := functional.IntegrationTestLeaf2IPLink1
	if intfMap, err := adapter.GetInterface(netconfClient, ifType, ifName); err == nil {
		//fmt.Println("Clear", client.Host, ifType, ifName, intfMap)
		log.Infoln("Delete interface ", ifType, ifName)
		if intfMap["address"] != "" {
			//Address is populated for Numbered interfaces
			adapter.UnconfigureInterfaceNumbered(netconfClient, ifType, ifName, intfMap["address"])
		}
		if intfMap["donor_type"] == "loopback" {
			adapter.UnconfigureInterfaceUnnumbered(netconfClient, ifType, ifName)
		}
	}
	msg, err := adapter.ConfigureInterfaceNumbered(netconfClient, ifType, ifName,
		"123.3.3.3/31", "Test IP")

	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.Error(t, err)

	//Configure with force should pass
	resp, err = devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, true)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})
	assert.NoError(t, err)

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	resp1, err := devUC.ConfigureFabric(context.Background(), FabricName, true, false)
	fmt.Println(err, resp1)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

}

func TestConfigure_Fabric_Single_Leaf(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)
	LeafIPList := []string{Leaf1IP}

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, []string{}, UserName, Password, false)
	assert.NoError(t, err)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	//Validate Fabric Topology
	vResp, err := devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.Equal(t, true, vResp.NoSpines)
	assert.NoError(t, err)
}

func TestConfigure_Fabric_Single_Spine(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)
	SpineIPList := []string{Spine1IP}

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})
	//Validate Fabric Topology
	vResp, err := devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.Equal(t, true, vResp.NoLeaves)
	assert.NoError(t, err)
}

func TestConfigure_Fabric_AddSpine_Then_Leaves(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)
	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear
	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Phase-1 -- Add Spines
	//Add Spines
	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	vResp, err := devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.Equal(t, true, vResp.NoLeaves)
	assert.NoError(t, err)

	//Phase-2 -- Add Leaves
	//Add Leaves
	resp, err = devUC.AddDevices(context.Background(), FabricName, LeafIPList, []string{}, UserName, Password, false)
	assert.NoError(t, err)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})

	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	//Configure
	_, err = devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

	//TODO Validation of the switch to see if they are indeed deleted Similat to ValidateFabricFetchResponse

}

func TestConfigure_Fabric_AddSpine_Then_Same_Node_as_Leaf(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)
	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear
	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Phase-1 -- Add node as Spine
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)

	//Phase-2 -- Add same node as Leaf
	resp, err = devUC.AddDevices(context.Background(), FabricName, SpineIPList, []string{}, UserName, Password, false)
	assert.Equal(t, fmt.Sprintln(Spine1IP, "already configured as Spine"), resp[0].Errors[0].Error())

}

func TestConfigure_Fabric_AddLeaf_Then_Same_Node_as_Spine(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)
	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear
	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Phase-1 -- Add node as Leaf
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)

	//Phase-2 -- Add same node as Leaf
	resp, err = devUC.AddDevices(context.Background(), FabricName, []string{}, LeafIPList, UserName, Password, false)
	assert.Equal(t, fmt.Sprintln(Leaf1IP, "already configured as Leaf"), resp[0].Errors[0].Error())
	assert.Equal(t, fmt.Sprintln(Leaf2IP, "already configured as Leaf"), resp[1].Errors[0].Error())

}
func TestConfigure_Fabric_AddLeaves_Then_Spines(t *testing.T) {
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)
	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear
	err := devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Phase-1 -- Add Leaves
	//Add Leaves
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, []string{}, UserName, Password, false)
	assert.NoError(t, err)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	vResp, err := devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.Equal(t, true, vResp.NoSpines)
	assert.NoError(t, err)

	//Phase-2 -- Add Spines
	//Add Spines
	resp, err = devUC.AddDevices(context.Background(), FabricName, []string{}, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	//Configure
	_, err = devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, false)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})

	//TODO Validation of the switch to see if they are indeed deleted Similat to ValidateFabricFetchResponse
}

func TestConfigure_Switches_Unnumbered_interface(t *testing.T) {
	var FabricUpdateReques = domain.FabricProperties{
		P2PIPType: domain.P2PIpTypeUnnumbered,
	}
	database.Setup(constants.TESTDBLocation + "Integration")
	defer cleanupDB(database.GetWorkingInstance())

	devUC := infra.GetUseCaseInteractor()
	devUC.AddFabric(context.Background(), FabricName)

	_, _, err := devUC.UpdateFabricProperties(context.Background(), FabricName, &FabricUpdateReques)
	assert.NoError(t, err)
	LeafIPList := []string{Leaf1IP, Leaf2IP}
	SpineIPList := []string{Spine1IP}
	DeviceIPList := []string{Spine1IP, Leaf1IP, Leaf2IP}

	//First Clear Configuration on Devices
	//efa debug clear

	err = devUC.AddDevicesAndClearFabric(context.Background(), DeviceIPList, DeviceIPList, UserName, Password)
	assert.NoError(t, err)

	//Add Devices
	resp, err := devUC.AddDevices(context.Background(), FabricName, LeafIPList, SpineIPList, UserName, Password, false)
	assert.NoError(t, err)
	//Validate the Response from AddDevices call
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, Role: usecase.LeafRole, FabricID: devUC.FabricID})

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, Role: usecase.SpineRole, FabricID: devUC.FabricID})

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), FabricName)
	assert.NoError(t, err)

	//Configure
	_, err = devUC.ConfigureFabric(context.Background(), FabricName, false, false)
	assert.NoError(t, err)

	ValidateFabricFetchResponse(t, devUC, FabricName, true)

	//Delete Case
	delResp, err := devUC.DeleteDevicesFromFabric(context.Background(), FabricName, DeviceIPList, UserName, Password, false, true, false)
	assert.NoError(t, err)
	//Validate the Response from DeleteDevicesFromFabric call
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf1IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Leaf2IP, FabricID: devUC.FabricID})
	assert.Contains(t, delResp, usecase.AddDeviceResponse{FabricName: FabricName, IPAddress: Spine1IP, FabricID: devUC.FabricID})
}

func ValidateFabricFetchResponse(t *testing.T, interactor *usecase.DeviceInteractor, FabricName string, unnumbered bool) {

	FabricFetchResponse, err := interactor.FetchFabricConfigs(context.Background(), FabricName, "all")
	SwitchConfigs := interactor.GetSwitchConfigs(context.Background(), FabricName)

	SwitchConfigIPMap := make(map[string]domain.SwitchConfig, 0)
	FabricFetchResponseMap := make(map[string]operation.ConfigSwitchResponse, 0)
	//Prepare SwitchConfig Map
	for _, sw := range SwitchConfigs {
		SwitchConfigIPMap[sw.DeviceIP] = sw
	}
	for _, fs := range FabricFetchResponse.SwitchResponse {
		FabricFetchResponseMap[fs.Host] = fs
	}

	assert.NoError(t, err)
	assert.Equal(t, FabricName, FabricFetchResponse.FabricName)
	assert.Equal(t, len(FabricFetchResponse.SwitchResponse), len(SwitchConfigs))

	//For each Switch in DB match the config on the Device
	for _, sw := range SwitchConfigs {
		fs, ok := FabricFetchResponseMap[sw.DeviceIP]
		assert.Equal(t, ok, true)
		assert.Equal(t, sw.DeviceIP, fs.Host)
		assert.Equal(t, sw.Role, fs.Role)
		assert.Equal(t, sw.LoopbackIP, fs.RouterID)

		//BGP Switch Configs
		assert.NotNil(t, fs.Bgp)
		assert.NotNil(t, fs.Bgp.PeerGroups)
		assert.NotNil(t, fs.Bgp.L2VPN)
		bgpConfigs := interactor.GetBGPSwitchConfigs(context.Background(), interactor.FabricID, sw.DeviceID)
		assert.Equal(t, sw.LocalAS, fs.Bgp.LocalAS)
		assert.Equal(t, interactor.FabricProperties.MaxPaths, fs.Bgp.MaxPaths)

		//TODO Interface Configs
		if sw.Role == usecase.LeafRole {
			assert.NotNil(t, fs.Ovg)
			assert.NotNil(t, fs.Evpn)
			//Verify Overlay Gateway
			assert.Equal(t, *fs.Ovg, operation.ConfigOVGResponse{Name: FabricName, GwType: "layer2-extension",
				LoopbackID: interactor.FabricProperties.VTEPLoopBackPortNumber, VNIAuto: "true", Activate: "true"})
			//Verify EVPN
			assert.Equal(t, *fs.Evpn, operation.ConfigEVPNRespone{Name: FabricName,
				DuplicageMacTimerValue: interactor.FabricProperties.DuplicateMacTimer,
				MaxCount:               interactor.FabricProperties.DuplicateMaxTimerMaxCount,
				TargetCommunity:        "auto",
				RouteTargetBoth:        "true",
				IgnoreAs:               "true"})

			assert.Equal(t, sw.VTEPLoopbackIP+"/32", fs.Bgp.Network)
			assert.Contains(t, fs.Bgp.PeerGroups, operation.ConfigBGPPeerGroupResponse{Name: interactor.FabricProperties.LeafPeerGroup, Description: "To Spine",
				RemoteAS: bgpConfigs[0].RemoteAS})
			//TODO allow-as-in in L2VPN
			assert.Contains(t, fs.Bgp.L2VPN.Neighbors, operation.ConfigBGPEVPNNeighborResponse{PeerGroup: interactor.FabricProperties.LeafPeerGroup, Activate: "true",
				Encapsulation: "vxlan"})
			for _, bgpC := range bgpConfigs {
				if unnumbered {
					assert.Contains(t, fs.Bgp.Neighbors, operation.ConfigBGPPeerGroupNeighborResponse{RemoteIP: bgpC.RemoteIPAddress,
						PeerGroup: interactor.FabricProperties.LeafPeerGroup, Multihop: "2"})
				} else {
					assert.Contains(t, fs.Bgp.Neighbors, operation.ConfigBGPPeerGroupNeighborResponse{RemoteIP: bgpC.RemoteIPAddress,
						PeerGroup: interactor.FabricProperties.LeafPeerGroup})
				}
			}

		} else {
			assert.Nil(t, fs.Ovg)
			assert.Nil(t, fs.Evpn)
			//No network configured on Spine
			assert.Equal(t, "", fs.Bgp.Network)
			assert.Contains(t, fs.Bgp.PeerGroups, operation.ConfigBGPPeerGroupResponse{Name: interactor.FabricProperties.SpinePeerGroup, Description: "To Leaf"})
			//TODO allow-as-in in L2VPN
			assert.Contains(t, fs.Bgp.L2VPN.Neighbors, operation.ConfigBGPEVPNNeighborResponse{PeerGroup: interactor.FabricProperties.SpinePeerGroup, Activate: "true",
				Encapsulation: "vxlan", NextHopUnchanged: "true"})
			for _, bgpC := range bgpConfigs {
				if unnumbered {
					assert.Contains(t, fs.Bgp.Neighbors, operation.ConfigBGPPeerGroupNeighborResponse{RemoteIP: bgpC.RemoteIPAddress,
						RemoteAS: bgpC.RemoteAS, PeerGroup: interactor.FabricProperties.SpinePeerGroup, Multihop: "2"})
				} else {
					assert.Contains(t, fs.Bgp.Neighbors, operation.ConfigBGPPeerGroupNeighborResponse{RemoteIP: bgpC.RemoteIPAddress,
						RemoteAS: bgpC.RemoteAS, PeerGroup: interactor.FabricProperties.SpinePeerGroup})
				}
			}
		}

	}
}

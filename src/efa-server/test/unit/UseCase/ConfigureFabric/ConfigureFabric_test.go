package configurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway"
	"efa-server/usecase"
	//"github.com/stretchr/testify/assert"
	"efa-server/infra/database"
	"efa-server/test/unit/mock"
	"testing"

	"efa-server/domain/operation"
	"efa-server/infra/constants"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"sort"
)

var MockSpine1IP = "SPINE1_IP"
var MockSpine2IP = "SPINE2_IP"
var MockLeaf1IP = "LEAF1_IP"
var MockFabricName = "test_fabric"
var MockClusterLeaf1IP = "CLUSTER_LEAF1_IP"
var MockClusterLeaf2IP = "CLUSTER_LEAF2_IP"
var MockClusterLeaf3IP = "CLUSTER_LEAF3_IP"
var MockClusterLeaf4IP = "CLUSTER_LEAF4_IP"

var UserName = "admin"
var Password = "password"

//This test case configures a two node fabric(one spine and one leaf)
//tests the output from the add method
func TestConfigure_OneSpine_OneLeaf(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "1/22", Mac: "M2", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	//Set the Database Location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface using DB and MockDevice Adapter
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	//Configure Fabric Should return no error
	cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	assert.NoError(t, err)
	fmt.Println(cresp, err)

}

//This method registers two devices, one spine and one leaf individually and also invokes validate topology
func TestConfigureSpine_And_Leaf(t *testing.T) {

	//Seperate Mock adapter for Spine
	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return "64512", nil
		},
	}
	//Seperate Mock adapter for Leaf
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/22", Mac: "M2", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2",
				RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil

		},
	}

	//Set the database location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface for Spine
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, UserName, Password, false)
	assert.NoError(t, err)

	//Create a Mock Interactor interface for Leaf
	devUC = usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	_, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{}, UserName, Password, false)
	assert.NoError(t, err)

	//Validate Fabric Topology
	_, err = devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err)

}

//Negative Test
//Register only one Spine
//Validate should indicate in the response saying that there are no leaf devices
func TestConfigureOnlySpine(t *testing.T) {

	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, UserName, Password, false)
	assert.NoError(t, err)

	//validation should return
	resp, err := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.Equal(t, true, resp.NoLeaves)

}

//Test case to make sure that credentials gets updated in DB when Add device is called
func TestConfigureDeviceAndUpdateCredentials(t *testing.T) {

	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, "admin", "pass1234", false)
	assert.NoError(t, err)

	// Check the DB contents password.
	var Device domain.Device
	Device, err = devUC.Db.GetDevice(MockFabricName, MockSpine1IP)
	assert.NoError(t, err)
	assert.Equal(t, "pass1234", Device.Password)

	// Again call AddDevice to update the password.
	_, err = devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, "admin", "password", false)
	assert.NoError(t, err)

	Device, err = devUC.Db.GetDevice(MockFabricName, MockSpine1IP)
	assert.NoError(t, err)
	assert.Equal(t, "password", Device.Password)

}

//Test case to make sure that credentials gets updated in DB when updateDevice is called.
func TestConfigureDeviceAndUpdateDeviceCredentials(t *testing.T) {

	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, "admin", "pass1234", false)
	assert.NoError(t, err)

	// Check the DB contents password.
	var Device domain.Device
	Device, err = devUC.Db.GetDevice(MockFabricName, MockSpine1IP)
	assert.NoError(t, err)
	assert.Equal(t, "pass1234", Device.Password)

	// Now call UpdateDevice to update the password.
	deviceIPs := []string{MockSpine1IP}
	output, err := devUC.UpdateDevices(context.Background(), deviceIPs, "admin", "password")
	assert.NoError(t, err)
	assert.Equal(t, "Successfully Updated Switch Credentials", output[0].Status)

	Device, err = devUC.Db.GetDevice(MockFabricName, MockSpine1IP)
	assert.NoError(t, err)
	assert.Equal(t, "password", Device.Password)

}

//Test case to make sure that credentials gets updated in DB when updateDevice is called.
func TestUpdateUnknownDeviceCredentials(t *testing.T) {

	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, "admin", "pass1234", false)
	assert.NoError(t, err)

	// Check the DB contents password.
	var Device domain.Device
	Device, err = devUC.Db.GetDevice(MockFabricName, MockSpine1IP)
	assert.NoError(t, err)
	assert.Equal(t, "pass1234", Device.Password)

	// Now call UpdateDevice to update the password.
	deviceIPs := []string{MockSpine2IP}
	output, err := devUC.UpdateDevices(context.Background(), deviceIPs, "admin", "password")
	assert.NoError(t, err)
	assert.Equal(t, "Error while retrieving Switch details from Database : record not found", output[0].Status)

}

//Negative Test
//Register only one Leaf
//Validate should indicate in the response saying that there are no Spine devices
func TestConfigureOnlyLeaf(t *testing.T) {

	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/22", Mac: "M2", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2",
				RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil

		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{}, UserName, Password, false)
	assert.NoError(t, err)

	resp, err := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.Equal(t, true, resp.NoSpines)

}

//Negative Test
//Leaf not connected to Spine
//Validate should indicate in the response saying that there are no leaf devices
func TestConfigure_Leaf_not_connected_to_spine(t *testing.T) {

	//There are missing links in the LLDP
	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{}, nil

		},
	}
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/22", Mac: "M2", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{}, nil

		},
	}
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP}, UserName, Password, false)
	assert.NoError(t, err)

	//Change Adapter to next Switch
	devUC = usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	resp1, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{}, UserName, Password, false)
	fmt.Println(resp1)
	assert.NoError(t, err)

	resp, err := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.Equal(t, "Leaf Device LEAF1_IP not connected to Spine Device SPINE1_IP", resp.MissingLinks[0])

}

//This method registers two devices, one spine and one leaf(MCT cluster) individually and also invokes validate topology
func TestConfigureSpine_And_Two_Node_Leaf_Cluster(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 1000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"}}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
			}

			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"}}, nil
			}
			//Leaf
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
			}

			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	MockFabricAdapter := mock.FabricAdapter{
		MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
			return true
		},
	}

	//Set the Database Location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface using DB and MockDevice Adapter
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}
	devUC.AddFabric(context.Background(), MockFabricName)

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err1)

	//Configure Fabric Should return no error
	cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	assert.NoError(t, err)
}

//This method registers two devices, one spine and one leaf(MCT cluster of 3 nodes) individually and also invokes validate topology

func TestConfigureSpine_And_Three_Leafs_Connected(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 1000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M1_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"}}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M2_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/21", Mac: "M2_21", ConfigState: "up"}}, nil
			}

			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M3_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/21", Mac: "M3_21", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"}}, nil
			}

			//ClusterLeaf3
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M4_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/21", Mac: "M4_21", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M1_10", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M2_10"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M3_10"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M4_10"}}, nil
			}

			//Leaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M2_10", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M1_10"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/21", LocalIntMac: "M2_21", RemoteIntType: "ethernet", RemoteIntName: "1/21", RemoteIntMac: "M3_21"}}, nil
			}

			//Leaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M3_10", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/21", LocalIntMac: "M3_21", RemoteIntType: "ethernet", RemoteIntName: "1/21", RemoteIntMac: "M2_21"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"}}, nil
			}

			//Leaf3
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M4_10", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	MockFabricAdapter := mock.FabricAdapter{
		MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
			return true
		},
	}

	//Set the Database Location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface using DB and MockDevice Adapter
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}
	devUC.AddFabric(context.Background(), MockFabricName)

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP}, []string{MockSpine1IP},
		UserName, Password, false)

	fmt.Println(resp)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	v, _ := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.Equal(t, 1, len(v.LeafLeafLinks))
	MockTwoDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 1000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M1_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"}}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M2_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/21", Mac: "M2_21", ConfigState: "up"}}, nil
			}

			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M3_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/21", Mac: "M3_21", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"}}, nil
			}

			//ClusterLeaf3
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/10", Mac: "M4_10", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/21", Mac: "M4_21", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M1_10", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M2_10"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M3_10"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M4_10"}}, nil
			}

			//Leaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M2_10", RemoteIntType: "ethernet", RemoteIntName: "1/10", RemoteIntMac: "M1_10"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/21", LocalIntMac: "M2_21", RemoteIntType: "ethernet", RemoteIntName: "1/21", RemoteIntMac: "M3_21"}}, nil
			}

			//Leaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M3_10", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/21", LocalIntMac: "M3_21", RemoteIntType: "ethernet", RemoteIntName: "1/21", RemoteIntMac: "M2_21"}},
					nil
			}

			//Leaf3
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/10", LocalIntMac: "M4_10", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
				}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	devTwoUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockTwoDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}

	//Create a Mock Interactor interface using DB and MockDevice Adapter

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err = devTwoUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP}, []string{MockSpine1IP},
		UserName, Password, false)

	fmt.Println(resp)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	v, _ = devTwoUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.Equal(t, 0, len(v.LeafLeafLinks))

}

//This method registers three devices, one spine, one leaf(non-MCT cluster) and one more leaf(MCT cluster) and also invokes validate topology and configures fabric
func TestConfigureSpine_One_Leaf_And_Two_Node_Leaf_Cluster(t *testing.T) {

	//Single MOCK adapter serving both the devices

	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 1000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"}}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
			}

			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"}}, nil
			}
			//Leaf
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
			}

			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	MockFabricAdapter := mock.FabricAdapter{
		MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
			return true
		},
	}

	//Set the Database Location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface using DB and MockDevice Adapter
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}
	devUC.AddFabric(context.Background(), MockFabricName)

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err1)

	//Configure Fabric Should return no error
	cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	assert.NoError(t, err)
}

func TestConfigureSpine_One_Leaf_And_Two_Two_Node_Leaf_Cluster_Update_Cluster_Copy_Default(t *testing.T) {
	var DatabaseRepository gateway.DatabaseRepository
	expected := map[uint][]operation.ConfigCluster{domain.MctCreate: []operation.ConfigCluster{
		operation.ConfigCluster{FabricName: "test_fabric", ClusterName: "test_fabric-cluster-1", ClusterID: "1", ClusterControlVlan: "4090", ClusterControlVe: "4090",
			ClusterMemberNodes: []operation.ClusterMemberNode{
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF1_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "1", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.2/31", NodePeerIP: "20.20.20.3", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF2_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "2", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.3/31", NodePeerIP: "20.20.20.2", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
			},
			OperationBitMap: 0x0},
		operation.ConfigCluster{FabricName: "test_fabric", ClusterName: "test_fabric-cluster-1", ClusterID: "1", ClusterControlVlan: "4090", ClusterControlVe: "4090",
			ClusterMemberNodes: []operation.ClusterMemberNode{
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF3_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "1", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.4/31", NodePeerIP: "20.20.20.5", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF4_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "2", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.5/31", NodePeerIP: "20.20.20.4", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
			},
			OperationBitMap: 0x0}}}
	{

		//Single MOCK adapter serving both the devices
		MockDeviceAdapter := mock.DeviceAdapter{
			MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
				return 1000000000, nil
			},
			MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
				//Spine
				if DeviceIP == MockSpine1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/14", Mac: "M1_14", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/15", Mac: "M1_14", ConfigState: "up"},
					}, nil
				}
				//ClusterLeaf1
				if DeviceIP == MockClusterLeaf1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
				}
				//ClusterLeaf2
				if DeviceIP == MockClusterLeaf2IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
				}

				//ClusterLeaf4
				if DeviceIP == MockClusterLeaf4IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M5_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M5_23", ConfigState: "up"}}, nil
				}

				if DeviceIP == MockClusterLeaf3IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M4_23", ConfigState: "up"}}, nil
				}

				if DeviceIP == MockLeaf1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
				}
				return []domain.Interface{}, nil
			},
			MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

				//Spine
				if DeviceIP == MockSpine1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M1_14", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M5_22"},
					}, nil
				}
				//Leaf
				if DeviceIP == MockClusterLeaf1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
				}

				if DeviceIP == MockClusterLeaf2IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
				}
				if DeviceIP == MockClusterLeaf3IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M4_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M5_23"}}, nil
				}

				if DeviceIP == MockClusterLeaf4IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M5_22", RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M1_14"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M5_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M4_23"}}, nil
				}

				if DeviceIP == MockLeaf1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
				}

				return []domain.LLDP{}, nil
			},
			MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

				return "", nil
			},
		}
		MockFabricAdapter := mock.FabricAdapter{
			MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
				return true
			},
		}
		//Set the Database Location
		database.Setup(constants.TESTDBLocation)
		defer cleanupDB(database.GetWorkingInstance())

		DatabaseRepository = gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

		//Create a Mock Interactor interface using DB and MockDevice Adapter
		devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
			FabricAdapter: &MockFabricAdapter}
		devUC.AddFabric(context.Background(), MockFabricName)

		//Call the Add Devices method passing in list of leaf and list of spine address
		resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP, MockLeaf1IP}, []string{MockSpine1IP},
			UserName, Password, false)

		//Leaf and spine  should be succesfully registered
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf4IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

		assert.NoError(t, err)

		_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
		assert.NoError(t, err1)

		//Configure Fabric Should return no error
		config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
		assert.NoError(t, err)
		keys := []uint{}
		for k := range config.MctCluster {
			keys = append(keys, k)
		}

		assert.Equal(t, len(expected), len(config.MctCluster))
		assert.Equal(t, []uint{domain.MctCreate}, keys)
		assert.Equal(t, len(expected[domain.MctCreate]), len(config.MctCluster[domain.MctCreate]))
		assert.Equal(t, len(expected[domain.MctCreate][0].ClusterMemberNodes), len(config.MctCluster[domain.MctCreate][0].ClusterMemberNodes))
		assert.Equal(t, len(expected[domain.MctCreate][1].ClusterMemberNodes), len(config.MctCluster[domain.MctCreate][1].ClusterMemberNodes))
		assert.Equal(t, expected[domain.MctCreate][0].OperationBitMap, config.MctCluster[domain.MctCreate][0].OperationBitMap)
		assert.Equal(t, expected[domain.MctCreate][1].OperationBitMap, config.MctCluster[domain.MctCreate][1].OperationBitMap)

		checkIP := func(clusterIndex, MemberIndex uint) bool {
			fmt.Println(expected[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP)
			fmt.Println(config.MctCluster[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP)
			return intersectIP(expected[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP,
				config.MctCluster[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP)
		}
		assert.True(t, checkIP(0, 0), true)
		assert.True(t, checkIP(0, 1), true)
		assert.True(t, checkIP(1, 0), true)
		assert.True(t, checkIP(1, 1), true)
		for _, sw := range config.Hosts {
			assert.Equal(t, 0, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			if sw.Host == MockSpine1IP || sw.Host == MockLeaf1IP {
				assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
				continue
			}
			fmt.Println("----> ", sw.Host, " -->", sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes)
			assert.Equal(t, 1, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			assert.Equal(t, domain.BGPEncapTypeForCluster, sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes[0].NodePeerEncapType)

		}
		err = devUC.CleanupDBAfterConfigureSuccess()
	}
	//Update
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 1000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/14", Mac: "M1_14", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/15", Mac: "M1_14", ConfigState: "up"},
				}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ve", IntName: "4090", IPAddress: "10.20.20.8/31", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
			}
			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ve", IntName: "4090", IPAddress: "10.20.20.9/31", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
			}

			//ClusterLeaf4
			if DeviceIP == MockClusterLeaf4IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M5_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M5_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockClusterLeaf3IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M4_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M1_14", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M5_22"},
				}, nil
			}
			//Leaf
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
			}

			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
			}
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			if DeviceIP == MockClusterLeaf4IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M5_22", RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M1_14"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	MockFabricAdapter := mock.FabricAdapter{
		MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
			return true
		},
	}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}
	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP, MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf4IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err1)

	//Configure Fabric Should return no error
	config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.NoError(t, err)
	keys := []int{}
	for k := range config.MctCluster {

		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	assert.Equal(t, 2, len(config.MctCluster))
	assert.Equal(t, []int{domain.MctCreate, domain.MctDelete}, keys)
	assert.Equal(t, 1, len(config.MctCluster[domain.MctCreate]))
	assert.Equal(t, 1, len(config.MctCluster[domain.MctDelete]))
	assert.Equal(t, 2, len(config.MctCluster[domain.MctCreate][0].ClusterMemberNodes))
	assert.Equal(t, 2, len(config.MctCluster[domain.MctDelete][0].ClusterMemberNodes))
	assert.Equal(t, 1<<domain.BitPositionForMctCreate, int(config.MctCluster[domain.MctCreate][0].OperationBitMap))
	assert.Equal(t, 0, int(config.MctCluster[domain.MctDelete][0].OperationBitMap))

	for _, sw := range config.Hosts {
		if sw.Host == MockLeaf1IP {
			assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			assert.Equal(t, 0, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			continue
		}
		if sw.Host == MockClusterLeaf3IP || sw.Host == MockClusterLeaf4IP {
			assert.Equal(t, 1, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
		}
		if sw.Host == MockSpine1IP || sw.Host == MockClusterLeaf3IP || sw.Host == MockClusterLeaf4IP {
			assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			continue
		}
		assert.Equal(t, 1, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
		assert.Equal(t, domain.BGPEncapTypeForCluster, sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes[0].NodePeerEncapType)

	}
	err = devUC.CleanupDBAfterConfigureSuccess()
	assert.NoError(t, err)
}

//This method registers three devices, one spine, one leaf(non-MCT cluster) and one more leaf(MCT cluster) and also invokes validate topology and configures fabric
func TestConfigureSpine_One_Leaf_And_Two_Node_Leaf_Cluster_Invalid_Speed(t *testing.T) {

	//Single MOCK adapter serving both the devices

	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 0, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"}}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
			}

			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"}}, nil
			}
			//Leaf
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
			}

			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	//Set the Database Location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface using DB and MockDevice Adapter
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)

	//Call the Add Devices method passing in list of leaf and list of spine address
	_, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	assert.Error(t, err)

}

func TestConfigureSpine_One_Leaf_And_Two_Two_Node_Leaf_Cluster_Update(t *testing.T) {
	var DatabaseRepository gateway.DatabaseRepository
	expected := map[uint][]operation.ConfigCluster{domain.MctCreate: []operation.ConfigCluster{
		operation.ConfigCluster{FabricName: "test_fabric", ClusterName: "test_fabric-cluster-1", ClusterID: "1", ClusterControlVlan: "4090", ClusterControlVe: "4090",
			ClusterMemberNodes: []operation.ClusterMemberNode{
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF1_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "1", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.2/31", NodePeerIP: "20.20.20.3", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF2_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "2", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.3/31", NodePeerIP: "20.20.20.2", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
			},
			OperationBitMap: 0x0},
		operation.ConfigCluster{FabricName: "test_fabric", ClusterName: "test_fabric-cluster-1", ClusterID: "1", ClusterControlVlan: "4090", ClusterControlVe: "4090",
			ClusterMemberNodes: []operation.ClusterMemberNode{
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF3_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "1", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.4/31", NodePeerIP: "20.20.20.5", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF4_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "2", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.5/31", NodePeerIP: "20.20.20.4", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
			},
			OperationBitMap: 0x0}}}
	{

		//Single MOCK adapter serving both the devices
		MockDeviceAdapter := mock.DeviceAdapter{
			MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
				return 1000000000, nil
			},
			MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
				//Spine
				if DeviceIP == MockSpine1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/14", Mac: "M1_14", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/15", Mac: "M1_14", ConfigState: "up"},
					}, nil
				}
				//ClusterLeaf1
				if DeviceIP == MockClusterLeaf1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
				}
				//ClusterLeaf2
				if DeviceIP == MockClusterLeaf2IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
				}

				//ClusterLeaf4
				if DeviceIP == MockClusterLeaf4IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M5_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M5_23", ConfigState: "up"}}, nil
				}

				if DeviceIP == MockClusterLeaf3IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M4_23", ConfigState: "up"}}, nil
				}

				if DeviceIP == MockLeaf1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
				}
				return []domain.Interface{}, nil
			},
			MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

				//Spine
				if DeviceIP == MockSpine1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M1_14", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M5_22"},
					}, nil
				}
				//Leaf
				if DeviceIP == MockClusterLeaf1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
				}

				if DeviceIP == MockClusterLeaf2IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
				}
				if DeviceIP == MockClusterLeaf3IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M4_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M5_23"}}, nil
				}

				if DeviceIP == MockClusterLeaf4IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M5_22", RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M1_14"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M5_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M4_23"}}, nil
				}

				if DeviceIP == MockLeaf1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
				}

				return []domain.LLDP{}, nil
			},
			MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

				return "", nil
			},
		}

		MockFabricAdapter := mock.FabricAdapter{
			MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
				return true
			},
		}
		//Set the Database Location
		database.Setup(constants.TESTDBLocation)
		defer cleanupDB(database.GetWorkingInstance())

		DatabaseRepository = gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

		//Create a Mock Interactor interface using DB and MockDevice Adapter
		devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
			FabricAdapter: &MockFabricAdapter}
		devUC.AddFabric(context.Background(), MockFabricName)

		//Call the Add Devices method passing in list of leaf and list of spine address
		resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP, MockLeaf1IP}, []string{MockSpine1IP},
			UserName, Password, false)

		//Leaf and spine  should be succesfully registered
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf4IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

		assert.NoError(t, err)

		_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
		assert.NoError(t, err1)

		//Configure Fabric Should return no error
		config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
		assert.NoError(t, err)
		keys := []uint{}
		for k := range config.MctCluster {
			keys = append(keys, k)
		}

		assert.Equal(t, len(expected), len(config.MctCluster))
		assert.Equal(t, []uint{domain.MctCreate}, keys)
		assert.Equal(t, len(expected[domain.MctCreate]), len(config.MctCluster[domain.MctCreate]))
		assert.Equal(t, len(expected[domain.MctCreate][0].ClusterMemberNodes), len(config.MctCluster[domain.MctCreate][0].ClusterMemberNodes))
		assert.Equal(t, len(expected[domain.MctCreate][1].ClusterMemberNodes), len(config.MctCluster[domain.MctCreate][1].ClusterMemberNodes))
		assert.Equal(t, expected[domain.MctCreate][0].OperationBitMap, config.MctCluster[domain.MctCreate][0].OperationBitMap)
		assert.Equal(t, expected[domain.MctCreate][1].OperationBitMap, config.MctCluster[domain.MctCreate][1].OperationBitMap)

		checkIP := func(clusterIndex, MemberIndex uint) bool {
			return intersectIP(expected[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP,
				config.MctCluster[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP)
		}
		assert.True(t, checkIP(0, 0), true)
		assert.True(t, checkIP(0, 1), true)
		assert.True(t, checkIP(1, 0), true)
		assert.True(t, checkIP(1, 1), true)
		for _, sw := range config.Hosts {
			assert.Equal(t, 0, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			if sw.Host == MockSpine1IP || sw.Host == MockLeaf1IP {
				assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
				continue
			}
			fmt.Println("----> ", sw.Host, " -->", sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes)
			assert.Equal(t, 1, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			assert.Equal(t, domain.BGPEncapTypeForCluster, sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes[0].NodePeerEncapType)

		}
		err = devUC.CleanupDBAfterConfigureSuccess()
	}
	//Update
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 4000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/14", Mac: "M1_14", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/15", Mac: "M1_14", ConfigState: "up"},
				}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ve", IntName: "4090", IPAddress: "10.20.20.8/31", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
			}
			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ve", IntName: "4090", IPAddress: "10.20.20.9/31", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
			}

			//ClusterLeaf4
			if DeviceIP == MockClusterLeaf4IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M5_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M5_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockClusterLeaf3IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M4_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M1_14", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M5_22"},
				}, nil
			}
			//Leaf
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
			}

			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
			}
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			if DeviceIP == MockClusterLeaf4IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M5_22", RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M1_14"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
		MockGetClusterByName: func(name string) (map[string]string, error) {
			c := make(map[string]string, 1)
			c["cluster-id"] = "1"
			return c, nil
		},
	}
	MockFabricAdapter := mock.FabricAdapter{
		MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
			return true
		},
	}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}
	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP, MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf4IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err1)

	//Configure Fabric Should return no error
	config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.NoError(t, err)
	keys := []int{}
	for k := range config.MctCluster {

		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	assert.Equal(t, 2, len(config.MctCluster))
	assert.Equal(t, []int{domain.MctDelete, domain.MctUpdate}, keys)
	assert.Equal(t, 1, len(config.MctCluster[domain.MctUpdate]))
	assert.Equal(t, 1, len(config.MctCluster[domain.MctDelete]))
	assert.Equal(t, 2, len(config.MctCluster[domain.MctUpdate][0].ClusterMemberNodes))
	assert.Equal(t, 2, len(config.MctCluster[domain.MctDelete][0].ClusterMemberNodes))
	//TODO Need to look why some times update is failing
	assert.Equal(t, 19, int(config.MctCluster[domain.MctUpdate][0].OperationBitMap))
	assert.Equal(t, 0, int(config.MctCluster[domain.MctDelete][0].OperationBitMap))

	for _, sw := range config.Hosts {
		if sw.Host == MockLeaf1IP {
			assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			assert.Equal(t, 0, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			continue
		}
		if sw.Host == MockClusterLeaf3IP || sw.Host == MockClusterLeaf4IP {
			assert.Equal(t, 1, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
		}
		if sw.Host == MockSpine1IP || sw.Host == MockClusterLeaf3IP || sw.Host == MockClusterLeaf4IP {
			assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			continue
		}
		assert.Equal(t, 1, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
		assert.Equal(t, domain.BGPEncapTypeForCluster, sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes[0].NodePeerEncapType)

	}
	err = devUC.CleanupDBAfterConfigureSuccess()
	assert.NoError(t, err)
}
func TestConfigureSpine_One_Leaf_And_Two_Two_Node_Leaf_Cluster_Delete_With_Switch_Error(t *testing.T) {
	var DatabaseRepository gateway.DatabaseRepository
	expected := map[uint][]operation.ConfigCluster{domain.MctCreate: []operation.ConfigCluster{
		operation.ConfigCluster{FabricName: "test_fabric", ClusterName: "test_fabric-cluster-1", ClusterID: "1", ClusterControlVlan: "4090", ClusterControlVe: "4090",
			ClusterMemberNodes: []operation.ClusterMemberNode{
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF1_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "1", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.2/31", NodePeerIP: "20.20.20.3", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF2_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "2", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.3/31", NodePeerIP: "20.20.20.2", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
			},
			OperationBitMap: 0x0},
		operation.ConfigCluster{FabricName: "test_fabric", ClusterName: "test_fabric-cluster-1", ClusterID: "1", ClusterControlVlan: "4090", ClusterControlVe: "4090",
			ClusterMemberNodes: []operation.ClusterMemberNode{
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF3_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "1", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.4/31", NodePeerIP: "20.20.20.5", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
				operation.ClusterMemberNode{NodeMgmtIP: "CLUSTER_LEAF4_IP", NodeModel: "", NodeMgmtUserName: "admin", NodeMgmtPassword: "password", NodeID: "2", NodePrincipalPriority: "0", RemoteNodePeerIP: "10.20.20.5/31", NodePeerIP: "20.20.20.4", NodePeerIntfType: "Port-channel", NodePeerIntfName: "1024", NodePeerIntfSpeed: "1000000000", RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{operation.InterNodeLinkPort{IntfType: "ethernet", IntfName: "1/23"}}},
			},
			OperationBitMap: 0x0}}}
	{

		//Single MOCK adapter serving both the devices
		MockDeviceAdapter := mock.DeviceAdapter{
			MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
				return 1000000000, nil
			},
			MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
				//Spine
				if DeviceIP == MockSpine1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/14", Mac: "M1_14", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/15", Mac: "M1_14", ConfigState: "up"},
					}, nil
				}
				//ClusterLeaf1
				if DeviceIP == MockClusterLeaf1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M2_23", ConfigState: "up"}}, nil
				}
				//ClusterLeaf2
				if DeviceIP == MockClusterLeaf2IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M3_23", ConfigState: "up"}}, nil
				}

				//ClusterLeaf4
				if DeviceIP == MockClusterLeaf4IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M5_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M5_23", ConfigState: "up"}}, nil
				}

				if DeviceIP == MockClusterLeaf3IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"},
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M4_23", ConfigState: "up"}}, nil
				}

				if DeviceIP == MockLeaf1IP {
					return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
				}
				return []domain.Interface{}, nil
			},
			MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

				//Spine
				if DeviceIP == MockSpine1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M1_14", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M5_22"},
					}, nil
				}
				//Leaf
				if DeviceIP == MockClusterLeaf1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M2_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M3_23"}}, nil
				}

				if DeviceIP == MockClusterLeaf2IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M3_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M2_23"}}, nil
				}
				if DeviceIP == MockClusterLeaf3IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M4_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M5_23"}}, nil
				}

				if DeviceIP == MockClusterLeaf4IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M5_22", RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M1_14"},
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/23", LocalIntMac: "M5_23", RemoteIntType: "ethernet", RemoteIntName: "1/23", RemoteIntMac: "M4_23"}}, nil
				}

				if DeviceIP == MockLeaf1IP {
					return []domain.LLDP{
						{FabricID: FabricID, DeviceID: DeviceID,
							LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
				}

				return []domain.LLDP{}, nil
			},
			MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

				return "", nil
			},
		}
		MockFabricAdapter := mock.FabricAdapter{
			MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
				return true
			},
		}
		//Set the Database Location
		database.Setup(constants.TESTDBLocation)
		defer cleanupDB(database.GetWorkingInstance())

		DatabaseRepository = gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

		//Create a Mock Interactor interface using DB and MockDevice Adapter
		devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
			FabricAdapter: &MockFabricAdapter}
		devUC.AddFabric(context.Background(), MockFabricName)

		//Call the Add Devices method passing in list of leaf and list of spine address
		resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP, MockLeaf1IP}, []string{MockSpine1IP},
			UserName, Password, false)

		//Leaf and spine  should be succesfully registered
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf4IP, Role: usecase.LeafRole})
		assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

		assert.NoError(t, err)

		_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
		assert.NoError(t, err1)

		//Configure Fabric Should return no error
		config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
		assert.NoError(t, err)
		keys := []uint{}
		for k := range config.MctCluster {
			keys = append(keys, k)
		}

		assert.Equal(t, len(expected), len(config.MctCluster))
		assert.Equal(t, []uint{domain.MctCreate}, keys)
		assert.Equal(t, len(expected[domain.MctCreate]), len(config.MctCluster[domain.MctCreate]))
		assert.Equal(t, len(expected[domain.MctCreate][0].ClusterMemberNodes), len(config.MctCluster[domain.MctCreate][0].ClusterMemberNodes))
		assert.Equal(t, len(expected[domain.MctCreate][1].ClusterMemberNodes), len(config.MctCluster[domain.MctCreate][1].ClusterMemberNodes))
		assert.Equal(t, expected[domain.MctCreate][0].OperationBitMap, config.MctCluster[domain.MctCreate][0].OperationBitMap)
		assert.Equal(t, expected[domain.MctCreate][1].OperationBitMap, config.MctCluster[domain.MctCreate][1].OperationBitMap)

		checkIP := func(clusterIndex, MemberIndex uint) bool {
			return intersectIP(expected[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP,
				config.MctCluster[domain.MctCreate][clusterIndex].ClusterMemberNodes[MemberIndex].RemoteNodePeerIP)
		}
		assert.True(t, checkIP(0, 0), true)
		assert.True(t, checkIP(0, 1), true)
		assert.True(t, checkIP(1, 0), true)
		assert.True(t, checkIP(1, 1), true)
		for _, sw := range config.Hosts {
			assert.Equal(t, 0, len(sw.UnconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			if sw.Host == MockSpine1IP || sw.Host == MockLeaf1IP {
				assert.Equal(t, 0, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
				continue
			}
			fmt.Println("----> ", sw.Host, " -->", sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes)
			assert.Equal(t, 1, len(sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes))
			assert.Equal(t, domain.BGPEncapTypeForCluster, sw.ConfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes[0].NodePeerEncapType)

		}
		err = devUC.CleanupDBAfterConfigureSuccess()
	}
	//Update
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 1000000000, nil
		},
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/11", Mac: "M1_11", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/12", Mac: "M1_12", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/13", Mac: "M1_13", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/14", Mac: "M1_14", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/15", Mac: "M1_14", ConfigState: "up"},
				}, nil
			}
			//ClusterLeaf1
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M2_22", ConfigState: "up"}},
					nil
			}
			//ClusterLeaf2
			if DeviceIP == MockClusterLeaf2IP {
				return []domain.Interface{
						{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M3_22", ConfigState: "up"}},
					nil
			}

			//ClusterLeaf4
			if DeviceIP == MockClusterLeaf4IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M5_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M5_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockClusterLeaf3IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"},
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/23", Mac: "M4_23", ConfigState: "up"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{
					{FabricID: FabricID, DeviceID: DeviceID, IntType: "ethernet", IntName: "1/22", Mac: "M4_22", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1_11", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/12", LocalIntMac: "M1_12", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M3_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/13", LocalIntMac: "M1_13", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M4_22"},
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M1_14", RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M5_22"},
				}, nil
			}
			//Leaf
			if DeviceIP == MockClusterLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2_22", RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1_11"}}, nil

			}

			if DeviceIP == MockClusterLeaf2IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M3_22", RemoteIntType: "ethernet", RemoteIntName: "1/12", RemoteIntMac: "M1_12"}}, nil

			}
			if DeviceIP == MockClusterLeaf3IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			if DeviceIP == MockClusterLeaf4IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M5_22", RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M1_14"}}, nil
			}

			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{
					{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M4_22", RemoteIntType: "ethernet", RemoteIntName: "1/13", RemoteIntMac: "M1_13"}}, nil
			}

			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	MockFabricAdapter := mock.FabricAdapter{
		MockIsMCTLeavesCompatible: func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
			return true
		},
	}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &MockFabricAdapter}
	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP, MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf1IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf2IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf3IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockClusterLeaf4IP, Role: usecase.LeafRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.NoError(t, err)

	_, err1 := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err1)

	//Configure Fabric Should return no error
	_, err = devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.NoError(t, err)
	DeviceIPList := []string{MockSpine1IP, MockLeaf1IP, MockClusterLeaf1IP, MockClusterLeaf2IP, MockClusterLeaf3IP, MockClusterLeaf4IP}
	Error := devUC.DeleteDevices(context.Background(), MockFabricName, DeviceIPList, true)
	err = devUC.CleanupDBAfterConfigureSuccess()
	fmt.Println(Error)
	assert.NoError(t, err)
}

func TestConfigure_SpineConnectedToAnotherSpine(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine-1
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "1/11", Mac: "M11", ConfigState: "up"},
					//link to Spine-2
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "1/5", Mac: "M5", ConfigState: "up"}}, nil
			}
			//Spine-2
			if DeviceIP == MockSpine2IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "1/14", Mac: "M14", ConfigState: "up"},
					//link to Spine-1
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "1/15", Mac: "M15", ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "1/22", Mac: "M22", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "1/24", Mac: "M24", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine-1
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M11",
					RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M22"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/5", LocalIntMac: "M5",
						RemoteIntType: "ethernet", RemoteIntName: "1/15", RemoteIntMac: "M15"}}, nil
			}
			//Spine-2
			if DeviceIP == MockSpine2IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "1/14", LocalIntMac: "M14",
					RemoteIntType: "ethernet", RemoteIntName: "1/24", RemoteIntMac: "M24"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/15", LocalIntMac: "M15",
						RemoteIntType: "ethernet", RemoteIntName: "1/5", RemoteIntMac: "M5"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M22",
					RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M11"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "1/24", LocalIntMac: "M24",
						RemoteIntType: "ethernet", RemoteIntName: "1/14", RemoteIntMac: "M14"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	//Set the Database Location
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	//Create a Mock Interactor interface using DB and MockDevice Adapter
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP, MockSpine2IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	//validation should return
	resp1, err := devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	fmt.Println(resp1)

	assert.Contains(t, resp1.SpineSpineLinks, "Spine Device SPINE2_IP connected to Spine Device SPINE1_IP")
	assert.Contains(t, resp1.SpineSpineLinks, "Spine Device SPINE1_IP connected to Spine Device SPINE2_IP")

}

func intersectIP(ip1, ip2 string) bool {
	//Already IP address are sanitized
	_, n1, _ := net.ParseCIDR(ip1)
	_, n2, _ := net.ParseCIDR(ip2)
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}

func cleanupDB(Database *database.Database) {
	Database.Close()
	os.Remove(constants.TESTDBLocation)
}

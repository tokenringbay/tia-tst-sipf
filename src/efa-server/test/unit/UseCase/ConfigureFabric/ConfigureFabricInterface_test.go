package configurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"efa-server/test/unit/mock"
	"efa-server/usecase"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

//Contains test-cases to test the affect of interface.ip address and lldp changes

//Obtain a new loopback for Leaf
func TestConfigure_NewInterface(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	SpineDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockSpine1IP)
	SpineSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, SpineDevice.ID)
	SpineInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)
	SpineBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)

	LeafDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockLeaf1IP)
	LeafSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, LeafDevice.ID)
	LeafInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)
	LeafBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)

	//Interface Configs
	assert.Contains(t, SpineInterfaceConfigs[0].IPAddress, "10.10.10")
	Spine1IP := SpineInterfaceConfigs[0].IPAddress
	assert.Equal(t, domain.IntfTypeEthernet, SpineInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/11", SpineInterfaceConfigs[0].IntName)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, SpineInterfaceConfigs[0].ConfigType)

	assert.Contains(t, LeafInterfaceConfigs[0].IPAddress, "10.10.10")
	Leaf1IP := LeafInterfaceConfigs[0].IPAddress

	assert.Equal(t, domain.IntfTypeEthernet, LeafInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/22", LeafInterfaceConfigs[0].IntName)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, LeafInterfaceConfigs[0].ConfigType)

	//BGP Neighbor Configs
	assert.Equal(t, Leaf1IP, SpineBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, LeafSwitchConfig.LocalAS, SpineBGPNeighborConfigs[0].RemoteAS)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, SpineBGPNeighborConfigs[0].ConfigType)

	assert.Equal(t, Spine1IP, LeafBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, SpineSwitchConfig.LocalAS, LeafBGPNeighborConfigs[0].RemoteAS)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, LeafBGPNeighborConfigs[0].ConfigType)

	//Used IP Table
	SpineusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, SpineDevice.ID)
	assert.Equal(t, 1, len(SpineusedIps))
	assert.Equal(t, "Loopback", SpineusedIps[0].IPType)
	assert.Equal(t, SpineSwitchConfig.LoopbackIP, SpineusedIps[0].IPAddress)
	assert.Contains(t, SpineSwitchConfig.LoopbackIP, "172.31.254")

	LeafusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, LeafDevice.ID)
	assert.Equal(t, 2, len(LeafusedIps))
	assert.Equal(t, "Loopback", LeafusedIps[0].IPType)
	assert.Equal(t, LeafSwitchConfig.LoopbackIP, LeafusedIps[0].IPAddress)
	assert.Contains(t, LeafSwitchConfig.LoopbackIP, "172.31.254")
	assert.Equal(t, "Loopback", LeafusedIps[1].IPType)
	assert.Equal(t, LeafSwitchConfig.VTEPLoopbackIP, LeafusedIps[1].IPAddress)
	assert.Contains(t, LeafSwitchConfig.VTEPLoopbackIP, "172.31.254")

	//P2P address are allocated in Pairs
	UsedIPPairs, err := DatabaseRepository.GetUsedIPPairsSOnDeviceAndType(Fabric.ID, SpineDevice.ID, LeafDevice.ID)
	assert.Equal(t, 1, len(UsedIPPairs))
	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{Spine1IP, Leaf1IP}, UsedIPPairs[0].IPAddressOne)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceOneID)

	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{Spine1IP, Leaf1IP}, UsedIPPairs[0].IPAddressTwo)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceTwoID)

}

//Obtain a new loopback for Leaf
func TestConfigure_ReserveValidIP(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.12"
	ExpectedLeaf1IP := "10.10.10.13"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}

	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	SpineDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockSpine1IP)

	SpineSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, SpineDevice.ID)
	SpineInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)
	SpineBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)

	LeafDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockLeaf1IP)
	LeafSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, LeafDevice.ID)
	LeafInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)
	LeafBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)

	//Interface Configs
	assert.Equal(t, ExpectedSpine1IP, SpineInterfaceConfigs[0].IPAddress)
	assert.Equal(t, domain.IntfTypeEthernet, SpineInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/11", SpineInterfaceConfigs[0].IntName)
	assert.Equal(t, domain.ConfigCreate, SpineInterfaceConfigs[0].ConfigType)

	assert.Equal(t, ExpectedLeaf1IP, LeafInterfaceConfigs[0].IPAddress)
	assert.Equal(t, domain.IntfTypeEthernet, LeafInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/22", LeafInterfaceConfigs[0].IntName)
	assert.Equal(t, domain.ConfigCreate, LeafInterfaceConfigs[0].ConfigType)

	//BGP Neighbor Configs
	assert.Equal(t, ExpectedLeaf1IP, SpineBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, LeafSwitchConfig.LocalAS, SpineBGPNeighborConfigs[0].RemoteAS)
	assert.Equal(t, domain.ConfigCreate, SpineBGPNeighborConfigs[0].ConfigType)

	assert.Equal(t, ExpectedSpine1IP, LeafBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, SpineSwitchConfig.LocalAS, LeafBGPNeighborConfigs[0].RemoteAS)
	assert.Equal(t, domain.ConfigCreate, LeafBGPNeighborConfigs[0].ConfigType)

	//Used IP Table
	SpineusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, SpineDevice.ID)
	assert.Equal(t, 1, len(SpineusedIps))
	assert.Equal(t, "Loopback", SpineusedIps[0].IPType)
	assert.Equal(t, SpineSwitchConfig.LoopbackIP, SpineusedIps[0].IPAddress)
	assert.Contains(t, SpineSwitchConfig.LoopbackIP, "172.31.254")

	LeafusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, LeafDevice.ID)
	assert.Equal(t, 2, len(LeafusedIps))
	assert.Equal(t, "Loopback", LeafusedIps[0].IPType)
	assert.Equal(t, LeafSwitchConfig.LoopbackIP, LeafusedIps[0].IPAddress)
	assert.Contains(t, LeafSwitchConfig.LoopbackIP, "172.31.254")
	assert.Equal(t, "Loopback", LeafusedIps[1].IPType)
	assert.Equal(t, LeafSwitchConfig.VTEPLoopbackIP, LeafusedIps[1].IPAddress)
	assert.Contains(t, LeafSwitchConfig.VTEPLoopbackIP, "172.31.254")

	//P2P address are allocated in Pairs
	UsedIPPairs, err := DatabaseRepository.GetUsedIPPairsSOnDeviceAndType(Fabric.ID, SpineDevice.ID, LeafDevice.ID)
	assert.Equal(t, 1, len(UsedIPPairs))
	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{ExpectedSpine1IP, ExpectedLeaf1IP}, UsedIPPairs[0].IPAddressOne)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceOneID)

	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{ExpectedSpine1IP, ExpectedLeaf1IP}, UsedIPPairs[0].IPAddressTwo)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceTwoID)

}

//Obtain a new loopback for Leaf
func TestConfigure_ReserveIPOnlyOnePeerConfigured(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.12"
	//EXPECTED_LEAF1_IP := "10.10.10.13"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: "", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

}

//Obtain a new loopback for Leaf
func TestConfigure_ReserveInValidIP(t *testing.T) {
	ExpectedSpine1IP := "1.10.10.12"
	ExpectedLeaf1IP := "1.10.10.20"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole,
		Errors: []error{errors.New("1.10.10.20 is not in 10.10.10.0/23 range on ethernet 1/22")}})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole,
		Errors: []error{errors.New("1.10.10.12 is not in 10.10.10.0/23 range on ethernet 1/11")}})

	assert.Equal(t, "Add Device Operation Failed", fmt.Sprint(err))

}

//check for existing
func TestConfigure_UseExistingInterface(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.12"
	ExpectedLeaf1IP := "10.10.10.13"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"}}, nil

			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}

	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	fmt.Println(resp)
	assert.NoError(t, err)

	//Simulating a success
	devUC.CleanupDBAfterConfigureSuccess()

	//do it for the second Time
	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	fmt.Println(resp)
	assert.NoError(t, err)

	Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	SpineDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockSpine1IP)
	SpineSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, SpineDevice.ID)
	SpineInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)
	SpineBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)

	LeafDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockLeaf1IP)
	LeafSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, LeafDevice.ID)
	LeafInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)
	LeafBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)

	//Interface Configs
	assert.Equal(t, ExpectedSpine1IP, SpineInterfaceConfigs[0].IPAddress)
	assert.Equal(t, domain.IntfTypeEthernet, SpineInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/11", SpineInterfaceConfigs[0].IntName)
	assert.Equal(t, domain.ConfigNone, SpineInterfaceConfigs[0].ConfigType)

	assert.Equal(t, ExpectedLeaf1IP, LeafInterfaceConfigs[0].IPAddress)
	assert.Equal(t, domain.IntfTypeEthernet, LeafInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/22", LeafInterfaceConfigs[0].IntName)
	assert.Equal(t, domain.ConfigNone, LeafInterfaceConfigs[0].ConfigType)

	//BGP Neighbor Configs
	assert.Equal(t, ExpectedLeaf1IP, SpineBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, LeafSwitchConfig.LocalAS, SpineBGPNeighborConfigs[0].RemoteAS)
	assert.Equal(t, domain.ConfigNone, SpineBGPNeighborConfigs[0].ConfigType)

	assert.Equal(t, ExpectedSpine1IP, LeafBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, SpineSwitchConfig.LocalAS, LeafBGPNeighborConfigs[0].RemoteAS)
	assert.Equal(t, domain.ConfigNone, LeafBGPNeighborConfigs[0].ConfigType)

	//Used IP Table
	SpineusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, SpineDevice.ID)
	assert.Equal(t, 1, len(SpineusedIps))
	assert.Equal(t, "Loopback", SpineusedIps[0].IPType)
	assert.Equal(t, SpineSwitchConfig.LoopbackIP, SpineusedIps[0].IPAddress)
	assert.Contains(t, SpineSwitchConfig.LoopbackIP, "172.31.254")

	LeafusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, LeafDevice.ID)
	assert.Equal(t, 2, len(LeafusedIps))
	assert.Equal(t, "Loopback", LeafusedIps[0].IPType)
	assert.Equal(t, LeafSwitchConfig.LoopbackIP, LeafusedIps[0].IPAddress)
	assert.Contains(t, LeafSwitchConfig.LoopbackIP, "172.31.254")
	assert.Equal(t, "Loopback", LeafusedIps[1].IPType)
	assert.Equal(t, LeafSwitchConfig.VTEPLoopbackIP, LeafusedIps[1].IPAddress)
	assert.Contains(t, LeafSwitchConfig.VTEPLoopbackIP, "172.31.254")

	//P2P address are allocated in Pairs
	UsedIPPairs, err := DatabaseRepository.GetUsedIPPairsSOnDeviceAndType(Fabric.ID, SpineDevice.ID, LeafDevice.ID)
	assert.Equal(t, 1, len(UsedIPPairs))
	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{ExpectedSpine1IP, ExpectedLeaf1IP}, UsedIPPairs[0].IPAddressOne)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceOneID)

	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{ExpectedSpine1IP, ExpectedLeaf1IP}, UsedIPPairs[0].IPAddressTwo)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceTwoID)
}

func TestConfigure_ReserveDifferentIP(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.2"
	ExpectedLeaf1IP := "10.10.10.3"

	ExpectedSpine1IP2 := "10.10.10.4"
	ExpectedLeaf1IP2 := "10.10.10.5"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"}}, nil

			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}
	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter2 := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP2 + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP2 + "/31",
					ConfigState: "up"}}, nil

			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	//Simulating a success
	devUC.CleanupDBAfterConfigureSuccess()

	//do it for the second Time
	devUC = usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter2),
		FabricAdapter: &mock.FabricAdapter{}}

	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	SpineDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockSpine1IP)
	SpineSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, SpineDevice.ID)
	SpineInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)
	SpineBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, SpineDevice.ID)

	LeafDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockLeaf1IP)
	LeafSwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, LeafDevice.ID)
	LeafInterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)
	LeafBGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, LeafDevice.ID)

	//Interface Configs
	fmt.Println(SpineInterfaceConfigs)
	assert.Equal(t, ExpectedSpine1IP2, SpineInterfaceConfigs[0].IPAddress)
	assert.Equal(t, domain.IntfTypeEthernet, SpineInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/11", SpineInterfaceConfigs[0].IntName)
	assert.Equal(t, domain.ConfigUpdate, SpineInterfaceConfigs[0].ConfigType)

	assert.Equal(t, ExpectedLeaf1IP2, LeafInterfaceConfigs[0].IPAddress)
	assert.Equal(t, domain.IntfTypeEthernet, LeafInterfaceConfigs[0].IntType)
	assert.Equal(t, "1/22", LeafInterfaceConfigs[0].IntName)
	assert.Equal(t, domain.ConfigUpdate, LeafInterfaceConfigs[0].ConfigType)

	//BGP Neighbor Configs
	assert.Equal(t, ExpectedLeaf1IP2, SpineBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, LeafSwitchConfig.LocalAS, SpineBGPNeighborConfigs[0].RemoteAS)
	assert.Equal(t, domain.ConfigUpdate, SpineBGPNeighborConfigs[0].ConfigType)

	assert.Equal(t, ExpectedSpine1IP2, LeafBGPNeighborConfigs[0].RemoteIPAddress)
	assert.Equal(t, SpineSwitchConfig.LocalAS, LeafBGPNeighborConfigs[0].RemoteAS)
	assert.Equal(t, domain.ConfigUpdate, LeafBGPNeighborConfigs[0].ConfigType)

	//Used IP Table
	SpineusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, SpineDevice.ID)
	assert.Equal(t, 1, len(SpineusedIps))
	assert.Equal(t, "Loopback", SpineusedIps[0].IPType)
	assert.Equal(t, SpineSwitchConfig.LoopbackIP, SpineusedIps[0].IPAddress)
	assert.Contains(t, SpineSwitchConfig.LoopbackIP, "172.31.254")

	LeafusedIps, err := DatabaseRepository.GetUsedIPSOnDeviceAndType(Fabric.ID, LeafDevice.ID)
	assert.Equal(t, 2, len(LeafusedIps))
	assert.Equal(t, "Loopback", LeafusedIps[0].IPType)
	assert.Equal(t, LeafSwitchConfig.LoopbackIP, LeafusedIps[0].IPAddress)
	assert.Contains(t, LeafSwitchConfig.LoopbackIP, "172.31.254")
	assert.Equal(t, "Loopback", LeafusedIps[1].IPType)
	assert.Equal(t, LeafSwitchConfig.VTEPLoopbackIP, LeafusedIps[1].IPAddress)
	assert.Contains(t, LeafSwitchConfig.VTEPLoopbackIP, "172.31.254")

	//P2P address are allocated in Pairs
	UsedIPPairs, err := DatabaseRepository.GetUsedIPPairsSOnDeviceAndType(Fabric.ID, SpineDevice.ID, LeafDevice.ID)
	assert.Equal(t, 1, len(UsedIPPairs))
	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{ExpectedSpine1IP2, ExpectedLeaf1IP2}, UsedIPPairs[0].IPAddressOne)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceOneID)

	assert.Equal(t, "P2P", UsedIPPairs[0].IPType)
	assert.Contains(t, []string{ExpectedSpine1IP2, ExpectedLeaf1IP2}, UsedIPPairs[0].IPAddressTwo)
	assert.Contains(t, []uint{SpineInterfaceConfigs[0].InterfaceID, LeafInterfaceConfigs[0].InterfaceID}, UsedIPPairs[0].InterfaceTwoID)

}

func TestConfigure_changeInInterfaceNumber(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.2"
	ExpectedLeaf1IP := "10.10.10.3"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"}}, nil

			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)

	//Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	//SpineDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockSpine1IP)
	//OldSpineInterfaceConfigs,err:=DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID,SpineDevice.ID)
	//OldSpineBGPNeighborConfigs,err:=DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID,SpineDevice.ID)

	//LeafDevice, _ := DatabaseRepository.GetDevice(MockFabricName, MockLeaf1IP)
	//OldLeafInterfaceConfigs,err:=DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID,LeafDevice.ID)
	//OldLeafBGPNeighborConfigs,err:=DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID,LeafDevice.ID)

	MockDeviceAdapter2 := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31", ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/44", Mac: "M4", IPAddress: ExpectedLeaf1IP + "/31", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/11", LocalIntMac: "M1",
					RemoteIntType: domain.IntfTypeEthernet, RemoteIntName: "1/44", RemoteIntMac: "M4"}}, nil
			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: domain.IntfTypeEthernet, LocalIntName: "1/44", LocalIntMac: "M4",
					RemoteIntType: "ethernet", RemoteIntName: "1/11", RemoteIntMac: "M1"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},
	}

	//Simulating a success
	devUC.CleanupDBAfterConfigureSuccess()

	//do it for the second Time
	devUC = usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter2),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	fmt.Println(resp)
	assert.Contains(t, fmt.Sprint(resp), "IPPair(10.10.10.2,10.10.10.3) not present in the Used IP Table for Fabric 1")
	assert.Contains(t, fmt.Sprint(resp), "IPPair(10.10.10.3,10.10.10.2) not present in the Used IP Table for Fabric 1")
	assert.Equal(t, "Add Device Operation Failed", err.Error())

}

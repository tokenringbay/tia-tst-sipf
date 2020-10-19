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

	"efa-server/infra/constants"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
)

var MockRack1IP1 = "RACK1_IP1"
var MockRack1IP2 = "RACK1_IP2"

var MockRack2IP1 = "RACK2_IP1"
var MockRack2IP2 = "RACK2_IP2"

//This test case configures two Racks
func TestConfigure_Rack(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_1_48", ConfigState: "up"},

					//Interface to be connected to Rack2IP1
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M1_1_1", ConfigState: "up"},
				}, nil
			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_2_48", ConfigState: "up"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_1_48", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M2_1_1", ConfigState: "up"},
				}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_2_48", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M1_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M2_1_1"}}, nil

			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_1_48"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M2_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M1_1_1"}}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_1_48"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},

		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 10000, nil
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
	resp, err := devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: MockRack1IP1},
		usecase.Rack{IP1: MockRack2IP2, IP2: MockRack2IP1}},
		UserName, Password, false)

	Rack1IPAddress := fmt.Sprintln(MockRack1IP2, ",", MockRack1IP1)
	Rack2IPAddress := fmt.Sprintln(MockRack2IP2, ",", MockRack2IP1)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack1IPAddress, Role: usecase.RackRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack2IPAddress, Role: usecase.RackRole})
	assert.NoError(t, err)

	//Configure Fabric Should return no error

	Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	Rack1IP1, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack1IP1)
	Rack1IP1SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack1IP1.ID)
	Rack1IP1InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP1.ID)
	Rack1IP1BGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP1.ID)
	Rack1IP1MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP1.ID)

	Rack1IP2, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack1IP2)
	Rack1IP2SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack1IP2.ID)
	Rack1IP2InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP2.ID)
	Rack1IP2MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP2.ID)

	Rack2IP1, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack2IP1)
	Rack2IP1SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack2IP1.ID)
	Rack2IP1InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP1.ID)
	Rack2IP1BGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP1.ID)
	Rack2IP1MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP1.ID)

	Rack2IP2, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack2IP2)
	Rack2IP2SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack2IP2.ID)
	Rack2IP2InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP2.ID)
	Rack2IP2MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP2.ID)

	Rack1, _ := DatabaseRepository.GetRack(Fabric.Name, Rack1IP1.IPAddress, Rack1IP2.IPAddress)
	Rack1EvpnNeighbors, _ := DatabaseRepository.GetRackEvpnConfig(Rack1.ID)

	Rack2, _ := DatabaseRepository.GetRack(Fabric.Name, Rack2IP1.IPAddress, Rack2IP2.IPAddress)
	Rack2EvpnNeighbors, _ := DatabaseRepository.GetRackEvpnConfig(Rack2.ID)

	assert.Equal(t, 4, len(Rack1EvpnNeighbors))
	assert.Equal(t, 4, len(Rack2EvpnNeighbors))
	for _, Rack1EvpnNeighbor := range Rack1EvpnNeighbors {
		assert.Equal(t, Rack1.ID, Rack1EvpnNeighbor.LocalRackID)
		assert.Equal(t, Rack2.ID, Rack1EvpnNeighbor.RemoteRackID)
		assert.Equal(t, Rack2IP1SwitchConfig.LocalAS, Rack1EvpnNeighbor.RemoteAS)
		assert.Condition(t, func() bool {
			if Rack1EvpnNeighbor.LocalDeviceID == Rack1IP1.ID || Rack1EvpnNeighbor.LocalDeviceID == Rack1IP2.ID {
				return true
			}
			return false
		})
		assert.Condition(t, func() bool {
			if Rack1EvpnNeighbor.RemoteDeviceID == Rack2IP1.ID || Rack1EvpnNeighbor.RemoteDeviceID == Rack2IP2.ID {
				return true
			}
			return false
		})
		if Rack1EvpnNeighbor.RemoteDeviceID == Rack2IP1.ID {
			assert.Equal(t, Rack2IP1SwitchConfig.LoopbackIP, Rack1EvpnNeighbor.EVPNAddress)
		} else {
			assert.Equal(t, Rack2IP2SwitchConfig.LoopbackIP, Rack1EvpnNeighbor.EVPNAddress)
		}

	}

	for _, Rack2EvpnNeighbor := range Rack2EvpnNeighbors {
		assert.Equal(t, Rack2.ID, Rack2EvpnNeighbor.LocalRackID)
		assert.Equal(t, Rack1.ID, Rack2EvpnNeighbor.RemoteRackID)
		assert.Equal(t, Rack1IP1SwitchConfig.LocalAS, Rack2EvpnNeighbor.RemoteAS)
		assert.Condition(t, func() bool {
			if Rack2EvpnNeighbor.LocalDeviceID == Rack2IP1.ID || Rack2EvpnNeighbor.LocalDeviceID == Rack2IP2.ID {
				return true
			}
			return false
		})
		assert.Condition(t, func() bool {
			if Rack2EvpnNeighbor.RemoteDeviceID == Rack1IP1.ID || Rack2EvpnNeighbor.RemoteDeviceID == Rack1IP2.ID {
				return true
			}
			return false
		})
		if Rack2EvpnNeighbor.RemoteDeviceID == Rack1IP1.ID {
			assert.Equal(t, Rack1IP1SwitchConfig.LoopbackIP, Rack2EvpnNeighbor.EVPNAddress)
		} else {
			assert.Equal(t, Rack1IP2SwitchConfig.LoopbackIP, Rack2EvpnNeighbor.EVPNAddress)
		}

	}

	Rack1IP1FabLink := ""
	//Interface Configs Rack1IP1
	for _, intf := range Rack1IP1InterfaceConfigs {
		if intf.IntName == "0/48" {
			assert.Contains(t, intf.IPAddress, "10.30.30")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
		} else {
			assert.Contains(t, intf.IPAddress, "10.10.10")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Equal(t, "0/1", intf.IntName)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
			Rack1IP1FabLink = intf.IPAddress
		}
	}

	assert.Contains(t, Rack1IP1SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack1IP1SwitchConfig.VTEPLoopbackIP, "172.31.254")
	assert.Contains(t, Rack1IP2SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack1IP2SwitchConfig.VTEPLoopbackIP, "172.31.254")
	//On Cluster both the IP's should be same
	assert.NotEqual(t, Rack1IP1SwitchConfig.LoopbackIP, Rack1IP2SwitchConfig.LoopbackIP)
	assert.Equal(t, Rack1IP1SwitchConfig.VTEPLoopbackIP, Rack1IP2SwitchConfig.VTEPLoopbackIP)
	//On Cluster ASN's should be same
	assert.Equal(t, Rack1IP1SwitchConfig.LocalAS, Rack1IP2SwitchConfig.LocalAS)
	//On Cluster MCT BGP Neighbors
	assert.Contains(t, Rack1IP1MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Contains(t, Rack1IP2MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Equal(t, Rack1IP2MCTBGPNeighborConfigs[0].RemoteAS, Rack1IP2MCTBGPNeighborConfigs[0].RemoteAS)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack1IP2MCTBGPNeighborConfigs[0].ConfigType)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack1IP2MCTBGPNeighborConfigs[0].ConfigType)
	//On L3 Backup
	for _, bgp := range Rack1IP1BGPNeighborConfigs {
		if bgp.Type == domain.MCTL3LBType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack1IP2InterfaceConfigs[0].IPAddress)
			assert.Equal(t, bgp.RemoteAS, Rack1IP2SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}
	}

	//Interface Configs Rack2IP1
	Rack2IP1FabLink := ""
	for _, intf := range Rack2IP1InterfaceConfigs {
		if intf.IntName == "0/48" {
			assert.Contains(t, intf.IPAddress, "10.30.30")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
		} else {
			assert.Contains(t, intf.IPAddress, "10.10.10")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Equal(t, "0/1", intf.IntName)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
			Rack2IP1FabLink = intf.IPAddress
		}
	}

	assert.Contains(t, Rack2IP1SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack2IP1SwitchConfig.VTEPLoopbackIP, "172.31.254")
	assert.Contains(t, Rack2IP2SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack2IP2SwitchConfig.VTEPLoopbackIP, "172.31.254")
	//On Cluster both the IP's should be same
	assert.NotEqual(t, Rack2IP1SwitchConfig.LoopbackIP, Rack2IP2SwitchConfig.LoopbackIP)
	assert.Equal(t, Rack2IP1SwitchConfig.VTEPLoopbackIP, Rack2IP2SwitchConfig.VTEPLoopbackIP)
	//On Cluster ASN's should be same
	assert.Equal(t, Rack2IP1SwitchConfig.LocalAS, Rack2IP2SwitchConfig.LocalAS)
	//On Cluster MCT BGP Neighbors
	assert.Contains(t, Rack2IP1MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Contains(t, Rack2IP2MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Equal(t, Rack2IP2MCTBGPNeighborConfigs[0].RemoteAS, Rack2IP2MCTBGPNeighborConfigs[0].RemoteAS)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack2IP2MCTBGPNeighborConfigs[0].ConfigType)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack2IP2MCTBGPNeighborConfigs[0].ConfigType)
	//On L3 Backup
	for _, bgp := range Rack2IP1BGPNeighborConfigs {
		if bgp.Type == domain.MCTL3LBType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack2IP2InterfaceConfigs[0].IPAddress)
			assert.Equal(t, bgp.RemoteAS, Rack2IP2SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}
	}

	//BGP Neighbor Configs between Rack1IP1 and Rack2IP1
	for _, bgp := range Rack1IP1BGPNeighborConfigs {
		if bgp.Type == domain.FabricBGPType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack2IP1FabLink)
			assert.Equal(t, bgp.RemoteAS, Rack2IP1SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}

	}
	//BGP Neighbor Configs between Rack1IP1 and Rack2IP1
	for _, bgp := range Rack2IP1BGPNeighborConfigs {
		if bgp.Type == domain.FabricBGPType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack1IP1FabLink)
			assert.Equal(t, bgp.RemoteAS, Rack1IP1SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}

	}

	cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	//assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	//assert.NoError(t, err)
	fmt.Println(cresp, err)

}

//This test case configures two Racks
func TestDeconfigure_Rack(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_1_48", ConfigState: "up"},

					//Interface to be connected to Rack2IP1
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M1_1_1", ConfigState: "up"},
				}, nil
			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_2_48", ConfigState: "up"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_1_48", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M2_1_1", ConfigState: "up"},
				}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_2_48", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M1_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M2_1_1"}}, nil

			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_1_48"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M2_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M1_1_1"}}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_1_48"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},

		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 10000, nil
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
	resp, err := devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: MockRack1IP1},
		usecase.Rack{IP1: MockRack2IP2, IP2: MockRack2IP1}},
		UserName, Password, false)

	Rack1IPAddress := fmt.Sprintln(MockRack1IP2, ",", MockRack1IP1)
	Rack2IPAddress := fmt.Sprintln(MockRack2IP2, ",", MockRack2IP1)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack1IPAddress, Role: usecase.RackRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack2IPAddress, Role: usecase.RackRole})
	assert.NoError(t, err)

	cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	//assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	//assert.NoError(t, err)
	fmt.Println(cresp, err)

	//DeConfigure
	dResp, err := devUC.DeleteDevicesFromNonCLOSFabric(context.Background(), MockFabricName,
		[]usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: MockRack1IP1}}, "", "", false, false, false)

	fmt.Println(dResp, err)

}

func TestConfigure_Rack_With_extra_links_in_rack(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_1_48", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/21", Mac: "M1_1_21", ConfigState: "up"},

					//Interface to be connected to Rack2IP1
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M1_1_1", ConfigState: "up"},
				}, nil
			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_2_48", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/21", Mac: "M1_2_21", ConfigState: "up"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_1_48", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M2_1_1", ConfigState: "up"},
				}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_2_48", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_2_48"},

					//extra-link
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/21", LocalIntMac: "M1_1_21",
						RemoteIntType: "ethernet", RemoteIntName: "0/21", RemoteIntMac: "M1_2_21"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M1_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M2_1_1"}}, nil

			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_1_48"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/21", LocalIntMac: "M1_2_21",
						RemoteIntType: "ethernet", RemoteIntName: "0/21", RemoteIntMac: "M1_1_21"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M2_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M1_1_1"}}, nil
			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_1_48"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},

		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 10000, nil
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
	resp, err := devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: MockRack1IP1},
		usecase.Rack{IP1: MockRack2IP2, IP2: MockRack2IP1}},
		UserName, Password, false)

	Rack1IPAddress := fmt.Sprintln(MockRack1IP2, ",", MockRack1IP1)
	Rack2IPAddress := fmt.Sprintln(MockRack2IP2, ",", MockRack2IP1)

	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack1IPAddress, Role: usecase.RackRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack2IPAddress, Role: usecase.RackRole})
	assert.NoError(t, err)

	//Configure Fabric Should return no error

	//cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	//assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	//assert.NoError(t, err)
	//fmt.Println(cresp, err)

	Fabric, _ := DatabaseRepository.GetFabric(MockFabricName)
	Rack1IP1, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack1IP1)
	Rack1IP1SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack1IP1.ID)
	Rack1IP1InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP1.ID)
	Rack1IP1BGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP1.ID)
	Rack1IP1MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP1.ID)

	Rack1IP2, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack1IP2)
	Rack1IP2SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack1IP2.ID)
	Rack1IP2InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP2.ID)
	Rack1IP2MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack1IP2.ID)

	Rack2IP1, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack2IP1)
	Rack2IP1SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack2IP1.ID)
	Rack2IP1InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP1.ID)
	Rack2IP1BGPNeighborConfigs, err := DatabaseRepository.GetBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP1.ID)
	Rack2IP1MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP1.ID)

	Rack2IP2, _ := DatabaseRepository.GetDevice(MockFabricName, MockRack2IP2)
	Rack2IP2SwitchConfig, _ := DatabaseRepository.GetSwitchConfigOnFabricIDAndDeviceID(Fabric.ID, Rack2IP2.ID)
	Rack2IP2InterfaceConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP2.ID)
	Rack2IP2MCTBGPNeighborConfigs, err := DatabaseRepository.GetMCTBGPSwitchConfigsOnDeviceID(Fabric.ID, Rack2IP2.ID)

	Rack1, _ := DatabaseRepository.GetRack(Fabric.Name, Rack1IP1.IPAddress, Rack1IP2.IPAddress)
	Rack1EvpnNeighbors, _ := DatabaseRepository.GetRackEvpnConfig(Rack1.ID)

	Rack2, _ := DatabaseRepository.GetRack(Fabric.Name, Rack2IP1.IPAddress, Rack2IP2.IPAddress)
	Rack2EvpnNeighbors, _ := DatabaseRepository.GetRackEvpnConfig(Rack2.ID)

	assert.Equal(t, 4, len(Rack1EvpnNeighbors))
	assert.Equal(t, 4, len(Rack2EvpnNeighbors))
	for _, Rack1EvpnNeighbor := range Rack1EvpnNeighbors {
		assert.Equal(t, Rack1.ID, Rack1EvpnNeighbor.LocalRackID)
		assert.Equal(t, Rack2.ID, Rack1EvpnNeighbor.RemoteRackID)
		assert.Equal(t, Rack2IP1SwitchConfig.LocalAS, Rack1EvpnNeighbor.RemoteAS)
		assert.Condition(t, func() bool {
			if Rack1EvpnNeighbor.LocalDeviceID == Rack1IP1.ID || Rack1EvpnNeighbor.LocalDeviceID == Rack1IP2.ID {
				return true
			}
			return false
		})
		assert.Condition(t, func() bool {
			if Rack1EvpnNeighbor.RemoteDeviceID == Rack2IP1.ID || Rack1EvpnNeighbor.RemoteDeviceID == Rack2IP2.ID {
				return true
			}
			return false
		})
		if Rack1EvpnNeighbor.RemoteDeviceID == Rack2IP1.ID {
			assert.Equal(t, Rack2IP1SwitchConfig.LoopbackIP, Rack1EvpnNeighbor.EVPNAddress)
		} else {
			assert.Equal(t, Rack2IP2SwitchConfig.LoopbackIP, Rack1EvpnNeighbor.EVPNAddress)
		}

	}

	for _, Rack2EvpnNeighbor := range Rack2EvpnNeighbors {
		assert.Equal(t, Rack2.ID, Rack2EvpnNeighbor.LocalRackID)
		assert.Equal(t, Rack1.ID, Rack2EvpnNeighbor.RemoteRackID)
		assert.Equal(t, Rack1IP1SwitchConfig.LocalAS, Rack2EvpnNeighbor.RemoteAS)
		assert.Condition(t, func() bool {
			if Rack2EvpnNeighbor.LocalDeviceID == Rack2IP1.ID || Rack2EvpnNeighbor.LocalDeviceID == Rack2IP2.ID {
				return true
			}
			return false
		})
		assert.Condition(t, func() bool {
			if Rack2EvpnNeighbor.RemoteDeviceID == Rack1IP1.ID || Rack2EvpnNeighbor.RemoteDeviceID == Rack1IP2.ID {
				return true
			}
			return false
		})
		if Rack2EvpnNeighbor.RemoteDeviceID == Rack1IP1.ID {
			assert.Equal(t, Rack1IP1SwitchConfig.LoopbackIP, Rack2EvpnNeighbor.EVPNAddress)
		} else {
			assert.Equal(t, Rack1IP2SwitchConfig.LoopbackIP, Rack2EvpnNeighbor.EVPNAddress)
		}

	}

	Rack1IP1FabLink := ""
	//Interface Configs Rack1IP1
	for _, intf := range Rack1IP1InterfaceConfigs {
		if intf.IntName == "0/48" {
			assert.Contains(t, intf.IPAddress, "10.30.30")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
		} else {
			assert.Contains(t, intf.IPAddress, "10.10.10")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Equal(t, "0/1", intf.IntName)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
			Rack1IP1FabLink = intf.IPAddress
		}
	}

	assert.Contains(t, Rack1IP1SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack1IP1SwitchConfig.VTEPLoopbackIP, "172.31.254")
	assert.Contains(t, Rack1IP2SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack1IP2SwitchConfig.VTEPLoopbackIP, "172.31.254")
	//On Cluster both the IP's should be same
	assert.NotEqual(t, Rack1IP1SwitchConfig.LoopbackIP, Rack1IP2SwitchConfig.LoopbackIP)
	assert.Equal(t, Rack1IP1SwitchConfig.VTEPLoopbackIP, Rack1IP2SwitchConfig.VTEPLoopbackIP)
	//On Cluster ASN's should be same
	assert.Equal(t, Rack1IP1SwitchConfig.LocalAS, Rack1IP2SwitchConfig.LocalAS)
	//On Cluster MCT BGP Neighbors
	assert.Contains(t, Rack1IP1MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Contains(t, Rack1IP2MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Equal(t, Rack1IP2MCTBGPNeighborConfigs[0].RemoteAS, Rack1IP2MCTBGPNeighborConfigs[0].RemoteAS)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack1IP2MCTBGPNeighborConfigs[0].ConfigType)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack1IP2MCTBGPNeighborConfigs[0].ConfigType)
	//On L3 Backup
	for _, bgp := range Rack1IP1BGPNeighborConfigs {
		if bgp.Type == domain.MCTL3LBType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack1IP2InterfaceConfigs[0].IPAddress)
			assert.Equal(t, bgp.RemoteAS, Rack1IP2SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}
	}

	//Interface Configs Rack2IP1
	Rack2IP1FabLink := ""
	for _, intf := range Rack2IP1InterfaceConfigs {
		if intf.IntName == "0/48" {
			assert.Contains(t, intf.IPAddress, "10.30.30")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
		} else {
			assert.Contains(t, intf.IPAddress, "10.10.10")
			assert.Equal(t, domain.IntfTypeEthernet, intf.IntType)
			assert.Equal(t, "0/1", intf.IntName)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, intf.ConfigType)
			Rack2IP1FabLink = intf.IPAddress
		}
	}

	assert.Contains(t, Rack2IP1SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack2IP1SwitchConfig.VTEPLoopbackIP, "172.31.254")
	assert.Contains(t, Rack2IP2SwitchConfig.LoopbackIP, "172.31.254")
	assert.Contains(t, Rack2IP2SwitchConfig.VTEPLoopbackIP, "172.31.254")
	//On Cluster both the IP's should be same
	assert.NotEqual(t, Rack2IP1SwitchConfig.LoopbackIP, Rack2IP2SwitchConfig.LoopbackIP)
	assert.Equal(t, Rack2IP1SwitchConfig.VTEPLoopbackIP, Rack2IP2SwitchConfig.VTEPLoopbackIP)
	//On Cluster ASN's should be same
	assert.Equal(t, Rack2IP1SwitchConfig.LocalAS, Rack2IP2SwitchConfig.LocalAS)
	//On Cluster MCT BGP Neighbors
	assert.Contains(t, Rack2IP1MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Contains(t, Rack2IP2MCTBGPNeighborConfigs[0].RemoteIPAddress, "10.20.20")
	assert.Equal(t, Rack2IP2MCTBGPNeighborConfigs[0].RemoteAS, Rack2IP2MCTBGPNeighborConfigs[0].RemoteAS)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack2IP2MCTBGPNeighborConfigs[0].ConfigType)
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, Rack2IP2MCTBGPNeighborConfigs[0].ConfigType)
	//On L3 Backup
	for _, bgp := range Rack2IP1BGPNeighborConfigs {
		if bgp.Type == domain.MCTL3LBType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack2IP2InterfaceConfigs[0].IPAddress)
			assert.Equal(t, bgp.RemoteAS, Rack2IP2SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}
	}

	//BGP Neighbor Configs between Rack1IP1 and Rack2IP1
	for _, bgp := range Rack1IP1BGPNeighborConfigs {
		if bgp.Type == domain.FabricBGPType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack2IP1FabLink)
			assert.Equal(t, bgp.RemoteAS, Rack2IP1SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}

	}
	//BGP Neighbor Configs between Rack1IP1 and Rack2IP1
	for _, bgp := range Rack2IP1BGPNeighborConfigs {
		if bgp.Type == domain.FabricBGPType {
			assert.Equal(t, bgp.RemoteIPAddress, Rack1IP1FabLink)
			assert.Equal(t, bgp.RemoteAS, Rack1IP1SwitchConfig.LocalAS)
			assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, bgp.ConfigType)
		}

	}
}

//This test case configures/adds two Racks(one after and other)
func TestConfigure_OneRack_Followed_By_Other(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_1_48", ConfigState: "up"},

					//Interface to be connected to Rack2IP1
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M1_1_1", ConfigState: "up"},
				}, nil
			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M1_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M1_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M1_2_48", ConfigState: "up"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_1_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_1_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_1_48", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/1", Mac: "M2_1_1", ConfigState: "up"},
				}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/46", Mac: "M2_2_46", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/47", Mac: "M2_2_47", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: "ethernet", IntName: "0/48", Mac: "M2_2_48", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Rack1_IP1
			if DeviceIP == MockRack1IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M1_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M2_1_1"}}, nil

			}
			//Rack1_IP2
			if DeviceIP == MockRack1IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M1_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M1_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M1_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M1_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M1_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M1_1_48"}}, nil
			}
			//Rack2_IP1
			if DeviceIP == MockRack2IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_1_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_2_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_1_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_2_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_1_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_2_48"},

					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/1", LocalIntMac: "M2_1_1",
						RemoteIntType: "ethernet", RemoteIntName: "0/1", RemoteIntMac: "M1_1_1"}}, nil

			}
			//Rack2_IP2
			if DeviceIP == MockRack2IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/46", LocalIntMac: "M2_2_46",
					RemoteIntType: "ethernet", RemoteIntName: "0/46", RemoteIntMac: "M2_1_46"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/47", LocalIntMac: "M2_2_47",
						RemoteIntType: "ethernet", RemoteIntName: "0/47", RemoteIntMac: "M2_1_47"},
					domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
						LocalIntType: "ethernet", LocalIntName: "0/48", LocalIntMac: "M2_2_48",
						RemoteIntType: "ethernet", RemoteIntName: "0/48", RemoteIntMac: "M2_1_48"}}, nil
			}
			return []domain.LLDP{}, nil
		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {

			return "", nil
		},

		MockGetInterfaceSpeed: func(InterfaceType string, InterfaceName string) (int, error) {
			return 10000, nil
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
	resp, err := devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: MockRack1IP1}},
		UserName, Password, false)

	Rack1, _ := DatabaseRepository.GetRack(MockFabricName, MockRack1IP1, MockRack1IP2)
	Rack1EvpnNeighbors, _ := DatabaseRepository.GetRackEvpnConfig(Rack1.ID)

	// after adding rack1 there should not be any neighbor
	assert.Equal(t, 0, len(Rack1EvpnNeighbors))

	Rack1IPAddress := fmt.Sprintln(MockRack1IP2, ",", MockRack1IP1)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack1IPAddress, Role: usecase.RackRole})

	//Call the Add Devices method passing in list of leaf and list of spine address
	resp, err = devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack2IP2, IP2: MockRack2IP1}},
		UserName, Password, false)

	Rack2, _ := DatabaseRepository.GetRack(MockFabricName, MockRack2IP2, MockRack2IP1)
	Rack2EvpnNeighbors, _ := DatabaseRepository.GetRackEvpnConfig(Rack2.ID)
	// after adding rack2 there should not be 4 neighbors for Rack2
	assert.Equal(t, 4, len(Rack2EvpnNeighbors))

	Rack2IPAddress := fmt.Sprintln(MockRack2IP2, ",", MockRack2IP1)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: Rack2IPAddress, Role: usecase.RackRole})
	assert.NoError(t, err)

	//Configure Fabric Should return no error
	cresp, err := devUC.ConfigureFabric(context.Background(), MockFabricName, false, true)
	assert.Equal(t, usecase.ConfigureFabricResponse{FabricName: MockFabricName}, cresp)
	assert.NoError(t, err)
	fmt.Println(cresp, err)

}

//This test case configures a two node fabric(one spine and one leaf)
//tests the output from the add method
func TestConfigure_DeviceConfigured_With_Other_Rack(t *testing.T) {

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockRack1IP1 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "0/11", Mac: "M1", ConfigState: "up"}}, nil
			}
			//Leaf
			if DeviceIP == MockRack1IP2 {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: "ethernet", IntName: "1/22", Mac: "M2", ConfigState: "up"}}, nil
			}
			return []domain.Interface{}, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			//Spine
			if DeviceIP == MockRack1IP1 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "0/11", LocalIntMac: "M1",
					RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil
			}
			//Leaf
			if DeviceIP == MockRack1IP2 {
				return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
					LocalIntType: "ethernet", LocalIntName: "1/22", LocalIntMac: "M2",
					RemoteIntType: "ethernet", RemoteIntName: "0/11", RemoteIntMac: "M1"}}, nil
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
	resp, err := devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: MockRack1IP1}},
		UserName, Password, false)

	resp, err = devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP2, IP2: "TEST_IP2"}},
		UserName, Password, false)
	fmt.Println(resp)
	assert.Error(t, err, errors.New("RACK1_IP2 already present in another Rack"))

	resp, err = devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: MockRack1IP1, IP2: "TEST_IP2"}},
		UserName, Password, false)
	assert.Error(t, err, errors.New("RACK1_IP2 already present in another Rack"))

	resp, err = devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: "TEST_IP2", IP2: MockRack1IP1}},
		UserName, Password, false)
	fmt.Println(resp)
	assert.Error(t, err, errors.New("RACK1_IP2 already present in another Rack"))

	resp, err = devUC.AddRacks(context.Background(), MockFabricName, []usecase.Rack{usecase.Rack{IP1: "TEST_IP2", IP2: MockRack1IP2}},
		UserName, Password, false)
	assert.Error(t, err, errors.New("RACK1_IP2 already present in another Rack"))

}

package configurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"efa-server/test/unit/mock"
	"efa-server/usecase"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

//Fabric with one leaf and and one spine
//Ensure that all ip address gets allocated for interfaces and loopbacks
//Neighbors with ASN is allocated
func TestConfigure_NewInterface_ActionRequest(t *testing.T) {

	Link1IP1 := "10.10.10.0"
	Link1IP2 := "10.10.10.1"
	LoopbackRange := "172.31.254"
	ExpectedSpine1ASN := "64512"
	ExpectedLeaf1ASN := "65000"

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
	FabricProperties, _ := DatabaseRepository.GetFabricProperties(Fabric.ID)
	config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.Nil(t, err)

	leafConfig := config.Hosts[0]
	spineConfig := config.Hosts[1]

	if config.Hosts[0].Host == MockSpine1IP {
		spineConfig = config.Hosts[0]
		leafConfig = config.Hosts[1]
	}

	//For Spine
	//Verifying Fabric Properties
	assert.Equal(t, MockFabricName, config.FabricName)

	assert.Equal(t, usecase.SpineRole, spineConfig.Role)
	assert.Equal(t, MockSpine1IP, spineConfig.Device)
	assert.Equal(t, MockSpine1IP, spineConfig.Host)

	//BGP fields
	assert.Equal(t, FabricProperties.MaxPaths, spineConfig.MaxPaths)
	assert.Equal(t, ExpectedSpine1ASN, spineConfig.BgpLocalAsn)
	assert.Equal(t, FabricProperties.SpinePeerGroup, spineConfig.SpinePeerGroup)
	assert.Equal(t, FabricProperties.LeafPeerGroup, spineConfig.LeafPeerGroup)

	//Interface fields
	assert.Equal(t, FabricProperties.P2PLinkRange, spineConfig.P2pLinkRange)
	assert.Equal(t, FabricProperties.LoopBackPortNumber, spineConfig.LoopbackPortNumber)
	assert.Equal(t, FabricProperties.BFDTx, spineConfig.BfdTx)
	assert.Equal(t, FabricProperties.BFDRx, spineConfig.BfdRx)
	assert.Equal(t, FabricProperties.BFDMultiplier, spineConfig.BfdMultiplier)

	//OVG Fields
	assert.Equal(t, FabricProperties.VTEPLoopBackPortNumber, spineConfig.VtepLoopbackPortNumber)
	assert.Equal(t, FormatYesNo(FabricProperties.VNIAutoMap), spineConfig.VlanVniAutoMap)
	assert.Equal(t, FabricProperties.AnyCastMac, spineConfig.AnycastMac)
	assert.Equal(t, FabricProperties.IPV6AnyCastMac, spineConfig.IPV6AnycastMac)

	//EVPN Fields
	assert.Equal(t, FabricProperties.ArpAgingTimeout, spineConfig.ArpAgingTimeout)
	assert.Equal(t, FabricProperties.MacAgingTimeout, spineConfig.MacAgingTimeout)
	assert.Equal(t, FabricProperties.MacAgingConversationalTimeout, spineConfig.MacAgingConversationalTimeout)
	assert.Equal(t, FabricProperties.MacMoveLimit, spineConfig.MacMoveLimit)
	assert.Equal(t, FabricProperties.DuplicateMacTimer, spineConfig.DuplicateMacTimer)
	assert.Equal(t, FabricProperties.DuplicateMaxTimerMaxCount, spineConfig.DuplicateMaxTimerMaxCount)

	//Interfaces
	assert.Equal(t, 2, len(spineConfig.Interfaces))
	//Interface towards Leaf Device
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, spineConfig.Interfaces[0].ConfigType)
	assert.Contains(t, spineConfig.Interfaces[0].IP, "10.10.10.")
	Spine1IP := spineConfig.Interfaces[0].IP
	//Loopback Interface
	assert.Equal(t, domain.ConfigCreate, spineConfig.Interfaces[1].ConfigType)
	assert.Contains(t, spineConfig.Interfaces[1].IP, LoopbackRange)

	//BGP Neighbors
	assert.Equal(t, 1, len(spineConfig.BgpNeighbors))
	//Neighbor towards Leaf
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, spineConfig.BgpNeighbors[0].ConfigType)
	assert.Contains(t, spineConfig.BgpNeighbors[0].NeighborAddress, "10.10.10.")
	assert.Equal(t, ExpectedLeaf1ASN, fmt.Sprint(spineConfig.BgpNeighbors[0].RemoteAs))

	//For Leaf
	//Interfaces
	assert.Equal(t, 3, len(leafConfig.Interfaces))
	//Interface towards Spine
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, leafConfig.Interfaces[0].ConfigType)
	assert.Contains(t, leafConfig.Interfaces[0].IP, "10.10.10.")
	Leaf1IP := leafConfig.Interfaces[0].IP

	//Loopback Interface 1
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[1].ConfigType)
	assert.Contains(t, leafConfig.Interfaces[1].IP, LoopbackRange)
	//Loopback Interface 2
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[2].ConfigType)
	assert.Contains(t, leafConfig.Interfaces[2].IP, LoopbackRange)

	//BGP Neighbors
	assert.Equal(t, 1, len(leafConfig.BgpNeighbors))
	//Neighbor towards Spine
	assert.Contains(t, []string{domain.ConfigCreate, domain.ConfigUpdate}, leafConfig.BgpNeighbors[0].ConfigType)
	assert.Contains(t, leafConfig.BgpNeighbors[0].NeighborAddress, "10.10.10.")
	assert.Equal(t, ExpectedSpine1ASN, fmt.Sprint(leafConfig.BgpNeighbors[0].RemoteAs))

	assert.NotEqual(t, Spine1IP, Leaf1IP)
	assert.Contains(t, []string{Link1IP1 + "/31", Link1IP2 + "/31"}, Spine1IP)
	assert.Contains(t, []string{Link1IP1 + "/31", Link1IP2 + "/31"}, Leaf1IP)
}

//Devices have pre-existing IP address on the interfaces
func TestConfigure_ReserveValidIP_ActionRequest(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.12"
	ExpectedLeaf1IP := "10.10.10.13"
	LoopbackRange := "172.31.254"
	ExpectedSpine1ASN := "64512"
	ExpectedLeaf1ASN := "65000"

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
	config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.Nil(t, err)

	leafConfig := config.Hosts[0]
	spineConfig := config.Hosts[1]

	if config.Hosts[0].Host == MockSpine1IP {
		spineConfig = config.Hosts[0]
		leafConfig = config.Hosts[1]
	}

	//Ensure that the IP address are allocated the same IP address as on the switch
	//Interfaces

	assert.Equal(t, 2, len(spineConfig.Interfaces))
	assert.Equal(t, domain.ConfigCreate, spineConfig.Interfaces[0].ConfigType)
	assert.Equal(t, ExpectedSpine1IP+"/31", spineConfig.Interfaces[0].IP)
	assert.Equal(t, domain.ConfigCreate, spineConfig.Interfaces[1].ConfigType)
	assert.Contains(t, spineConfig.Interfaces[1].IP, LoopbackRange)

	//BGP Neighbors
	assert.Equal(t, 1, len(spineConfig.BgpNeighbors))
	assert.Equal(t, domain.ConfigCreate, spineConfig.BgpNeighbors[0].ConfigType)
	assert.Equal(t, ExpectedLeaf1IP, spineConfig.BgpNeighbors[0].NeighborAddress)
	assert.Equal(t, ExpectedLeaf1ASN, fmt.Sprint(spineConfig.BgpNeighbors[0].RemoteAs))

	//For Leaf
	//Interfaces
	assert.Equal(t, 3, len(leafConfig.Interfaces))
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[0].ConfigType)
	assert.Equal(t, ExpectedLeaf1IP+"/31", leafConfig.Interfaces[0].IP)
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[1].ConfigType)
	assert.Contains(t, leafConfig.Interfaces[1].IP, LoopbackRange)
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[2].ConfigType)

	assert.Contains(t, leafConfig.Interfaces[2].IP, LoopbackRange)

	//BGP Neighbors
	assert.Equal(t, 1, len(leafConfig.BgpNeighbors))
	assert.Equal(t, domain.ConfigCreate, leafConfig.BgpNeighbors[0].ConfigType)
	assert.Equal(t, ExpectedSpine1IP, leafConfig.BgpNeighbors[0].NeighborAddress)
	assert.Equal(t, ExpectedSpine1ASN, fmt.Sprint(leafConfig.BgpNeighbors[0].RemoteAs))

}

//Interface IP Address getting Update as part of second call to Add Device
//So only Physical Interface objects should have ConfigCreate
func TestConfigure_UseExistingInterface_Action_Request(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.12"
	ExpectedLeaf1IP := "10.10.10.13"
	ExpectedSpine1LP1IP := "172.31.254.1"
	ExpectedLeaf1LP1IP := "172.31.254.2"
	ExpectedLeaf1LP2IP := "172.31.254.3"
	ExpectedSpine1ASN := "64512"
	ExpectedLeaf1ASN := "65000"
	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedSpine1LP1IP + "/32", ConfigState: "up"},
				}, nil

			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLeaf1LP1IP + "/32", ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLeaf1LP2IP + "/32", ConfigState: "up"},
				}, nil
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

	//Simulating a success
	devUC.CleanupDBAfterConfigureSuccess()

	//do it for the second Time
	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP},
		UserName, Password, false)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: "test_fabric", FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.NoError(t, err)
	config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.Nil(t, err)

	leafConfig := config.Hosts[0]
	spineConfig := config.Hosts[1]

	if config.Hosts[0].Host == MockSpine1IP {
		spineConfig = config.Hosts[0]
		leafConfig = config.Hosts[1]
	}

	//Interfaces
	//Second time discovery so the interfaces should be None config
	assert.Equal(t, 2, len(spineConfig.Interfaces))
	assert.Equal(t, domain.ConfigNone, spineConfig.Interfaces[0].ConfigType)
	assert.Equal(t, ExpectedSpine1IP+"/31", spineConfig.Interfaces[0].IP)
	assert.Equal(t, domain.ConfigNone, spineConfig.Interfaces[1].ConfigType)
	assert.Equal(t, ExpectedSpine1LP1IP+"/32", spineConfig.Interfaces[1].IP)

	//BGP Neighbors
	//Second time discovery so the Neighbors should be None config
	assert.Equal(t, 1, len(spineConfig.BgpNeighbors))
	assert.Equal(t, domain.ConfigNone, spineConfig.BgpNeighbors[0].ConfigType)
	assert.Equal(t, ExpectedLeaf1IP, spineConfig.BgpNeighbors[0].NeighborAddress)
	assert.Equal(t, ExpectedLeaf1ASN, fmt.Sprint(spineConfig.BgpNeighbors[0].RemoteAs))

	//For Leaf
	//Interfaces
	//Second time discovery so the interfaces should be None config
	assert.Equal(t, 3, len(leafConfig.Interfaces))
	assert.Equal(t, domain.ConfigNone, leafConfig.Interfaces[0].ConfigType)
	assert.Equal(t, ExpectedLeaf1IP+"/31", leafConfig.Interfaces[0].IP)
	assert.Equal(t, domain.ConfigNone, leafConfig.Interfaces[1].ConfigType)
	assert.Equal(t, ExpectedLeaf1LP1IP+"/32", leafConfig.Interfaces[1].IP)
	assert.Equal(t, domain.ConfigNone, leafConfig.Interfaces[2].ConfigType)
	assert.Equal(t, ExpectedLeaf1LP2IP+"/32", leafConfig.Interfaces[2].IP)

	//BGP Neighbors
	//Second time discovery so the Neighbors should be None config
	assert.Equal(t, 1, len(leafConfig.BgpNeighbors))
	assert.Equal(t, domain.ConfigNone, leafConfig.BgpNeighbors[0].ConfigType)
	assert.Equal(t, ExpectedSpine1IP, leafConfig.BgpNeighbors[0].NeighborAddress)
	assert.Equal(t, ExpectedSpine1ASN, fmt.Sprint(leafConfig.BgpNeighbors[0].RemoteAs))

}

//Change IP address of Phy Interfaces on Seoond call for Add Device
func TestConfigure_ReserveDifferentIP_Action_Request(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.2"
	ExpectedLeaf1IP := "10.10.10.3"
	ExpectedSpine1IP2 := "10.10.10.4"
	ExpectedLeaf1IP2 := "10.10.10.5"
	LoopbackRange := "172.31.254"
	ExpectedSpine1ASN := "64512"
	ExpectedLeaf1ASN := "65000"

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

	MockDeviceAdapter2 := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				//No Loopback address - simulate the case of delete
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP2 + "/31",
					ConfigState: "up"}}, nil

			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				//No Loopback address - simulate the case of delete
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

	config, err := devUC.GetActionRequestObject(context.Background(), MockFabricName, false)
	assert.Nil(t, err)

	leafConfig := config.Hosts[0]
	spineConfig := config.Hosts[1]

	if config.Hosts[0].Host == MockSpine1IP {
		spineConfig = config.Hosts[0]
		leafConfig = config.Hosts[1]
	}
	//Interfaces
	assert.Equal(t, 2, len(spineConfig.Interfaces))
	//Physical Interface is updated
	assert.Equal(t, domain.ConfigUpdate, spineConfig.Interfaces[0].ConfigType)
	assert.Equal(t, ExpectedSpine1IP2+"/31", spineConfig.Interfaces[0].IP)
	//IP address got deleted on Switch
	assert.Equal(t, domain.ConfigCreate, spineConfig.Interfaces[1].ConfigType)
	assert.Contains(t, spineConfig.Interfaces[1].IP, LoopbackRange)

	//BGP Neighbors
	assert.Equal(t, 1, len(spineConfig.BgpNeighbors))
	//Physical Interface is updated so neighbor is updated
	assert.Equal(t, domain.ConfigUpdate, spineConfig.BgpNeighbors[0].ConfigType)
	assert.Equal(t, ExpectedLeaf1IP2, spineConfig.BgpNeighbors[0].NeighborAddress)
	assert.Equal(t, ExpectedLeaf1ASN, fmt.Sprint(spineConfig.BgpNeighbors[0].RemoteAs))

	//For Leaf
	//Interfaces
	assert.Equal(t, 3, len(leafConfig.Interfaces))
	//Physical Interface is updated
	assert.Equal(t, domain.ConfigUpdate, leafConfig.Interfaces[0].ConfigType)
	assert.Equal(t, ExpectedLeaf1IP2+"/31", leafConfig.Interfaces[0].IP)
	//IP address got deleted on Switch
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[1].ConfigType)
	assert.Contains(t, leafConfig.Interfaces[1].IP, LoopbackRange)
	//IP address got deleted on Switch
	assert.Equal(t, domain.ConfigCreate, leafConfig.Interfaces[2].ConfigType)
	assert.Contains(t, leafConfig.Interfaces[2].IP, LoopbackRange)

	//BGP Neighbors
	assert.Equal(t, 1, len(leafConfig.BgpNeighbors))
	//Physical Interface is updated so neighbor is updated
	assert.Equal(t, domain.ConfigUpdate, leafConfig.BgpNeighbors[0].ConfigType)
	assert.Equal(t, ExpectedSpine1IP2, leafConfig.BgpNeighbors[0].NeighborAddress)
	assert.Equal(t, ExpectedSpine1ASN, fmt.Sprint(leafConfig.BgpNeighbors[0].RemoteAs))

}

func TestConfigure_changeInInterfaceNumber_Action_Request(t *testing.T) {
	ExpectedSpine1IP := "10.10.10.2"
	ExpectedLeaf1IP := "10.10.10.3"
	ExpectedSpine1LP1IP := "172.31.254.1"
	ExpectedLeaf1LP1IP := "172.31.254.2"
	ExpectedLeaf1LP2IP := "172.31.254.3"

	//Single MOCK adapter serving both the devices
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			//Spine
			if DeviceIP == MockSpine1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedSpine1LP1IP + "/32",
						ConfigState: "up"}}, nil

			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/22", Mac: "M2", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLeaf1LP1IP + "/32",
						ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLeaf1LP2IP + "/32",
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
					IntType: domain.IntfTypeEthernet, IntName: "1/11", Mac: "M1", IPAddress: ExpectedSpine1IP + "/31",
					ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedSpine1LP1IP + "/32",
						ConfigState: "up"}}, nil

			}
			//Leaf
			if DeviceIP == MockLeaf1IP {
				return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeEthernet, IntName: "1/44", Mac: "M4", IPAddress: ExpectedLeaf1IP + "/31",
					ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLeaf1LP1IP + "/32",
						ConfigState: "up"},
					domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
						IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLeaf1LP2IP + "/32",
						ConfigState: "up"}}, nil
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
	fmt.Println(resp)
	//Leaf and spine  should be succesfully registered
	assert.Contains(t, fmt.Sprint(resp), "IPPair(10.10.10.2,10.10.10.3) not present in the Used IP Table for Fabric 1")
	assert.Contains(t, fmt.Sprint(resp), "IPPair(10.10.10.3,10.10.10.2) not present in the Used IP Table for Fabric 1")
	assert.Equal(t, "Add Device Operation Failed", err.Error())

}

func FormatYesNo(data string) bool {
	if data == "Yes" {
		return true
	}
	return false
}

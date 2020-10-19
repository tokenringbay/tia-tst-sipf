package deconfigurefabric

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
	"github.com/stretchr/testify/assert"
	//"os"
	"fmt"
	"os"
	"strconv"
)

var MockSpine1IP = "SPINE1_IP"
var MockLeaf1IP = "LEAF1_IP"
var MockFabricName = "test_fabric"
var UserName = "admin"
var Password = "password"
var DeConfigureDBName = constants.TESTDBLocation + "Deconf"

func TestDeConfigureSpineAndLeaf(t *testing.T) {

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
	database.Setup(DeConfigureDBName)
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

	SpineSwitchConfigAfterAdd, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockSpine1IP)
	SpineInterfaceSwitchConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(devUC.FabricID, SpineSwitchConfigAfterAdd.DeviceID)
	SpineLLDPNeighbors, err := DatabaseRepository.GetLLDPNeighborsOnDevice(devUC.FabricID, SpineSwitchConfigAfterAdd.DeviceID)
	fmt.Println(SpineLLDPNeighbors)

	LeafSwitchConfigAfterAdd, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	LeafInterfaceSwitchConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(devUC.FabricID, LeafSwitchConfigAfterAdd.DeviceID)
	LeafLLDPNeighbors, err := DatabaseRepository.GetLLDPNeighborsOnDevice(devUC.FabricID, LeafSwitchConfigAfterAdd.DeviceID)
	fmt.Println(LeafLLDPNeighbors)

	_, err = devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err)

	//Delete the device
	DeviceIPList := []string{MockSpine1IP, MockLeaf1IP}
	Error := devUC.DeleteDevices(context.Background(), MockFabricName, DeviceIPList, true)
	assert.NoError(t, Error.Error)

	validateAfterDelete(t, &devUC, MockFabricName, MockSpine1IP, SpineSwitchConfigAfterAdd, SpineInterfaceSwitchConfigs)
	validateAfterDelete(t, &devUC, MockFabricName, MockLeaf1IP, LeafSwitchConfigAfterAdd, LeafInterfaceSwitchConfigs)

}

func TestDeConfigureSpineAndLeafOneByOne(t *testing.T) {

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
	database.Setup(DeConfigureDBName)
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

	SpineSwitchConfigAfterAdd, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockSpine1IP)
	SpineInterfaceSwitchConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(devUC.FabricID, SpineSwitchConfigAfterAdd.DeviceID)
	SpineLLDPNeighbors, err := DatabaseRepository.GetLLDPNeighborsOnDevice(devUC.FabricID, SpineSwitchConfigAfterAdd.DeviceID)
	fmt.Println(SpineLLDPNeighbors)

	LeafSwitchConfigAfterAdd, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	LeafInterfaceSwitchConfigs, err := DatabaseRepository.GetInterfaceSwitchConfigsOnDeviceID(devUC.FabricID, LeafSwitchConfigAfterAdd.DeviceID)
	LeafLLDPNeighbors, err := DatabaseRepository.GetLLDPNeighborsOnDevice(devUC.FabricID, LeafSwitchConfigAfterAdd.DeviceID)
	fmt.Println(LeafLLDPNeighbors)

	_, err = devUC.ValidateFabricTopology(context.Background(), MockFabricName)
	assert.NoError(t, err)

	//Delete the device
	DeviceIPList := []string{MockSpine1IP}
	Error := devUC.DeleteDevices(context.Background(), MockFabricName, DeviceIPList, true)
	assert.NoError(t, Error.Error)
	validateAfterDelete(t, &devUC, MockFabricName, MockSpine1IP, SpineSwitchConfigAfterAdd, SpineInterfaceSwitchConfigs)

	DeviceIPList = []string{MockLeaf1IP}
	Error = devUC.DeleteDevices(context.Background(), MockFabricName, DeviceIPList, true)
	assert.NoError(t, Error.Error)
	validateAfterDelete(t, &devUC, MockFabricName, MockLeaf1IP, LeafSwitchConfigAfterAdd, LeafInterfaceSwitchConfigs)

}

func validateAfterDelete(t *testing.T, interactor *usecase.DeviceInteractor, FabricName string, IPAddress string,
	switchconfigAfterAdd domain.SwitchConfig, interfaceSwitchConfigs []domain.InterfaceSwitchConfig) {
	//No Device should be present and hence all tables will be cascade deleted
	DatabaseRepository := interactor.Db
	_, err := DatabaseRepository.GetDevice(FabricName, IPAddress)
	assert.NotNil(t, err)

	//There should be no LLDPS -- should be deleted by Cascade
	LLDPS, err := DatabaseRepository.GetLLDPsonDevice(interactor.FabricID, switchconfigAfterAdd.DeviceID)
	assert.Equal(t, 0, len(LLDPS))

	//There should be no Interfaces -- should be deleted by Cascade
	Interfaces, err := DatabaseRepository.GetInterfacesonDevice(interactor.FabricID, switchconfigAfterAdd.DeviceID)
	assert.Equal(t, 0, len(Interfaces))

	//ASN should be sent back to the Pool
	asn, _ := strconv.ParseUint(switchconfigAfterAdd.LocalAS, 10, 64)
	asnCount, _ := DatabaseRepository.GetASNAndCountOnASNAndRole(interactor.FabricID, asn, switchconfigAfterAdd.Role)
	assert.Equal(t, int64(1), asnCount)

	//Loopback Interface should be sent back to the Pool
	loopbackCount, _, _ := DatabaseRepository.GetIPEntryAndCountOnIPAddressAndType(interactor.FabricID, switchconfigAfterAdd.LoopbackIP, "Loopback")
	assert.Equal(t, int64(1), loopbackCount)

	if switchconfigAfterAdd.Role == usecase.LeafRole {
		//VTEP Loopback Interface should be sent back to the Pool
		vteploopbackCount, _, _ := DatabaseRepository.GetIPEntryAndCountOnIPAddressAndType(interactor.FabricID, switchconfigAfterAdd.VTEPLoopbackIP, "Loopback")
		assert.Equal(t, int64(1), vteploopbackCount)
	}

	//Check that all interface IP address has gone back to IP Pair Pool
	for _, interfaceConfig := range interfaceSwitchConfigs {
		fmt.Println(interfaceConfig.IPAddress)
		count, _, _ := DatabaseRepository.GetIPPairEntryAndCountOnEitherIPAddressAndType(interactor.FabricID, interfaceConfig.IPAddress, "P2P")
		assert.Equal(t, int64(1), count)
	}

}

func cleanupDB(Database *database.Database) {
	Database.Close()
	os.Remove(DeConfigureDBName)
}

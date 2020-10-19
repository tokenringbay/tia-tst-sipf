package clearfabric

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
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
)

var MockSpine1IP = "SPINE1_IP"
var MockLeaf1IP = "LEAF1_IP"
var MockFabricName = "efa"
var UserName = "admin"
var DBName = constants.TESTDBLocation + "Clear"
var Password = "password"

func TestClearSpine_And_Leaf_Test(t *testing.T) {

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

	database.Setup(DBName)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), MockFabricName)

	err := devUC.AddDevicesAndClearFabric(context.Background(), []string{MockLeaf1IP, MockSpine1IP}, []string{MockLeaf1IP, MockSpine1IP},
		UserName, Password)
	assert.NoError(t, err)

	//Ensure the clear fabric has been cleared

	_, err = DatabaseRepository.GetFabric("dummy_clear_fabric")
	assert.NotNil(t, err)

}

//Test Request objects sent for clear config
func TestClearSpine_And_Leaf_TestRequestObjects(t *testing.T) {

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

	database.Setup(DBName)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}

	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter),
		FabricAdapter: &mock.FabricAdapter{}}
	devUC.AddFabric(context.Background(), "default")

	err := devUC.DiscoverDevicesForClear(context.Background(), []string{MockLeaf1IP, MockSpine1IP}, []string{MockLeaf1IP, MockSpine1IP},
		UserName, Password)
	assert.NoError(t, err)

	clearFabricRequest, _, err := devUC.GenerateFabricConfigsForClear(context.Background(), []string{MockLeaf1IP, MockSpine1IP})
	//Two hosts
	assert.Equal(t, len(clearFabricRequest.Hosts), 2)
	//Leaf

	leafIndex := 0
	spineIndex := 1

	if clearFabricRequest.Hosts[1].Host == MockLeaf1IP {
		leafIndex = 1
		spineIndex = 0
	}
	fmt.Println(clearFabricRequest.Hosts)
	assert.Equal(t, len(clearFabricRequest.Hosts[leafIndex].Interfaces), 3)
	assert.Equal(t, clearFabricRequest.Hosts[leafIndex].Interfaces[0].InterfaceName, "1/22")
	assert.Equal(t, clearFabricRequest.Hosts[leafIndex].Interfaces[1].InterfaceName, "1")
	assert.Equal(t, clearFabricRequest.Hosts[leafIndex].Interfaces[2].InterfaceName, "2")

	//Spine
	assert.Equal(t, clearFabricRequest.Hosts[spineIndex].Host, MockSpine1IP)
	assert.Equal(t, len(clearFabricRequest.Hosts[spineIndex].Interfaces), 3)
	assert.Equal(t, clearFabricRequest.Hosts[spineIndex].Interfaces[0].InterfaceName, "1/11")
	assert.Equal(t, clearFabricRequest.Hosts[spineIndex].Interfaces[1].InterfaceName, "1")
	assert.Equal(t, clearFabricRequest.Hosts[spineIndex].Interfaces[2].InterfaceName, "2")

	assert.NoError(t, err)

}

func cleanupDB(Database *database.Database) {
	Database.Close()
	os.Remove(DBName)
}

package fetchconfig

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
var DBName = constants.TESTDBLocation + "fetch"

func TestFetchFabricConfigSpine_And_Leaf_Test(t *testing.T) {

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

	devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{MockSpine1IP}, "admin", "password", false)

	Fabric, _ := devUC.Db.GetFabric(MockFabricName)
	FabricRequest, err := devUC.PrepareActionFabricFetchRequest(context.Background(), &Fabric, "all")
	fmt.Println(FabricRequest, err)
	devUC.FetchFabricConfigs(context.Background(), MockFabricName, "all")
	assert.NoError(t, err)

}

func cleanupDB(Database *database.Database) {
	Database.Close()
	os.Remove(DBName)
}

package ippool

import (
	"context"
	"efa-server/gateway"
	"efa-server/infra/database"
	"efa-server/test/unit/mock"
	"efa-server/usecase"
	"testing"

	"efa-server/domain"
	"efa-server/infra/constants"
	"github.com/stretchr/testify/assert"
	"os"
)

var (
	IPPoolRange        = "10.3.0.24/24"
	IPPoolRangeSmall   = "10.3.0.0/31"
	IPPoolRangeInvalid = "10.340.0.23/24"
	IPDBName           = constants.TESTDBLocation + "ip"
	MockFabricName     = "efa-test"
)

func TestPool_PopulateIP(t *testing.T) {

	database.Setup(IPDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIP(context.Background(), MockFabricName, Fabric.ID, IPPoolRange, "Loopback", false)
	assert.Nil(t, err, "")
}

func TestPool_PopulateIP_Invalid(t *testing.T) {

	database.Setup(IPDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIP(context.Background(), MockFabricName, Fabric.ID, IPPoolRangeInvalid, "Loopback", false)
	assert.NotNil(t, err, "")
}
func TestPool_PopulateIP_withoutFabric(t *testing.T) {
	database.Setup(IPDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	err := devUC.PopulateIP(context.Background(), MockFabricName, Fabric.ID, IPPoolRange, "Loopback", false)
	assert.NotNil(t, err, "")
}

func TestPool_GetIPExhausted(t *testing.T) {
	database.Setup(IPDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIP(context.Background(), MockFabricName, Fabric.ID, IPPoolRangeSmall, "Loopback", true)
	assert.Nil(t, err, "")

	//2 - Add Dummy devices and interfaces, call GetIPPair which allocates IP's to the pair of interfaces
	//Add Dummy Devices with Interfaces
	DeviceOne := domain.Device{IPAddress: "IPAddress1", FabricID: Fabric.ID}
	DeviceTwo := domain.Device{IPAddress: "IPAddress2", FabricID: Fabric.ID}
	devUC.Db.CreateDevice(&DeviceOne)
	devUC.Db.CreateDevice(&DeviceTwo)

	InterfaceOne := domain.Interface{IntName: "0/25", FabricID: Fabric.ID, DeviceID: DeviceOne.ID}
	InterfaceTwo := domain.Interface{IntName: "0/26", FabricID: Fabric.ID, DeviceID: DeviceTwo.ID}
	devUC.Db.CreateInterface(&InterfaceOne)
	devUC.Db.CreateInterface(&InterfaceTwo)

	//Get a new set of IP Adddess for the interfaces
	_, err = devUC.GetIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", InterfaceOne.ID)
	_, err = devUC.GetIP(context.Background(), Fabric.ID, DeviceTwo.ID, "Loopback", InterfaceTwo.ID)
	assert.NotNil(t, err)
}

func TestPool_GetAlreadyAllocatedIP(t *testing.T) {

	database.Setup(IPDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIP(context.Background(), MockFabricName, Fabric.ID, IPPoolRange, "Loopback", false)
	assert.Nil(t, err, "")

	//1 - Interface does not exist, so it should return error
	//Both the Device and Interface does not exist, hence return Error
	_, err = devUC.GetAlreadyAllocatedIP(context.Background(), Fabric.ID, 1, "Loopback", 1)
	assert.NotNil(t, err, "")

	//2 - Add Dummy devices and interfaces, call GetIPPair which allocates IP's to the pair of interfaces
	//Add Dummy Devices with Interfaces
	DeviceOne := domain.Device{IPAddress: "IPAddress1", FabricID: Fabric.ID}
	devUC.Db.CreateDevice(&DeviceOne)

	InterfaceOne := domain.Interface{IntName: "0/25", FabricID: Fabric.ID, DeviceID: DeviceOne.ID}
	devUC.Db.CreateInterface(&InterfaceOne)

	//Get a new set of IP Adddess for the interfaces
	ip, err := devUC.GetIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", InterfaceOne.ID)
	//First set of IP address should be returned
	assert.Equal(t, "10.3.0.0", ip)
	assert.Nil(t, err)

	//Fetch already allocated IP for these interfaces
	aip, err := devUC.GetAlreadyAllocatedIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", InterfaceOne.ID)
	assert.Equal(t, "10.3.0.0", aip)
	assert.Nil(t, err)

	//Ensure that the IP address are not in the available Pool
	count, _ := devUC.GetIPCountInPool(context.Background(), Fabric.ID, aip, "Loopback")
	assert.Equal(t, int64(0), count)

	//3 - Release the IP for the Pair of Interfaces and fetch again
	rerr := devUC.ReleaseIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", ip, InterfaceOne.ID)
	assert.Nil(t, rerr)
	_, raerr := devUC.GetAlreadyAllocatedIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", InterfaceOne.ID)
	assert.NotNil(t, raerr)
	//Ensure that the IP Pair is back in the pool because of Release operation
	count, _ = devUC.GetIPCountInPool(context.Background(), Fabric.ID, aip, "Loopback")
	assert.Equal(t, int64(1), count)

}

func TestPool_ReserveIPSuccess(t *testing.T) {
	database.Setup(IPDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIP(context.Background(), MockFabricName, Fabric.ID, IPPoolRange, "Loopback", false)
	assert.Nil(t, err, "")

	//2 - Add Dummy devices and interfaces, Try Reserving a Valid IP Address
	DeviceOne := domain.Device{IPAddress: "IPAddress1", FabricID: Fabric.ID}
	devUC.Db.CreateDevice(&DeviceOne)

	InterfaceOne := domain.Interface{IntName: "0/25", FabricID: Fabric.ID, DeviceID: DeviceOne.ID}
	devUC.Db.CreateInterface(&InterfaceOne)

	ReserverIP := "10.3.0.2"
	//Reserve a new set of IP Adddess for the interfaces
	err = devUC.ReserveIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", ReserverIP, InterfaceOne.ID)
	assert.Nil(t, err)

	//Fetch already allocated IP for these interfaces
	aip, err := devUC.GetAlreadyAllocatedIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", InterfaceOne.ID)
	assert.Equal(t, ReserverIP, aip)
	assert.Nil(t, err)

	//Ensure that the IP address are not in the available Pool
	//Ensure that the IP address are not in the available Pool
	count, _ := devUC.GetIPCountInPool(context.Background(), Fabric.ID, aip, "Loopback")
	assert.Equal(t, int64(0), count)

	//3 - Release the IP for the Pair of Interfaces and fetch again
	rerr := devUC.ReleaseIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", aip, InterfaceOne.ID)
	assert.Nil(t, rerr)
	_, raerr := devUC.GetAlreadyAllocatedIP(context.Background(), Fabric.ID, DeviceOne.ID, "Loopback", InterfaceOne.ID)
	assert.NotNil(t, raerr)
	//Ensure that the IP Pair is back in the pool because of Release operation
	count, _ = devUC.GetIPCountInPool(context.Background(), Fabric.ID, aip, "Loopback")
	assert.Equal(t, int64(1), count)
}
func cleanupDB(Database *database.Database, DBName string) {
	Database.Close()
	os.Remove(DBName)
}

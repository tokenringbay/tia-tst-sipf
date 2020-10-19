package ippairpool

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
	IPPairPoolRange        = "10.3.0.24/24"
	IPPairPoolRangeSmall   = "10.3.0.0/31"
	IPPairPoolRangeInvalid = "10.340.0.23/24"
	MockFabricName         = "efa-test"
	IPPairDBName           = constants.TESTDBLocation + "ip-pair"
	FabricID               = 1
)

func TestIPPairPool_PopulateIPPair(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, Fabric.ID, IPPairPoolRange, "P2P", false)
	assert.Nil(t, err, "")
}

func TestIPPairPool_PopulateIPPair_Invalid(t *testing.T) {

	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, Fabric.ID, IPPairPoolRangeInvalid, "P2P", false)
	assert.NotNil(t, err, "")
}

func TestIPPairPool_PopulateIPPair_withoutFabric(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, Fabric.ID, IPPairPoolRange, "P2P", false)
	assert.NotNil(t, err, "")
}
func TestIPPairPool_GetIPPairExhausted(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, Fabric.ID, IPPairPoolRangeSmall, "P2P", true)
	assert.Nil(t, err, "")

	//1 - Interface does not exist, so it should return error
	//Both the Device and Interface does not exist, hence return Error
	_, _, err = devUC.GetAlreadyAllocatedIPPair(context.Background(), Fabric.ID, 1, 2, "P2P", 1, 2)
	assert.NotNil(t, err, "")

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
	_, _, err = devUC.GetIPPair(context.Background(), Fabric.ID, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.NotNil(t, err)
}
func TestIPPairPool_GetAlreadyAllocatedIPPair(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	Fabric := domain.Fabric{Name: MockFabricName}
	devUC.Db.CreateFabric(&Fabric)
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, Fabric.ID, IPPairPoolRange, "P2P", false)
	assert.Nil(t, err, "")

	//1 - Interface does not exist, so it should return error
	//Both the Device and Interface does not exist, hence return Error
	_, _, err = devUC.GetAlreadyAllocatedIPPair(context.Background(), Fabric.ID, 1, 2, "P2P", 1, 2)
	assert.NotNil(t, err, "")

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
	ip1, ip2, err := devUC.GetIPPair(context.Background(), Fabric.ID, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	//First set of IP address should be returned
	assert.Equal(t, "10.3.0.0", ip1)
	assert.Equal(t, "10.3.0.1", ip2)
	assert.Nil(t, err)

	//Fetch already allocated IP for these interfaces
	aip1, aip2, err := devUC.GetAlreadyAllocatedIPPair(context.Background(), Fabric.ID, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.Equal(t, "10.3.0.0", aip1)
	assert.Equal(t, "10.3.0.1", aip2)
	assert.Nil(t, err)

	//Ensure that the IP address are not in the available Pool
	count, _ := devUC.GetIPPairCountInPool(context.Background(), Fabric.ID, aip1, aip2, "P2P")
	assert.Equal(t, int64(0), count)

	//3 - Release the IP for the Pair of Interfaces and fetch again
	rerr := devUC.ReleaseIPPair(context.Background(), Fabric.ID, DeviceOne.ID, DeviceTwo.ID, "P2P", ip1, ip2, InterfaceOne.ID, InterfaceTwo.ID)
	assert.Nil(t, rerr)
	_, _, raerr := devUC.GetAlreadyAllocatedIPPair(context.Background(), Fabric.ID, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.NotNil(t, raerr)
	//Ensure that the IP Pair is back in the pool because of Release operation
	count, _ = devUC.GetIPPairCountInPool(context.Background(), Fabric.ID, aip1, aip2, "P2P")
	assert.Equal(t, int64(1), count)

}

func TestIPPairPool_ReserveIPPairSuccess(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	devUC.Db.CreateFabric(&domain.Fabric{Name: MockFabricName})
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, uint(FabricID), IPPairPoolRange, "P2P", false)
	assert.Nil(t, err, "")

	//2 - Add Dummy devices and interfaces, Try Reserving a Valid IP Address
	DeviceOne := domain.Device{IPAddress: "IPAddress1", FabricID: 1}
	DeviceTwo := domain.Device{IPAddress: "IPAddress2", FabricID: 1}
	devUC.Db.CreateDevice(&DeviceOne)
	devUC.Db.CreateDevice(&DeviceTwo)

	InterfaceOne := domain.Interface{IntName: "0/25", FabricID: 1, DeviceID: DeviceOne.ID}
	InterfaceTwo := domain.Interface{IntName: "0/26", FabricID: 1, DeviceID: DeviceTwo.ID}
	devUC.Db.CreateInterface(&InterfaceOne)
	devUC.Db.CreateInterface(&InterfaceTwo)

	ReserverIP1 := "10.3.0.2"
	ReserverIP2 := "10.3.0.3"
	//Reserve a new set of IP Adddess for the interfaces
	err = devUC.ReserveIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", ReserverIP1, ReserverIP2, InterfaceOne.ID, InterfaceTwo.ID)
	assert.Nil(t, err)

	//Fetch already allocated IP for these interfaces
	aip1, aip2, err := devUC.GetAlreadyAllocatedIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.Equal(t, ReserverIP1, aip1)
	assert.Equal(t, ReserverIP2, aip2)
	assert.Nil(t, err)

	//Ensure that the IP address are not in the available Pool
	count, _ := devUC.GetIPPairCountInPool(context.Background(), 1, aip1, aip2, "P2P")
	assert.Equal(t, int64(0), count)

	//3 - Release the IP for the Pair of Interfaces and fetch again
	rerr := devUC.ReleaseIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", ReserverIP1, ReserverIP2, InterfaceOne.ID, InterfaceTwo.ID)
	assert.Nil(t, rerr)
	_, _, raerr := devUC.GetAlreadyAllocatedIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.NotNil(t, raerr)
	//Ensure that the IP Pair is back in the pool because of Release operation
	count, _ = devUC.GetIPPairCountInPool(context.Background(), 1, aip1, aip2, "P2P")
	assert.Equal(t, int64(1), count)

}

func TestIPPairPool_ReserveIPPairFailDueToInvalidPair(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	devUC.Db.CreateFabric(&domain.Fabric{Name: MockFabricName})
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, uint(FabricID), IPPairPoolRange, "P2P", false)
	assert.Nil(t, err, "")

	//2 - Add Dummy devices and interfaces, Try Reserving a Invalid IP Address Pair
	DeviceOne := domain.Device{IPAddress: "IPAddress1", FabricID: 1}
	DeviceTwo := domain.Device{IPAddress: "IPAddress2", FabricID: 1}
	devUC.Db.CreateDevice(&DeviceOne)
	devUC.Db.CreateDevice(&DeviceTwo)

	InterfaceOne := domain.Interface{IntName: "0/25", FabricID: 1, DeviceID: DeviceOne.ID}
	InterfaceTwo := domain.Interface{IntName: "0/26", FabricID: 1, DeviceID: DeviceTwo.ID}
	devUC.Db.CreateInterface(&InterfaceOne)
	devUC.Db.CreateInterface(&InterfaceTwo)

	ReserverIP1 := "10.3.0.21"
	ReserverIP2 := "10.3.0.3"
	//Reserve a new set of IP Adddess for the interfaces
	err = devUC.ReserveIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", ReserverIP1, ReserverIP2, InterfaceOne.ID, InterfaceTwo.ID)
	assert.Equal(t, "IPPair(10.3.0.21,10.3.0.3) not present in the Used IP Table for Fabric 1 Device (1,2)", err.Error())
}

func TestIPPairPool_ReserveIPPairSuccessOutOfOrder(t *testing.T) {
	database.Setup(IPPairDBName)
	defer cleanupDB(database.GetWorkingInstance(), IPPairDBName)

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}
	devUC.Db.CreateFabric(&domain.Fabric{Name: MockFabricName})
	err := devUC.PopulateIPPairs(context.Background(), MockFabricName, uint(FabricID), IPPairPoolRange, "P2P", false)
	assert.Nil(t, err, "")

	//2 - Add Dummy devices and interfaces, Try Reserving a Valid IP Address
	DeviceOne := domain.Device{IPAddress: "IPAddress1", FabricID: 1}
	DeviceTwo := domain.Device{IPAddress: "IPAddress2", FabricID: 1}
	devUC.Db.CreateDevice(&DeviceOne)
	devUC.Db.CreateDevice(&DeviceTwo)

	InterfaceOne := domain.Interface{IntName: "0/25", FabricID: 1, DeviceID: DeviceOne.ID}
	InterfaceTwo := domain.Interface{IntName: "0/26", FabricID: 1, DeviceID: DeviceTwo.ID}
	devUC.Db.CreateInterface(&InterfaceOne)
	devUC.Db.CreateInterface(&InterfaceTwo)

	//Get a new set of IP Adddess for the interfaces
	ip1, ip2, err := devUC.GetIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	//First set of IP address should be returned
	assert.Equal(t, "10.3.0.0", ip1)
	assert.Equal(t, "10.3.0.1", ip2)
	assert.Nil(t, err)

	ReserverIP1 := "10.3.0.0"
	ReserverIP2 := "10.3.0.1"
	//Reserve the same set of Interfaces where the order of interfaces and devices are interchanged
	err = devUC.ReserveIPPair(context.Background(), 1, DeviceTwo.ID, DeviceOne.ID, "P2P", ReserverIP2, ReserverIP1, InterfaceTwo.ID, InterfaceOne.ID)
	assert.Nil(t, err)

	//Fetch already allocated IP for these interfaces
	aip1, aip2, err := devUC.GetAlreadyAllocatedIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.Equal(t, ReserverIP1, aip1)
	assert.Equal(t, ReserverIP2, aip2)
	assert.Nil(t, err)

	//Change the Order of Query for the already allocated interfaces
	aip1, aip2, err = devUC.GetAlreadyAllocatedIPPair(context.Background(), 1, DeviceTwo.ID, DeviceOne.ID, "P2P", InterfaceTwo.ID, InterfaceOne.ID)
	//Ensure IP address are recieved in the order they are queried
	assert.Equal(t, ReserverIP2, aip1)
	assert.Equal(t, ReserverIP1, aip2)
	assert.Nil(t, err)

	//Ensure that the IP address are not in the available Pool
	count, _ := devUC.GetIPPairCountInPool(context.Background(), 1, aip1, aip2, "P2P")
	assert.Equal(t, int64(0), count)

	//3 - Release the IP for the Pair of Interfaces Out of Order and fetch again
	rerr := devUC.ReleaseIPPair(context.Background(), 1, DeviceTwo.ID, DeviceOne.ID, "P2P", ReserverIP2, ReserverIP1, InterfaceTwo.ID, InterfaceOne.ID)
	assert.Nil(t, rerr)
	_, _, raerr := devUC.GetAlreadyAllocatedIPPair(context.Background(), 1, DeviceOne.ID, DeviceTwo.ID, "P2P", InterfaceOne.ID, InterfaceTwo.ID)
	assert.NotNil(t, raerr)
	//Ensure that the IP Pair is back in the pool because of Release operation
	count, _ = devUC.GetIPPairCountInPool(context.Background(), 1, aip1, aip2, "P2P")
	assert.Equal(t, int64(1), count)

}

func cleanupDB(Database *database.Database, DBName string) {
	Database.Close()
	os.Remove(DBName)
}

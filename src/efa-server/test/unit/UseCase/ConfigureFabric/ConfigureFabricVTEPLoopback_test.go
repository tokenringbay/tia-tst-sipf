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
	"github.com/stretchr/testify/assert"
	"testing"
)

//Test cases for addition/update/delete of VTEPLoop back

//Obtain a new loopback for Leaf
//Create Config on Switch
func TestConfigure_NewLeafVTEPLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.2"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first loopback from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfig.VTEPLoopbackIPConfigType)

}

//Reserve a valid loopback for Leaf
//No need to push to switch
func TestConfigure_ReserveValidVTEPLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLoopbackIP + "/32"}}, nil

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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.VTEPLoopbackIPConfigType)

}

//Reserve a Invalid loopback for Leaf
func TestConfigure_ReserveInValidVTEPLoopback(t *testing.T) {
	ExpectedLoopbackIP := "171.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLoopbackIP + "/32"}}, nil

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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole,
		Errors: []error{errors.New("171.31.254.12 loopback on 2 not in 172.31.254.0/24 Range")}})

	assert.NotNil(t, err)

}

//Reserve a valid loopback for Leaf
//Multiple times
func TestConfigure_useExistingValidVTEPLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLoopbackIP + "/32"}}, nil

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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//No need to push to switch
	assert.Equal(t, domain.ConfigNone, switchConfig.VTEPLoopbackIPConfigType)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.VTEPLoopbackIPConfigType)

}

func TestConfigure_SecondCallWithEmptyVTEPLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.2"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
			}, nil

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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)
	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//Need to push to switch
	assert.Equal(t, domain.ConfigCreate, switchConfig.VTEPLoopbackIPConfigType)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfig.VTEPLoopbackIPConfigType)

}

func TestConfigure_FirstCallReserveSecondCallWithEmptyVTEPLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//No need to push to switch
	assert.Equal(t, domain.ConfigNone, switchConfig.VTEPLoopbackIPConfigType)

	MockLeafDeviceAdapterEmptyVTEPLoopback := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
			}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}
	devUC.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockLeafDeviceAdapterEmptyVTEPLoopback)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//Need to push to switch
	assert.Equal(t, domain.ConfigCreate, switchConfig.VTEPLoopbackIPConfigType)

}

//Reserve a valid loopback for Leaf
//Multiple times
func TestConfigure_ReserveDifferentValidVTEPLoopbackSecondTime(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	ExpectedLoopbackIP2 := "172.31.254.13"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.VTEPLoopbackIP)
	//No need to push to switch
	assert.Equal(t, domain.ConfigNone, switchConfig.VTEPLoopbackIPConfigType)

	MockLeafSecondDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "2", IPAddress: ExpectedLoopbackIP2 + "/32", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}
	devUC.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockLeafSecondDeviceAdapter)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP2, switchConfig.VTEPLoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigUpdate, switchConfig.VTEPLoopbackIPConfigType)

}

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

//Test-cases for addition/updation of loopback address

//Obtain a new loopback for Leaf
//Config Type would be ConfigCreate
func TestConfigure_NewspineLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.1"
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
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfig.LoopbackIPConfigType)

}

//Reserve a valid loopback for Leaf
//Since Switch already has the Loopback IP, so no need to push
func TestConfigure_ReserveValidLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.LoopbackIPConfigType)

}

//Reserve a Invalid loopback for Leaf
func TestConfigure_ReserveInValidLoopback(t *testing.T) {
	ExpectedLoopbackIP := "171.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
		Errors: []error{errors.New("171.31.254.12 loopback on 1 not in 172.31.254.0/24 Range")}})
	assert.NotNil(t, err)

}

//Reserve a valid loopback for Leaf
//Multiple times
func TestConfigure_useExistingValidLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//Verify loopback already on Switch
	assert.Equal(t, domain.ConfigNone, switchConfig.LoopbackIPConfigType)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.LoopbackIPConfigType)

}

func TestConfigure_SecondCallWithEmptyLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.1"
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
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//First Create the Loopback
	assert.Equal(t, domain.ConfigCreate, switchConfig.LoopbackIPConfigType)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//Need to push to switch (as it is not there on the switch)
	assert.Equal(t, domain.ConfigCreate, switchConfig.LoopbackIPConfigType)

}

func TestConfigure_FirstCallReserveSecondCallWithEmptyLoopback(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//No need to push as it already exists on Switch
	assert.Equal(t, domain.ConfigNone, switchConfig.LoopbackIPConfigType)

	MockLeafDeviceAdapterEmtpyLoopback := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: "", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
	}

	devUC.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockLeafDeviceAdapterEmtpyLoopback)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})
	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//Need to push to switch (as it is not there on the switch)
	assert.Equal(t, domain.ConfigCreate, switchConfig.LoopbackIPConfigType)

}

//Reserve a valid loopback for Leaf
//Multiple times
func TestConfigure_ReserveDifferentValidLoopbackSecondTime(t *testing.T) {
	ExpectedLoopbackIP := "172.31.254.12"
	ExpectedLoopbackIP2 := "172.31.254.13"
	MockLeafDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLoopbackIP + "/32", ConfigState: "up"}}, nil

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
	assert.Equal(t, ExpectedLoopbackIP, switchConfig.LoopbackIP)
	//No need to push as it is already there on Switch
	assert.Equal(t, domain.ConfigNone, switchConfig.LoopbackIPConfigType)

	MockLeafSecondDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"},
				domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
					IntType: domain.IntfTypeLoopback, IntName: "1", IPAddress: ExpectedLoopbackIP2 + "/32", ConfigState: "up"}}, nil

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
	assert.Nil(t, err)
	switchConfig, err = DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	//Verify loopback is the first ASN from the Pool
	assert.Equal(t, ExpectedLoopbackIP2, switchConfig.LoopbackIP)
	//Verify loopback is to be created
	assert.Equal(t, domain.ConfigUpdate, switchConfig.LoopbackIPConfigType)

}

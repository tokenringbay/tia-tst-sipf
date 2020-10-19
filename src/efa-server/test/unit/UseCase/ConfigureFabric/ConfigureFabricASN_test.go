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

//Contains Test Cases to test allocation of ASN's when Devices are added or Update

//Obtain a new ASN for Leaf -- It should pick up 65000 (as per default Fabric Properties)
//Config Type has to be ConfigCreate
func TestConfigure_NewLeafASN(t *testing.T) {
	ExpectedASN := "65000"

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
	//Verify ASN is the first ASN from the Pool
	assert.Equal(t, ExpectedASN, switchConfig.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfig.ASConfigType)

}

//Obtain a new ASN for Leaf -- It should pick up 64512 (as per default Fabric Properties)
//Config Type has to be ConfigCreate
func TestConfigure_NewSpineASN(t *testing.T) {
	ExpectedASN := "64512"

	MockSpineDeviceAdapter := mock.DeviceAdapter{
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
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})

	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockSpine1IP)
	//Verify ASN is the first ASN from the Pool
	assert.Equal(t, ExpectedASN, switchConfig.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfig.ASConfigType)
}

//Reserve using a Valid ASN on Leaf
//As the Leaf already has the ASN config should be NONE (Nothing to be pushed to switch)
func TestConfigure_ValidLeafASN(t *testing.T) {
	ASNOnDevice := "65001"

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
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
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
	assert.Equal(t, ASNOnDevice, switchConfig.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.ASConfigType)
}

//Try Reserve using a InValid ASN on Leaf
func TestConfigure_InvalidLeafASN(t *testing.T) {
	ASNOnDevice := "64519"
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
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
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
		Errors: []error{errors.New("ASN 64519 not in range 65000-65534")}})

	assert.NotNil(t, err)

}

//Reserve using a Valid ASN on Spine,
//As the switch already has the ASN config_type has to be ConfigNone
func TestConfigure_ValidSpineASN(t *testing.T) {
	ASNOnDevice := "64512"
	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole})

	assert.Nil(t, err)
	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockSpine1IP)
	assert.Equal(t, ASNOnDevice, switchConfig.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.ASConfigType)

}

//Reserve using a InValid ASN on Spine
func TestConfigure_InvalidSpineASN(t *testing.T) {
	ASNOnDevice := "64519"
	MockSpineDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{}, []string{MockSpine1IP},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockSpine1IP, Role: usecase.SpineRole,
		Errors: []error{errors.New("ASN 64519 not in range 64512")}})

	assert.NotNil(t, err)

}

//Reserve ASN from Valid Range (Device having an ASN)
//Call next time asking for Same ASN,Config Type has to be NONE
func TestConfigure_ReserveSameASN(t *testing.T) {
	ASNOnDevice := "65000"

	MockLeafDeviceAdapterInitial := mock.DeviceAdapter{
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
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
		},
	}

	database.Setup(constants.TESTDBLocation)
	defer cleanupDB(database.GetWorkingInstance())

	DatabaseRepository := gateway.DatabaseRepository{Database: database.GetWorkingInstance()}
	devUC := usecase.DeviceInteractor{Db: &DatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockLeafDeviceAdapterInitial)}
	devUC.AddFabric(context.Background(), MockFabricName)

	resp, err := devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)

	switchConfig, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	assert.Equal(t, ASNOnDevice, switchConfig.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfig.ASConfigType)

	//Call again requesting for SAME ASN
	devUC.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockLeafDeviceAdapter)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfigSecond, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	assert.Equal(t, ASNOnDevice, switchConfigSecond.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigNone, switchConfigSecond.ASConfigType)

}

//Reserve ASN from Valid Range (Device having an ASN)
//Call next time without ASN on the Device. The existing ASN should be reserved.
func TestConfigure_UseExistingASN(t *testing.T) {
	ASNOnDevice := "65001"
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
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
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
	assert.Equal(t, ASNOnDevice, switchConfig.LocalAS)
	//Verify ASN is already on the Switch, so need to push
	assert.Equal(t, domain.ConfigNone, switchConfig.ASConfigType)

	//Next Call without Device having ASN
	MockLeafDeviceAdapterWithoutASN := mock.DeviceAdapter{
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
	devUC.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockLeafDeviceAdapterWithoutASN)
	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfigSecond, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	assert.Equal(t, ASNOnDevice, switchConfigSecond.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigCreate, switchConfigSecond.ASConfigType)
}

//Reserve ASN from Valid Range (Device having an ASN)
//Call next time to Reserve ASN from another Valid Range
func TestConfigure_ReserveDifferentASNSecondTime(t *testing.T) {
	ASNOnDevice := "65001"
	ASNOnDevice2 := "65002"

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
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice, nil
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
	assert.Equal(t, ASNOnDevice, switchConfig.LocalAS)
	//Verify ASN is to be created
	assert.Equal(t, domain.ConfigNone, switchConfig.ASConfigType)

	//Next Call without Device having ASN
	MockLeafDeviceAdapterWithAnotherASN := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {

			return []domain.Interface{domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
				IntType: "ethernet", IntName: "1/11", Mac: "M1", ConfigState: "up"}}, nil

		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {

			return []domain.LLDP{domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
				LocalIntType: "ethernet", LocalIntName: "1/11", LocalIntMac: "M1",
				RemoteIntType: "ethernet", RemoteIntName: "1/22", RemoteIntMac: "M2"}}, nil

		},
		MockGetASN: func(FabricID uint, Device uint, DeviceIP string) (string, error) {
			return ASNOnDevice2, nil
		},
	}
	devUC.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockLeafDeviceAdapterWithAnotherASN)

	resp, err = devUC.AddDevices(context.Background(), MockFabricName, []string{MockLeaf1IP}, []string{},
		UserName, Password, false)
	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: MockFabricName, FabricID: 1, IPAddress: MockLeaf1IP, Role: usecase.LeafRole})

	assert.Nil(t, err)
	switchConfigSecond, err := DatabaseRepository.GetSwitchConfigOnDeviceIP(MockFabricName, MockLeaf1IP)
	assert.Equal(t, ASNOnDevice2, switchConfigSecond.LocalAS)
	//Since switch already has the ASN no need to push to switch
	assert.Equal(t, domain.ConfigNone, switchConfigSecond.ASConfigType)
}

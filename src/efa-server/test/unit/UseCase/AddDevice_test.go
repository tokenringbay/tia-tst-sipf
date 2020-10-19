package usecase

import (
	"context"
	"efa-server/domain"
	ad "efa-server/infra/device/adapter"
	"efa-server/test/unit/mock"
	"efa-server/usecase"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var UserName = "admin"
var Password = "password"
var IPAddress = "test"

func TestDeviceAdd_FabricDoesnotExist(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			err := errors.New("Fabric deoes not exist")
			return domain.Fabric{}, err
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{IPAddress: IPAddress, FabricID: 0,
		Errors: []error{errors.New("Fabric default does not exist")}})

	assert.NotNil(t, err)
}

func TestDeviceAdd_SwitchAlreadyExist(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetFabric: func(FabricName string) (domain.Fabric, error) {
			return domain.Fabric{Name: FabricName, ID: 1}, nil
		},
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{IPAddress: IPAddress, UserName: UserName}, nil
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, FabricID: 1, IPAddress: IPAddress, Role: usecase.SpineRole})

	//assert.EqualError(t, err, "Switch test already exist", "")
	assert.Nil(t, err, "")
}

func TestDeviceAdd_OpenTransactionFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, errors.New("Switch does not exist")
		},
		MockOpenTransaction: func() error {
			return errors.New("Failed to Open Transaction")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	fmt.Println(resp)
	assert.NotNil(t, err)

}

func TestDeviceAdd_DeviceCreateFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, errors.New("Switch does not exist")
		},
		MockCreateDevice: func(Device *domain.Device) error {
			return errors.New("Device Save Failed")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactory}

	_, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Equal(t, err, errors.New("Switch test create Failed with error : Device Save Failed"))

}
func TestDeviceAdd_DeviceConnectionFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, errors.New("Switch does not exist")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.DeviceAdapterFactoryFailed}
	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, FabricID: 1, IPAddress: IPAddress, Role: usecase.SpineRole,
		Errors: []error{errors.New("Switch test connection Failed : Failed to Open Connection")}})
	assert.NotNil(t, err)

}

func TestDeviceAdd_InterfaceFetchFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, errors.New("Switch does not exist")
		},
	}
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			var Interfaces []domain.Interface
			return Interfaces, errors.New("Cannot fetch Interfaces")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter)}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, FabricID: 1, IPAddress: IPAddress, Role: usecase.SpineRole,
		Errors: []error{errors.New("Interface fetch for switch test Failed")}})
	assert.NotNil(t, err)

}

func TestDeviceAdd_InterfacePersistFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, nil
		},
		MockCreateInterface: func(Interface *domain.Interface) error {
			return errors.New("Cannot Persist Interface")
		},
	}
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			var Interfaces = []domain.Interface{
				domain.Interface{
					IntType: "ethernet",
					IntName: "1/11",
				},
			}
			return Interfaces, nil
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter)}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, FabricID: 1, IPAddress: IPAddress, Role: usecase.SpineRole,
		Errors: []error{errors.New("Failed to create Physical Interface ethernet 1/11")}})
	assert.NotNil(t, err)

}

func TestDeviceAdd_FetchLLDPFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, nil
		},
	}
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			var Interfaces = []domain.Interface{
				domain.Interface{
					IntType: "ethernet",
					IntName: "1/11",
				},
			}
			return Interfaces, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {
			var LLDPS []domain.LLDP
			return LLDPS, errors.New("Failed to fetch LLDPS")
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter)}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, FabricID: 1, IPAddress: IPAddress, Role: usecase.SpineRole,
		Errors: []error{errors.New("Failed to fetch LLDP for test")}})
	assert.NotNil(t, err)

}

func TestDeviceAdd_PersistLLDPFailed(t *testing.T) {
	MockDatabaseRepository := mock.DatabaseRepository{
		MockGetDevice: func(FabricName string, IPAddress string) (domain.Device, error) {
			return domain.Device{}, nil
		},
		MockCreateLLDP: func(Interface *domain.LLDP) error {
			return errors.New("Failed to Persist LLDP Data")
		},
	}
	MockDeviceAdapter := mock.DeviceAdapter{
		MockGetInterfaces: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error) {
			var Interfaces = []domain.Interface{
				domain.Interface{
					IntType: "ethernet",
					IntName: "1/11",
				},
			}
			return Interfaces, nil
		},
		MockGetLLDPs: func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error) {
			var LLDPS = []domain.LLDP{
				domain.LLDP{
					LocalIntType: "ethernet",
					LocalIntName: "1/11",
				},
			}
			return LLDPS, nil
		},
	}

	devUC := usecase.DeviceInteractor{Db: &MockDatabaseRepository, DeviceAdapterFactory: mock.GetDeviceAdapterFactory(MockDeviceAdapter)}

	resp, err := devUC.AddDevices(context.Background(), FabricName, []string{}, []string{IPAddress},
		UserName, Password, false)

	assert.Contains(t, resp, usecase.AddDeviceResponse{FabricName: FabricName, FabricID: 1, IPAddress: IPAddress, Role: usecase.SpineRole,
		Errors: []error{errors.New("Failed to create LLDP for ethernet 1/11")}})
	assert.NotNil(t, err)

}

func TestDeviceSupported_FirmwareVersion(t *testing.T) {
	model := fmt.Sprint(ad.AvalancheType, "_", "18r.1.01")
	err := ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.1.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18r.2.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)
	model = fmt.Sprint(ad.AvalancheType, "_", "18.2.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FusionType, "_", "18r.1.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FusionType, "_", "18.1.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FusionType, "_", "18r.2.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FusionType, "_", "18.2.01")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.OrcaType, "_", "18x.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.OrcaType, "_", "18.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.OrcaTType, "_", "18x.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.OrcaTType, "_", "18.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FreedomType, "_", "17s.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FreedomType, "_", "17.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.CedarType, "_", "17s.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.CedarType, "_", "17.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FreedomType, "_", "18s.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.FreedomType, "_", "18.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.CedarType, "_", "18s.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.CedarType, "_", "18.1.00")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18r.1.01aa")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.1.01aa")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18r.2.01aa")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.2.01aa")
	err = ad.CheckSupportedVersion(model)
	assert.NoError(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.1.01slxos_123")
	err = ad.CheckSupportedVersion(model)
	fmt.Println(err)
	assert.NoError(t, err)

	//Error Scenarios follow

	model = fmt.Sprint(ad.AvalancheType, "_", "17r.2.01aa")
	err = ad.CheckSupportedVersion(model)
	assert.Error(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "17.2.01aa")
	err = ad.CheckSupportedVersion(model)
	assert.Error(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "")
	err = ad.CheckSupportedVersion(model)
	assert.Error(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18")
	err = ad.CheckSupportedVersion(model)
	assert.Error(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.1")
	err = ad.CheckSupportedVersion(model)
	assert.Error(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.01.01_cr")
	err = ad.CheckSupportedVersion(model)
	fmt.Println(err)
	assert.Error(t, err)

	model = fmt.Sprint(ad.AvalancheType, "_", "18.01.01_123213")
	err = ad.CheckSupportedVersion(model)
	fmt.Println(err)
	assert.Error(t, err)

}

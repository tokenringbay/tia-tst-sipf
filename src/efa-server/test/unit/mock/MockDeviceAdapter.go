package mock

import (
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	Interactor "efa-server/usecase/interactorinterface"
	"errors"
	"fmt"
)

//DeviceAdapter represents a mock DeviceAdapter
type DeviceAdapter struct {
	IPAddress                  string
	MockGetInterfaces          func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.Interface, error)
	MockGetLLDPs               func(FabricID uint, DeviceID uint, DeviceIP string) ([]domain.LLDP, error)
	MockConfigureFabric        func(config operation.ConfigFabricRequest) error
	MockGetASN                 func(FabricID uint, Device uint, DeviceIP string) (string, error)
	MockEnableInterfaces       func(InterfaceNames []string) (string, error)
	MockGetInterfaceSpeed      func(InterfaceType string, InterfaceName string) (int, error)
	MockGetInterfaceVe         func(name string) (map[string]string, error)
	MockGetClusterByName       func(name string) (map[string]string, error)
	MockGetInterfacePoMember   func(name string) (map[string]string, error)
	MockGetDeviceDetail        func(FabricID uint, DeviceID uint, DeviceIP string) (domain.DeviceDetail, error)
	MockGetSwitchHostName      func(DeviceIP string) (string, error)
	MockCheckSupportedFirmware func(DeviceIP string) error
}

//DeviceAdapterFactory represents a mock DeviceAdapterFactory returning a mock DeviceAdapter
func DeviceAdapterFactory(ctx context.Context, IPAddress string, UserName string, Password string) (Interactor.DeviceAdapter, error) {
	return &DeviceAdapter{}, nil
}

//GetDeviceAdapterFactory represents a mock DeviceAdapterFactory
// returning a mock DeviceAdapter (whose properties are input as another DeviceAdapter)
func GetDeviceAdapterFactory(deviceAdapter DeviceAdapter) func(ctx context.Context, IPAddress string, UserName string, Password string) (Interactor.DeviceAdapter, error) {
	return func(ctx context.Context, IPAddress string, UserName string, Password string) (Interactor.DeviceAdapter, error) {

		//clone of DeviceAdapter
		CloneDeviceAdapter := DeviceAdapter{IPAddress: IPAddress, MockGetInterfaces: deviceAdapter.MockGetInterfaces,
			MockGetLLDPs: deviceAdapter.MockGetLLDPs, MockConfigureFabric: deviceAdapter.MockConfigureFabric,
			MockGetASN: deviceAdapter.MockGetASN, MockEnableInterfaces: deviceAdapter.MockEnableInterfaces,
			MockGetInterfaceSpeed: deviceAdapter.MockGetInterfaceSpeed, MockGetInterfaceVe: deviceAdapter.MockGetInterfaceVe,
			MockGetClusterByName: deviceAdapter.MockGetClusterByName, MockGetInterfacePoMember: deviceAdapter.MockGetInterfacePoMember,
		}
		return &CloneDeviceAdapter, nil
	}
}

//DeviceAdapterFactoryFailed returns a DeviceAdapter with a connection error
func DeviceAdapterFactoryFailed(ctx context.Context, IPAddress string, UserName string, Password string) (Interactor.DeviceAdapter, error) {
	return &DeviceAdapter{}, errors.New("Failed to Open Connection")
}

//CloseConnection represents a mock CloseConnection
func (ad *DeviceAdapter) CloseConnection(ctx context.Context) error {
	fmt.Println("Mock Close Connection Called")
	return nil
}

//EnableInterfaces represents a mock EnableInterfaces
func (ad *DeviceAdapter) EnableInterfaces(InterfaceNames []string) (string, error) {
	if ad.MockEnableInterfaces != nil {
		return ad.MockEnableInterfaces(InterfaceNames)
	}
	return "", nil
}

//GetInterfaces represents a mock GetInterfaces
func (ad *DeviceAdapter) GetInterfaces(FabricID uint, DeviceID uint, ControlVlan string) ([]domain.Interface, error) {
	if ad.MockGetInterfaces != nil {
		return ad.MockGetInterfaces(FabricID, DeviceID, ad.IPAddress)
	}
	var Interfaces []domain.Interface
	return Interfaces, nil
}

//GetDeviceDetail gets device details about the device
func (ad *DeviceAdapter) GetDeviceDetail(FabricID uint, DeviceID uint, DeviceIP string) (
	domain.DeviceDetail, error) {
	if ad.MockGetDeviceDetail != nil {
		return ad.MockGetDeviceDetail(FabricID, DeviceID, DeviceIP)
	}

	return domain.DeviceDetail{}, nil
}

//GetLLDPs represents a mock GetLLDPs
func (ad *DeviceAdapter) GetLLDPs(FabricID uint, DeviceID uint) ([]domain.LLDP, error) {
	if ad.MockGetLLDPs != nil {
		return ad.MockGetLLDPs(FabricID, DeviceID, ad.IPAddress)
	}
	var LLDPS []domain.LLDP
	return LLDPS, nil
}

//GetASN represents a mock GetASN
func (ad *DeviceAdapter) GetASN(FabricID uint, DeviceID uint) (string, error) {
	if ad.MockGetASN != nil {
		return ad.MockGetASN(FabricID, DeviceID, ad.IPAddress)
	}
	return "", nil
}

//ConfigureFabric represents a ConfigureFabric
func (ad *DeviceAdapter) ConfigureFabric(config operation.ConfigFabricRequest) error {
	if ad.MockConfigureFabric != nil {
		return ad.MockConfigureFabric(config)
	}
	return nil
}

//GetInterfaceSpeed represents a GetInterfaceSpeed
func (ad *DeviceAdapter) GetInterfaceSpeed(InterfaceType string, InterfaceName string) (int, error) {
	if ad.MockGetInterfaceSpeed != nil {
		return ad.MockGetInterfaceSpeed(InterfaceType, InterfaceName)
	}
	return 0, nil
}

//GetInterfaceVe represents a mock GetInterfaceVe
func (ad *DeviceAdapter) GetInterfaceVe(name string) (map[string]string, error) {
	var m map[string]string
	if ad.MockGetInterfaceVe != nil {
		return ad.MockGetInterfaceVe(name)
	}
	return m, nil
}

//GetClusterByName represents a mock GetClusterByName
func (ad *DeviceAdapter) GetClusterByName(name string) (map[string]string, error) {
	var m map[string]string
	if ad.MockGetClusterByName != nil {
		return ad.MockGetClusterByName(name)
	}
	return m, nil
}

//GetInterfacePoMember represents a mock GetInterfacePoMember
func (ad *DeviceAdapter) GetInterfacePoMember(name string) (map[string]string, error) {
	var m map[string]string
	if ad.MockGetInterfacePoMember != nil {
		return ad.MockGetInterfacePoMember(name)
	}
	return m, nil
}

//GetSwitchHostName gets HostName of the device
func (ad *DeviceAdapter) GetSwitchHostName(DeviceIP string) (
	string, error) {
	hostName := ""
	if ad.MockGetSwitchHostName != nil {
		return ad.MockGetSwitchHostName(DeviceIP)
	}

	return hostName, nil
}

//CheckSupportedFirmware checks for model and firmware version
func (ad *DeviceAdapter) CheckSupportedFirmware(DeviceIP string) error {
	if ad.MockCheckSupportedFirmware != nil {
		return ad.MockCheckSupportedFirmware(DeviceIP)
	}
	return nil
}

//Mock Functions

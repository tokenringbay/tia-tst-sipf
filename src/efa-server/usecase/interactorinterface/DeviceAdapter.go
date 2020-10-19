package interactorinterface

import (
	"context"
	"efa-server/domain"
)

//DeviceAdapter provides methods to interact with Device
type DeviceAdapter interface {
	//CloseConnection closess the Connection to the Device
	CloseConnection(ctx context.Context) error

	//GetInterfaces fetches the Interfaces from the Device
	GetInterfaces(FabricID uint, DeviceID uint, ControlVlan string) ([]domain.Interface, error)

	//EnableInterfaces enables the interfaces on the Device
	EnableInterfaces(InterfaceNames []string) (string, error)

	//Get Device Details about the device
	GetDeviceDetail(FabricID uint, DeviceID uint, DeviceIP string) (domain.DeviceDetail, error)

	//GetLLDPs fetches the LLDP data from the Device
	GetLLDPs(FabricID uint, DeviceID uint) ([]domain.LLDP, error)

	//GetASN fetches the ASN for the Device
	GetASN(FabricID uint, Device uint) (string, error)

	//GetInterfaceSpeed  fetches the Interface speed from the Device
	GetInterfaceSpeed(InterfaceType string, InterfaceName string) (int, error)

	//GetInterfaceVe fetches the VE interface from the switch
	GetInterfaceVe(name string) (map[string]string, error)

	// GetClusterByName fetches the MCT cluster from the switch
	GetClusterByName(name string) (map[string]string, error)

	//GetInterfacePoMember fetches the PO interface member belongs
	GetInterfacePoMember(name string) (map[string]string, error)

	//Get HostName of the device
	GetSwitchHostName(DeviceIP string) (string, error)

	//CheckSupportedFirmware checks for model and firmware version
	CheckSupportedFirmware(DeviceIP string) error
}

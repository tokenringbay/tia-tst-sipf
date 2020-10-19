package interfaces

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
	"efa-server/infra/device/models"
)

//Interface provides collection of methods supported for Interface
type Interface interface {
	GetInterfaces(client *client.NetconfClient, ControlVlan string) ([]models.InterfaceSwitchResponse, error)
	GetInterfaceSpeed(client *client.NetconfClient, interfaceType string, interfaceName string) (int, error)
	PersistConfig(client *client.NetconfClient) (map[string]string, error)
	ConfigureInterfaceLoopback(client *client.NetconfClient, loopbackID string, ipAddress string) (string, error)
	DeleteInterfaceLoopback(client *client.NetconfClient, loopbackID string) (string, error)
	ConfigureInterfaceNumbered(client *client.NetconfClient, intType string, intName string,
		ipAddress string, description string) (string, error)
	UnconfigureInterfaceNumbered(client *client.NetconfClient, intType string, intName string,
		ipAddress string) (string, error)
	UnconfigureInterfaceDesc(client *client.NetconfClient, intType string, intName string) (string, error)
	UnconfigureInterfaceSpeed(client *client.NetconfClient, intType string, intName string) (string, error)
	ConfigureInterfaceUnnumbered(client *client.NetconfClient, intType string, intName string, donorType string,
		donorName string) (string, error)
	UnconfigureInterfaceUnnumbered(client *client.NetconfClient, intType string, intName string) (string, error)
	GetInterface(client *client.NetconfClient, InterfaceType string, InterfaceName string) (map[string]string, error)
	GetEthernetInterfaceConfigs(client *client.NetconfClient, intfNames []string) ([]operation.ConfigIntfResponse, error)
	GetLoopbackInterfaceConfigs(client *client.NetconfClient, loopbackIds []string) ([]operation.ConfigIntfResponse, error)
	GetVEInterfaceConfigs(client *client.NetconfClient, intfNames []string) ([]operation.ConfigIntfResponse, error)
	GetInterfaceConfigs(client *client.NetconfClient, intfs map[string][]string) ([]operation.ConfigIntfResponse, error)
	ConfigureInterfaceVe(client *client.NetconfClient, name string,
		ipAddress string, bfdrx string, bfdtx string, bfdmultiplier string) (string, error)
	UnconfigureInterfaceVeIP(client *client.NetconfClient, name string,
		ipAddress string) (string, error)
	ConfigureInterfaceVeIP(client *client.NetconfClient, name string,
		ipAddress string) (string, error)
	GetInterfaceVe(client *client.NetconfClient, name string) (map[string]string, error)

	DeleteInterfaceVe(client *client.NetconfClient, name string) (string, error)
	CreateInterfacePo(client *client.NetconfClient, name string,
		speed string, description string, controlVlan string) (string, error)
	CreateBasicInterfacePo(client *client.NetconfClient, name string,
		speed string, description string) (string, error)
	ConfigureInterfacePoSpeed(client *client.NetconfClient, name string, speed string) (string, error)
	DeleteInterfacePo(client *client.NetconfClient, name string) (string, error)
	GetInterfacePo(client *client.NetconfClient, name string) (map[string]string, error)
	GetInterfacePoDetails(client *client.NetconfClient, name string) (map[string]string, error)
	AddInterfaceToPo(client *client.NetconfClient, name string, portChannelDescription string,
		portChannel string, portChannelMode string, portChannelType string, speed string) (string, error)
	GetInterfacePoMember(client *client.NetconfClient, name string) (map[string]string, error)
	DeleteInterfaceFromPo(client *client.NetconfClient, name string, portChannel string) (string, error)
	EnableInterfaces(client *client.NetconfClient, interfaceNames []string) (string, error)
	DisableInterface(client *client.NetconfClient, ifName string) (string, error)
	DisableInterfaces(client *client.NetconfClient, interfaceNames []string) (string, error)
}

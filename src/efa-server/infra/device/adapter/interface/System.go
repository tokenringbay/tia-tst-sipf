package interfaces

import (
	"efa-server/infra/device/client"
)

//System provides collection of methods supported for System
type System interface {
	//ConfigureSystemL2Mtu is used to configure System L2 MTU on the switching device
	ConfigureSystemL2Mtu(client *client.NetconfClient, l2MTUValue string) (string, error)

	//UnconfigureSystemL2Mtu is used to unconfigure System L2 MTU on the switching device
	UnconfigureSystemL2Mtu(client *client.NetconfClient) (string, error)

	//ConfigureSystemIPMtu is used to configure System IP MTU on the switching device
	ConfigureSystemIPMtu(client *client.NetconfClient, ipMTUValue string) (string, error)

	//UnconfigureSystemIPMtu is used to unconfugure System IP MTU from the switching device
	UnconfigureSystemIPMtu(client *client.NetconfClient) (string, error)

	//ConfigureMacAndArp is used to configure MAC and ARP parameters on the switching device
	ConfigureMacAndArp(client *client.NetconfClient, arpAgingTimeout string,
		macAgingTimeout string, macAgingConversationalTimeout string, macMoveLimit string) (string, error)

	//UnconfigureMacAndArp is used to unconfigure MAC and ARP parameters from the switching device
	UnconfigureMacAndArp(client *client.NetconfClient) (string, error)

	//GetArp is used to get the running-config of the ARP parameters from the switching device
	GetArp(client *client.NetconfClient) (map[string]string, error)

	//GetMac is used to get the running-config of the MAC parameters from the switching device
	GetMac(client *client.NetconfClient) (map[string]string, error)

	//ConfigureAnycastGateway is used to configure static anycast-gateway on the switching device
	ConfigureAnycastGateway(client *client.NetconfClient, ipv4AnycastGatewayMac string,
		ipv6AnycastGatewayMac string) (string, error)

	//GetAnycastGateway is used to get the running-config of AnyCast Gateway from the device
	GetAnycastGateway(client *client.NetconfClient) (map[string]string, error)

	//UnconfigureAnycastGateway is used to unconfigure static anycast-gateway from the switching device
	UnconfigureAnycastGateway(client *client.NetconfClient) (string, error)

	//ConfigureRouterID is used to configure "router-id" config on the switching device
	ConfigureRouterID(client *client.NetconfClient, routerID string) (string, error)

	//UnconfigureRouterID is used to uncofigure "router-id" from the switching device
	UnconfigureRouterID(client *client.NetconfClient) (string, error)

	//GetRouterID is used to get the running-config of "router-id" from the switching device
	GetRouterID(client *client.NetconfClient) (string, error)

	//GetDevice detail provides detail of the device like SWBD, Version
	GetDeviceDetail(client *client.NetconfClient) (map[string]string, error)

	//GetIPRoutes is used to get the running-config of "overlay-gateway" from the switching device
	GetIPRoutes(client *client.NetconfClient) (map[string]string, error)

	//ConfigureNumberedRouteMap Route Map for Numbered Interfaces
	ConfigureNumberedRouteMap(client *client.NetconfClient) (string, error)

	//UnConfigureNumberedRouteMap Route Map for Numbered Interfaces
	UnConfigureNumberedRouteMap(client *client.NetconfClient) (string, error)

	//ConfigureSwitchHostName is used to configure Host name on the switching device
	ConfigureSwitchHostName(client *client.NetconfClient, hostName string) (string, error)

	//UnconfigureSwitchHostName is used to unconfigure System Host name on the switching device
	UnconfigureSwitchHostName(client *client.NetconfClient) (string, error)

	//GetSwitchHostName is used to get System Host name on the switching device
	GetSwitchHostName(client *client.NetconfClient) (string, error)
}

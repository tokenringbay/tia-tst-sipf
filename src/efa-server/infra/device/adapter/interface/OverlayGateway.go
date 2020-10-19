package interfaces

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
)

//OverlayGateway provides collection of methods supported for OverlayGateway
type OverlayGateway interface {
	//CreateOverlayGateway is used to create "overlay-gateway" and its subconfig on the switching device
	CreateOverlayGateway(client *client.NetconfClient, gwName string, gwType string,
		loopbackID string, mapVNIAuto string) (string, error)

	//GetOverlayGateway is used to get the running-config of "overlay-gateway" from the switching device
	GetOverlayGateway(client *client.NetconfClient) (operation.ConfigOVGResponse, error)

	//DeleteOverlayGateway is used to delete the "overlay-gateway" from the switching device
	DeleteOverlayGateway(client *client.NetconfClient, gwName string) (string, error)

	//CreateEvpnInstance is used to create "evpn" instance on the switching device
	CreateEvpnInstance(client *client.NetconfClient, eviName string, duplicateMacTimer string,
		duplicateMacTimerMaxCount string) (string, error)

	//DeleteEvpnInstance is used to delete "evpn" instance from the switching device
	DeleteEvpnInstance(client *client.NetconfClient, eviName string) (string, error)

	//GetEvpnInstance is used to get the running-config of "evpn" instance from the switching device
	GetEvpnInstance(client *client.NetconfClient) (operation.ConfigEVPNRespone, error)
}

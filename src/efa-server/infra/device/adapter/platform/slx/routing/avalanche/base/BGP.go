package base

import (
	"efa-server/infra/device/adapter/platform/slx"
	"efa-server/infra/device/adapter/platform/slx/routing/base"
	"efa-server/infra/device/client"
)

//SLXAvalancheBase structure for cedar base SLX
type SLXAvalancheBase struct {
	base.SLXRoutingBase
}

//GetRouterBgpL2EVPNNeighborEncapType is used to get the EncapType that need to be configured.
func (base *SLXAvalancheBase) GetRouterBgpL2EVPNNeighborEncapType(client *client.NetconfClient) (string, error) {
	return slx.BGPEncapTypeForRoutingAvalanche, nil
}

//ConfigureRouterBgpL2EVPNNeighbor is used to configure the MCT BGP neighbour and enable BFD on the neighbour, on the switching device
func (base *SLXAvalancheBase) ConfigureRouterBgpL2EVPNNeighbor(client *client.NetconfClient, neighborAddress string,
	neighborLoopBackAddress string, loopbackNumber string,
	remoteAs string, encapType string, bfdEnabled string) (string, error) {
	var bgpMap = map[string]interface{}{"neighborAddress": neighborLoopBackAddress, "loopbackNumber": loopbackNumber, "remoteAs": remoteAs, "encapType": encapType,
		"bfdEnabled": bfdEnabled, "peer_group_name": ""}
	config, templateError := base.GetStringFromTemplate(routerBgpMctNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(routerBgpMctL2EvpnNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(routerBgpNeighborDeactivateInIpv4UnicastAF, bgpMap)

	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	return resp, err
}

//UnconfigureRouterBgpL2EVPNNeighbor is used to unconfigure the MCT BGP neighbour from the switching device
func (base *SLXAvalancheBase) UnconfigureRouterBgpL2EVPNNeighbor(client *client.NetconfClient, neighborAddress string, neighborLoopBackAddress string) (string, error) {
	var bgpMap = map[string]interface{}{"neighbor_address": neighborLoopBackAddress}
	config, templateError := base.GetStringFromTemplate(routerBgpNeighborDelete, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureIPRoute is used to configure the static route
func (base *SLXAvalancheBase) ConfigureIPRoute(client *client.NetconfClient, loopbackIP string,
	veIP string) (string, error) {
	var bgpMap = map[string]interface{}{"loopbackIP": loopbackIP, "VEIP": veIP}
	config, templateError := base.GetStringFromTemplate(configureIPRoute, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	return resp, err
}

//DeConfigureIPRoute is used to configure the static route
func (base *SLXAvalancheBase) DeConfigureIPRoute(client *client.NetconfClient, loopbackIP string,
	veIP string) (string, error) {
	var bgpMap = map[string]interface{}{"loopbackIP": loopbackIP, "VEIP": veIP}
	config, templateError := base.GetStringFromTemplate(deconfigureIPRoute, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	return resp, err
}

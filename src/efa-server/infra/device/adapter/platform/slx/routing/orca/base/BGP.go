package base

import (
	"efa-server/infra/device/adapter/platform/slx"
	"efa-server/infra/device/client"
)

//GetRouterBgpL2EVPNNeighborEncapType is used to get the EncapType that need to be configured.
func (base *SLXOrcaBase) GetRouterBgpL2EVPNNeighborEncapType(client *client.NetconfClient) (string, error) {
	return slx.BGPEncapTypeForRoutingOrca, nil
}

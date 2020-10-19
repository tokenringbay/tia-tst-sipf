package interfaces

import (
	"efa-server/infra/device/client"
	"efa-server/infra/device/models"
)

//LLDP provides collection of methods supported for LLDP
type LLDP interface {
	//GetLLDPNeighbors is used to get the operational state of the LLDP neighbours, from the switching device
	GetLLDPNeighbors(client *client.NetconfClient) ([]models.InterfaceLLDPResponse, error)
}

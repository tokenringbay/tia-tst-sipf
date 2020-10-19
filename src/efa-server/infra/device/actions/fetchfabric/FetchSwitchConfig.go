package fetchfabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	"efa-server/usecase"
	"sync"
)

//FetchSwitchConfig fetches config from the switch.
func FetchSwitchConfig(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.SwitchIdentity,
	switchResponses chan operation.ConfigSwitchResponse, errs chan actions.OperationError) {
	defer fabricGate.Done()
	adapter := ad.GetAdapter(sw.Model)
	netconfClient := &client.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := netconfClient.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Interface Login", Error: err, Host: sw.Host}
		return
	}
	defer netconfClient.Close()

	switchResponse := operation.ConfigSwitchResponse{}
	switchResponse.Host = sw.Host
	switchResponse.Role = sw.Role

	switchResponse.Bgp, _ = adapter.GetRouterBgp(netconfClient)

	switchResponse.RouterID, _ = adapter.GetRouterID(netconfClient)
	if sw.Role == usecase.LeafRole || sw.Role == usecase.RackRole {
		ovg, _ := adapter.GetOverlayGateway(netconfClient)
		switchResponse.Ovg = &ovg
		Evpn, _ := adapter.GetEvpnInstance(netconfClient)
		switchResponse.Evpn = &Evpn
		//TODO Get ANyCast Gateway

		// Fetch the Cluster details and update the switch response for the display
		clusterName, clusterID, _, _, clusterPeerIntfName, clusterPeerIP, _ := adapter.GetCluster(netconfClient)

		switchResponse.ClusterDetails.ID = clusterID
		switchResponse.ClusterDetails.Name = clusterName
		//switchResponse.ClusterDetails.Vlan = clusterVlanID
		switchResponse.ClusterDetails.PeerIP = clusterPeerIP
		switchResponse.ClusterDetails.PortChannel.ID = clusterPeerIntfName

		resp, _ := adapter.GetInterfacePo(netconfClient, clusterPeerIntfName)
		switchResponse.ClusterDetails.PortChannel.Speed = resp["speed"]
		switchResponse.ClusterDetails.PortChannel.Description = resp["description"]

		if resp["shutdown"] == "" {
			switchResponse.ClusterDetails.PortChannel.Shutdown = "False"
		} else {
			switchResponse.ClusterDetails.PortChannel.Shutdown = "True"
		}

		respPo, _ := adapter.GetInterfacePoDetails(netconfClient, clusterPeerIntfName)
		switchResponse.ClusterDetails.PortChannel.AggregatorMode = respPo["aggregator-mode"]
		switchResponse.ClusterDetails.PortChannel.AggregatorType = respPo["aggregator-type"]
		switchResponse.ClusterDetails.PortChannel.MemberPorts = respPo["intfs"]

	}

	switchResponse.Interfaces, _ = adapter.GetInterfaceConfigs(netconfClient, sw.Interfaces)

	switchResponses <- switchResponse
}

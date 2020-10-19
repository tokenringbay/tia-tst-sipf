package interfaces

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
)

//BGP provides collection of methods supported for BGP
type BGP interface {
	//ConfigureRouterBgp is used to configure "router bgp" on the switching device
	ConfigureRouterBgp(client *client.NetconfClient, localAs string, peerGroupName string,
		description string, networkAddress string, maxPaths string,
		evpn string, allowasIn string, retainRouteTargetAll string, nextHopUnchanged string, bfdEnable string,
		bfdMinTx string, bfdMinRx string, bfdMultiplier string, detrisibuteConnected string, detrisibuteConnectedWithRouteMap string) (string, error)

	//UnconfigureRouterBgp is used to unconfigure "router bgp" from the switching device
	UnconfigureRouterBgp(client *client.NetconfClient) (string, error)

	//ConfigureRouterBgpNeighbor is used to configure "router bgp neighbour" and "associate the neighbour with a peer-group
	// on the switching device
	ConfigureRouterBgpNeighbor(client *client.NetconfClient, remoteAs string, peerGroupName string,
		neighborAddress string, bgpMultihop string, unnumberedInterface bool, isLeaf string, nextHopLeaf bool) (string, error)

	//UnconfigureRouterBgpNeighbor is used to unconfigure "router bgp neighbour" from the switching device
	UnconfigureRouterBgpNeighbor(client *client.NetconfClient, remoteAs string, peerGroupName string,
		neighborAddress string) (string, error)

	//IsRouterBgpPresent is used to check whether "router bgp" config is present on a switching device
	IsRouterBgpPresent(client *client.NetconfClient) error

	//GetRouterBgp is used to get the running-config of "router bgp"
	GetRouterBgp(client *client.NetconfClient) (operation.ConfigBGPResponse, error)

	//GetLocalAsn is used to get the "router bgp" Local ASN configuration from the switching device
	GetLocalAsn(client *client.NetconfClient) (string, error)

	//GetRouterBgpL2EVPNNeighborEncapType is used to get the right Encapsulation type for different type of devices.
	GetRouterBgpL2EVPNNeighborEncapType(client *client.NetconfClient) (string, error)

	//ConfigureRouterBgpL2EVPNNeighbor is used to configure the MCT BGP neighbour and enable BFD on the neighbour, on the switching device
	ConfigureRouterBgpL2EVPNNeighbor(client *client.NetconfClient, neighborAddress string,
		neighborLoopBackAddress string, loopbackNumber string,
		remoteAs string, encapType string, bfdEnabled string) (string, error)

	//UnconfigureRouterBgpL2EVPNNeighbor is used to unconfigure the MCT BGP neighbour from the switching device
	UnconfigureRouterBgpL2EVPNNeighbor(client *client.NetconfClient, neighborAddress string, neighborLoopBackAddress string) (string, error)

	//ExecuteClearBgpEvpnNeighbourAll is used to execute "clear bgp evpn neighbor all" on the switching device
	ExecuteClearBgpEvpnNeighbourAll(sshClient *client.SSHClient) error

	//ExecuteClearBgpEvpnNeighbour is used to execute "clear bgp evpn neighbor <neighbour-ip>" on the switching device
	ExecuteClearBgpEvpnNeighbour(sshClient *client.SSHClient, neighbourIP string) error

	//ConfigureIPRoute is used to configure the static route
	ConfigureIPRoute(client *client.NetconfClient, loopbackIP string, veIP string) (string, error)

	//DeConfigureIPRoute is used to configure the static route
	DeConfigureIPRoute(client *client.NetconfClient, loopbackIP string, veIP string) (string, error)

	//ConfigureNonClosRouterBgp is used to configure "router bgp" on the switching device for Non-clos use-case
	ConfigureNonClosRouterBgp(client *client.NetconfClient, localAs string, Network []string,
		bfdMinTx string, bfdMinRx string, bfdMultiplier string,
		eBGPPeerGroup string, eBGPPeerGroupDescription string, eBGPBFD bool,
		EVPNPeerGroup string, EVPNPeerGroupDescription string, EVPNPeerGroupMultiHop string,
		MaxPaths string, encap string, nextHopUnChanged bool,
		retainRTAll bool) (string, error)

	//ConfigureNonClosRouterEvpnNeighbor is used to configure Evpn Neighbors on the switching device for Non-clos use-case
	ConfigureNonClosRouterEvpnNeighbor(client *client.NetconfClient,
		remoteAs string, peerGroup string, peerGroupDescription string, neighborAddress string,
		loopbackNumber string, multiHop string) (string, error)

	//ConfigureNonClosRouterBgpNeighbor is used to configure Evpn Neighbors on the switching device for Non-clos use-case
	ConfigureNonClosRouterBgpNeighbor(client *client.NetconfClient,
		remoteAs string, peerGroup string, peerGroupDesc string, neighborAddress string,
		isBfdEnabled bool, isNextHopSelf bool) (string, error)
}

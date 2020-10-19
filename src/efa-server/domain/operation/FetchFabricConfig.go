package operation

import (
	"fmt"
	"github.com/ghodss/yaml"
)

//FabricFetchRequest is a request Object for Fetching details of Fabric
type FabricFetchRequest struct {
	FabricName string           `json:"fabric_name"`
	Hosts      []SwitchIdentity `json:"hosts"`
}

//SwitchIdentity represents all the attributes to identify a switch
type SwitchIdentity struct {
	Host       string              `json:"switch"`
	Role       string              `json:"role"`
	UserName   string              `json:"user"`
	Password   string              `json:"password"`
	Model      string              `json:"model"`
	Interfaces map[string][]string `json:"interfaces"` // Map of interface type to interface names.
}

//ConfigIntfRequest is a request object representing the interface
type ConfigIntfRequest struct {
	Type string `json:"interface_type"`
	Name string `json:"interface_name"`
}

//FabricFetchResponse is a response object representing the fabric config
type FabricFetchResponse struct {
	FabricName     string                 `json:"fabric_name"`
	SwitchResponse []ConfigSwitchResponse `json:"switches"`
}

//ConfigSwitchResponse is a response object representing the switch config
type ConfigSwitchResponse struct {
	Host           string                `json:"switch"`
	Role           string                `json:"role"`
	Bgp            ConfigBGPResponse     `json:"bgp"`
	RouterID       string                `json:"router_id"`
	Ovg            *ConfigOVGResponse    `json:"overlay_gateway"`
	Evpn           *ConfigEVPNRespone    `json:"evpn"`
	Interfaces     []ConfigIntfResponse  `json:"interfaces"`
	ClusterDetails ConfigClusterResponse `json:"cluster_details"`
}

//ConfigClusterResponse is a response object representing the switch config
type ConfigClusterResponse struct {
	Name string `json:"cluster_name"`
	ID   string `json:"cluster_id"`
	//Vlan      string           `json:"cluster_vlan"`
	PeerIP      string           `json:"peer_ip"`
	PortChannel ConfigPoResponse `json:"port_channel"`
}

//ConfigPoResponse is a response object representing the switch config
type ConfigPoResponse struct {
	ID             string `json:"port_channel_id"`
	Description    string `json:"port_channel_description"`
	Speed          string `json:"speed"`
	Shutdown       string `json:"shutdown"`
	MemberPorts    string `json:"member_ports"`
	AggregatorMode string `json:"port_channel_mode"`
	AggregatorType string `json:"port_channel_type"`
}

//ConfigIntfResponse is a response object representing the interfaces
type ConfigIntfResponse struct {
	Name        string `json:"interface_name"`
	Type        string `json:"interface_type"`
	IPAddress   string `json:"interface_ip_address"`
	Description string `json:"interface_description"`
}

//ConfigEVPNRespone is a response object representing the "evpn" config
type ConfigEVPNRespone struct {
	Name                   string `json:"name"`
	DuplicageMacTimerValue string `json:"duplicate_mac_timer_value"`
	MaxCount               string `json:"max_count"`
	TargetCommunity        string `json:"target_community"`
	RouteTargetBoth        string `json:"route_target_both"`
	IgnoreAs               string `json:"ignore_as"`
}

//ConfigOVGResponse is a response object representing the "overlay-gateway" config
type ConfigOVGResponse struct {
	Name       string `json:"name"`
	GwType     string `json:"gateway_type"`
	LoopbackID string `json:"loopback_id"`
	VNIAuto    string `json:"vni_auto"`
	Activate   string `json:"activate"`
}

//ConfigBGPResponse is a response object representing the "router bgp" config
type ConfigBGPResponse struct {
	//BGP Switch Fields
	LocalAS        string                               `json:"local_as"`
	Network        string                               `json:"network"`
	MaxPaths       string                               `json:"max_paths"`
	BFDRx          string                               `json:"bfd_rx"`
	BFDTx          string                               `json:"bfd_tx"`
	BFDMultiplier  string                               `json:"bfd_multiplier"`
	PeerGroups     []ConfigBGPPeerGroupResponse         `json:"peer_groups"`
	Neighbors      []ConfigBGPPeerGroupNeighborResponse `json:"neighbors"`
	L2VPN          ConfigBGPL2VPNResponse               `json:"l2vpn"`
	NetworkList    []string                             `json:"network_list"`
	IPv4PeerGroups []string                             `json:"ipv4_peer_groups"`
}

//ConfigBGPPeerGroupResponse is a response object representing the router bgp neighbour config
type ConfigBGPPeerGroupResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	BFD         string `json:"bfd"`
	RemoteAS    string `json:"remote_as"`
	Multihop    string `json:"multihop"`
}

//ConfigBGPPeerGroupNeighborResponse is a response object representing the router bgp neighbour group config
type ConfigBGPPeerGroupNeighborResponse struct {
	RemoteIP    string `json:"remote_ip"`
	RemoteAS    string `json:"remote_as"`
	PeerGroup   string `json:"peer_group"`
	Multihop    string `json:"bgp_multihop"`
	NextHopSelf string `json:"next_hop_self"`
}

//ConfigBGPL2VPNResponse is a response object representing the group bgp evpn neighbour
type ConfigBGPL2VPNResponse struct {
	RetainRTAll  string                          `json:"retain_route_target_all"`
	GraceRestart string                          `json:"graceful_restart"`
	Neighbors    []ConfigBGPEVPNNeighborResponse `json:"evpn_neighbors"`
}

//ConfigBGPEVPNNeighborResponse is a response object representing the bgp evpn neighbour
type ConfigBGPEVPNNeighborResponse struct {
	PeerGroup        string `json:"peer_group"`
	Encapsulation    string `json:"encapsulation"`
	AllowASIn        string `json:"allowas_in"`
	Activate         string `json:"activate"`
	IPAddress        string `json:"ip_address"`
	NextHopUnchanged string `json:"next_hop_unchanged"`
}

func (c ConfigSwitchResponse) String() string {
	y, err := yaml.Marshal(c)
	if err == nil {
		return string(y)
	}
	return fmt.Sprint(err)
}

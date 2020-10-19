package domain

import (
	"errors"
)

const (
	//P2PIpTypeNumbered represents point to point numbered interface
	P2PIpTypeNumbered = "numbered"

	//P2PIpTypeUnnumbered represents point to point unnumbered interface
	P2PIpTypeUnnumbered = "unnumbered"
)

const (
	//CLOSFabricType represent CLOS Fabric
	CLOSFabricType = "clos"

	//NonCLOSFabricType represent CLOS Fabric
	NonCLOSFabricType = "non-clos"
)

var (
	//ErrFabricNotFound implies the input fabric is not found
	ErrFabricNotFound = errors.New("A fabric with the specified name was not found")

	//ErrFabricActive implies the input fabric is active and cannot be updated
	ErrFabricActive = errors.New("A fabric is already Active and Cannot be updated ")

	//ErrFabricIncorrectValues implies the input values of fabric setting are incorrect
	ErrFabricIncorrectValues = errors.New("Incorrect values specified for Fabric setting")

	//ErrFabricInternalError implies an internal error
	ErrFabricInternalError = errors.New("Internal error")
)

//Fabric represents DC Fabric table
type Fabric struct {
	Name             string
	ID               uint
	FabricProperties FabricProperties
}

//FabricProperties represents a table containing properties of the fabric
type FabricProperties struct {
	ID       uint `json:"id"`
	FabricID uint `json:"fabric_id"`

	//Switch Level Fields
	ConfigureOverlayGateway string `json:"configure_overlay_gateway"`
	MTU                     string `json:"mtu"`
	IPMTU                   string `json:"ip_mtu"`

	//BGP Fields
	SpineASNBlock  string `json:"spine_asn_block"`
	LeafASNBlock   string `json:"leaf_asn_block"`
	RackASNBlock   string `json:"rack_asn_block"`
	BGPMultiHop    string `json:"bgp_multihop"`
	MaxPaths       string `json:"max_paths"`
	AllowASIn      string `json:"allow_as_in"`
	LeafPeerGroup  string `json:"leaf_peer_group"`
	SpinePeerGroup string `json:"spine_peer_group"`

	//Interface Fields
	P2PLinkRange       string `json:"p2p_link_range"`
	P2PIPType          string `json:"p2p_ip_type"`
	LoopBackIPRange    string `json:"loopback_ip_range"`
	LoopBackPortNumber string `json:"loopback_port_number"`
	MCTLinkIPRange     string `json:"mct_link_ip_range"`
	MCTL3LBIPRange     string `json:"mct_l3_lb_ip_range"`
	BFDEnable          string `json:"bfd_enable"`
	BFDTx              string `json:"bfd_tx"`
	BFDRx              string `json:"bfd_rx"`
	BFDMultiplier      string `json:"bfd_multiplier"`

	//OVG Fields
	VTEPLoopBackPortNumber string `json:"vtep_loopback_port_number"`
	VNIAutoMap             string `json:"vni_auto_map"`
	AnyCastMac             string `json:"any_cast_mac"`
	IPV6AnyCastMac         string `json:"ipv6_any_cast_mac"`

	//EVPN Fields
	ArpAgingTimeout               string `json:"arp_aging_timeout"`
	MacAgingTimeout               string `json:"mac_aging_timeout"`
	MacAgingConversationalTimeout string `json:"mac_legacy_aging_timeout"`
	MacMoveLimit                  string `json:"mac_move_limit"`
	DuplicateMacTimer             string `json:"duplicate_mac_timer"`
	DuplicateMaxTimerMaxCount     string `json:"duplicate_mac_timer_max_count"`

	//MCT Configs
	ControlVlan           string `json:"control_vlan"`
	ControlVE             string `json:"control_ve"`
	MctPortChannel        string `json:"mct_port_channel"`
	RoutingMctPortChannel string `json:"routing_mct_port_channel"`

	// Fabric Type
	FabricType string `json:"fabric_type"`

	// NON ClOS Fields
	RackPeerEBGPGroup string `json:"rack_peer_ebgp_group"`
	RackPeerOvgGroup  string `json:"rack_peer_overlay_evpn_group"`
}

/*type FabricOperations interface {
	AddFabric(FabricName string)
	DeleteFabric(FabricName string)
	ListFabrics() []Fabric
}*/

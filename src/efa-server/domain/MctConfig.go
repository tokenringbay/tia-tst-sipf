package domain

const (
	//MctCreate represents the create operation of the MCT cluster
	MctCreate = 1

	//MctDelete represents the delete operation of the MCT cluster
	MctDelete = 2

	//MctUpdate represents the update operation of the MCT cluster
	MctUpdate = 3
)

const (
	//DefaultControlVlan represents the default cluster control VLAN, user can choose to override the cluster control VLAN to non-default
	DefaultControlVlan = "4090"

	//DefaultControlVE represents the default cluster control VE, user can choose to override the cluster control VE to non-default
	DefaultControlVE = "4090"

	//DefaultMctPortChannel represents the default cluster peer interface, user can choose to override the cluster peer interface to non-default
	DefaultMctPortChannel = "1024"

	//RoutingDefaultMctPortChannel represents the default cluster peer interface, user can choose to override the cluster peer interface to non-default
	RoutingDefaultMctPortChannel = "64"

	//DefaultMctLinkRange represents the default , cluster Default IP Range,  user can choose to override the cluster Default Mct Link Range
	DefaultMctLinkRange = "10.20.20.0/24"

	//DefaultMctL3LBRange represents the default , cluster Default IP Range,  user can choose to override the cluster Default Mct Link Range
	DefaultMctL3LBRange = "10.30.30.0/24"

	//MCTPoolName represents name of MCT IP Pool
	MCTPoolName = "MCTLink"

	//RackL3LoopBackPoolName represents name of MCT IP Pool
	RackL3LoopBackPoolName = "RACKL3LB"

	//BGPEncapTypeForCluster represents the place holder for BGP Neighbour encapsulation type.
	BGPEncapTypeForCluster = "nsh" //This can be "nsh" or "mpls" or "mct"(it is internally mapped to either of these in actions)

)

const (
	//BitPositionForPeerOneIP represnts BIT Position for PeerOne IP
	BitPositionForPeerOneIP = 0

	//BitPositionForPeerTwoIP represents BIT Position for PeerTwo IP
	BitPositionForPeerTwoIP = 1

	//BitPositionForForPortAdd represents BIT Position to Add port
	BitPositionForForPortAdd = 2

	//BitPositionForForPortDelete represents BIT Position to Delete  port
	BitPositionForForPortDelete = 3

	//BitPositionForPeerSpeed represents BIT Position for PeerInterface Speed
	BitPositionForPeerSpeed = 4

	//BitPositionForMctCreate represents BIT Position for cluster UPDATE operation
	BitPositionForMctCreate = 30
)

//MCTClusterDetails represents a table
type MCTClusterDetails struct {
	ID                 uint64
	DeviceID           uint
	PrincipalSwitchMac string
	NodeInternalIP     string
	NodePublicIP       string
	NodePrincipal      string
	NodeIsLocal        string
	SerialNnum         string
	NodeCondition      string
	NodeStatus         string
	NodeID             uint8
	FirmwareVersion    string
	NodeMac            string
	NodeSwitchType     string
}

//MCTMemberPorts represents the physical ports of MCT member nodes
type MCTMemberPorts struct {
	ID                   uint64
	ClusterID            uint16
	FabricID             uint
	DeviceID             uint
	RemoteDeviceID       uint
	InterfaceID          uint
	InterfaceType        string
	InterfaceName        string
	InterfaceSpeed       int
	RemoteInterfaceID    uint
	RemoteInterfaceType  string
	RemoteInterfaceName  string
	RemoteInterfaceSpeed int
	ConfigType           string
}

//MctClusterConfig represents a table with brief info of MCT cluster
type MctClusterConfig struct {
	ID                  uint16
	ClusterID           uint16
	FabricID            uint
	DeviceID            uint
	MCTNeighborDeviceID uint
	DeviceOneMgmtIP     string
	DeviceTwoMgmtIP     string
	PeerInterfaceName   string
	PeerInterfacetype   string
	PeerInterfaceSpeed  string
	PeerOneIP           string
	PeerTwoIP           string
	ControlVlan         string
	ControlVE           string
	VEInterfaceOneID    uint
	VEInterfaceTwoID    uint
	ConfigType          string
	LocalNodeID         string
	RemoteNodeID        string
	//Updated Fields 0x1 PeerIP 0x2 Speed
	UpdatedAttributes uint64
}

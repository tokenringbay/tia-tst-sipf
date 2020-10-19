package operation

import (
	"efa-server/domain"
)

//ConfigFabricRequest is a request object to represent configuration of the fabric
type ConfigFabricRequest struct {
	Hosts          []ConfigSwitch
	MctCluster     map[uint][]ConfigCluster
	FabricSettings domain.FabricProperties
	FabricName     string
}

//ConfigSwitch is used by the device actions to configure the fabric
type ConfigSwitch struct {
	Host                     string
	Device                   string
	UserName                 string
	Password                 string
	Afi                      string
	AllowasIn                string
	AnycastMac               string
	IPV6AnycastMac           string
	ConfigureOverlayGateway  string
	BFDEnable                string
	BfdMultiplier            string
	BfdRx                    string
	BfdTx                    string
	Mtu                      string
	IPMtu                    string
	BgpLocalAsn              string
	BgpMultihop              string
	BgpNeighbors             []ConfigBgpNeighbor
	BgpVrf                   string
	Chassis                  string
	Interfaces               []ConfigInterface
	LoopbackPortNumber       string
	P2PIPType                string
	MaxPaths                 string
	P2pLinkRange             string
	RbridgeID                string
	Model                    string
	Principal                bool
	Recursion                bool
	Role                     string
	Source                   string
	UpdateSource             string
	Fabric                   string
	MacMoveThreshold         string
	EnableVf                 string
	LeafPeerGroup            string
	SpinePeerGroup           string
	PeerGroup                string
	PeerGroupDescription     string
	EvpnPeerGroup            string
	EvpnPeerGroupDescription string
	Network                  string
	NonCLOSNetwork           string
	SingleSpineAs            bool

	//OVG Fields
	VtepLoopbackPortNumber string
	VlanVniAutoMap         bool

	//EVPN Fields
	ArpAgingTimeout               string
	MacAgingTimeout               string
	MacAgingConversationalTimeout string
	MacMoveLimit                  string
	DuplicateMacTimer             string
	DuplicateMaxTimerMaxCount     string

	//MCT DATA PLANE
	UnconfigureMCTBGPNeighbors ConfigDataPlaneCluster
	ConfigureMCTBGPNeighbors   ConfigDataPlaneCluster
	//MCT MANAGEMENT
	MctSecondaryNode bool

	// NON ClOS Fields
	RackPeerEBGPGroup string
	RackPeerOvgGroup  string
}

//ConfigBgpNeighbor is used by the device actions to configure BGP Neighbor
type ConfigBgpNeighbor struct {
	CreationTime    interface{} `json:"creation_time"`
	NeighborAddress string      `json:"neighbor_address"`
	NeighborUpTime  string      `json:"neighbor_up_time"`
	RbridgeID       interface{} `json:"rbridge_id"`
	RemoteAs        int64       `json:"remote_as"`
	State           string      `json:"state"`
	ConfigType      string      `json:"config_type"`
	NeighborType    string      `json:"neighbor_type"`
}

//ConfigInterface is used by the device actions to configure interface
type ConfigInterface struct {
	Donor         string
	DonorPort     string
	Description   string
	InterfaceName string
	InterfaceType string
	IP            string
	RbridgeID     string
	ConfigType    string `json:"config_type"`
}

//MgmtClusterNodePrinciPrioDefault represents the default value of management cluster node principal priority
var MgmtClusterNodePrinciPrioDefault = "0"

//ConfigCluster is used by the device actions to configure management cluster
type ConfigCluster struct {
	FabricName         string
	ClusterName        string
	ClusterID          string
	ClusterControlVlan string
	ClusterControlVe   string
	ClusterMemberNodes []ClusterMemberNode
	OperationBitMap    uint64
}

//ClusterMemberNode is used by the device actions to configure management cluster node
type ClusterMemberNode struct {
	NodeMgmtIP                string
	NodeModel                 string
	NodeMgmtUserName          string
	NodeMgmtPassword          string
	NodeID                    string
	NodePrincipalPriority     string
	RemoteNodePeerIP          string
	RemoteNodePeerLoopbackIP  string
	NodePeerIP                string
	NodePeerLoopbackIP        string
	NodePeerIntfType          string
	NodePeerIntfName          string
	NodePeerIntfSpeed         string
	BFDEnable                 string
	BFDTx                     string
	BFDRx                     string
	BFDMultiplier             string
	RemoteNodeConnectingPorts []InterNodeLinkPort
}

//InterNodeLinkPort is used by the device actions to configure the inter node link of the management cluster
type InterNodeLinkPort struct {
	IntfType string
	IntfName string
}

//MgmtClusterStatus is used by the device actions to get the status of management cluster
type MgmtClusterStatus struct {
	TotalMemberNodeCount        string
	DisconnectedMemberNodeCount string
	PrincipalNodeMac            string
	MemberNodes                 []MgmtClusterMemberNode
}

//MgmtClusterMemberNode is used by the device actions to get the status of management cluster node
type MgmtClusterMemberNode struct {
	NodeSerial      string
	NodeMac         string
	NodeMgmtIP      string
	NodeInternalIP  string
	NodeID          string
	NodeCondition   string
	NodeStatus      string
	NodeIsPrincipal string
	NodeIsLOcal     string
	NodeSwitchType  string
	NodeFwVersion   string
}

//ConfigDataPlaneCluster is used by the device actions to configure data plane cluster
type ConfigDataPlaneCluster struct {
	FabricName                  string
	DataPlaneClusterMemberNodes []DataPlaneClusterMemberNode
	OperationBitMap             uint64
}

//DataPlaneClusterMemberNode is used by the device actions to configure data plane cluster node
type DataPlaneClusterMemberNode struct {
	NodeMgmtIP         string
	NodeMgmtUserName   string
	NodeModel          string
	NodeMgmtPassword   string
	NodePeerIP         string
	NodePeerLoopbackIP string
	NodeLoopBackNumber string
	NodePeerASN        string
	NodePeerEncapType  string
	NodePeerBFDEnabled string
}

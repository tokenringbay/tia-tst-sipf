package database

//Fabric struct represents a DataCenter Fabric
type Fabric struct {
	ID               uint `gorm:"primary_key"`
	Name             string
	FabricProperties FabricProperties `gorm:"ForeignKey:FabricID;AssociationForeignKey:Refer"`
}

//FabricProperties represents properties of a DataCenter Fabric
type FabricProperties struct {
	ID                      uint `gorm:"primary_key"`
	FabricID                uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	P2PLinkRange            string
	P2PIPType               string
	LoopBackIPRange         string
	LoopBackPortNumber      string
	SpineASNBlock           string
	LeafASNBlock            string
	RackASNBlock            string `gorm:"default:'4200000000-4200065534'"`
	VTEPLoopBackPortNumber  string
	AnyCastMac              string
	IPV6AnyCastMac          string
	ConfigureOverlayGateway string
	VNIAutoMap              string
	BFDEnable               string `gorm:"default:'Yes'"` //Previous release BFD was enabled
	BFDTx                   string
	BFDRx                   string
	BFDMultiplier           string
	BGPMultiHop             string
	MaxPaths                string
	AllowASIn               string
	MTU                     string
	IPMTU                   string
	LeafPeerGroup           string
	SpinePeerGroup          string

	//EVPN Fields
	ArpAgingTimeout               string
	MacAgingTimeout               string
	MacAgingConversationalTimeout string
	MacMoveLimit                  string
	DuplicateMacTimer             string
	DuplicateMaxTimerMaxCount     string

	//MCT Related
	MCTLinkIPRange string
	MCTL3LBIPRange string `gorm:"default:'10.30.30.0/24'"`
	ControlVlan    string
	ControlVE      string
	MctPortChannel string
	//The default value is provided here so that value gets initialized in case of upgrade
	RoutingMctPortChannel string `gorm:"default:'64'"`

	// Fabric Type
	FabricType string `gorm:"default:'clos'"`

	// NON ClOS Fields
	RackPeerEBGPGroup string `gorm:"default:'underlay-ebgp-group'"`
	RackPeerOvgGroup  string `gorm:"default:'overlay-ebgp-group'"`
}

//Device represents a switching device
type Device struct {
	ID              uint `gorm:"primary_key"`
	FabricID        uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	Name            string
	IPAddress       string
	RbridgeID       string
	DeviceRole      string
	LocalAs         string
	UserName        string
	Password        string
	FirmwareVersion string
	Model           string
	DeviceType      string
	LLDPData        []LLDPData      `gorm:"ForeignKey:DeviceOneID;AssociationForeignKey:Refer"`
	PhysInterface   []PhysInterface `gorm:"ForeignKey:DeviceOneID;AssociationForeignKey:Refer"`
}

//Rack represents a rack containing two Leaf nodes
type Rack struct {
	ID          uint `gorm:"primary_key"`
	FabricID    uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceOneID uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	DeviceTwoID uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	DeviceOneIP string
	DeviceTwoIP string
	RackName    string
}

//RackEvpnNeighbors represent evpn neighbor per rack
type RackEvpnNeighbors struct {
	ID             uint `gorm:"primary_key"`
	LocalRackID    uint `sql:"type:integer REFERENCES racks(id) ON DELETE CASCADE"`
	LocalDeviceID  uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	RemoteRackID   uint `sql:"type:integer REFERENCES racks(id) ON DELETE CASCADE"`
	RemoteDeviceID uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	EVPNAddress    string
	RemoteAS       string
	ConfigType     string
}

//LLDPData represents LLDP neighbour info of a switching device
type LLDPData struct {
	ID                      uint `gorm:"primary_key"`
	DeviceID                uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	FabricID                uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	LocalIntName            string
	LocalIntType            string
	LocalIntMac             string
	RemoteIntName           string
	RemoteIntType           string
	RemoteIntMac            string
	RemoteChassisID         string
	RemoteSystemName        string
	RemoteManagementAddress string
	ConfigType              string
}

//LLDPNeighbor represents the LLDP switching device pair
type LLDPNeighbor struct {
	ID               uint `gorm:"primary_key"`
	FabricID         uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceOneID      uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	DeviceTwoID      uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	InterfaceOneID   uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	InterfaceTwoID   uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	DeviceOneRole    string
	DeviceTwoRole    string
	InterfaceOneName string
	InterfaceOneType string
	InterfaceOneIP   string
	InterfaceTwoName string
	InterfaceTwoType string
	InterfaceTwoIP   string
	ConfigType       string
}

//PhysInterface represents physical interface of a switching device
type PhysInterface struct {
	ID             uint `gorm:"primary_key"`
	FabricID       uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceID       uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	IntType        string
	IntName        string
	InterfaceSpeed string
	IPAddress      string
	Identifier     string
	Mac            string
	role           string
	ConfigType     string
}

//ASNAllocationPool represents the unallocated ASN of a switching device
type ASNAllocationPool struct {
	ID         uint `gorm:"primary_key"`
	FabricID   uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	ASN        uint64
	DeviceRole string
}

//UsedASN represents the allocated ASN of a switching device
type UsedASN struct {
	ID         uint `gorm:"primary_key"`
	FabricID   uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceID   uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	ASN        uint64
	DeviceRole string
}

//IPAllocationPool represents unallocated IP of a switching device
type IPAllocationPool struct {
	ID        uint `gorm:"primary_key"`
	FabricID  uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	IPAddress string
	IPType    string
}

//UsedIP represents allocated IP of a switching device
type UsedIP struct {
	ID          uint `gorm:"primary_key"`
	FabricID    uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceID    uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	IPAddress   string
	IPType      string
	InterfaceID uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
}

//IPPairAllocationPool represents unallocated IP pair of a switching device
type IPPairAllocationPool struct {
	ID           uint `gorm:"primary_key"`
	FabricID     uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	IPAddressOne string
	IPAddressTwo string
	IPType       string
}

//UsedIPPair represents allocated IP pair of a switching device
type UsedIPPair struct {
	ID             uint `gorm:"primary_key"`
	FabricID       uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceOneID    uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	DeviceTwoID    uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	IPAddressOne   string
	IPAddressTwo   string
	IPType         string
	InterfaceOneID uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	InterfaceTwoID uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
}

//SwitchConfig represents config to be pushed to a switching device
type SwitchConfig struct {
	ID                       uint `gorm:"primary_key"`
	DeviceID                 uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	FabricID                 uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceIP                 string
	LocalAS                  string
	LoopbackIP               string
	VTEPLoopbackIP           string
	Role                     string
	ASConfigType             string
	LoopbackIPConfigType     string
	VTEPLoopbackIPConfigType string
}

//InterfaceSwitchConfig represents interface config to be pushed to a switching device
type InterfaceSwitchConfig struct {
	ID          uint `gorm:"primary_key"`
	FabricID    uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceID    uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	InterfaceID uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	IntType     string
	IntName     string
	DonorType   string
	DonorName   string
	IPAddress   string
	ConfigType  string
	Description string
}

//RemoteNeighborSwitchConfig represents the config of a remote neighbour switching device
type RemoteNeighborSwitchConfig struct {
	ID                uint `gorm:"primary_key"`
	FabricID          uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceID          uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	RemoteInterfaceID uint `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	RemoteDeviceID    uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	EncapsulationType string
	RemoteIPAddress   string
	RemoteAS          string
	Type              string
	ConfigType        string
}

//ExecutionLog represents detailed info of the executed operations w.r.t. the application
type ExecutionLog struct {
	ID        uint `gorm:"primary_key"`
	UUID      string
	Command   string
	Params    string
	Status    string
	StartTime string
	EndTime   string
	Duration  string
}

//MCT Related Tables
//Gorm Convention - Column name will be the lower snake case fields name
//DONT CHANGE NAMES OF STRUCT FIELDS THEY ARE USED IN DOMAIN LAYER

//MCTClusterDetail represents detailed status of the MCT cluster w.r.t. a switching device
type MCTClusterDetail struct {
	ID                 uint64 `gorm:"primary_key" gorm:"type:bigint; DEFAULT:id_generator()"`
	NodeID             uint8
	DeviceID           uint   `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	PrincipalSwitchMac string `gorm:"size:64"` // set field size to 64
	NodeInternalIP     string `gorm:"size:30"`
	NodePublicIP       string `gorm:"size:30"`
	NodePrincipal      string `gorm:"size:30"`
	NodeIsLocal        string `gorm:"size:20"`
	SerialNnum         string `gorm:"size:30"`
	NodeCondition      string `gorm:"size:30"`
	NodeStatus         string `gorm:"size:60"`
	FirmwareVersion    string `gorm:"size:128"`
	NodeMac            string `gorm:"size:30"`
	NodeSwitchType     string `gorm:"size:30"`
}

//ClusterMember represents the MCT cluster physical topology
type ClusterMember struct {
	ID                   uint64 `gorm:"primary_key"`
	FabricID             uint   `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	ClusterID            uint16 `sql:"type:integer REFERENCES mct_cluster_configs(id) ON DELETE CASCADE"`
	InterfaceID          uint   `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	DeviceID             uint   `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	RemoteDeviceID       uint   `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	InterfaceName        string `gorm:"size:255"`
	InterfaceType        string `gorm:"size:64"`
	InterfaceSpeed       int
	RemoteInterfaceID    uint   `sql:"type:integer REFERENCES phys_interfaces(id) ON DELETE CASCADE"`
	RemoteInterfaceName  string `gorm:"size:255"`
	RemoteInterfaceType  string `gorm:"size:64"`
	RemoteInterfaceSpeed int
	ConfigType           string
}

//MctClusterConfig represents MCT cluster config of a switching device
type MctClusterConfig struct {
	ID                  uint16 `gorm:"primary_key" gorm:"type:bigint; DEFAULT:id_generator()"`
	ClusterID           uint16
	FabricID            uint `sql:"type:integer REFERENCES fabrics(id) ON DELETE CASCADE"`
	DeviceID            uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
	MCTNeighborDeviceID uint `sql:"type:integer REFERENCES devices(id) ON DELETE CASCADE"`
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

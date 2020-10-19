package domain

const (
	//ConfigCreate represents a create config operation
	ConfigCreate = "CREATE CONFIG"

	//ConfigUpdate represents an update config operation
	ConfigUpdate = "UPDATE CONFIG"

	//ConfigDelete represents a delete config operation
	ConfigDelete = "DELETE CONFIG"

	//ConfigNone represents a no operation
	ConfigNone = "NONE CONFIG"
)

const (
	//MCTL3LBType represents MCT L3 BGP Neighbor
	MCTL3LBType = "RACKL3LB"

	//MCTBGPType represents MCT BGP Neighbor
	MCTBGPType = "MCTBGP"

	//FabricBGPType represents  BGP Neighbor
	FabricBGPType = "FABBGP"

	//EVPNENIGHBORType represents EVPN Neighbor
	EVPNENIGHBORType = "EVPNBGP"
)

//SwitchConfig represents a table to store the switch config
type SwitchConfig struct {
	ID                       uint
	DeviceID                 uint
	FabricID                 uint
	DeviceIP                 string
	LocalAS                  string
	LoopbackIP               string
	VTEPLoopbackIP           string
	Role                     string
	ASConfigType             string
	LoopbackIPConfigType     string
	VTEPLoopbackIPConfigType string
	//userName and Password Fields are not stored in DB
	UserName string
	Password string
}

//InterfaceSwitchConfig represents a table to store the switch interface config
type InterfaceSwitchConfig struct {
	ID          uint
	FabricID    uint
	DeviceID    uint
	InterfaceID uint
	IntType     string
	IntName     string
	DonorType   string
	DonorName   string
	IPAddress   string
	ConfigType  string
	Description string
}

//RemoteNeighborSwitchConfig represents a table to store the remote neighbour switch config
type RemoteNeighborSwitchConfig struct {
	ID                uint
	FabricID          uint
	DeviceID          uint
	RemoteInterfaceID uint
	EncapsulationType string
	RemoteDeviceID    uint
	RemoteIPAddress   string
	RemoteAS          string
	Type              string
	ConfigType        string
}

//RackEvpnNeighbors represent evpn neighbor per rack
type RackEvpnNeighbors struct {
	ID             uint
	LocalRackID    uint
	LocalDeviceID  uint
	RemoteRackID   uint
	RemoteDeviceID uint
	EVPNAddress    string
	RemoteAS       string
	ConfigType     string
}

package domain

const (
	//IntfTypeLoopback represents the logical loopback interface
	IntfTypeLoopback = "loopback"

	//IntfTypeEthernet represents the physical ethernet interface
	IntfTypeEthernet = "ethernet"

	//IntfTypeVe represents the logical VE/SVI interface
	IntfTypeVe = "ve"
)

//LLDP contains the LLDP neighbour info
type LLDP struct {
	ID                      uint
	DeviceID                uint
	FabricID                uint
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

//Interface contains detailed interface info
type Interface struct {
	ID             uint
	FabricID       uint
	DeviceID       uint
	IntType        string
	IntName        string
	InterfaceSpeed string
	IPAddress      string
	Identifier     string
	Mac            string
	role           string
	ConfigType     string
	//ConfigState indicate where the port is in shutdown or up state
	ConfigState string
}

//LLDPNeighbor contains prepared info from Interface and LLDP
type LLDPNeighbor struct {
	ID               uint
	FabricID         uint
	DeviceOneID      uint
	DeviceTwoID      uint
	DeviceOneRole    string
	DeviceTwoRole    string
	InterfaceOneID   uint
	InterfaceTwoID   uint
	InterfaceOneName string
	InterfaceOneType string
	InterfaceOneIP   string
	InterfaceTwoName string
	InterfaceTwoType string
	InterfaceTwoIP   string
	ConfigType       string
}

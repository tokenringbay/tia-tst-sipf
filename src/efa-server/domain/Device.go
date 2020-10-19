package domain

//Device represents a switch device table
type Device struct {
	ID                  uint
	FabricID            uint
	Name                string
	IPAddress           string
	RbridgeID           string
	DeviceRole          string
	LocalAs             string
	UserName            string
	Password            string
	FirmwareVersion     string
	Model               string
	DeviceType          string
	LLDPS               []LLDP
	Interfaces          []Interface
	IsPasswordEncrypted bool
}

//DeviceDetail holds the Device details
type DeviceDetail struct {
	FabricID        uint
	DeviceID        uint
	FirmwareVersion string
	Model           string
}

//DeviceOperations represents
/*type DeviceOperations interface {
	AddDevice(FabricName string, IPAddress string, UserID string, Password string) (string, error)
	DeleteDevice(FabricName string, IPAddress string) (string, error)
	ListDevices() ([]Device, error)
}*/

//Rack represents a rack containing two Leaf nodes
type Rack struct {
	ID          uint
	FabricID    uint
	DeviceOneID uint
	DeviceTwoID uint
	DeviceOneIP string
	DeviceTwoIP string
	RackName    string
}

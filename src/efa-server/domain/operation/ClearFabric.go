package operation

//ClearFabricRequest is a request Object for clearing the config of fabric
type ClearFabricRequest struct {
	FabricName string              `json:"fabric_name"`
	Hosts      []ClearSwitchDetail `json:"hosts"`
}

//ClearSwitchDetail is a request object for clearing the config of switch
type ClearSwitchDetail struct {
	Host            string                 `json:"switch"`
	Role            string                 `json:"role"`
	Model           string                 `json:"model"`
	UserName        string                 `json:"user"`
	Password        string                 `json:"password"`
	LoopBackIPRange string                 `json:"loopback_ip_range"`
	Interfaces      []ClearInterfaceDetail `json:"inerfaces"`
}

//ClearInterfaceDetail is a request object for clearing the config of interface
type ClearInterfaceDetail struct {
	InterfaceName string `json:"name"`
	InterfaceType string `json:"type"`
	IP            string `json:"ip"`
}

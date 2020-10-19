package models

//InterfaceSwitchResponse structure represents Interface response from the switching device
type InterfaceSwitchResponse struct {
	InterfaceName  string
	InterfaceType  string
	InterfaceSpeed string
	IPAddress      string
	InterfaceMac   string
	ConfigState    string
}

//InterfaceLLDPResponse structure represents LLDP Response from the switching device
type InterfaceLLDPResponse struct {
	LocalInterfaceName  string
	LocalInterfaceType  string
	LocalInterfaceMac   string
	RemoteInterfaceName string
	RemoteInterfaceType string
	RemoteInterfaceMac  string
}

//OverlayGatewayResponse structure represents overlay-gateway response
type OverlayGatewayResponse struct {
	Data struct {
		OverlayGateway struct {
			Xmlns  string `xml:"-xmlns"`
			Name   string `xml:"name"`
			GwType string `xml:"gw-type"`
			IP     struct {
				Interface struct {
					Loopback struct {
						LoopbackID string `xml:"loopback-id"`
					} `xml:"loopback"`
				} `xml:"interface"`
			} `xml:"ip"`
			Map struct {
				VlanAndBd struct {
					Vni struct {
					} `xml:"vni"`
				} `xml:"vlan-and-bd"`
			} `xml:"map"`
		} `xml:"overlay-gateway"`
	} `xml:"data"`
}

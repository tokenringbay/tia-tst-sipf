package base

import (
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
	"efa-server/infra/device/models"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"strings"
)

var phyIntfSpeed100MBPS = 100
var phyIntfSpeed1GBPS = 1000
var phyIntfSpeed10GBPS = 10000
var phyIntfSpeed25GBPS = 25000
var phyIntfSpeed40GBPS = 40000
var phyIntfSpeed100GBPS = 100000

func (base *SLXBase) getInterfaceDetailRequest(lastInterfaceName string, lastInterfaceType string) string {
	var requestInterface string
	if lastInterfaceName == "" {
		requestInterface = `<get-interface-detail xmlns="urn:brocade.com:mgmt:brocade-interface-ext"></get-interface-detail>`
	} else {
		requestInterfaceFmt := `<get-interface-detail xmlns="urn:brocade.com:mgmt:brocade-interface-ext">
		                      <last-rcvd-interface>
			                      <interface-type>%s</interface-type>
			                      <interface-name>%s</interface-name>
		                      </last-rcvd-interface>
							  </get-interface-detail>`
		requestInterface = fmt.Sprintf(requestInterfaceFmt, lastInterfaceType, lastInterfaceName)

	}
	return requestInterface
}

func (base *SLXBase) getIPInterfaceRequest(lastInterfaceName string, lastInterfaceType string) string {
	requestInterface := `<get-ip-interface xmlns="urn:brocade.com:mgmt:brocade-interface-ext"></get-ip-interface>`
	return requestInterface
}

func (base *SLXBase) getInterfacesMac(client *client.NetconfClient) (map[string]string, error) {
	var InterfaceMac map[string]string
	InterfaceMac = make(map[string]string)
	hasMore := "true"
	lastInterfaceName := ""
	lastInterfaceType := ""

	for hasMore == "true" {
		request := base.getInterfaceDetailRequest(lastInterfaceName, lastInterfaceType)
		resp, err := client.ExecuteRPC(request)
		if err != nil {
			return InterfaceMac, err
		}
		//log.Infoln(resp)
		doc := etree.NewDocument()
		if err := doc.ReadFromBytes([]byte(resp)); err != nil {
			return InterfaceMac, err
		}
		if elems := doc.FindElements("//interface"); elems != nil {
			for _, elem := range elems {
				if de := elem.FindElement(".//interface-name"); de != nil {

					lastInterfaceName = de.Text()
				}
				if de := elem.FindElement(".//interface-type"); de != nil {

					lastInterfaceType = de.Text()
				}
				if de := elem.FindElement(".//current-hardware-address"); de != nil {
					InterfaceMac[lastInterfaceType+"_"+lastInterfaceName] = de.Text()

				}

			}
			if de := doc.FindElement("//has-more"); de != nil {

				hasMore = de.Text()
			}
		}
	}

	return InterfaceMac, nil
}

//GetInterfaces gets the operational state of interfaces, from switching device
func (base *SLXBase) GetInterfaces(client *client.NetconfClient, ControlVlan string) ([]models.InterfaceSwitchResponse, error) {
	Interfaces := make([]models.InterfaceSwitchResponse, 0)
	InterfaceMac, _ := base.getInterfacesMac(client)

	request := base.getIPInterfaceRequest("", "")
	resp, err := client.ExecuteRPC(request)
	//log.Infoln(resp)

	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return Interfaces, err
	}

	if elems := doc.FindElements("//interface"); elems != nil {
		var Interface models.InterfaceSwitchResponse
		for _, elem := range elems {

			if de := elem.FindElement(".//interface-type"); de != nil {
				Interface.InterfaceType = strings.ToLower(de.Text())
				if !(strings.Contains(Interface.InterfaceType, domain.IntfTypeEthernet) ||
					strings.Contains(Interface.InterfaceType, domain.IntfTypeLoopback) ||
					strings.Contains(Interface.InterfaceType, domain.IntfTypeVe)) {
					continue
				}
			}
			if de := elem.FindElement(".//if-state"); de != nil {
				Interface.ConfigState = de.Text()
			}
			if de := elem.FindElement(".//interface-name"); de != nil {
				Interface.InterfaceName = de.Text()
			}
			if strings.Contains(Interface.InterfaceType, domain.IntfTypeVe) && ControlVlan != Interface.InterfaceName {
				continue
			}
			if de := elem.FindElement(".//ipv4"); de != nil {
				ipAddress := de.Text()
				if ipAddress != "unassigned" {
					Interface.IPAddress = ipAddress
				} else {
					Interface.IPAddress = ""
				}

			}
			Interface.InterfaceMac = InterfaceMac[Interface.InterfaceType+"_"+Interface.InterfaceName]
			Interfaces = append(Interfaces, Interface)
		}
	}

	return Interfaces, err
}

//GetInterfaceSpeed is used to get the "speed" of the physical interface
func (base *SLXBase) GetInterfaceSpeed(client *client.NetconfClient, interfaceType string, interfaceName string) (int, error) {
	var requestInterface string
	if interfaceName == "" || interfaceType == "" {
		return 0, errors.New("Invalid arguments")
	}

	if interfaceType == "Eth" {
		interfaceType = "ethernet"
	}

	requestInterfaceFmt := `<get-interface-detail xmlns="urn:brocade.com:mgmt:brocade-interface-ext">
      			               <interface-type>%s</interface-type>
                               <interface-name>%s</interface-name>
                             </get-interface-detail>`
	requestInterface = fmt.Sprintf(requestInterfaceFmt, interfaceType, interfaceName)

	RawResponse, err := client.ExecuteRPC(requestInterface)

	if err != nil {
		fmt.Println("Error from get-media-detail RPC execution: ", err)
		return 0, err
	}

	//fmt.Println("Response of get-media-detail: ", RawResponse)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(RawResponse)); err != nil {
		fmt.Println("Error in reading the get-interface-detail RPC o/p")
		return 0, err
	}

	de := doc.FindElement("//actual-line-speed")
	if de != nil {
		return base.getSpeedInt(de.Text()), nil
	}

	return 0, nil
}

func (base *SLXBase) getMinInterfaceSpeed(client *client.NetconfClient, inputInterfaces []operation.InterNodeLinkPort) (int, error) {
	if len(inputInterfaces) == 0 {
		return 0, errors.New("Input interface array is empty")
	}
	var minSpeed = phyIntfSpeed100GBPS
	for _, phyInterface := range inputInterfaces {
		speed, err := base.GetInterfaceSpeed(client, phyInterface.IntfType, phyInterface.IntfName)
		if err != nil {
			return 0, err
		}
		if minSpeed > speed {
			minSpeed = speed
		}
	}
	return minSpeed, nil
}

func (base *SLXBase) getSpeedInt(speedStr string) int {
	switch speedStr {
	case "100Mbps":
		return phyIntfSpeed100MBPS
	case "1Gbps":
		return phyIntfSpeed1GBPS
	case "10Gbps":
		return phyIntfSpeed10GBPS
	case "25Gbps":
		return phyIntfSpeed25GBPS
	case "40Gbps":
		return phyIntfSpeed40GBPS
	case "100Gbps":
		return phyIntfSpeed100GBPS
	default:
		return 0
	}
}

//PersistConfig is used to persist the running-config to startup-config, on the switching device
func (base *SLXBase) PersistConfig(client *client.NetconfClient) (map[string]string, error) {
	ResultMap := make(map[string]string)
	var persistMap = map[string]interface{}{}
	config, _ := base.GetStringFromTemplate(persistConfig, persistMap)

	resp, err := client.ExecuteRPC(config)

	if err != nil {
		return ResultMap, err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}
	if elem := doc.FindElement("//session-id"); elem != nil {
		ResultMap["session-id"] = elem.Text()
	}
	if elem := doc.FindElement("//status"); elem != nil {
		ResultMap["status"] = elem.Text()
	}
	return ResultMap, err

}

//ConfigureInterfaceLoopback is used to configure the "loopback" interface and its subconfig on the switching device
func (base *SLXBase) ConfigureInterfaceLoopback(client *client.NetconfClient, loopbackID string, ipAddress string) (string, error) {
	var loopbackMap = map[string]interface{}{"loopback_id": loopbackID, "ipaddress": ipAddress}

	config, templateError := base.GetStringFromTemplate(ipLoopbackCreate, loopbackMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(ipLoopbackActivate, loopbackMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	return resp, err
}

//DeleteInterfaceLoopback is used to delete the "loopback" interface from the switching device
func (base *SLXBase) DeleteInterfaceLoopback(client *client.NetconfClient, loopbackID string) (string, error) {
	var loopbackMap = map[string]interface{}{"loopback_id": loopbackID}

	config, templateError := base.GetStringFromTemplate(ipLoopbackDelete, loopbackMap)

	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err
}

//ConfigureInterfaceNumbered is used to configure numbered interface on the switching device
func (base *SLXBase) ConfigureInterfaceNumbered(client *client.NetconfClient, intType string, intName string,
	ipAddress string, description string) (string, error) {

	var interfaceMap = map[string]interface{}{"int_name": intName, "int_type": intType,
		"ipaddress": ipAddress, "description": description}

	config, templateError := base.GetStringFromTemplate(interfaceNumberedCreate, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(interfaceActivate, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	return resp, err
}

//UnconfigureInterfaceNumbered is used to unconfigure numbered interface from the switching device
func (base *SLXBase) UnconfigureInterfaceNumbered(client *client.NetconfClient, intType string, intName string,
	ipAddress string) (string, error) {

	var interfaceMap = map[string]interface{}{"int_name": intName, "int_type": intType,
		"ipaddress": ipAddress}

	config, templateError := base.GetStringFromTemplate(interfaceNumberedDelete, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureInterfaceDesc is used to unconfigure interface description from the switching device
func (base *SLXBase) UnconfigureInterfaceDesc(client *client.NetconfClient, intType string, intName string) (string, error) {

	var interfaceMap = map[string]interface{}{"int_name": intName, "int_type": intType}

	config, templateError := base.GetStringFromTemplate(interfaceDescDelete, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureInterfaceSpeed is used to unconfigure interface speed from the switching device
func (base *SLXBase) UnconfigureInterfaceSpeed(client *client.NetconfClient, intType string, intName string) (string, error) {

	var interfaceMap = map[string]interface{}{"int_name": intName, "int_type": intType}

	config, templateError := base.GetStringFromTemplate(interfaceSpeedDelete, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err
}

//ConfigureInterfaceUnnumbered is used to configure unnumbered interface on the switching device
func (base *SLXBase) ConfigureInterfaceUnnumbered(client *client.NetconfClient, intType string, intName string, donorType string,
	donorName string) (string, error) {

	var interfaceMap = map[string]interface{}{"int_name": intName, "int_type": intType,
		"donor_interface_type": donorType, "donor_interface_name": donorName}

	config, templateError := base.GetStringFromTemplate(interfaceUnnumberedCreate, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(interfaceActivate, interfaceMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	return resp, err
}

//UnconfigureInterfaceUnnumbered is used to unconfigure unnumbered interface from the switching device
func (base *SLXBase) UnconfigureInterfaceUnnumbered(client *client.NetconfClient, intType string, intName string) (string, error) {

	var interfaceMap = map[string]interface{}{"int_name": intName, "int_type": intType}

	config, templateError := base.GetStringFromTemplate(interfaceUnnumberedDelete, interfaceMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err
}

//GetInterface is used to get the running-config of interface from the switching device
func (base *SLXBase) GetInterface(client *client.NetconfClient, InterfaceType string, InterfaceName string) (map[string]string, error) {
	response := make(map[string]string)
	request := fmt.Sprintf("/interface/%s[name='%s']", InterfaceType, InterfaceName)
	if InterfaceType == domain.IntfTypeLoopback {
		request = fmt.Sprintf("/routing-system/interface/%s[id='%s']", InterfaceType, InterfaceName)
	}

	resp, err := client.GetConfig(request)
	doc := etree.NewDocument()

	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		return response, err
	}

	if elem := doc.FindElement("//id"); elem != nil {
		response["Name"] = elem.Text()
	}
	if elem := doc.FindElement("//address//address"); elem != nil {
		response["address"] = elem.Text()
	}
	if elem := doc.FindElement("//min-rx"); elem != nil {
		response["min-rx"] = elem.Text()
	}
	if elem := doc.FindElement("//multiplier"); elem != nil {
		response["multiplier"] = elem.Text()
	}
	if elem := doc.FindElement("//min-tx"); elem != nil {
		response["min-tx"] = elem.Text()
	}
	if elem := doc.FindElement("//ip-donor-interface-type"); elem != nil {
		response["donor_type"] = elem.Text()
	}
	if elem := doc.FindElement("//ip-donor-interface-name"); elem != nil {
		response["donor_name"] = elem.Text()
	}
	if elem := doc.FindElement("//description"); elem != nil {
		response["description"] = elem.Text()
	}
	if elem := doc.FindElement("//speed"); elem != nil {
		response["speed"] = elem.Text()
	}
	return response, err

}

//GetEthernetInterfaceConfigs is used to get the running-config of ethernet interfaces from the switching device
func (base *SLXBase) GetEthernetInterfaceConfigs(client *client.NetconfClient, intfNames []string) ([]operation.ConfigIntfResponse, error) {
	response := make([]operation.ConfigIntfResponse, 0)
	request := fmt.Sprintf("/interface/%s", domain.IntfTypeEthernet)

	resp, err := client.GetConfig(request)
	doc := etree.NewDocument()
	if err = doc.ReadFromString(resp); err != nil {
		return response, err
	}

	if elem := doc.FindElement("//interface"); elem != nil {
		for _, ethernet := range elem.SelectElements("ethernet") {
			if name := ethernet.SelectElement("name"); name != nil {
				ethIntfName := name.Text()
				for _, val := range intfNames {
					if val == ethIntfName {
						var addr string
						var description string
						if addrElem := ethernet.FindElement(".//address//address"); addrElem != nil {
							addr = addrElem.Text()
						}
						if descrElem := ethernet.FindElement(".//description"); descrElem != nil {
							description = descrElem.Text()
						}
						intfConfig := operation.ConfigIntfResponse{Name: ethIntfName, Type: domain.IntfTypeEthernet, IPAddress: addr, Description: description}
						response = append(response, intfConfig)
						continue
					}
				}
			}
		}
	}
	return response, err
}

//GetLoopbackInterfaceConfigs is used to get the running-config of loopback interfaces from the switching device
func (base *SLXBase) GetLoopbackInterfaceConfigs(client *client.NetconfClient, loopbackIds []string) ([]operation.ConfigIntfResponse, error) {
	response := make([]operation.ConfigIntfResponse, 0)
	request := fmt.Sprintf("/routing-system/interface/%s", domain.IntfTypeLoopback)
	resp, err := client.GetConfig(request)
	doc := etree.NewDocument()

	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		return response, err
	}
	if elem := doc.FindElement("//interface"); elem != nil {
		for _, ethernet := range elem.SelectElements("loopback") {
			if id := ethernet.SelectElement("id"); id != nil {
				loopbackID := id.Text()
				for _, val := range loopbackIds {
					if val == loopbackID {
						var addr string
						description := "Not Applicable"
						if addrElem := ethernet.FindElement(".//address//address"); addrElem != nil {
							addr = addrElem.Text()
						}
						intfConfig := operation.ConfigIntfResponse{Name: loopbackID, Type: domain.IntfTypeLoopback, IPAddress: addr, Description: description}
						response = append(response, intfConfig)
						continue
					}
				}
			}
		}
	}

	return response, err
}

//GetVEInterfaceConfigs is used to get the running-config of VE interfaces from the switching device
func (base *SLXBase) GetVEInterfaceConfigs(client *client.NetconfClient, intfNames []string) ([]operation.ConfigIntfResponse, error) {
	fmt.Println("Get VE Interface configs")
	return nil, nil
}

//GetInterfaceConfigs is used to get the running-config of interface from the switching device
func (base *SLXBase) GetInterfaceConfigs(client *client.NetconfClient, intfs map[string][]string) ([]operation.ConfigIntfResponse, error) {
	response := make([]operation.ConfigIntfResponse, 0)
	var err error
	var responses []operation.ConfigIntfResponse
	for intfType, intfNames := range intfs {
		if domain.IntfTypeEthernet == intfType {
			responses, err = base.GetEthernetInterfaceConfigs(client, intfNames)
			for _, resp := range responses {
				response = append(response, resp)
			}
		} else if domain.IntfTypeLoopback == intfType {
			responses, err = base.GetLoopbackInterfaceConfigs(client, intfNames)
			for _, resp := range responses {
				response = append(response, resp)
			}
		} else if domain.IntfTypeVe == intfType {
			/*			fmt.Println("Fetch intf configs of type VE")
						GetVEInterfaceConfigs(client, intfNames)*/
		}
	}
	return response, err
}

//AddInterfaceToPo is used to add member port to the port-channel, on the switching device
func (base *SLXBase) AddInterfaceToPo(client *client.NetconfClient, name string, portChannelDescription string,
	portChannel string, portChannelMode string, portChannelType string, speed string) (string, error) {

	var portChannelMap = map[string]interface{}{"name": name, "port_channel_description": portChannelDescription,
		"port_channel": portChannel, "port_channel_mode": portChannelMode, "port_channel_type": portChannelType,
		"speed": speed}

	/*Add member to a port-channel. */
	config, templateError := base.GetStringFromTemplate(intAddToPortChannel, portChannelMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	/*config, templateError = base.GetStringFromTemplate(intPhySpeed, portChannelMap)
	if templateError != nil {
		return "", templateError

	}
	resp, err = client.EditConfig(config)

	if err != nil {
		return "", err
	}*/

	var poMemberActivate = map[string]interface{}{"name": name}
	config, templateError = base.GetStringFromTemplate(intPhyActivate, poMemberActivate)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	return resp, err
}

//GetInterfacePoMember is used to get running-config of the port-channel member port, from the switching device
func (base *SLXBase) GetInterfacePoMember(client *client.NetconfClient, name string) (map[string]string, error) {
	ResultMap := make(map[string]string)
	RequestMsg := fmt.Sprintf("/interface/ethernet[name='%s']", name)

	resp, err := client.GetConfig(RequestMsg)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElement("//description"); elem != nil {
		ResultMap["description"] = elem.Text()
	}
	if elem := doc.FindElement("//channel-group/port-int"); elem != nil {
		ResultMap["port-channel"] = elem.Text()
	}
	if elem := doc.FindElement("//channel-group/mode"); elem != nil {
		ResultMap["port-channel-mode"] = elem.Text()
	}
	if elem := doc.FindElement("//channel-group/type"); elem != nil {
		ResultMap["port-channel-type"] = elem.Text()
	}
	if elem := doc.FindElement("//shutdown"); elem != nil {
		ResultMap["shutdown"] = "true"
	}
	return ResultMap, err
}

//DeleteInterfaceFromPo is used to remove the member port from the port-channel interface, from the switching device
func (base *SLXBase) DeleteInterfaceFromPo(client *client.NetconfClient, name string, portChannel string) (string, error) {
	var portChannelMap = map[string]interface{}{"name": name, "port_channel": portChannel}

	config, templateError := base.GetStringFromTemplate(intRemoveFromPortChannel, portChannelMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//EnableInterfaces is used to "no shutdown" all the physical interfaces, on the switching device
func (base *SLXBase) EnableInterfaces(client *client.NetconfClient, interfaceNames []string) (string, error) {
	var intfMap = map[string]interface{}{"interface_names": interfaceNames}
	config, templateError := base.GetStringFromTemplate(intEnableInterfaces, intfMap)

	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//DisableInterface is used to "shutdown" a physical interface, on the switching device
func (base *SLXBase) DisableInterface(client *client.NetconfClient, ifName string) (string, error) {
	var intfMap = map[string]interface{}{"intf_name": ifName}
	config, templateError := base.GetStringFromTemplate(intPhyDeactivate, intfMap)

	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//DisableInterfaces is used to "shutdown" all the physical interfaces, on the switching device
func (base *SLXBase) DisableInterfaces(client *client.NetconfClient, interfaceNames []string) (string, error) {
	var intfMap = map[string]interface{}{"interface_names": interfaceNames}
	config, templateError := base.GetStringFromTemplate(intDisableInterfaces, intfMap)

	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureInterfaceVe is used to configure "interface Ve" and its sub config, on the switching device
func (base *SLXBase) ConfigureInterfaceVe(client *client.NetconfClient, name string,
	ipAddress string, bfdrx string, bfdtx string, bfdmultiplier string) (string, error) {
	var veMap = map[string]interface{}{"name": name, "ip_address": ipAddress}
	config, templateError := base.GetStringFromTemplate(intVeCreate, veMap)

	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(intVeActivate, veMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	return resp, err
}

//UnconfigureInterfaceVeIP is used to unconfigure "ip address" of the "interface Ve", from the switching device
func (base *SLXBase) UnconfigureInterfaceVeIP(client *client.NetconfClient, name string,
	ipAddress string) (string, error) {
	var veMap = map[string]interface{}{"name": name, "ip_address": ipAddress}
	config, templateErr := base.GetStringFromTemplate(intVeDeleteIP, veMap)

	if templateErr != nil {
		return "", templateErr
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureInterfaceVeIP is used to configure "ip address" of the "interface Ve", on the switching device
func (base *SLXBase) ConfigureInterfaceVeIP(client *client.NetconfClient, name string,
	ipAddress string) (string, error) {
	var veMap = map[string]interface{}{"name": name, "ip_address": ipAddress}

	config, templateErr := base.GetStringFromTemplate(intVeSetIP, veMap)
	if templateErr != nil {
		return "", templateErr
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//GetInterfaceVe is used to get running-config of the "interface Ve <id>" from the switching device
func (base *SLXBase) GetInterfaceVe(client *client.NetconfClient, name string) (map[string]string, error) {
	ResultMap := make(map[string]string)
	RequestMsg := fmt.Sprintf("/routing-system/interface/ve[name='%s']", name)

	resp, err := client.GetConfig(RequestMsg)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElement("//address/address"); elem != nil {
		ResultMap["ip-address"] = elem.Text()
	}
	if elem := doc.FindElement("//shutdown"); elem != nil {
		ResultMap["shutdown"] = "true"
	}
	return ResultMap, err
}

//DeleteInterfaceVe is used to delete the "interface Ve <id>" from the switching device
func (base *SLXBase) DeleteInterfaceVe(client *client.NetconfClient, name string) (string, error) {
	var veMap = map[string]interface{}{"name": name}

	config, templateError := base.GetStringFromTemplate(intVeDelete, veMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//CreateBasicInterfacePo is used to create "interface Port-Channel <po-id>" and its sub config on the switching device. This doesnt create switchport config under PO.
func (base *SLXBase) CreateBasicInterfacePo(client *client.NetconfClient, name string,
	speed string, description string) (string, error) {

	var portChannel = map[string]interface{}{"name": name, "speed": speed,
		"description": description}

	//Create port-channel
	config, templateError := base.GetStringFromTemplate(intPortChannelCreate, portChannel)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	//Set Speed
	config, templateError = base.GetStringFromTemplate(intPortChannelSpeedSet, portChannel)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	//Set Description
	config, templateError = base.GetStringFromTemplate(intPortChannelDescriptionSet, portChannel)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(intPortChannelActivate, portChannel)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	return resp, err
}

//CreateInterfacePo is used to create "interface Port-Channel <po-id>" and its sub config on the switching device
func (base *SLXBase) CreateInterfacePo(client *client.NetconfClient, name string,
	speed string, description string, controlVlan string) (string, error) {
	return base.CreateBasicInterfacePo(client, name, speed, description)
}

//ConfigureInterfacePoSpeed is used to configure "speed" on the "interface Port-Channel <po-id>", on the switching device
func (base *SLXBase) ConfigureInterfacePoSpeed(client *client.NetconfClient, name string, speed string) (string, error) {

	var portChannel = map[string]interface{}{"name": name, "speed": speed}
	config, templateError := base.GetStringFromTemplate(intPortChannelDeactivate, portChannel)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(intPortChannelSpeedSet, portChannel)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(intPortChannelActivate, portChannel)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	return resp, err
}

//DeleteInterfacePo is used to delete "interface Port-Channel <po-id>" from the switching device
func (base *SLXBase) DeleteInterfacePo(client *client.NetconfClient, name string) (string, error) {
	var portChannel = map[string]interface{}{"name": name}
	config, templateError := base.GetStringFromTemplate(intPortChannelDelete, portChannel)

	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//GetInterfacePo is used to get running-config of "interface Port-channel <po-id>" from the switching device
func (base *SLXBase) GetInterfacePo(client *client.NetconfClient, name string) (map[string]string, error) {
	ResultMap := make(map[string]string)
	RequestMsg := fmt.Sprintf("/interface/port-channel[name='%s']", name)

	resp, err := client.GetConfig(RequestMsg)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElement("//po-speed"); elem != nil {
		ResultMap["speed"] = elem.Text()
	}
	if elem := doc.FindElement("//description"); elem != nil {
		ResultMap["description"] = elem.Text()
	}
	if elem := doc.FindElement("//shutdown"); elem != nil {
		ResultMap["shutdown"] = "true"
	}
	if elem := doc.FindElement("//mode//vlan-mode"); elem != nil {
		ResultMap["vlan-mode"] = elem.Text()
	}
	if elem := doc.FindElement("//vlan/add"); elem != nil {
		ResultMap["vlan"] = elem.Text()
	}
	return ResultMap, err
}

//GetInterfacePoDetailRequest is used to get port channel details based on the PO ID from the switching device
func (base *SLXBase) GetInterfacePoDetailRequest(name string) string {
	var requestInterface string

	requestInterfaceFmt := `<get-port-channel-detail xmlns="urn:brocade.com:mgmt:brocade-lag">
                              <aggregator-id>%s</aggregator-id>
                              </get-port-channel-detail>
		                      `
	requestInterface = fmt.Sprintf(requestInterfaceFmt, name)

	return requestInterface
}

//GetInterfacePoDetails is used to get port channel details including the member ports from the switching device
func (base *SLXBase) GetInterfacePoDetails(client *client.NetconfClient, name string) (map[string]string, error) {
	ResultMap := make(map[string]string)

	request := base.GetInterfacePoDetailRequest(name)
	resp, err := client.ExecuteRPC(request)

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElements("//aggr-member//interface-name"); elem != nil {
		var intfName string
		for _, intf := range elem {
			intfName = intfName + " " + intf.Text()
		}
		ResultMap["intfs"] = intfName
	}

	if elem := doc.FindElement("//aggregator-mode"); elem != nil {
		ResultMap["aggregator-mode"] = elem.Text()
	}

	if elem := doc.FindElement("//aggregator-type"); elem != nil {
		ResultMap["aggregator-type"] = elem.Text()
	}

	return ResultMap, err
}

package base

import (
	"efa-server/infra/device/client"
	"fmt"
	"github.com/beevik/etree"
)

//ConfigureSystemL2Mtu is used to configure System L2 MTU on the switching device
func (base *SLXBase) ConfigureSystemL2Mtu(client *client.NetconfClient, l2MTUValue string) (string, error) {
	var l2mtuMap = map[string]interface{}{"l2_mtu_value": l2MTUValue}

	config, templateError := base.GetStringFromTemplate(configureSystemWideL2Mtu, l2mtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureSystemL2Mtu is used to unconfigure System L2 MTU on the switching device
func (base *SLXBase) UnconfigureSystemL2Mtu(client *client.NetconfClient) (string, error) {
	var l2mtuMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(unconfigureSystemWideL2Mtu, l2mtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureSystemIPMtu is used to configure System IP MTU on the switching device
func (base *SLXBase) ConfigureSystemIPMtu(client *client.NetconfClient, ipMTUValue string) (string, error) {
	var ipmtuMap = map[string]interface{}{"ip_mtu_value": ipMTUValue}

	config, templateError := base.GetStringFromTemplate(configureSystemWideIPMtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(configureSystemWideIPv6Mtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}
	return resp, err
}

//UnconfigureSystemIPMtu is used to unconfugure System IP MTU from the switching device
func (base *SLXBase) UnconfigureSystemIPMtu(client *client.NetconfClient) (string, error) {
	var ipmtuMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(unconfigureSystemWideIPMtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(unconfigureSystemWideIPv6Mtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}
	return resp, err
}

//GetArp is used to get the running-config of the ARP parameters from the switching device
func (base *SLXBase) GetArp(client *client.NetconfClient) (map[string]string, error) {
	response := make(map[string]string)
	resp, err := client.GetConfig("/host-table")
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//aging-mode/conversational"); elem != nil {
		response["arp-aging-mode-conversational"] = "true"
	}
	if elem := doc.FindElement("//conversational-timeout"); elem != nil {
		response["arp-conversational-timeout"] = elem.Text()
	}
	return response, err
}

//GetMac is used to get the running-config of the MAC parameters from the switching device
func (base *SLXBase) GetMac(client *client.NetconfClient) (map[string]string, error) {
	response := make(map[string]string)
	resp, err := client.GetConfig("/mac-address-table")
	doc := etree.NewDocument()

	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//learning-mode"); elem != nil {
		response["mac-learning-mode"] = elem.Text()
	}
	if elem := doc.FindElement("//conversational-time-out"); elem != nil {
		response["mac-conversational-timeout"] = elem.Text()
	}
	if elem := doc.FindElement("//legacy-time-out"); elem != nil {
		response["mac-aging-timeout"] = elem.Text()
	}
	if elem := doc.FindElement("//mac-move-limit"); elem != nil {
		response["mac-move-limit"] = elem.Text()
	}

	return response, err

}

//GetAnycastGateway is used to get the running-config of AnyCast Gateway from the device
func (base *SLXBase) GetAnycastGateway(client *client.NetconfClient) (map[string]string, error) {

	response := make(map[string]string)
	resp, err := client.GetConfig("/routing-system")
	doc := etree.NewDocument()
	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//ip-anycast-gateway-mac"); elem != nil {
		response["ip-anycast-gateway-mac"] = elem.Text()
	}

	if elem := doc.FindElement("//ipv6-anycast-gateway-mac"); elem != nil {
		response["ipv6-anycast-gateway-mac"] = elem.Text()
	}

	return response, err

}

//ConfigureAnycastGateway is used to configure static anycast-gateway on the switching device
func (base *SLXBase) ConfigureAnycastGateway(client *client.NetconfClient, ipv4AnycastGatewayMac string,
	ipv6AnycastGatewayMac string) (string, error) {

	var anycastMap = map[string]interface{}{"ipv4_anycast_gateway_mac": ipv4AnycastGatewayMac,
		"ipv6_anycast_gateway_mac": ipv6AnycastGatewayMac}

	config, templateError := base.GetStringFromTemplate(ipAnycastGatewayCreate, anycastMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//UnconfigureAnycastGateway is used to unconfigure static anycast-gateway from the switching device
func (base *SLXBase) UnconfigureAnycastGateway(client *client.NetconfClient) (string, error) {
	var anycastMap = map[string]interface{}{}
	config, templateError := base.GetStringFromTemplate(ipAnycastGatewayDelete, anycastMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureRouterID is used to configure "router-id" config on the switching device
func (base *SLXBase) ConfigureRouterID(client *client.NetconfClient, routerID string) (string, error) {
	var routerMap = map[string]interface{}{"router_id": routerID}

	config, templateError := base.GetStringFromTemplate(routerIDCreate, routerMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//UnconfigureRouterID is used to uncofigure "router-id" from the switching device
func (base *SLXBase) UnconfigureRouterID(client *client.NetconfClient) (string, error) {
	var routerMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(routerIDDelete, routerMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//GetRouterID is used to get the running-config of "router-id" from the switching device
func (base *SLXBase) GetRouterID(client *client.NetconfClient) (string, error) {
	rotuerID := ""
	resp, err := client.GetConfig("/ip/rtm-config")

	doc := etree.NewDocument()

	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return rotuerID, err
	}
	if elem := doc.FindElement("//router-id"); elem != nil {
		rotuerID = elem.Text()
	}
	return rotuerID, nil
}

//GetDeviceDetail provides detail of the device like SWBD, Version
func (base *SLXBase) GetDeviceDetail(client *client.NetconfClient) (map[string]string, error) {

	request := `
      <action xmlns="http://tail-f.com/ns/netconf/actions/1.0">
      	<data>
          <show xmlns="urn:brocade.com:mgmt:brocade-common-def">
	    	<infra  xmlns="urn:brocade.com:mgmt:brocade-ras-ext"> 	
				<chassis  xmlns="urn:brocade.com:mgmt:brocade-ras-ext">
            	</chassis> 
            </infra> 
          </show>
      	</data>
      </action>`

	response := make(map[string]string)
	resp, err := client.ExecuteRPC(request)
	if err != nil {
		fmt.Println(err)
		return response, err
	}
	doc := etree.NewDocument()
	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//switch-type"); elem != nil {
		response["switch-type"] = elem.Text()
	}

	request = `<show-firmware-version xmlns="urn:brocade.com:mgmt:brocade-firmware-ext"></show-firmware-version>`

	resp, err = client.ExecuteRPC(request)
	if err != nil {
		fmt.Println(err)
		return response, err
	}
	doc = etree.NewDocument()
	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//firmware-full-version"); elem != nil {
		response["firmware-full-version"] = elem.Text()
	}

	if elem := doc.FindElement("//os-version"); elem != nil {
		response["os-version"] = elem.Text()
	}
	return response, err
}

//GetIPRoutes is used to get the running-config of "overlay-gateway" from the switching device
func (base *SLXBase) GetIPRoutes(client *client.NetconfClient) (map[string]string, error) {

	iprouteMap := make(map[string]string)

	return iprouteMap, nil

}

//ConfigureNumberedRouteMap Route Map for Numbered Interfaces
func (base *SLXBase) ConfigureNumberedRouteMap(client *client.NetconfClient) (string, error) {
	var routeMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(routeMapIPPrefixCreate, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(routeMapPermitCreate, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(routeMapDenyCreate, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(routeMapDenyIPPrefixCreate, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	return resp, err
}

//UnConfigureNumberedRouteMap routemap for numbered interfaces
func (base *SLXBase) UnConfigureNumberedRouteMap(client *client.NetconfClient) (string, error) {
	var routeMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(routeMapIPPrefixDelete, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(routeMapPermitDelete, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(routeMapDenyDelete, routeMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}
	return resp, err
}

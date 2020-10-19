package base

import (
	"bytes"
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
	"fmt"
	"github.com/beevik/etree"
	"text/template"
)

//SLXBase structure for base SLX
type SLXBase struct {
}

//GetStringFromTemplate fetches string from the template
func (base *SLXBase) GetStringFromTemplate(templateName string, dataMap map[string]interface{}) (string, error) {
	t := template.Must(template.New("ovg").Parse(templateName))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, dataMap)
	return tpl.String(), err
}

//CreateOverlayGateway is used to create "overlay-gateway" and its subconfig on the switching device
func (base *SLXBase) CreateOverlayGateway(client *client.NetconfClient, gwName string, gwType string,
	loopbackID string, mapVNIAuto string) (string, error) {
	var ovgMap = map[string]interface{}{"gw_name": gwName, "gw_type": gwType,
		"loopback_id": loopbackID, "map_vni_auto": mapVNIAuto}

	config, templateError := base.GetStringFromTemplate(overlayGatewayCreate, ovgMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err

}

//GetOverlayGateway is used to get the running-config of "overlay-gateway" from the switching device
func (base *SLXBase) GetOverlayGateway(client *client.NetconfClient) (operation.ConfigOVGResponse, error) {
	response := operation.ConfigOVGResponse{}
	resp, err := client.GetConfig("overlay-gateway")
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//name"); elem != nil {
		response.Name = elem.Text()
	}
	if elem := doc.FindElement("//gw-type"); elem != nil {
		response.GwType = elem.Text()
	}
	if elem := doc.FindElement("//loopback-id"); elem != nil {
		response.LoopbackID = elem.Text()
	}
	if elem := doc.FindElement("//auto"); elem != nil {
		response.VNIAuto = "true"
	}
	if elem := doc.FindElement("//activate"); elem != nil {
		response.Activate = "true"
	}

	return response, err

}

//DeleteOverlayGateway is used to delete the "overlay-gateway" from the switching device
func (base *SLXBase) DeleteOverlayGateway(client *client.NetconfClient, gwName string) (string, error) {

	var ovgMap = map[string]interface{}{"gw_name": gwName}

	config, templateError := base.GetStringFromTemplate(overlayGatewayDelete, ovgMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err

}

//CreateEvpnInstance is used to create "evpn" instance on the switching device
func (base *SLXBase) CreateEvpnInstance(client *client.NetconfClient, eviName string, duplicateMacTimer string,
	duplicateMacTimerMaxCount string) (string, error) {

	var evpnMap = map[string]interface{}{"evi_name": eviName, "duplicate_mac_timer": duplicateMacTimer,
		"duplicate_mac_timer_max_count": duplicateMacTimerMaxCount}

	config, templateError := base.GetStringFromTemplate(evpnInstanceCreate, evpnMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//DeleteEvpnInstance is used to delete "evpn" instance from the switching device
func (base *SLXBase) DeleteEvpnInstance(client *client.NetconfClient, eviName string) (string, error) {

	var evpnMap = map[string]interface{}{"evi_name": eviName}

	config, templateError := base.GetStringFromTemplate(evpnInstanceDelete, evpnMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//GetEvpnInstance is used to get the running-config of "evpn" instance from the switching device
func (base *SLXBase) GetEvpnInstance(client *client.NetconfClient) (operation.ConfigEVPNRespone, error) {
	response := operation.ConfigEVPNRespone{}
	resp, err := client.GetConfig("/routing-system/evpn-config")
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return response, err
	}

	if elem := doc.FindElement("//instance-name"); elem != nil {
		response.Name = elem.Text()
	}
	if elem := doc.FindElement("//duplicate-mac-timer-value"); elem != nil {
		response.DuplicageMacTimerValue = elem.Text()
	}
	if elem := doc.FindElement("//max-count"); elem != nil {
		response.MaxCount = elem.Text()
	}
	if elem := doc.FindElement("//target-community"); elem != nil {
		response.TargetCommunity = elem.Text()
	}
	if elem := doc.FindElement("//route-target/both"); elem != nil {
		response.RouteTargetBoth = "true"
	}
	if elem := doc.FindElement("//ignore-as"); elem != nil {
		response.IgnoreAs = "true"
	}

	return response, err

}

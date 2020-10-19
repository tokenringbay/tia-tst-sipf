package base

import (
	"efa-server/infra/device/adapter/platform/slx/base"
	"efa-server/infra/device/client"
	"fmt"
	"github.com/beevik/etree"
)

//SLXSwitchingBase structure for switching base SLX
type SLXSwitchingBase struct {
	base.SLXBase
}

//ConfigureMacAndArp is used to configure MAC and ARP parameters on the switching device
func (base *SLXSwitchingBase) ConfigureMacAndArp(client *client.NetconfClient, arpAgingTimeout string,
	macAgingTimeout string, macAgingConversationalTimeout string, macMoveLimit string) (string, error) {

	var macArpMap = map[string]interface{}{"arp_aging_timeout": arpAgingTimeout, "mac_aging_timeout": macAgingTimeout,
		"mac_aging_conversational_timeout": macAgingConversationalTimeout, "mac_move_limit": macMoveLimit}

	config, templateError := base.GetStringFromTemplate(macMoveDetectCreate, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(conversationPropertyCreate, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	var legMap = map[string]interface{}{"mac_aging_timeout": macAgingTimeout}

	config, templateError = base.GetStringFromTemplate(configureLegacyMacTimeout, legMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	return resp, err
}

//UnconfigureMacAndArp is used to unconfigure MAC and ARP parameters from the switching device
func (base *SLXSwitchingBase) UnconfigureMacAndArp(client *client.NetconfClient) (string, error) {

	var macArpMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(macConfigDelete, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(macMoveDetectDelete, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(conversationMacDelete, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(conversationArpDelete, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)

	return resp, err
}

//ConfigureSwitchHostName is used to configure Host name on the switching device
func (base *SLXSwitchingBase) ConfigureSwitchHostName(client *client.NetconfClient, hostName string) (string, error) {
	var hostNameMap = map[string]interface{}{"host_name": hostName}

	config, templateError := base.GetStringFromTemplate(configureSwitchHostName, hostNameMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureSwitchHostName is used to unconfigure Host name on the switching device
func (base *SLXSwitchingBase) UnconfigureSwitchHostName(client *client.NetconfClient) (string, error) {
	var hostNameMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(unconfigureSwitchHostName, hostNameMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//GetSwitchHostName is used to get the running-config of "Host name" from the switching device
func (base *SLXSwitchingBase) GetSwitchHostName(client *client.NetconfClient) (string, error) {
	hostName := ""
	resp, err := client.GetConfig("/system/switch-attributes")

	doc := etree.NewDocument()

	if err = doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return hostName, err
	}
	if elem := doc.FindElement("//host-name"); elem != nil {
		hostName = elem.Text()
	}
	return hostName, nil
}

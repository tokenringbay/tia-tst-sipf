package base

import (
	"efa-server/infra/device/adapter/platform/slx/base"
	"efa-server/infra/device/client"
	"fmt"
	"github.com/beevik/etree"
)

//SLXRoutingBase structure for routing base SLX
type SLXRoutingBase struct {
	base.SLXBase
}

//ConfigureMacAndArp is used to configure MAC and ARP parameters on the switching device
func (base *SLXRoutingBase) ConfigureMacAndArp(client *client.NetconfClient, arpAgingTimeout string,
	macAgingTimeout string, macAgingConversationalTimeout string, macMoveLimit string) (string, error) {

	var legMap = map[string]interface{}{"mac_aging_timeout": macAgingTimeout}

	config, templateError := base.GetStringFromTemplate(configureLegacyMacTimeout, legMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//UnconfigureMacAndArp is used to unconfigure MAC and ARP parameters from the switching device
func (base *SLXRoutingBase) UnconfigureMacAndArp(client *client.NetconfClient) (string, error) {

	var macArpMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(macConfigDelete, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	return resp, err
}

//ConfigureSwitchHostName is used to configure Host name on the switching device
func (base *SLXRoutingBase) ConfigureSwitchHostName(client *client.NetconfClient, hostName string) (string, error) {
	var hostNameMap = map[string]interface{}{"host_name": hostName}

	config, templateError := base.GetStringFromTemplate(configureSwitchHostName, hostNameMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureSwitchHostName is used to unconfigure Host name on the switching device
func (base *SLXRoutingBase) UnconfigureSwitchHostName(client *client.NetconfClient) (string, error) {
	var hostNameMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(unconfigureSwitchHostName, hostNameMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//GetSwitchHostName is used to get the running-config of "Host name" from the switching device
func (base *SLXRoutingBase) GetSwitchHostName(client *client.NetconfClient) (string, error) {
	hostName := ""
	resp, err := client.GetConfig("/system-ras/switch-attributes")

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

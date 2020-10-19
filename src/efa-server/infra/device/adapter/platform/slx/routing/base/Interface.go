package base

import (
	"efa-server/infra/device/client"
)

//CreateInterfacePo is used to create "interface Port-Channel <po-id>" and its sub config on the switching device
func (base *SLXRoutingBase) CreateInterfacePo(client *client.NetconfClient, name string,
	speed string, description string, controlVlan string) (string, error) {

	resp, err := base.CreateBasicInterfacePo(client, name, speed, description)

	if err != nil {
		return "", err
	}
	var portChannel = map[string]interface{}{"name": name, "vlan": controlVlan}

	//Add Switchport config to port-channel
	config, templateError := base.GetStringFromTemplate(intPortChannelAddSwitchPortBasic, portChannel)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(intPortChannelSwitchPortVlanMode, portChannel)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(intPortChannelSwitchPortAddVlan, portChannel)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	return resp, err
}

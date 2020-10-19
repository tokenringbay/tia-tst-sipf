package fv18r200

import (
	"efa-server/infra/device/adapter/platform/slx/routing/avalanche/base"
	"efa-server/infra/device/client"
)

//SLXAvalancheFV18R200 structure for cedar FV18200 SLX
type SLXAvalancheFV18R200 struct {
	base.SLXAvalancheBase
}

//ConfigureSystemIPMtu is used to configure System IP MTU on the switching device
func (base *SLXAvalancheFV18R200) ConfigureSystemIPMtu(client *client.NetconfClient, ipMTUValue string) (string, error) {
	var ipmtuMap = map[string]interface{}{"ip_mtu_value": ipMTUValue}

	config, templateError := base.GetStringFromTemplate(configureSystemWideIPMtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureSystemIPMtu is used to unconfugure System IP MTU from the switching device
func (base *SLXAvalancheFV18R200) UnconfigureSystemIPMtu(client *client.NetconfClient) (string, error) {
	var ipmtuMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(unconfigureSystemWideIPMtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

package base

import "efa-server/infra/device/client"

//ConfigureMacAndArp is used to configure MAC and ARP parameters on the switching device
func (base *SLXOrcaBase) ConfigureMacAndArp(client *client.NetconfClient, arpAgingTimeout string,
	macAgingTimeout string, macAgingConversationalTimeout string, macMoveLimit string) (string, error) {

	var macArpMap = map[string]interface{}{"arp_aging_timeout": arpAgingTimeout, "mac_aging_timeout": macAgingTimeout}

	config, templateError := base.GetStringFromTemplate(conversationPropertyCreate, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
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
func (base *SLXOrcaBase) UnconfigureMacAndArp(client *client.NetconfClient) (string, error) {

	var macArpMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(macConfigDelete, macArpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
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

//ConfigureSystemIPMtu is used to configure System IP MTU on the switching device
func (base *SLXOrcaBase) ConfigureSystemIPMtu(client *client.NetconfClient, ipMTUValue string) (string, error) {
	var ipmtuMap = map[string]interface{}{"ip_mtu_value": ipMTUValue}

	config, templateError := base.GetStringFromTemplate(configureSystemWideIPMtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

//UnconfigureSystemIPMtu is used to unconfugure System IP MTU from the switching device
func (base *SLXOrcaBase) UnconfigureSystemIPMtu(client *client.NetconfClient) (string, error) {
	var ipmtuMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(unconfigureSystemWideIPMtu, ipmtuMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	return resp, err
}

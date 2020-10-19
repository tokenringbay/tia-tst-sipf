package base

import "efa-server/infra/device/client"

//ConfigureInterfaceVe is used to configure "interface Ve" and its sub config, on the switching device
func (base *SLXOrcaBase) ConfigureInterfaceVe(client *client.NetconfClient, name string,
	ipAddress string, bfdrx string, bfdtx string, bfdmultiplier string) (string, error) {
	var veMap = map[string]interface{}{"name": name, "ip_address": ipAddress,
		"bfd_min_tx": bfdtx, "bfd_min_rx": bfdrx, "bfd_multiplier": bfdmultiplier}
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

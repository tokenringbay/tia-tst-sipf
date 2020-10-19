package base

import "efa-server/infra/device/client"

import "efa-server/infra/device/adapter/platform/slx/switching/base"

//SLXCedarBase structure for cedar base SLX
type SLXCedarBase struct {
	base.SLXSwitchingBase
}

//ConfigureAnycastGateway is used to configure static anycast-gateway on the switching device
func (base *SLXCedarBase) ConfigureAnycastGateway(client *client.NetconfClient, ipv4AnycastGatewayMac string,
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
func (base *SLXCedarBase) UnconfigureAnycastGateway(client *client.NetconfClient) (string, error) {
	var anycastMap = map[string]interface{}{}
	config, templateError := base.GetStringFromTemplate(ipAnycastGatewayDelete, anycastMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

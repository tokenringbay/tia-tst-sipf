package freedom

import (
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/actions/deconfigurefabric"

	"efa-server/domain/operation"
	"testing"
	//"fmt"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"fmt"
	"github.com/stretchr/testify/assert"
)

//Test EVPN Create and verify using NetConf
func TestEVPN(t *testing.T) {
	//Open up the Client and set other attributes

	ctx, fabricGate, fabricErrors, Errors, client := initializeTest()

	detail, err := ad.GetDeviceDetail(client)
	assert.Contains(t, detail.Model, Model)
	adapter := ad.GetAdapter(detail.Model)

	//cleanup EVPN before testing
	cleanEVPN(client)
	//defer client.Close()

	sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
		ArpAgingTimeout: arpAgingTimeout, MacAgingTimeout: macAgingTimeout, MacAgingConversationalTimeout: macAgingConversationalTimeout,
		MacMoveLimit: macMoveLimit, DuplicateMacTimer: duplicateMacTimer, DuplicateMaxTimerMaxCount: duplicateMacTimerMaxCount, Model: detail.Model}

	//Call the Actions
	configurefabric.ConfigureEvpn(ctx, fabricGate, &sw, false, fabricErrors)
	//Setup for cleanup
	defer cleanEVPN(client)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	//Fetch EVPN details using NetConf
	evpnResponse, err := adapter.GetEvpnInstance(client)

	assert.Nil(t, err)

	assert.Equal(t, operation.ConfigEVPNRespone{Name: sw.Fabric, DuplicageMacTimerValue: duplicateMacTimer,
		MaxCount: "5", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, evpnResponse)

	//Fetch  ARP
	arpMap, err := adapter.GetArp(client)
	assert.Equal(t, map[string]string{"arp-aging-mode-conversational": "true", "arp-conversational-timeout": arpAgingTimeout}, arpMap)

	macMap, err := adapter.GetMac(client)
	//fmt.Println(macMap)

	assert.Equal(t, map[string]string{"mac-conversational-timeout": macAgingConversationalTimeout, "mac-aging-timeout": macAgingTimeout,
		"mac-move-limit": macMoveLimit, "mac-learning-mode": "conversational"}, macMap)

}

//Test EVPN Force Create and verify using NetConf
func TestEVPNNameMismatch(t *testing.T) {
	//Open up the Client and set other attributes
	ctx, fabricGate, fabricErrors, Errors, client := initializeTest()
	detail, _ := ad.GetDeviceDetail(client)
	assert.Contains(t, detail.Model, Model)
	adapter := ad.GetAdapter(detail.Model)
	//cleanup EVPN before testing
	cleanEVPN(client)

	adapter.CreateEvpnInstance(client, "efa-1", duplicateMacTimer, duplicateMacTimerMaxCount)

	sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
		ArpAgingTimeout: arpAgingTimeout, MacAgingTimeout: macAgingTimeout, MacAgingConversationalTimeout: macAgingConversationalTimeout,
		MacMoveLimit: macMoveLimit, DuplicateMacTimer: duplicateMacTimer, DuplicateMaxTimerMaxCount: duplicateMacTimerMaxCount, Model: detail.Model}

	//Call the Actions
	configurefabric.ConfigureEvpn(ctx, fabricGate, &sw, false, fabricErrors)
	//Setup for cleanup
	defer cleanEVPN(client)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

	//assert.Empty(t,Errors)
	assert.Equal(t, "EVPN efa-1 already configured on switch", fmt.Sprint(Errors[0].Error))

}

//Test EVPN Force Create and verify using NetConf
func TestEVPNForce(t *testing.T) {
	//Open up the Client and set other attributes
	ctx, fabricGate, fabricErrors, Errors, client := initializeTest()
	detail, _ := ad.GetDeviceDetail(client)
	assert.Contains(t, detail.Model, Model)
	adapter := ad.GetAdapter(detail.Model)
	//cleanup EVPN before testing
	cleanEVPN(client)

	adapter.CreateEvpnInstance(client, "efa-1", duplicateMacTimer, duplicateMacTimerMaxCount)

	sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
		ArpAgingTimeout: arpAgingTimeout, MacAgingTimeout: macAgingTimeout, MacAgingConversationalTimeout: macAgingConversationalTimeout,
		MacMoveLimit: macMoveLimit, DuplicateMacTimer: duplicateMacTimer, DuplicateMaxTimerMaxCount: duplicateMacTimerMaxCount, Model: detail.Model}

	//Call the Actions
	configurefabric.ConfigureEvpn(ctx, fabricGate, &sw, true, fabricErrors)
	//Setup for cleanup
	defer cleanEVPN(client)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	//Fetch EVPN details using NetConf
	evpnResponse, err := adapter.GetEvpnInstance(client)

	assert.Nil(t, err)

	assert.Equal(t, operation.ConfigEVPNRespone{Name: sw.Fabric, DuplicageMacTimerValue: duplicateMacTimer,
		MaxCount: "5", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, evpnResponse)

	//Fetch  ARP
	arpMap, err := adapter.GetArp(client)
	assert.Equal(t, map[string]string{"arp-aging-mode-conversational": "true", "arp-conversational-timeout": arpAgingTimeout}, arpMap)

	macMap, err := adapter.GetMac(client)
	//fmt.Println(macMap)

	assert.Equal(t, map[string]string{"mac-conversational-timeout": macAgingConversationalTimeout, "mac-aging-timeout": macAgingTimeout,
		"mac-move-limit": macMoveLimit, "mac-learning-mode": "conversational"}, macMap)

}

func cleanEVPN(client *netconf.NetconfClient) (string, error) {
	detail, _ := ad.GetDeviceDetail(client)

	adapter := ad.GetAdapter(detail.Model)
	evpnRespone, _ := adapter.GetEvpnInstance(client)
	evpnOnSwitch := evpnRespone.Name
	if evpnOnSwitch != "" {
		return adapter.DeleteEvpnInstance(client, evpnOnSwitch)
	}
	return "", nil
}

func TestDeleteEVPN(t *testing.T) {
	//Open up the Client and set other attributes
	ctx, fabricGate, fabricErrors, Errors, client := initializeTest()
	detail, _ := ad.GetDeviceDetail(client)
	assert.Contains(t, detail.Model, Model)
	adapter := ad.GetAdapter(detail.Model)
	//cleanup EVPN before testing
	cleanEVPN(client)
	//defer client.Close()

	sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
		ArpAgingTimeout: arpAgingTimeout, MacAgingTimeout: macAgingTimeout, MacAgingConversationalTimeout: macAgingConversationalTimeout,
		MacMoveLimit: macMoveLimit, DuplicateMacTimer: duplicateMacTimer, DuplicateMaxTimerMaxCount: duplicateMacTimerMaxCount, Model: detail.Model}

	//Call the Actions
	configurefabric.ConfigureEvpn(ctx, fabricGate, &sw, false, fabricErrors)
	//Setup for cleanup

	//Fetch EVPN details using NetConf
	evpnResponse, err := adapter.GetEvpnInstance(client)

	assert.Nil(t, err)

	assert.Equal(t, operation.ConfigEVPNRespone{Name: sw.Fabric, DuplicageMacTimerValue: duplicateMacTimer,
		MaxCount: "5", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, evpnResponse)

	//Fetch  ARP
	arpMap, err := adapter.GetArp(client)
	assert.Equal(t, map[string]string{"arp-aging-mode-conversational": "true", "arp-conversational-timeout": arpAgingTimeout}, arpMap)

	macMap, err := adapter.GetMac(client)
	//fmt.Println(macMap)

	assert.Equal(t, map[string]string{"mac-conversational-timeout": macAgingConversationalTimeout, "mac-aging-timeout": macAgingTimeout,
		"mac-move-limit": macMoveLimit, "mac-learning-mode": "conversational"}, macMap)

	//delete EVPN Instance
	fabricGate.Add(1)
	deconfigurefabric.UnconfigureEvpn(ctx, fabricGate, &sw, fabricErrors)

	//Fetch EVPN details using NetConf
	evpnResponse, err = adapter.GetEvpnInstance(client)

	assert.Nil(t, err)

	assert.Equal(t, operation.ConfigEVPNRespone{Name: "", DuplicageMacTimerValue: "",
		MaxCount: "", TargetCommunity: "", RouteTargetBoth: "", IgnoreAs: ""}, evpnResponse)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

}

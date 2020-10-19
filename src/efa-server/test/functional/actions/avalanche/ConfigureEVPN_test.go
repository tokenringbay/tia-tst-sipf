package avalanche

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
	"os"
)

//Test EVPN Create and verify using NetConf
func TestEVPN(t *testing.T) {
	//Open up the Client and set other attributes
	if os.Getenv("SKIP_AV") == "1" {
		fmt.Println("Skipped")
		t.Skip()
	}
	detail, err := ad.GetDeviceDetail(client)
	assert.Contains(t, detail.Model, Model)
	adapter := ad.GetAdapter(detail.Model)

	ctx, fabricGate, fabricErrors, Errors, client := initializeTest()
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

	macMap, err := adapter.GetMac(client)
	//fmt.Println(macMap)

	assert.Equal(t, map[string]string{"mac-aging-timeout": macAgingTimeout, "mac-conversational-timeout": "300"}, macMap)

}

//Test EVPN Force Create and verify using NetConf
func TestEVPNNameMismatch(t *testing.T) {
	if os.Getenv("SKIP_AV") == "1" {
		fmt.Println("Skipped")
		t.Skip()
	}
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
	if os.Getenv("SKIP_AV") == "1" {
		fmt.Println("Skipped")
		t.Skip()
	}
	//Open up the Client and set other attributes
	ctx, fabricGate, fabricErrors, Errors, client := initializeTest()
	detail, err := ad.GetDeviceDetail(client)
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

	macMap, err := adapter.GetMac(client)
	//fmt.Println(macMap)

	assert.Equal(t, map[string]string{"mac-aging-timeout": macAgingTimeout, "mac-conversational-timeout": "300"}, macMap)

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
	if os.Getenv("SKIP_AV") == "1" {
		fmt.Println("Skipped")
		t.Skip()
	}
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

	//Fetch EVPN details using NetConf
	evpnResponse, err := adapter.GetEvpnInstance(client)

	assert.Nil(t, err)

	assert.Equal(t, operation.ConfigEVPNRespone{Name: sw.Fabric, DuplicageMacTimerValue: duplicateMacTimer,
		MaxCount: "5", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, evpnResponse)

	macMap, err := adapter.GetMac(client)
	//fmt.Println(macMap)

	assert.Equal(t, map[string]string{"mac-aging-timeout": macAgingTimeout, "mac-conversational-timeout": "300"}, macMap)

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

package common

import (
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/actions/deconfigurefabric"

	"testing"

	"efa-server/domain/operation"
	//"fmt"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/adapter/interface"
	netconf "efa-server/infra/device/client"
	_ "efa-server/test/functional"
	"fmt"
	"github.com/stretchr/testify/assert"
)

var VNIAUTOMAP = true
var IPV4MAC = "0201.0101.0101"
var IPV6MAC = "0201.0101.0102"

//Test Overlay Gateway Create and verify using NetConf
func TestConfigureOverlay_Gateway(t *testing.T) {

	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			//cleanup Overlay Gateway before testing
			cleanOverlayGateway(client, Model)
			//defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				VlanVniAutoMap: VNIAUTOMAP, VtepLoopbackPortNumber: vteploopbackPortnumber,
				AnycastMac: IPV4MAC, IPV6AnycastMac: IPV6MAC, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureOverlayGateway(ctx, fabricGate, &sw, false, fabricErrors)
			//Setup for cleanup
			defer cleanOverlayGateway(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Fetch Overlay Gateway details using NetConf
			ovgResponse, err := adapter.GetOverlayGateway(client)

			assert.Nil(t, err)

			assert.Equal(t, operation.ConfigOVGResponse{Name: FabricName, GwType: "layer2-extension",
				LoopbackID: vteploopbackPortnumber, VNIAuto: "true", Activate: "true"}, ovgResponse)

		})
	}
}

//Test Overlay Gateway Create when its already configured with a different name
func TestConfigureOverlay_GatewayNameNotMatching(t *testing.T) {

	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			//cleanup Overlay Gateway before testing
			cleanOverlayGateway(client, Model)
			detail, _ := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			adapter.CreateOverlayGateway(client, "efa-1", "layer2-extension",
				vteploopbackPortnumber, "true")
			//defer client.Close()

			//Call the Actions
			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				VlanVniAutoMap: VNIAUTOMAP, VtepLoopbackPortNumber: vteploopbackPortnumber,
				AnycastMac: IPV4MAC, IPV6AnycastMac: IPV6MAC, Model: detail.Model}

			configurefabric.ConfigureOverlayGateway(ctx, fabricGate, &sw, false, fabricErrors)

			//Setup for cleanup
			defer cleanOverlayGateway(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Equal(t, "Overlay gateway efa-1 already configured on switch", fmt.Sprint(Errors[0].Error))

		})
	}
}

//Test Overlay Gateway Create when its already configured with a same name
func TestConfigureOverlay_GatewayNameMatching(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		client = &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
		client.Login()
		defer client.Close()

		t.Run(name, func(t *testing.T) {
			//Open up the Client and set other attributes
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			//cleanup Overlay Gateway before testing
			cleanOverlayGateway(client, Model)
			detail, _ := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			adapter.CreateOverlayGateway(client, FabricName, "layer2-extension",
				vteploopbackPortnumber, "true")
			//defer client.Close()

			//Call the Actions
			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				VlanVniAutoMap: VNIAUTOMAP, VtepLoopbackPortNumber: vteploopbackPortnumber,
				AnycastMac: IPV4MAC, IPV6AnycastMac: IPV6MAC, Model: detail.Model}

			configurefabric.ConfigureOverlayGateway(ctx, fabricGate, &sw, false, fabricErrors)

			//Setup for cleanup
			defer cleanOverlayGateway(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Fetch Overlay Gateway details using NetConf
			ovgResponse, err := adapter.GetOverlayGateway(client)

			assert.Nil(t, err)

			assert.Equal(t, operation.ConfigOVGResponse{Name: sw.Fabric, GwType: "layer2-extension",
				LoopbackID: vteploopbackPortnumber, VNIAuto: "true", Activate: "true"}, ovgResponse)

		})
	}
}

//Test Overlay Gateway Create when its already configured with a different name and then use force option
func TestConfigureOverlay_GatewayForce(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			//cleanup Overlay Gateway before testing
			cleanOverlayGateway(client, Model)
			detail, _ := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			adapter.CreateOverlayGateway(client, "efa-1", "layer2-extension",
				vteploopbackPortnumber, "true")
			//defer client.Close()

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				VlanVniAutoMap: VNIAUTOMAP, VtepLoopbackPortNumber: vteploopbackPortnumber,
				AnycastMac: IPV4MAC, IPV6AnycastMac: IPV6MAC, Model: detail.Model}

			sw.Fabric = FabricName
			configurefabric.ConfigureOverlayGateway(ctx, fabricGate, &sw, true, fabricErrors)

			//Setup for cleanup
			defer cleanOverlayGateway(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Fetch Overlay Gateway details using NetConf
			ovgResponse, err := adapter.GetOverlayGateway(client)

			assert.Nil(t, err)
			assert.Equal(t, operation.ConfigOVGResponse{Name: sw.Fabric, GwType: "layer2-extension",
				LoopbackID: vteploopbackPortnumber, VNIAuto: "true", Activate: "true"}, ovgResponse)

		})
	}
}

func cleanupCluster(client *netconf.NetconfClient, adapter interfaces.Switch) {
	//cleanup existing cluster
	exclusterName, exclusterID, _, _, _, _, _ := adapter.GetCluster(client)

	if exclusterName != "" {
		adapter.DeleteCluster(client, exclusterName, exclusterID)
	}
}

func cleanOverlayGateway(client *netconf.NetconfClient, Model string) (string, error) {
	detail, _ := ad.GetDeviceDetail(client)
	adapter := ad.GetAdapter(detail.Model)
	cleanupCluster(client, adapter)

	adapter.UnconfigureAnycastGateway(client)

	ovgResponse, _ := adapter.GetOverlayGateway(client)
	gwOnSwitch := ovgResponse.Name
	if gwOnSwitch != "" {
		return adapter.DeleteOverlayGateway(client, gwOnSwitch)
	}

	return "", nil
}

//Test Overlay Gateway Create and verify using NetConf
func TestDelete_Gateway(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			//cleanup Overlay Gateway before testing
			cleanOverlayGateway(client, Model)
			//defer client.Close()
			detail, _ := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				VlanVniAutoMap: VNIAUTOMAP, VtepLoopbackPortNumber: vteploopbackPortnumber,
				AnycastMac: IPV4MAC, IPV6AnycastMac: IPV6MAC, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureOverlayGateway(ctx, fabricGate, &sw, false, fabricErrors)

			//Fetch Overlay Gateway details using NetConf
			ovgResponse, err := adapter.GetOverlayGateway(client)

			assert.Nil(t, err)

			assert.Equal(t, operation.ConfigOVGResponse{Name: FabricName, GwType: "layer2-extension",
				LoopbackID: vteploopbackPortnumber, VNIAuto: "true", Activate: "true"}, ovgResponse)

			fabricGate.Add(1)
			// Trigger Cleanup
			deconfigurefabric.UnConfigureOverlayGateway(ctx, fabricGate, &sw, false, fabricErrors)

			ovgResponse, err = adapter.GetOverlayGateway(client)
			assert.Equal(t, operation.ConfigOVGResponse{Name: "", GwType: "",
				LoopbackID: "", VNIAuto: "", Activate: ""}, ovgResponse)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)
		})
	}
}

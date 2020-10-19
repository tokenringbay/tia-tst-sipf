package common

import (
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions/fetchfabric"
	"github.com/stretchr/testify/assert"
	"testing"

	actions "efa-server/infra/device/actions/configurefabric"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"efa-server/usecase"
	"fmt"
)

func TestFetchSwitchResponse(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			//------------------- EVPN
			//Open up the Client and set other attributes
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, _ := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			ctx, fabricGate, fabricErrors, Errors := initializeTest()

			//cleanup EVPN before testing
			cleanEVPN(client, detail.Model)

			//Configure EVPN
			swEVPN := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				ArpAgingTimeout: arpAgingTimeout, MacAgingTimeout: macAgingTimeout, MacAgingConversationalTimeout: macAgingConversationalTimeout,
				MacMoveLimit: macMoveLimit, DuplicateMacTimer: duplicateMacTimer, DuplicateMaxTimerMaxCount: duplicateMacTimerMaxCount,
				Model: detail.Model}

			//Call the Actions
			actions.ConfigureEvpn(ctx, fabricGate, &swEVPN, false, fabricErrors)
			//Setup for cleanup
			defer cleanEVPN(client, detail.Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//------------------- OVG
			//Configure OVERlay Gateway
			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			//cleanup Overlay Gateway before testing
			cleanOverlayGateway(client, Model)
			adapter.CreateOverlayGateway(client, FabricName, "layer2-extension",
				vteploopbackPortnumber, "true")
			//defer client.Close()

			//Call the Actions
			swOVG := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				VlanVniAutoMap: VNIAUTOMAP, VtepLoopbackPortNumber: vteploopbackPortnumber,
				AnycastMac: IPV4MAC, IPV6AnycastMac: IPV6MAC, Model: detail.Model}

			actions.ConfigureOverlayGateway(ctx, fabricGate, &swOVG, false, fabricErrors)

			//Setup for cleanup
			defer cleanOverlayGateway(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//------------------- BGP
			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			//cleanup EVPN before testing
			cleanBGP(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber, InterfaceType: domain.IntfTypeLoopback,
				IP: bgpNetwork, ConfigType: domain.ConfigCreate})

			Neighbors := make([]operation.ConfigBgpNeighbor, 0)
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor1Add, RemoteAs: int64(bgpNeighbor1Asn),
				ConfigType: domain.ConfigCreate})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor2Add, RemoteAs: int64(bgpNeighbor2Asn),
				ConfigType: domain.ConfigCreate})

			swBGP := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				Role: bgpSw1Role, BgpLocalAsn: bgpSw2Asn, PeerGroup: bgpLeafPeerGroup, PeerGroupDescription: bgpLeafPeerGroupDesc,
				MaxPaths: bgpMaxPaths, Network: bgpNetwork, ConfigureOverlayGateway: bgpEvpnEnabled, AllowasIn: bgpAllowasIn,
				LoopbackPortNumber: bgpLoopbackNumber, Interfaces: Interfaces, BgpNeighbors: Neighbors,
				BfdTx: BFDTx, BfdRx: BFDRx, BfdMultiplier: BFDMultiplier, Model: detail.Model, BFDEnable: "Yes"}

			//Call the Actions
			actions.ConfigureBGP(ctx, fabricGate, &swBGP, false, fabricErrors)
			//Setup for cleanup
			defer cleanBGP(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Fetch Switch

			ctx, fabricGate, switchResponsChannel, fabricErrors, client := initializeFabricTest()

			switchIdentity := operation.SwitchIdentity{Host: Host, UserName: UserName, Password: Password, Role: usecase.LeafRole, Model: detail.Model}

			//Call the Actions
			fetchfabric.FetchSwitchConfig(ctx, fabricGate, switchIdentity, switchResponsChannel, fabricErrors)

			switchResponses, Errors := closeFetchActionChannel(fabricGate, fabricErrors, switchResponsChannel)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			assert.Equal(t, 1, len(switchResponses))
			switchResponse := switchResponses[0]
			fmt.Println(switchResponse)

			//Verify EVPN
			assert.Equal(t, operation.ConfigEVPNRespone{Name: swEVPN.Fabric, DuplicageMacTimerValue: duplicateMacTimer,
				MaxCount: "5", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, *switchResponse.Evpn)

			//Verify OVG
			assert.Equal(t, operation.ConfigOVGResponse{Name: swOVG.Fabric, GwType: "layer2-extension",
				LoopbackID: vteploopbackPortnumber, VNIAuto: "true", Activate: "true"}, *switchResponse.Ovg)

			//Verify BGP
			bgpResponse := switchResponse.Bgp
			assert.Equal(t, bgpSw2Asn, bgpResponse.LocalAS)
			assert.Equal(t, bgpNetwork, bgpResponse.Network)
			assert.Equal(t, bgpMaxPaths, bgpResponse.MaxPaths)

			assert.Equal(t, 1, len(bgpResponse.PeerGroups))
			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.PeerGroups[0].Name)
			assert.Equal(t, bgpLeafPeerGroupDesc, bgpResponse.PeerGroups[0].Description)
			assert.Equal(t, "true", bgpResponse.PeerGroups[0].BFD)

			assert.Equal(t, 2, len(bgpResponse.Neighbors))
			assert.Equal(t, bgpNeighbor1Add, bgpResponse.Neighbors[0].RemoteIP)

			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.Neighbors[0].PeerGroup)
			assert.Equal(t, bgpNeighbor2Add, bgpResponse.Neighbors[1].RemoteIP)

			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.Neighbors[0].PeerGroup)

			assert.Equal(t, 1, len(bgpResponse.L2VPN.Neighbors))
			assert.Equal(t, "vxlan", bgpResponse.L2VPN.Neighbors[0].Encapsulation)
			assert.Equal(t, bgpAllowasIn, bgpResponse.L2VPN.Neighbors[0].AllowASIn)
			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.L2VPN.Neighbors[0].PeerGroup)
			assert.Equal(t, "true", bgpResponse.L2VPN.Neighbors[0].Activate)

			assert.Equal(t, bgpNetwork, switchResponse.RouterID+"/32")

		})
	}
}

func cleanEVPN(client *netconf.NetconfClient, Model string) (string, error) {
	adapter := ad.GetAdapter(Model)
	evpnRespone, _ := adapter.GetEvpnInstance(client)
	evpnOnSwitch := evpnRespone.Name
	if evpnOnSwitch != "" {
		return adapter.DeleteEvpnInstance(client, evpnOnSwitch)
	}
	return "", nil
}

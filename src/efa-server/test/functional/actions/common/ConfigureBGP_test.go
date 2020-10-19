package common

import (
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/actions/deconfigurefabric"

	"testing"

	"efa-server/domain"
	"efa-server/domain/operation"
	//"fmt"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"github.com/stretchr/testify/assert"

	"efa-server/test/functional"
	"efa-server/usecase"
	"fmt"
	"os"
)

var BFDRx = "200"
var BFDTx = "100"
var BFDMultiplier = "3"
var bgpSw1Role = usecase.LeafRole
var bgpSw2Asn = "65000"
var bgpSw2AsnNew = "65001"
var bgpLeafPeerGroup = "spine-group"
var bgpLeafPeerGroupDesc = "To Spine"
var bgpMaxPaths = "8"
var bgpMaxPathsNew = "4"
var bgpNetwork = "4.4.4.4/32"
var bgpNetworkNew = "4.4.3.4/32"
var bgpEvpnEnabled = "Yes"
var bgpAllowasIn = "1"
var bgpAllowasInNew = "2"
var bgpLoopbackNumber = "1"
var bgpNeighbor1Add = "31.31.31.31"
var bgpNeighbor1Asn = 67890
var bgpNeighbor2Add = "31.31.31.35"
var bgpNeighbor2Asn = 67891

var bgpNeighbor1AddNew = "3.3.4.3"
var bgpNeighbor1AsnNew = 67894

type platform struct {
	IP    string
	Model string
}

var platforms = map[string]platform{

	"Freedom": {functional.NetConfFreedomIP, ad.FreedomType},
	"Cedar":   {functional.NetConfCedarIP, ad.CedarType},
}

func init() {
	if os.Getenv("SKIP_AV") != "1" {
		platforms["Avalanche"] = platform{functional.NetConfAvalancheIP, ad.AvalancheType}
	}
	if os.Getenv("SKIP_OR") != "1" {
		platforms["Orca"] = platform{functional.NetConfAvalancheIP, ad.AvalancheType}
	}
}

func TestBGP_LeafCreate(t *testing.T) {
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
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			//cleanup EVPN before testing
			cleanBGP(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber, InterfaceType: domain.IntfTypeLoopback,
				IP: bgpNetwork, ConfigType: domain.ConfigCreate})

			Neighbors := make([]operation.ConfigBgpNeighbor, 0)
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor1Add, RemoteAs: int64(bgpNeighbor1Asn),
				ConfigType: domain.ConfigCreate})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor2Add, RemoteAs: int64(bgpNeighbor1Asn),
				ConfigType: domain.ConfigCreate})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				Role: bgpSw1Role, BgpLocalAsn: bgpSw2Asn, PeerGroup: bgpLeafPeerGroup, PeerGroupDescription: bgpLeafPeerGroupDesc,
				MaxPaths: bgpMaxPaths, Network: bgpNetwork, ConfigureOverlayGateway: bgpEvpnEnabled, AllowasIn: bgpAllowasIn,
				LoopbackPortNumber: bgpLoopbackNumber, Interfaces: Interfaces, BgpNeighbors: Neighbors,
				BfdTx: BFDTx, BfdRx: BFDRx, BfdMultiplier: BFDMultiplier, Model: detail.Model, BFDEnable: "Yes"}

			//Call the Actions
			configurefabric.ConfigureBGP(ctx, fabricGate, &sw, false, fabricErrors)
			//Setup for cleanup
			defer cleanBGP(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Fetch BGP details using NetConf
			bgpResponse, err := adapter.GetRouterBgp(client)
			fmt.Println(bgpResponse)
			assert.Nil(t, err)
			assert.Equal(t, bgpSw2Asn, bgpResponse.LocalAS)
			assert.Equal(t, bgpNetwork, bgpResponse.Network)
			assert.Equal(t, bgpMaxPaths, bgpResponse.MaxPaths)

			assert.Equal(t, BFDTx, bgpResponse.BFDTx)
			assert.Equal(t, BFDRx, bgpResponse.BFDRx)
			assert.Equal(t, BFDMultiplier, bgpResponse.BFDMultiplier)

			assert.Equal(t, 1, len(bgpResponse.PeerGroups))
			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.PeerGroups[0].Name)
			assert.Equal(t, bgpLeafPeerGroupDesc, bgpResponse.PeerGroups[0].Description)
			assert.Equal(t, fmt.Sprint(bgpNeighbor1Asn), bgpResponse.PeerGroups[0].RemoteAS)
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

			routerID, _ := adapter.GetRouterID(client)
			assert.Equal(t, bgpNetwork, routerID+"/32")

		})
	}
}

//Test EVPN Create and verify using NetConf

func TestBGP_LeafUpdate(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			//Open up the Client and set other attributes
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			{
				//cleanup EVPN before testing

				cleanBGP(client, Model)
				//defer client.Close()

				Interfaces := make([]operation.ConfigInterface, 0)
				Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber, InterfaceType: domain.IntfTypeLoopback,
					IP: bgpNetwork, ConfigType: domain.ConfigCreate})

				Neighbors := make([]operation.ConfigBgpNeighbor, 0)
				Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor1Add, RemoteAs: int64(bgpNeighbor1Asn),
					ConfigType: domain.ConfigCreate})
				Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor2Add, RemoteAs: int64(bgpNeighbor1Asn),
					ConfigType: domain.ConfigCreate})

				sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
					Role: bgpSw1Role, BgpLocalAsn: bgpSw2Asn, PeerGroup: bgpLeafPeerGroup, PeerGroupDescription: bgpLeafPeerGroupDesc,
					MaxPaths: bgpMaxPaths, Network: bgpNetwork, ConfigureOverlayGateway: bgpEvpnEnabled, AllowasIn: bgpAllowasIn,
					LoopbackPortNumber: bgpLoopbackNumber, Interfaces: Interfaces, BgpNeighbors: Neighbors,
					BfdTx: BFDTx, BfdRx: BFDRx, BfdMultiplier: BFDMultiplier, Model: detail.Model, BFDEnable: "Yes"}
				//Call the Actions
				configurefabric.ConfigureBGP(ctx, fabricGate, &sw, false, fabricErrors)

				closeActionChannel(fabricGate, fabricErrors, Errors)
			}
			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber, InterfaceType: domain.IntfTypeLoopback,
				IP: bgpNetwork, ConfigType: domain.ConfigCreate})
			Neighbors := make([]operation.ConfigBgpNeighbor, 0)
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor1Add, RemoteAs: int64(bgpNeighbor1Asn),
				ConfigType: domain.ConfigDelete})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor1AddNew, RemoteAs: int64(bgpNeighbor1AsnNew),
				ConfigType: domain.ConfigUpdate})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor2Add, RemoteAs: int64(bgpNeighbor2Asn),
				ConfigType: domain.ConfigDelete})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor2Add, RemoteAs: int64(bgpNeighbor1AsnNew),
				ConfigType: domain.ConfigUpdate})
			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				Role: bgpSw1Role, BgpLocalAsn: bgpSw2AsnNew, PeerGroup: bgpLeafPeerGroup, PeerGroupDescription: bgpLeafPeerGroupDesc,
				MaxPaths: bgpMaxPathsNew, Network: bgpNetworkNew, ConfigureOverlayGateway: bgpEvpnEnabled, AllowasIn: bgpAllowasInNew,
				LoopbackPortNumber: bgpLoopbackNumber, Interfaces: Interfaces, BgpNeighbors: Neighbors,
				BfdTx: BFDTx, BfdRx: BFDRx, BfdMultiplier: BFDMultiplier, Model: detail.Model}

			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			//Call the Actions
			configurefabric.ConfigureBGP(ctx, fabricGate, &sw, false, fabricErrors)
			//Setup for cleanup
			defer cleanBGP(client, Model)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Fetch BGP details using NetConf
			bgpResponse, err := adapter.GetRouterBgp(client)
			fmt.Println(bgpResponse)
			assert.Nil(t, err)
			assert.Equal(t, bgpSw2AsnNew, bgpResponse.LocalAS)
			assert.Equal(t, bgpNetworkNew, bgpResponse.Network)
			assert.Equal(t, bgpMaxPathsNew, bgpResponse.MaxPaths)
			assert.Equal(t, 1, len(bgpResponse.PeerGroups))
			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.PeerGroups[0].Name)
			assert.Equal(t, bgpLeafPeerGroupDesc, bgpResponse.PeerGroups[0].Description)
			assert.Equal(t, fmt.Sprint(bgpNeighbor1AsnNew), bgpResponse.PeerGroups[0].RemoteAS)
			assert.Equal(t, "true", bgpResponse.PeerGroups[0].BFD)
			//TODO Delete was failing
			//assert.Equal(t,3,len(bgpResponse.Neighbors))
			assert.Contains(t, bgpResponse.Neighbors, operation.ConfigBGPPeerGroupNeighborResponse{RemoteIP: bgpNeighbor1AddNew,
				PeerGroup: bgpLeafPeerGroup})
			assert.Contains(t, bgpResponse.Neighbors, operation.ConfigBGPPeerGroupNeighborResponse{RemoteIP: bgpNeighbor2Add,
				PeerGroup: bgpLeafPeerGroup})
			assert.Equal(t, 1, len(bgpResponse.L2VPN.Neighbors))
			assert.Equal(t, "vxlan", bgpResponse.L2VPN.Neighbors[0].Encapsulation)
			assert.Equal(t, bgpAllowasInNew, bgpResponse.L2VPN.Neighbors[0].AllowASIn)
			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.L2VPN.Neighbors[0].PeerGroup)
			assert.Equal(t, "true", bgpResponse.L2VPN.Neighbors[0].Activate)
			routerID, _ := adapter.GetRouterID(client)
			assert.Equal(t, bgpNetwork, routerID+"/32")
		})
	}

}

func cleanBGP(client *netconf.NetconfClient, Model string) (string, error) {
	detail, _ := ad.GetDeviceDetail(client)

	adapter := ad.GetAdapter(detail.Model)
	adapter.UnconfigureRouterBgp(client)
	return "", nil
}

func TestBGP_Unconfigure(t *testing.T) {

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
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanBGP(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber, InterfaceType: domain.IntfTypeLoopback,
				IP: bgpNetwork, ConfigType: domain.ConfigCreate})

			Neighbors := make([]operation.ConfigBgpNeighbor, 0)
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor1Add, RemoteAs: int64(bgpNeighbor1Asn),
				ConfigType: domain.ConfigCreate})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor2Add, RemoteAs: int64(bgpNeighbor1Asn),
				ConfigType: domain.ConfigCreate})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				Role: bgpSw1Role, BgpLocalAsn: bgpSw2Asn, PeerGroup: bgpLeafPeerGroup, PeerGroupDescription: bgpLeafPeerGroupDesc,
				MaxPaths: bgpMaxPaths, Network: bgpNetwork, ConfigureOverlayGateway: bgpEvpnEnabled, AllowasIn: bgpAllowasIn,
				LoopbackPortNumber: bgpLoopbackNumber, Interfaces: Interfaces, BgpNeighbors: Neighbors,
				BfdTx: BFDTx, BfdRx: BFDRx, BfdMultiplier: BFDMultiplier, Model: detail.Model, BFDEnable: "Yes"}

			//Call the Actions
			configurefabric.ConfigureBGP(ctx, fabricGate, &sw, false, fabricErrors)

			// Verify it is configured
			bgpResponse, err := adapter.GetRouterBgp(client)
			fmt.Println(bgpResponse)
			assert.Nil(t, err)
			assert.Equal(t, bgpSw2Asn, bgpResponse.LocalAS)
			assert.Equal(t, bgpNetwork, bgpResponse.Network)
			assert.Equal(t, bgpMaxPaths, bgpResponse.MaxPaths)

			assert.Equal(t, 1, len(bgpResponse.PeerGroups))
			assert.Equal(t, bgpLeafPeerGroup, bgpResponse.PeerGroups[0].Name)
			assert.Equal(t, bgpLeafPeerGroupDesc, bgpResponse.PeerGroups[0].Description)
			assert.Equal(t, fmt.Sprint(bgpNeighbor1Asn), bgpResponse.PeerGroups[0].RemoteAS)
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

			routerID, _ := adapter.GetRouterID(client)
			assert.Equal(t, bgpNetwork, routerID+"/32")

			fabricGate.Add(1)
			//Unconfigure BGP Now
			deconfigurefabric.UnconfigureBGP(ctx, fabricGate, &sw, fabricErrors)
			bgpResponse, err = adapter.GetRouterBgp(client)
			fmt.Println(bgpResponse)

			assert.Nil(t, err)
			assert.Equal(t, "", bgpResponse.LocalAS)
			assert.Equal(t, "", bgpResponse.Network)
			assert.Equal(t, "", bgpResponse.MaxPaths)
			assert.Equal(t, 0, len(bgpResponse.PeerGroups))
			assert.Equal(t, 0, len(bgpResponse.Neighbors))
			assert.Equal(t, 0, len(bgpResponse.L2VPN.Neighbors))

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)
		})
	}
}

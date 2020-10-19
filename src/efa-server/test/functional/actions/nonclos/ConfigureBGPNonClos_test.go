package nonclos

import (
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/actions/deconfigurefabric"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"efa-server/test/functional"
	"efa-server/usecase"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var BFDRx = "300"
var BFDTx = "300"
var BFDMultiplier = "3"
var bgpSw1Role = usecase.LeafRole
var bgpSw2Asn = "42000000"
var bgpSw2AsnNew = "42000001"
var bgpPeerGroup = "peer-group"
var bgpPeerGroupDesc = "BGP Peers"
var bgpEvpnPeerGroup = "evpn-peer-group"
var bgpEvpnPeerGroupDesc = "Evpn Peers"
var bgpMaxPaths = "8"
var bgpMaxPathsNew = "4"
var bgpNetwork = "4.4.4.4/32"
var bgpNetworkNew = "5.5.5.5/32"
var loopbackIP = "9.9.9.9/32"
var bgpEvpnEnabled = "Yes"
var bgpAllowasIn = "1"
var bgpAllowasInNew = "2"
var bgpLoopbackNumber = "1"
var bgpLoopbackNumber1 = "2"
var bgpNeighbor = []string{"31.31.31.31", "31.31.31.35", "31.31.31.39"}
var bgpNeighborAsn = []int{67890, 67891, 67892}

var bgpMultihop = "4"

type platform struct {
	IP    string
	Model string
}

var platforms = map[string]platform{

	"Freedom": {functional.NetConfFreedomIP, ad.FreedomType},
	//"Cedar":   {functional.NetConfCedarIP, ad.CedarType},
}

//Test TestBGP Create and verify using NetConf
func TestNonClosBGP(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			adapter.ConfigureInterfaceLoopback(client, bgpLoopbackNumber, loopbackIP)

			//cleanup EVPN before testing
			cleanBGP(client, detail.Model)

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber, InterfaceType: domain.IntfTypeLoopback,
				IP: bgpNetwork, ConfigType: domain.ConfigCreate})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: bgpLoopbackNumber1, InterfaceType: domain.IntfTypeLoopback,
				IP: bgpNetworkNew, ConfigType: domain.ConfigCreate})

			Neighbors := make([]operation.ConfigBgpNeighbor, 0)

			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor[0], RemoteAs: int64(bgpNeighborAsn[0]),
				ConfigType: domain.ConfigCreate, NeighborType: domain.EVPNENIGHBORType})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor[1], RemoteAs: int64(bgpNeighborAsn[1]),
				ConfigType: domain.ConfigCreate, NeighborType: domain.FabricBGPType})
			Neighbors = append(Neighbors, operation.ConfigBgpNeighbor{NeighborAddress: bgpNeighbor[2], RemoteAs: int64(bgpNeighborAsn[2]),
				ConfigType: domain.ConfigCreate, NeighborType: domain.MCTL3LBType})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password,
				Role: bgpSw1Role, BgpLocalAsn: bgpSw2Asn, PeerGroup: bgpPeerGroup, PeerGroupDescription: bgpPeerGroupDesc,
				MaxPaths: bgpMaxPaths, Network: bgpNetwork, ConfigureOverlayGateway: bgpEvpnEnabled, AllowasIn: bgpAllowasIn,
				LoopbackPortNumber: bgpLoopbackNumber, Interfaces: Interfaces, BgpNeighbors: Neighbors,
				BfdTx: BFDTx, BfdRx: BFDRx, BfdMultiplier: BFDMultiplier, Model: detail.Model, BFDEnable: "Yes",
				EvpnPeerGroup: bgpEvpnPeerGroup, EvpnPeerGroupDescription: bgpEvpnPeerGroupDesc, BgpMultihop: bgpMultihop}

			//Call the Actions
			configurefabric.ConfigureNonClosBGP(ctx, fabricGate, &sw, false, fabricErrors)

			//Fetch BGP details using NetConf
			bgpResponse, err := adapter.GetRouterBgp(client)
			assert.Nil(t, err)

			assert.Equal(t, bgpSw2Asn, bgpResponse.LocalAS)

			assert.Equal(t, BFDTx, bgpResponse.BFDTx)
			assert.Equal(t, BFDRx, bgpResponse.BFDRx)
			assert.Equal(t, BFDMultiplier, bgpResponse.BFDMultiplier)

			assert.Equal(t, bgpMaxPaths, bgpResponse.MaxPaths)

			assert.Equal(t, 2, len(bgpResponse.PeerGroups))

			assert.Equal(t, bgpEvpnPeerGroup, bgpResponse.PeerGroups[0].Name)
			if bgpEvpnPeerGroup == bgpResponse.PeerGroups[0].Name {
				assert.Equal(t, bgpMultihop, bgpResponse.PeerGroups[0].Multihop)
				assert.Equal(t, bgpEvpnPeerGroupDesc, bgpResponse.PeerGroups[0].Description)
			}

			assert.Equal(t, bgpPeerGroup, bgpResponse.PeerGroups[1].Name)
			if bgpPeerGroup == bgpResponse.PeerGroups[1].Name {
				assert.Equal(t, "true", bgpResponse.PeerGroups[1].BFD)
				assert.Equal(t, bgpPeerGroupDesc, bgpResponse.PeerGroups[1].Description)
			}

			assert.Equal(t, len(Neighbors), len(bgpResponse.Neighbors))

			for j := 0; j < len(bgpResponse.Neighbors); j++ {
				assert.Equal(t, bgpNeighbor[j], bgpResponse.Neighbors[j].RemoteIP)
				assert.Equal(t, fmt.Sprint(bgpNeighborAsn[j]), bgpResponse.Neighbors[j].RemoteAS)
				if j == 0 {
					assert.Equal(t, bgpEvpnPeerGroup, bgpResponse.Neighbors[j].PeerGroup)
				} else if j == 1 {
					assert.Equal(t, bgpPeerGroup, bgpResponse.Neighbors[j].PeerGroup)
				}
			}

			assert.Equal(t, 1, len(bgpResponse.L2VPN.Neighbors))
			assert.Equal(t, "vxlan", bgpResponse.L2VPN.Neighbors[0].Encapsulation)
			assert.Equal(t, bgpEvpnPeerGroup, bgpResponse.L2VPN.Neighbors[0].PeerGroup)
			assert.Equal(t, "true", bgpResponse.L2VPN.Neighbors[0].Activate)
			assert.Equal(t, "true", bgpResponse.L2VPN.GraceRestart)
			assert.Equal(t, "true", bgpResponse.L2VPN.RetainRTAll)

			routerID, _ := adapter.GetRouterID(client)
			assert.Equal(t, bgpNetwork, routerID+"/32")

			for k := 0; k < len(bgpResponse.NetworkList); k++ {
				if bgpResponse.NetworkList[k] == bgpNetwork || bgpResponse.NetworkList[k] == bgpNetworkNew {
				} else {
					assert.Equal(t, "true", bgpResponse.NetworkList[k])
				}
			}
			for k := 0; k < len(bgpResponse.IPv4PeerGroups); k++ {
				if bgpResponse.IPv4PeerGroups[k] == bgpPeerGroup || bgpResponse.IPv4PeerGroups[k] == bgpEvpnPeerGroup {
				} else {
					assert.Equal(t, "true", bgpResponse.IPv4PeerGroups[k])
				}
			}
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Check that action throws no Errors
			assert.Empty(t, Errors)

			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			deconfigurefabric.UnconfigureNonClosBGP(ctx, fabricGate, &sw, fabricErrors)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Check that action throws no Errors
			assert.Empty(t, Errors)

			adapter.DeleteInterfaceLoopback(client, bgpLoopbackNumber)
		})
	}

}

func cleanBGP(client *netconf.NetconfClient, Model string) (string, error) {
	adapter := ad.GetAdapter(Model)
	adapter.UnconfigureRouterBgp(client)
	return "", nil
}

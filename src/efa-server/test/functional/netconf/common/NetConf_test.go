package common

import (
	"efa-server/domain/operation"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	"efa-server/test/functional"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var UserName = "admin"
var Password = functional.DeviceAdminPassword
var FabricName = "test_fabric"
var vteploopbackPortnumber = "1"

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
		platforms["Orca"] = platform{functional.NetConfOrcaIP, ad.OrcaType}
	}
}
func TestGetModel(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			fmt.Println(detail)
			assert.Contains(t, detail.Model, Model)
			fmt.Println(detail, err)
		})
	}

}

//OverlayGateway Test Cases =================================
func TestOverlayGateway(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			netconfClient := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			netconfClient.Login()
			defer netconfClient.Close()

			detail, err := ad.GetDeviceDetail(netconfClient)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			cleanup := func(client *client.NetconfClient) (string, error) {

				ovg, _ := adapter.GetOverlayGateway(client)
				gwOnSwitch := ovg.Name
				if gwOnSwitch != "" {
					return adapter.DeleteOverlayGateway(client, gwOnSwitch)
				}
				return "", nil
			}

			cleanup(netconfClient)
			//Create Overlay Gateway
			msg, err := adapter.CreateOverlayGateway(netconfClient, FabricName, "layer2-extension",
				vteploopbackPortnumber, "true")
			fmt.Println(msg)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			ovgResponse, err := adapter.GetOverlayGateway(netconfClient)
			assert.Nil(t, err, "")
			assert.Equal(t, operation.ConfigOVGResponse{Name: FabricName, GwType: "layer2-extension",
				LoopbackID: vteploopbackPortnumber, VNIAuto: "true", Activate: "true"}, ovgResponse)

			//Delete
			msg, err = adapter.DeleteOverlayGateway(netconfClient, FabricName)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			ovgResponse, err = adapter.GetOverlayGateway(netconfClient)
			assert.Nil(t, err, "")
			assert.Equal(t, operation.ConfigOVGResponse{}, ovgResponse)
		})
	}
}

func TestEVPN(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			netconfClient := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			netconfClient.Login()
			defer netconfClient.Close()
			detail, err := ad.GetDeviceDetail(netconfClient)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			cleanup := func(client *client.NetconfClient) (string, error) {

				evpn, _ := adapter.GetEvpnInstance(client)
				evpnOnSwitch := evpn.Name
				if evpnOnSwitch != "" {
					return adapter.DeleteEvpnInstance(client, evpnOnSwitch)
				}
				return "", nil
			}

			cleanup(netconfClient)
			//Create Overlay Gateway
			msg, err := adapter.CreateEvpnInstance(netconfClient, FabricName, "100",
				"3")
			fmt.Println(msg)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			evpnResponse, err := adapter.GetEvpnInstance(netconfClient)
			assert.Nil(t, err, "")
			assert.Equal(t, operation.ConfigEVPNRespone{Name: FabricName, DuplicageMacTimerValue: "100",
				MaxCount: "3", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, evpnResponse)

			//Delete
			msg, err = adapter.DeleteEvpnInstance(netconfClient, FabricName)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			evpnResponse, err = adapter.GetEvpnInstance(netconfClient)
			assert.Nil(t, err, "")
			assert.Equal(t, operation.ConfigEVPNRespone{}, evpnResponse)
		})
	}
}

//System Test Cases =================================
func TestSetL2Mtu(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			msg, err := adapter.ConfigureSystemL2Mtu(client, "4000")
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			msg, err = adapter.UnconfigureSystemL2Mtu(client)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

		})
	}
}

func TestSetIpMtu(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			msg, err := adapter.ConfigureSystemIPMtu(client, "4000")
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			msg, err = adapter.UnconfigureSystemIPMtu(client)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

		})
	}
}

func TestRouterID(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			netconfClient := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			netconfClient.Login()
			defer netconfClient.Close()
			detail, err := ad.GetDeviceDetail(netconfClient)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			cleanup := func(client *client.NetconfClient) (string, error) {

				routerID, _ := adapter.GetRouterID(client)

				if routerID != "" {
					return adapter.UnconfigureRouterID(client)
				}
				return "", nil
			}

			cleanup(netconfClient)
			//Create Overlay Gateway
			msg, err := adapter.ConfigureRouterID(netconfClient, "10.32.45.2")

			fmt.Println(msg)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			routerID, err := adapter.GetRouterID(netconfClient)
			assert.Nil(t, err, "")
			assert.Equal(t, "10.32.45.2", routerID)

			//Delete
			msg, err = adapter.UnconfigureRouterID(netconfClient)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			routerID, err = adapter.GetRouterID(netconfClient)
			assert.Nil(t, err, "")
			assert.Equal(t, "", routerID)
		})
	}
}

//System Test Cases =================================
//BGP Test Cases =================================
func TestConfigureRouterBGP(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//Cleanup Router-BGP before starting testing
			adapter.UnconfigureRouterBgp(client)

			LocalAS := "65201"
			RemoteAS := "65210"
			BFDMultiplier := "3"
			BFDRx := "300"
			BFDTx := "300"
			MaxPaths := "2"
			Networks := "10.10.10.10/32"
			PeerGroupName := "test-group"
			//RemotePeerGroupName := "remote-group"
			RemoteAddress := "11.11.11.11"
			EVPNRemoteAddress := "12.12.12.12"
			PeerGroupDescription := "test-description"
			Multihop := "2"
			NextHopSelf := true
			loopbackNumber := "1"
			loopbackAddress := "13.13.13.12"
			RemoteLoopbackAddress := "13.13.13.13"

			BFD := "true"
			AllowASIn := "1"
			Encapsulation := "vxlan"

			adapter.ConfigureInterfaceLoopback(client, loopbackNumber, loopbackAddress+"/32")
			defer adapter.DeleteInterfaceLoopback(client, loopbackNumber)
			msg, err := adapter.ConfigureRouterBgp(client, LocalAS, PeerGroupName, PeerGroupDescription,
				Networks, MaxPaths, "Yes", AllowASIn, "Yes", "true",
				"Yes", BFDTx, BFDRx, BFDMultiplier, "No", "No")

			assert.Equal(t, "<ok/>", msg, "")

			assert.Nil(t, err, "")

			err = adapter.IsRouterBgpPresent(client)
			assert.Nil(t, err, "")

			asn, err := adapter.GetLocalAsn(client)
			assert.Equal(t, LocalAS, asn)

			msg, err = adapter.ConfigureRouterBgpNeighbor(client, RemoteAS, PeerGroupName, RemoteAddress, Multihop, true, "Yes", NextHopSelf)

			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			msg, err = adapter.ConfigureRouterBgpL2EVPNNeighbor(client, EVPNRemoteAddress, RemoteLoopbackAddress, loopbackNumber,
				RemoteAS, Encapsulation, BFD)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			bgpResponse, err := adapter.GetRouterBgp(client)

			assert.Equal(t, LocalAS, bgpResponse.LocalAS)
			assert.Equal(t, BFDMultiplier, bgpResponse.BFDMultiplier)
			assert.Equal(t, BFDRx, bgpResponse.BFDRx)
			assert.Equal(t, BFDTx, bgpResponse.BFDTx)
			assert.Equal(t, Networks, bgpResponse.Network)
			assert.Equal(t, MaxPaths, bgpResponse.MaxPaths)
			assert.Equal(t, PeerGroupName, bgpResponse.PeerGroups[0].Name)
			assert.Equal(t, PeerGroupDescription, bgpResponse.PeerGroups[0].Description)
			assert.Equal(t, BFD, bgpResponse.PeerGroups[0].BFD)
			assert.Equal(t, RemoteAS, bgpResponse.PeerGroups[0].RemoteAS)

			assert.Equal(t, PeerGroupName, bgpResponse.Neighbors[0].PeerGroup)
			assert.Equal(t, RemoteAddress, bgpResponse.Neighbors[0].RemoteIP)
			assert.Equal(t, Multihop, bgpResponse.Neighbors[0].Multihop)
			assert.Equal(t, "true", bgpResponse.Neighbors[0].NextHopSelf)

			assert.Equal(t, "true", bgpResponse.L2VPN.Neighbors[0].Activate)
			assert.Equal(t, AllowASIn, bgpResponse.L2VPN.Neighbors[0].AllowASIn)
			assert.Equal(t, Encapsulation, bgpResponse.L2VPN.Neighbors[0].Encapsulation)
			assert.Equal(t, PeerGroupName, bgpResponse.L2VPN.Neighbors[0].PeerGroup)

			//Deconfigure
			msg, err = adapter.UnconfigureRouterBgpL2EVPNNeighbor(client, EVPNRemoteAddress, RemoteLoopbackAddress)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			msg, err = adapter.UnconfigureRouterBgpNeighbor(client, RemoteAS, PeerGroupName, RemoteAddress)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			msg, err = adapter.UnconfigureRouterBgp(client)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

		})
	}
}

//BGP Test Cases =================================
func TestInterfaceEnable(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//Enable Multiple Interfaces
			msg, err := adapter.DisableInterfaces(client, []string{"0/31", "0/30"})
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Enable Multiple Interfaces
			msg, err = adapter.EnableInterfaces(client, []string{"0/31", "0/30"})
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")
		})
	}
}

func TestConfigureVE(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			veName := "111"
			ipAddress := "5.4.3.22/31"
			BFDMultiplier := "3"
			BFDRx := "300"
			BFDTx := "300"
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//create
			msg, err := adapter.ConfigureInterfaceVe(client, veName, ipAddress, BFDRx, BFDTx, BFDMultiplier)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			ResultMap, err := adapter.GetInterfaceVe(client, veName)
			assert.Nil(t, err, "")
			assert.Equal(t, map[string]string{"ip-address": ipAddress}, ResultMap, "")

			//Delete
			DeleteMsg, err := adapter.DeleteInterfaceVe(client, veName)
			assert.Equal(t, "<ok/>", DeleteMsg, "")
			assert.Nil(t, err, "")

			//Get
			ResultMap, err = adapter.GetInterfaceVe(client, veName)
			assert.Nil(t, err, "")
			assert.Equal(t, map[string]string{}, ResultMap, "")

		})
	}
}

func TestConfigurePortChannel(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			portChannel := "12"
			speed := "40000"
			controlVlan := "400"
			controlVe := "400"
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)

			if Model == ad.AvalancheType || Model == ad.OrcaType {
				// Create controlVlan as its used in PO
				msg, err := adapter.CreateClusterControlVlan(client, controlVlan, controlVe, "mct vlan")
				assert.Equal(t, "<ok/>", msg, "")
				assert.Nil(t, err, "")

				controlVlanMap, err2 := adapter.GetClusterControlVlan(client, controlVlan)
				assert.Nil(t, err2)
				assert.Equal(t, map[string]string{"control-ve": controlVe, "description": "mct vlan"}, controlVlanMap)
			}
			//create
			msg, err := adapter.CreateInterfacePo(client, portChannel, speed, "mct pc", controlVlan)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			ResultMap, err := adapter.GetInterfacePo(client, portChannel)
			assert.Nil(t, err, "")
			if Model == ad.AvalancheType || Model == ad.OrcaType {
				assert.Equal(t, map[string]string{"description": "mct pc", "speed": speed, "vlan-mode": "trunk-no-default-native", "vlan": controlVlan}, ResultMap, "")
			} else {
				assert.Equal(t, map[string]string{"description": "mct pc", "speed": speed}, ResultMap, "")
			}

			//Delete
			DeleteMsg, err := adapter.DeleteInterfacePo(client, portChannel)
			assert.Equal(t, "<ok/>", DeleteMsg, "")
			assert.Nil(t, err, "")

			//Get
			ResultMap, err = adapter.GetInterfacePo(client, portChannel)
			assert.Nil(t, err, "")
			assert.Equal(t, map[string]string{}, ResultMap, "")

			if Model == ad.AvalancheType {
				DeleteMsg, err = adapter.DeleteClusterControlVlan(client, controlVlan)
				assert.Equal(t, "<ok/>", DeleteMsg, "")
				assert.Nil(t, err, "")

				controlVlanMap, err2 := adapter.GetClusterControlVlan(client, controlVlan)
				assert.Nil(t, err2)
				assert.Equal(t, map[string]string{}, controlVlanMap, "")
			}
		})
	}
}

func TestConfigureInterfaceOnPortChannel(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			name := "0/13"
			speed := "40000"
			if Model == ad.FreedomType {
				speed = "10000"
			}
			portChannelDescription := "MCT Lag"
			portChannel := "2"
			portChannelMode := "active"
			portChannelType := "standard"
			controlVlan := "400"
			controlVe := "400"
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//Pre-Req
			premsg, err := adapter.DeleteInterfacePo(client, portChannel)
			premsg, err = adapter.DisableInterface(client, name)

			if Model == ad.AvalancheType {
				// Create controlVlan as its used in PO
				msg, err := adapter.CreateClusterControlVlan(client, controlVlan, controlVe, "mct vlan")
				assert.Equal(t, "<ok/>", msg, "")
				assert.Nil(t, err, "")

				controlVlanMap, err2 := adapter.GetClusterControlVlan(client, controlVlan)
				assert.Nil(t, err2)
				assert.Equal(t, map[string]string{"control-ve": controlVe, "description": "mct vlan"}, controlVlanMap)
			}

			premsg, err = adapter.CreateInterfacePo(client, portChannel, speed, "mct pc", controlVlan)
			assert.Equal(t, "<ok/>", premsg, "")
			assert.Nil(t, err, "")

			//create
			msg, err := adapter.AddInterfaceToPo(client, name, portChannelDescription, portChannel, portChannelMode, portChannelType, speed)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			//Get
			ResultMap, err := adapter.GetInterfacePoMember(client, name)
			assert.Nil(t, err, "")
			assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "MCT Lag", "port-channel": "2"}, ResultMap, "")

			//Delete
			DeleteMsg, err := adapter.DeleteInterfaceFromPo(client, name, portChannel)
			assert.Equal(t, "<ok/>", DeleteMsg, "")
			assert.Nil(t, err, "")

			//Get
			ResultMap, err = adapter.GetInterfacePoMember(client, name)
			assert.Nil(t, err, "")
			assert.Equal(t, map[string]string{}, ResultMap, "")

			//Cleanup of Pre-requisite
			DeleteMsg, err = adapter.DeleteInterfacePo(client, portChannel)
			assert.Equal(t, "<ok/>", DeleteMsg, "")
			assert.Nil(t, err, "")

			if Model == ad.AvalancheType {
				DeleteMsg, err = adapter.DeleteClusterControlVlan(client, controlVlan)
				assert.Equal(t, "<ok/>", DeleteMsg, "")
				assert.Nil(t, err, "")

				controlVlanMap, err2 := adapter.GetClusterControlVlan(client, controlVlan)
				assert.Nil(t, err2)
				assert.Equal(t, map[string]string{}, controlVlanMap, "")
			}
		})
	}
}

func TestPersistConfig(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			PersistMap, err := adapter.PersistConfig(client)
			fmt.Println(PersistMap, err)
		})
	}
}

func TestConfigureUnnumbered(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			unconigmsg, err := adapter.UnconfigureInterfaceUnnumbered(client, "ethernet", "0/51")
			fmt.Println(unconigmsg, err)
		})
	}
}

func TestManagementClusterStatus(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			_, mgmtClusterStatus, _, err := adapter.GetManagementClusterStatus(client)
			for _, node := range mgmtClusterStatus.MemberNodes {
				fmt.Println(node.NodeMgmtIP)
				fmt.Println(node.NodeSwitchType)
				fmt.Println(node.NodeFwVersion)
			}
			fmt.Println(mgmtClusterStatus)
			assert.NoError(t, err)

		})
	}
}

func TestConfigureNumberedRouteMap(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			_, err = adapter.ConfigureNumberedRouteMap(client)
			assert.NoError(t, err)

			assert.Contains(t, detail.Model, Model)
			adapter = ad.GetAdapter(detail.Model)

			_, err = adapter.UnConfigureNumberedRouteMap(client)

			assert.NoError(t, err)
		})
	}
}

//System Test Case for Host Name =================================
func TestSetHostName(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			hostName := "EFATestDevice"
			msg, err := adapter.ConfigureSwitchHostName(client, hostName)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			getHostName, _ := adapter.GetSwitchHostName(client)
			assert.Equal(t, getHostName, hostName)

			msg, err = adapter.UnconfigureSwitchHostName(client)
			assert.Equal(t, "<ok/>", msg, "")
			assert.Nil(t, err, "")

			getHostName, _ = adapter.GetSwitchHostName(client)
			assert.NotEqual(t, getHostName, hostName)

		})
	}
}

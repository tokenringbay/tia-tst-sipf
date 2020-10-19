package base

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/adapter/platform/slx"
	"efa-server/infra/device/client"
	"errors"
	"fmt"

	"github.com/beevik/etree"
)

//ConfigureRouterBgp is used to configure "router bgp" on the switching device
func (base *SLXBase) ConfigureRouterBgp(client *client.NetconfClient, localAs string, peerGroupName string,
	description string, networkAddress string, maxPaths string,
	evpn string, allowasIn string, retainRouteTargetAll string, nextHopUnchanged string, bfdEnable string,
	bfdMinTx string, bfdMinRx string, bfdMultiplier string, detrisibuteConnected string, detrisibuteConnectedWithRouteMap string) (string, error) {
	var bgpMap = map[string]interface{}{"local_as": localAs, "peer_group_name": peerGroupName,
		"description":     description,
		"network_address": networkAddress, "max_paths": maxPaths,
		"evpn": evpn, "allowas_in": allowasIn,
		"retain_route_target_all":          retainRouteTargetAll,
		"next_hop_unchanged":               nextHopUnchanged,
		"bfd_enable":                       bfdEnable,
		"bfd_min_tx":                       bfdMinTx,
		"bfd_min_rx":                       bfdMinRx,
		"bfd_multiplier":                   bfdMultiplier,
		"detrisibuteConnected":             detrisibuteConnected,
		"detrisibuteConnectedWithRouteMap": detrisibuteConnectedWithRouteMap,
	}

	config, templateError := base.GetStringFromTemplate(bgpRouterCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err

}

//UnconfigureRouterBgp is used to unconfigure "router bgp" from the switching device
func (base *SLXBase) UnconfigureRouterBgp(client *client.NetconfClient) (string, error) {
	var bgpMap = map[string]interface{}{}

	config, templateError := base.GetStringFromTemplate(routerBgpDelete, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureRouterBgpNeighbor is used to configure "router bgp neighbour" and "associate the neighbour with a peer-group
// on the switching device
func (base *SLXBase) ConfigureRouterBgpNeighbor(client *client.NetconfClient, remoteAs string, peerGroupName string,
	neighborAddress string, bgpMultihop string, unnumberedInterface bool, isLeaf string, nextHopSelf bool) (string, error) {
	var bgpMap = map[string]interface{}{"remote_as": remoteAs, "peer_group_name": peerGroupName,
		"neighbor_address": neighborAddress, "bgp_multihop": bgpMultihop, "is_leaf": isLeaf, "next_hop_self": nextHopSelf}

	config, templateError := base.GetStringFromTemplate(routerBgpNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(routerBgpNeighborAssociateWithPeerGroup, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}
	if unnumberedInterface {
		config, templateError = base.GetStringFromTemplate(routerBgpNeighborMultihop, bgpMap)
		if templateError != nil {
			return "", templateError
		}

		resp, err = client.EditConfig(config)
		if err != nil {
			return "", err
		}
	}

	return resp, err
}

//UnconfigureRouterBgpNeighbor is used to unconfigure "router bgp neighbour" from the switching device
func (base *SLXBase) UnconfigureRouterBgpNeighbor(client *client.NetconfClient, remoteAs string, peerGroupName string,
	neighborAddress string) (string, error) {
	var bgpMap = map[string]interface{}{"neighbor_address": neighborAddress}

	config, templateError := base.GetStringFromTemplate(routerBgpNeighborDelete, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//IsRouterBgpPresent is used to check whether "router bgp" config is present on a switching device
func (base *SLXBase) IsRouterBgpPresent(client *client.NetconfClient) error {
	resp, err := client.GetConfig("/routing-system/router/router-bgp")
	if err != nil {
		return err
	}
	if resp == "<data></data>" {
		return errors.New("No Router BGP")
	}
	return nil
}

//GetRouterBgp is used to get the running-config of "router bgp"
func (base *SLXBase) GetRouterBgp(client *client.NetconfClient) (operation.ConfigBGPResponse, error) {
	bgpResponse := operation.ConfigBGPResponse{}
	resp, err := client.GetConfig("/routing-system/router/router-bgp")

	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return bgpResponse, err
	}

	if elem := doc.FindElement("//local-as"); elem != nil {

		bgpResponse.LocalAS = elem.Text()
	}
	if Networks := doc.FindElements("//network/network-ipv4-address"); len(Networks) != 0 {
		bgpResponse.NetworkList = make([]string, 0)
		for _, Network := range Networks {
			bgpResponse.NetworkList = append(bgpResponse.NetworkList, Network.Text())
		}
	}
	if elem := doc.FindElement("//network-ipv4-address"); elem != nil {

		bgpResponse.Network = elem.Text()
	}
	if elem := doc.FindElement("//af-ipv4-neighbor-peergroup-holder"); elem != nil {
		if PeerGroups := doc.FindElements("//af-ipv4-neighbor-peergroup/af-ipv4-neighbor-peergroup-name"); len(PeerGroups) != 0 {
			for _, PeerGroup := range PeerGroups {
				bgpResponse.IPv4PeerGroups = append(bgpResponse.IPv4PeerGroups, PeerGroup.Text())
			}
		}
	}
	if elem := doc.FindElement("//load-sharing-value"); elem != nil {

		bgpResponse.MaxPaths = elem.Text()
	}
	if elem := doc.FindElement("//min-tx"); elem != nil {

		bgpResponse.BFDTx = elem.Text()
	}
	if elem := doc.FindElement("//min-rx"); elem != nil {

		bgpResponse.BFDRx = elem.Text()
	}
	if elem := doc.FindElement("//multiplier"); elem != nil {

		bgpResponse.BFDMultiplier = elem.Text()
	}

	if peerGroups := doc.FindElements("//neighbor-peer-grp"); len(peerGroups) != 0 {
		bgpResponse.PeerGroups = make([]operation.ConfigBGPPeerGroupResponse, 0, len(peerGroups))
		for _, peerGroupelem := range peerGroups {
			peerGroupResponse := operation.ConfigBGPPeerGroupResponse{}
			if de := peerGroupelem.FindElement(".//router-bgp-neighbor-peer-grp"); de != nil {
				peerGroupResponse.Name = de.Text()
			}
			if de := peerGroupelem.FindElement(".//description"); de != nil {
				peerGroupResponse.Description = de.Text()
			}
			if de := peerGroupelem.FindElement(".//bfd-enable"); de != nil {
				peerGroupResponse.BFD = "true"
			}
			if de := peerGroupelem.FindElement(".//remote-as"); de != nil {
				peerGroupResponse.RemoteAS = de.Text()
			}
			if de := peerGroupelem.FindElement(".//ebgp-multihop-count"); de != nil {
				peerGroupResponse.Multihop = de.Text()
			}
			bgpResponse.PeerGroups = append(bgpResponse.PeerGroups, peerGroupResponse)
		}
	}

	if neighbors := doc.FindElements("//neighbor-ips/neighbor-addr"); len(neighbors) != 0 {
		bgpResponse.Neighbors = make([]operation.ConfigBGPPeerGroupNeighborResponse, 0, len(neighbors))
		for _, elem := range neighbors {
			peerGroupNeighbors := operation.ConfigBGPPeerGroupNeighborResponse{}
			if de := elem.FindElement(".//router-bgp-neighbor-address"); de != nil {
				peerGroupNeighbors.RemoteIP = de.Text()
			}
			if de := elem.FindElement(".//remote-as"); de != nil {
				peerGroupNeighbors.RemoteAS = de.Text()
			}
			if de := elem.FindElement(".//associate-peer-group"); de != nil {
				peerGroupNeighbors.PeerGroup = de.Text()
			}
			if de := elem.FindElement(".//ebgp-multihop-count"); de != nil {
				peerGroupNeighbors.Multihop = de.Text()
			}
			if de := elem.FindElement(".//next-hop-self"); de != nil {
				peerGroupNeighbors.NextHopSelf = "true"
			}

			bgpResponse.Neighbors = append(bgpResponse.Neighbors, peerGroupNeighbors)
		}
	}

	if elem := doc.FindElement("//l2vpn/evpn"); elem != nil {
		bgpResponse.L2VPN = operation.ConfigBGPL2VPNResponse{}
		if de := elem.FindElement(".//graceful-restart/graceful-restart-status"); de != nil {
			bgpResponse.L2VPN.GraceRestart = "true"
		}
		if de := elem.FindElement(".//retain/route-target/all"); de != nil {
			bgpResponse.L2VPN.RetainRTAll = "true"
		}
		if neighbors := elem.FindElements(".//neighbor/evpn-peer-group"); len(neighbors) != 0 {
			bgpResponse.L2VPN.Neighbors = make([]operation.ConfigBGPEVPNNeighborResponse, 0, len(neighbors))
			for _, EvpnElem := range neighbors {
				evpnNeighbor := operation.ConfigBGPEVPNNeighborResponse{}
				if de := EvpnElem.FindElement(".//encapsulation"); de != nil {
					evpnNeighbor.Encapsulation = de.Text()
				}
				if de := EvpnElem.FindElement(".//evpn-neighbor-peergroup-name"); de != nil {
					evpnNeighbor.PeerGroup = de.Text()
				}
				if de := EvpnElem.FindElement(".//allowas-in"); de != nil {
					evpnNeighbor.AllowASIn = de.Text()
				}
				if de := EvpnElem.FindElement(".//next-hop-unchanged"); de != nil {
					evpnNeighbor.NextHopUnchanged = "true"
				}
				if de := EvpnElem.FindElement(".//activate"); de != nil {
					evpnNeighbor.Activate = "true"
				}
				bgpResponse.L2VPN.Neighbors = append(bgpResponse.L2VPN.Neighbors, evpnNeighbor)
			}
		} else if neighbors := elem.FindElements(".//neighbor/evpn-neighbor-ipv4"); len(neighbors) != 0 {
			bgpResponse.L2VPN.Neighbors = make([]operation.ConfigBGPEVPNNeighborResponse, 0, len(neighbors))
			for _, EvpnElem := range neighbors {
				evpnNeighbor := operation.ConfigBGPEVPNNeighborResponse{}
				if de := EvpnElem.FindElement(".//encapsulation"); de != nil {
					evpnNeighbor.Encapsulation = de.Text()
				}
				if de := EvpnElem.FindElement(".//evpn-neighbor-ipv4-address"); de != nil {
					evpnNeighbor.IPAddress = de.Text()
				}
				if de := EvpnElem.FindElement(".//next-hop-unchanged"); de != nil {
					evpnNeighbor.NextHopUnchanged = "true"
				}
				if de := EvpnElem.FindElement(".//activate"); de != nil {
					evpnNeighbor.Activate = "true"
				}
				bgpResponse.L2VPN.Neighbors = append(bgpResponse.L2VPN.Neighbors, evpnNeighbor)
			}
		}
	}

	return bgpResponse, err
}

//GetLocalAsn is used to get the "router bgp" Local ASN configuration from the switching device
func (base *SLXBase) GetLocalAsn(client *client.NetconfClient) (string, error) {
	resp, err := client.GetConfig("/routing-system/router/router-bgp")

	var LocalASN string
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return LocalASN, err
	}
	if localas := doc.FindElement("//local-as"); localas != nil {
		LocalASN = localas.Text()
	}

	return LocalASN, err
}

//GetRouterBgpL2EVPNNeighborEncapType is used to get the EncapType that need to be configured.
func (base *SLXBase) GetRouterBgpL2EVPNNeighborEncapType(client *client.NetconfClient) (string, error) {
	return slx.BGPEncapTypeForSwitching, nil
}

//ConfigureRouterBgpL2EVPNNeighbor is used to configure the MCT BGP neighbour and enable BFD on the neighbour, on the switching device
func (base *SLXBase) ConfigureRouterBgpL2EVPNNeighbor(client *client.NetconfClient, neighborAddress string,
	neighborLoopBackAddress string, loopbackNumber string,
	remoteAs string, encapType string, bfdEnabled string) (string, error) {
	var bgpMap = map[string]interface{}{"neighborAddress": neighborAddress, "remoteAs": remoteAs, "encapType": encapType, "bfdEnabled": bfdEnabled,
		"nextHopEnable": false, "peer_group_name": ""}
	config, templateError := base.GetStringFromTemplate(routerBgpMctNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(routerBgpMctL2EvpnNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(routerBgpNeighborDeactivateInIpv4UnicastAF, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	return resp, err
}

//UnconfigureRouterBgpL2EVPNNeighbor is used to unconfigure the MCT BGP neighbour from the switching device
func (base *SLXBase) UnconfigureRouterBgpL2EVPNNeighbor(client *client.NetconfClient, neighborAddress string, neighborLoopBackAddress string) (string, error) {
	var bgpMap = map[string]interface{}{"neighbor_address": neighborAddress}
	config, templateError := base.GetStringFromTemplate(routerBgpNeighborDelete, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)
	return resp, err
}

//ExecuteClearBgpEvpnNeighbourAll is used to execute "clear bgp evpn neighbor all" on the switching device
func (base *SLXBase) ExecuteClearBgpEvpnNeighbourAll(sshClient *client.SSHClient) error {
	command := "clear bgp evpn neighbor  all"
	output := sshClient.ExecuteOperationalCommand(command)
	if output != "" {
		return errors.New(output)
	}
	return nil
}

//ExecuteClearBgpEvpnNeighbour is used to execute "clear bgp evpn neighbor <neighbour-ip>" on the switching device
func (base *SLXBase) ExecuteClearBgpEvpnNeighbour(sshClient *client.SSHClient, neighbourIP string) error {
	command := "clear bgp evpn neighbor " + neighbourIP
	output := sshClient.ExecuteOperationalCommand(command)
	if output != "" {
		return errors.New(output)
	}
	return nil
}

//ConfigureIPRoute is used to configure the static route
func (base *SLXBase) ConfigureIPRoute(client *client.NetconfClient, loopbackIP string,
	veIP string) (string, error) {

	return "<ok/>", nil
}

//DeConfigureIPRoute is used to configure the static route
func (base *SLXBase) DeConfigureIPRoute(client *client.NetconfClient, loopbackIP string,
	veIP string) (string, error) {

	return "<ok/>", nil
}

//ConfigureNonClosRouterBgp is used to configure "router bgp" on the switching device for Non-clos use-case
func (base *SLXBase) ConfigureNonClosRouterBgp(client *client.NetconfClient, localAs string,
	networkAddress []string, bfdMinTx string, bfdMinRx string, bfdMultiplier string,
	eBGPPeerGroup string, eBGPPeerGroupDescription string, eBGPBFD bool,
	EVPNPeerGroup string, EVPNPeerGroupDescription string, EVPNPeerGroupMultiHop string,
	MaxPaths string, encap string, nextHopUnChanged bool,
	retainRTAll bool) (string, error) {

	var bgpMap = map[string]interface{}{"local_as": localAs,
		"bfd_min_tx":               bfdMinTx,
		"bfd_min_rx":               bfdMinRx,
		"bfd_multiplier":           bfdMultiplier,
		"max_paths":                MaxPaths,
		"eBGPPeerGroup":            eBGPPeerGroup,
		"eBGPPeerGroupDescription": eBGPPeerGroupDescription,
		"eBGPBFD":                  eBGPBFD,
		"EVPNPeerGroup":            EVPNPeerGroup,
		"EVPNPeerGroupDescription": EVPNPeerGroupDescription,
		"EVPNPeerGroupMultiHop":    EVPNPeerGroupMultiHop,
		"encap":                    encap,
		"next_hop_unchanged":       nextHopUnChanged,
		"retain_rt_all":            retainRTAll,
	}

	config, templateError := base.GetStringFromTemplate(bgpRouterNonClosCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err := client.EditConfig(config)

	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(bgpRouterNonClosEVPNPeerGroupCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(bgpRouterNonClosEBGPPeerGroupCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	config, templateError = base.GetStringFromTemplate(bgpRouterNonClosEVPNNeighborProperties, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	for index := 0; index < len(networkAddress); index++ {
		bgpMap = map[string]interface{}{"network_address": networkAddress[index]}
		config, templateError = base.GetStringFromTemplate(bgpRouterNonClosNetworkCreate, bgpMap)
		if templateError != nil {
			return "", templateError
		}
		resp, err = client.EditConfig(config)
	}
	return resp, err
}

//ConfigureNonClosRouterEvpnNeighbor is used to configure Evpn Neighbors on the switching device for Non-clos use-case
func (base *SLXBase) ConfigureNonClosRouterEvpnNeighbor(client *client.NetconfClient,
	remoteAs string, peerGroup string, peerGroupDescription string, neighborAddress string,
	loopbackNumber string, multiHop string) (string, error) {
	var bgpMap = map[string]interface{}{"remote_as": remoteAs, "neighborAddress": neighborAddress,
		"loopback_number": loopbackNumber, "peer_group_name": peerGroup,
		"peer_group_desc": peerGroupDescription, "multi_hop": multiHop,
	}

	config, templateError := base.GetStringFromTemplate(bgpRouterNonClosEVPNNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(bgpRouterNonClosEVPNNeighborAssociate, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(routerBgpNeighborDeactivateInIpv4UnicastAF, bgpMap)
	if templateError != nil {
		return "", templateError
	}
	resp, err = client.EditConfig(config)

	return resp, err
}

//ConfigureNonClosRouterBgpNeighbor is used to configure Evpn Neighbors on the switching device for Non-clos use-case
func (base *SLXBase) ConfigureNonClosRouterBgpNeighbor(client *client.NetconfClient,
	remoteAs string, peerGroup string, peerGroupDesc string, neighborAddress string,
	isBfdEnabled bool, isNextHopSelf bool) (string, error) {

	var bgpMap = map[string]interface{}{
		"remote_as": remoteAs, "neighbor_address": neighborAddress,
		"peer_group_name": peerGroup, "peer_group_desc": peerGroupDesc,
		"next_hop_self": isNextHopSelf, "bfd_enabled": isBfdEnabled,
	}

	config, templateError := base.GetStringFromTemplate(bgpRouterNonClosBGPNeighborCreate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(bgpRouterNonClosBGPNeighborAssociate, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	if err != nil {
		return "", err
	}

	config, templateError = base.GetStringFromTemplate(bgpRouterNonClosBGPNeighborProperties, bgpMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)

	return resp, err
}

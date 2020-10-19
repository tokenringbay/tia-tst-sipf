package configurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	"strconv"
	"strings"
	"sync"

	nlog "github.com/sirupsen/logrus"
)

//ConfigureNonClosBGP is used to configure "router bgp" config on the switch.
func ConfigureNonClosBGP(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Configure BGP",
	})
	adapter := ad.GetAdapter(sw.Model)

	netconfClient := &client.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := netconfClient.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Interface Login", Error: err, Host: sw.Host}
		return
	}
	defer netconfClient.Close()

	//BGP
	Operation := "BGP Operation"
	log.Info("Configure BGP")

	Network := make([]string, 2)
	Network[0] = sw.Network
	Network[1] = sw.NonCLOSNetwork

	encap := "vxlan"
	nextHopUnChanged := true
	retainRTAll := true

	x, err := adapter.ConfigureNonClosRouterBgp(netconfClient, sw.BgpLocalAsn,
		Network, sw.BfdTx, sw.BfdRx, sw.BfdMultiplier, sw.PeerGroup, sw.PeerGroupDescription, formatYesNo(sw.BFDEnable),
		sw.EvpnPeerGroup, sw.EvpnPeerGroupDescription, sw.BgpMultihop, sw.MaxPaths, encap,
		nextHopUnChanged, retainRTAll)

	if err == nil {
		log.Infof("BGP Operation Completed: %s\n", x)
	} else {
		log.Errorf("BGP Operation Failed: %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
	}

	//BGP Neighbor
	Operation = "BGP Neighbor"
	//First Delete Neighbors
	for _, neigh := range sw.BgpNeighbors {
		if neigh.ConfigType == domain.ConfigDelete {
			remoteAs := strconv.FormatInt(neigh.RemoteAs, 10)
			log.Infof("Delete BGP Neighbor RemoteAs=%d,IP =%s", neigh.RemoteAs, neigh.NeighborAddress)
			peerGroup := ""
			// TODO: Use Global Variable for EVPN_NEIGHBOR_TYPE
			if neigh.NeighborType == domain.EVPNENIGHBORType {
				peerGroup = sw.EvpnPeerGroup
			} else if neigh.NeighborType == domain.FabricBGPType {
				peerGroup = sw.PeerGroup
			} else {
				peerGroup = ""
			}
			_, err = adapter.UnconfigureRouterBgpNeighbor(netconfClient, remoteAs,
				peerGroup, neigh.NeighborAddress)
			if err != nil {
				errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
			}
		}
	}

	//Then Create Neighbors
	for _, neigh := range sw.BgpNeighbors {
		if neigh.ConfigType != domain.ConfigDelete {
			log.Infof("Create BGP Neighbor RemoteAs=%d,IP =%s,Type=%s", neigh.RemoteAs,
				neigh.NeighborAddress, neigh.NeighborType)
			remoteAs := strconv.FormatInt(neigh.RemoteAs, 10)
			// TODO: Use Global Variable for EVPN_NEIGHBOR_TYPE/BGP_NEIGHBOR_TYPE
			if neigh.NeighborType == domain.EVPNENIGHBORType {
				_, err = adapter.ConfigureNonClosRouterEvpnNeighbor(netconfClient, remoteAs,
					sw.EvpnPeerGroup, sw.EvpnPeerGroupDescription, neigh.NeighborAddress,
					sw.LoopbackPortNumber, sw.BgpMultihop)
			} else if neigh.NeighborType == domain.FabricBGPType {
				_, err = adapter.ConfigureNonClosRouterBgpNeighbor(netconfClient, remoteAs,
					sw.PeerGroup, sw.PeerGroupDescription, neigh.NeighborAddress, formatYesNo(sw.BFDEnable), false)
			} else if neigh.NeighborType == domain.MCTL3LBType {
				_, err = adapter.ConfigureNonClosRouterBgpNeighbor(netconfClient, remoteAs,
					"", "", neigh.NeighborAddress, formatYesNo(sw.BFDEnable), true)
			}
			if err != nil {
				errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
			}
		}

	}

	if err == nil {
		log.Infof("BGP Neighbor Operation Completed: %s\n", x)
	} else {
		log.Errorf("BGP Neighbor Operation Failed: %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
	}

	//Router ID
	Operation = "Router ID"
	var routerID string
	for _, intf := range sw.Interfaces {
		if sw.LoopbackPortNumber == intf.InterfaceName && domain.IntfTypeLoopback == intf.InterfaceType {
			s := strings.Split(intf.IP, "/")
			routerID = s[0]
		}
	}

	if routerID == "" {
		log.Error("Router ID could not be computed")
	}
	log.Infof("Configure Router ID=%s", routerID)
	x, err = adapter.ConfigureRouterID(netconfClient, routerID)
	if err == nil {
		log.Infof("Router ID Operation Completed: %s\n", x)
	} else {
		log.Errorf("Router ID Operation Failed: %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
	}
}

func formatYesNo(data string) bool {
	if data == "Yes" {
		return true
	}
	return false
}

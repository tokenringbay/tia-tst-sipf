package configurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	nlog "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"

	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
	"efa-server/usecase"
)

//ConfigureBGP is used to configure "router bgp" config on the switch.
func ConfigureBGP(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
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
	retainRouteTargetAll := "No"
	nextHopUnChanged := "No"
	isLeaf := "Yes"

	numberedInterface := (sw.P2PIPType == domain.P2PIpTypeNumbered)
	if sw.Role == usecase.SpineRole {
		retainRouteTargetAll = "Yes"
		nextHopUnChanged = "Yes"
		isLeaf = "No"
	}

	detrisibuteConnected := "No"
	detrisibuteConnectedWithRouteMap := "No"
	//If OverlayGateway is no
	if sw.Role == usecase.LeafRole && sw.ConfigureOverlayGateway == "No" {
		//For numbered interface, we need to do redistribute connected with route Map
		if numberedInterface {
			detrisibuteConnectedWithRouteMap = "Yes"
			x, err := adapter.ConfigureNumberedRouteMap(netconfClient)

			if err == nil {
				log.Infof("BGP Operation Route Map Completed: %s\n", x)
			} else {
				log.Errorf("BGP Operation Route Map Failed: %s\n", err)
				errs <- actions.OperationError{Operation: "BGP Operation Route Map", Error: err, Host: sw.Host}
			}
		} else {
			////For unnumbered interface, we need to do redistribute connected
			detrisibuteConnected = "Yes"
		}
	} //else go with the default

	x, err := adapter.ConfigureRouterBgp(netconfClient, sw.BgpLocalAsn,
		sw.PeerGroup, sw.PeerGroupDescription, sw.Network, sw.MaxPaths,
		sw.ConfigureOverlayGateway, sw.AllowasIn, retainRouteTargetAll,
		nextHopUnChanged, sw.BFDEnable, sw.BfdTx, sw.BfdRx, sw.BfdMultiplier, detrisibuteConnected, detrisibuteConnectedWithRouteMap)

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
			_, err = adapter.UnconfigureRouterBgpNeighbor(netconfClient, remoteAs,
				sw.PeerGroup, neigh.NeighborAddress)
			if err != nil {
				errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
			}
		}
	}

	//Then Create Neighbors
	for _, neigh := range sw.BgpNeighbors {
		if neigh.ConfigType != domain.ConfigDelete {
			log.Infof("Create BGP Neighbor RemoteAs=%d,IP =%s", neigh.RemoteAs, neigh.NeighborAddress)
			remoteAs := strconv.FormatInt(neigh.RemoteAs, 10)
			//For Un-numbered set BGP Multihop
			unnumberedInterface := (sw.P2PIPType == domain.P2PIpTypeUnnumbered)
			nextHopSelf := false
			_, err = adapter.ConfigureRouterBgpNeighbor(netconfClient, remoteAs,
				sw.PeerGroup, neigh.NeighborAddress, sw.BgpMultihop, unnumberedInterface, isLeaf, nextHopSelf)
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

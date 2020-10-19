package deconfigurefabric

import (
	"context"
	"efa-server/gateway/appcontext"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	nlog "github.com/sirupsen/logrus"
	"sync"

	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
	"errors"
	"strconv"
)

//UnconfigureNonClosBGP is used to unconfigure "router bgp" config on the switch.
func UnconfigureNonClosBGP(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    sw.Fabric,
		"Operation": "UnConfigure BGP",
		"Switch":    sw.Host,
	})
	adapter := ad.GetAdapter(sw.Model)
	client := &client.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	client.Login()
	defer client.Close()

	Operation := "UnConfigure BGP Neighbor"
	bgpResponse, _ := adapter.GetRouterBgp(client)
	bgpNeighborResponseMap := make(map[string]bool, 0)
	for _, swNeigh := range bgpResponse.Neighbors {
		bgpNeighborResponseMap[swNeigh.RemoteIP] = true
	}
	//First Delete Neighbors
	for _, neigh := range sw.BgpNeighbors {
		//If neighbors present on Switch,only then delete the neighbors
		//there is a bug on the switch netconf if try to delete a non-existent neighbor
		if _, ok := bgpNeighborResponseMap[neigh.NeighborAddress]; ok {
			delete(bgpNeighborResponseMap, neigh.NeighborAddress)
			log.Infof("BGP Neighbor RemoteAs=%d,IP =%s", neigh.RemoteAs, neigh.NeighborAddress)
			remoteAs := strconv.FormatInt(neigh.RemoteAs, 10)
			_, err := adapter.UnconfigureRouterBgpNeighbor(client, remoteAs,
				sw.PeerGroup, neigh.NeighborAddress)
			if err != nil {
				log.Errorf("BGP Neighbor Operation Failed: %s\n", err)
				errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
			}
		} else {
			log.Infof("BGP Neighbor RemoteAs not present on switch,Ignore=%d,IP =%s", neigh.RemoteAs, neigh.NeighborAddress)
		}

	}

	//Fetch BGP Neighbors again using NetConf. If there is no other neighbor delete the router bgp
	if len(bgpNeighborResponseMap) == 0 {

		Operation = "UnConfigure Router BGP Operation"
		x, err := adapter.UnconfigureRouterBgp(client)
		if err == nil {
			log.Infof("BGP Unconfiguration Completed: %s\n", x)
		} else {
			log.Errorf("BGP Unconfiguration Failed: %s\n", err)
			errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
			return
		}
	}

	log.Info("UnConfigure Router ID")
	Operation = "UnConfigure Router ID"
	x, err := adapter.UnconfigureRouterID(client)
	if err == nil {
		log.Infof("Router ID Operation Completed: %s\n", x)
	} else {
		log.Errorf("Router ID Operation Failed: %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
	}
	return
}

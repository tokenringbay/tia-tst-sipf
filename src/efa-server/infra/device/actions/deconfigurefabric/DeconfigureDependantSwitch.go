package deconfigurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"strconv"
	"sync"

	"efa-server/infra/device/adapter/interface"
)

//UnconfigureDependantSwitch cleans up interfaces and bgp neighbors from dependant Device
func UnconfigureDependantSwitch(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.ConfigSwitch,
	force bool, errs chan actions.OperationError, persist bool) {
	defer fabricGate.Done()
	ctx = context.WithValue(ctx, appcontext.DeviceName, sw.Host)
	log := appcontext.Logger(ctx)
	adapter := ad.GetAdapter(sw.Model)
	netconfClient := &client.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := netconfClient.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Interface Login", Error: err, Host: sw.Host}
		return
	}
	defer netconfClient.Close()

	if err := deleteRouterBGPNeighbors(log, adapter, netconfClient, &sw); err != nil {
		errs <- actions.OperationError{Operation: "Delete Router BGP Neighbors Failed", Error: err, Host: sw.Host}
	}

	if err := deleteInterfaces(log, adapter, netconfClient, &sw); err != nil {
		errs <- actions.OperationError{Operation: "Configure Switch Interfaces", Error: err, Host: sw.Host}
	}
	if persist {
		if _, err := adapter.PersistConfig(netconfClient); err != nil {
			msg := fmt.Sprintf("Persist Config Failed %s:Error %s", sw.Host, err)
			log.Errorln(msg)
			errs <- actions.OperationError{Operation: "Configure Switch Properties", Error: errors.New(msg), Host: sw.Host}
		}
	}

}

func deleteRouterBGPNeighbors(log *nlog.Entry, adapter interfaces.Switch, client *client.NetconfClient, sw *operation.ConfigSwitch) error {
	//Coding yet to be done based on neighbor type
	//fmt.Println(sw.BgpNeighbors)

	bgpResponse, _ := adapter.GetRouterBgp(client)
	bgpNeighborResponseMap := make(map[string]bool, 0)
	for _, swNeigh := range bgpResponse.Neighbors {
		bgpNeighborResponseMap[swNeigh.RemoteIP] = true
	}

	for _, neighbor := range sw.BgpNeighbors {
		//If neighbors present on Switch,only then delete the neighbors
		//there is a bug on the switch netconf if try to delete a non-existent neighbor
		if _, ok := bgpNeighborResponseMap[neighbor.NeighborAddress]; ok {
			delete(bgpNeighborResponseMap, neighbor.NeighborAddress)

			remoteAs := strconv.FormatInt(neighbor.RemoteAs, 10)
			_, err := adapter.UnconfigureRouterBgpNeighbor(client, remoteAs,
				sw.PeerGroup, neighbor.NeighborAddress)
			if err != nil {
				log.Errorf("failed to delete BGP Neighbor RemoteAs=%d,IP =%s", neighbor.RemoteAs, neighbor.NeighborAddress)
			} else {
				log.Infof("deleted BGP Neighbor RemoteAs=%d,IP =%s", neighbor.RemoteAs, neighbor.NeighborAddress)
			}
		}
	}
	return nil
}

func deleteInterfaces(log *nlog.Entry, adapter interfaces.Switch, client *client.NetconfClient,
	sw *operation.ConfigSwitch) error {

	for _, intf := range sw.Interfaces {
		ifName := intf.InterfaceName
		ifType := intf.InterfaceType
		var err error
		var x string

		//Try to fetch the interface details of the interface before clear
		if intfMap, err := adapter.GetInterface(client, ifType, ifName); err == nil {
			//fmt.Println("Clear", client.Host, ifType, ifName, intfMap)
			log.Infoln("Delete interface ", ifType, ifName)
			if intfMap["address"] != "" {
				//Address is populated for Numbered interfaces
				adapter.UnconfigureInterfaceNumbered(client, ifType, ifName, intfMap["address"])
			}
			if intfMap["donor_type"] == "loopback" {
				adapter.UnconfigureInterfaceUnnumbered(client, ifType, ifName)
			}

		}

		if err != nil {
			msg := fmt.Sprintf("Interface %s %s with IP address %s:Error %s", ifType, ifName, intf.IP, err)
			log.Errorln(msg, x)
			return errors.New(msg)
		}
	}
	return nil
}

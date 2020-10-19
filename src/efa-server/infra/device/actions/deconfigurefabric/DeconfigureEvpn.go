package deconfigurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"errors"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//UnconfigureEvpn is used to unconfigure "evpn <evpn-instance-name>" on the switch.
func UnconfigureEvpn(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    sw.Fabric,
		"Operation": "Configure EVPN",
		"Switch":    sw.Host,
	})

	adapter := ad.GetAdapter(sw.Model)

	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	client.Login()
	defer client.Close()

	log.Info("Mac & Arp")
	Operation := "Mac & Arp"

	x, err := adapter.UnconfigureMacAndArp(client)
	if err == nil {
		log.Infof("Mac & Arp Unconfiguration Completed %s\n", x)
	} else {
		log.Errorf("Mac & Arp Unconfiguration Error %s\n", err)
		errs <- actions.OperationError{Operation: Operation,
			Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
		return
	}

	Operation = "EVPN"
	x, err = adapter.DeleteEvpnInstance(client, sw.Fabric)
	if err == nil {
		log.Printf("EVPN Unconfiguration Completed %s\n", x)
	} else {
		log.Errorf("EVPN Unconfiguration Error %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
		return
	}

	//Execute "clear bgp evpn neighbour all"
	Operation = "Clearing All BGP EVPN Neighbour"
	err = actions.ExecuteClearBgpEvpnNeighbourAll(sw)
	if err == nil {
		log.Info("Clearing All BGP EVPN Neighbour completed\n")
	} else {
		log.Error("Clearing All BGP EVPN Neighbour Failed\n")
		errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
	}
	return
}

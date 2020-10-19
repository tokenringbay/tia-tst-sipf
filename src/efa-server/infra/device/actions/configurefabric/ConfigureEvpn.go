package configurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//ConfigureEvpn is used to configure "evpn <evpn-instance-name>" on the switch.
func ConfigureEvpn(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Configure EVPN",
	})

	adapter := ad.GetAdapter(sw.Model)

	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := client.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Interface Login", Error: err, Host: sw.Host}
		return
	}
	defer client.Close()
	Operation := "Mac & Arp"
	//There is only Create operation. No Update or delete of fields wihin  Mac and Arp

	log.Info("Mac & Arp")
	//fmt.Println(sw.ArpAgingTimeout,sw.MacAgingTimeout,sw.MacAgingConversationalTimeout,sw.MacMoveLimit)
	x, err := adapter.ConfigureMacAndArp(client, sw.ArpAgingTimeout, sw.MacAgingTimeout, sw.MacAgingConversationalTimeout, sw.MacMoveLimit)
	if err == nil {
		log.Infof("Mac & Arp Completed %s\n", x)
	} else {
		log.Errorf("Mac & Arp Error %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
	}

	Operation = "EVPN"
	//There is only Create operation. No Update or delete of fields wihin  EVPN
	log.Info("Configure EVPN")
	x, err = createEVPN(log, client, sw, force)
	if err == nil {
		log.Printf("EVPN Completed %s\n", x)
	} else {
		log.Errorf("EVPN Error %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
	}

	//Execute "clear bgp evpn neighbour all"
	Operation = "Clearing All BGP EVPN Neighbour"
	log.Info("Clearing All BGP EVPN Neighbour")
	err = actions.ExecuteClearBgpEvpnNeighbourAll(sw)
	if err == nil {
		log.Info("Clearing All BGP EVPN Neighbour completed\n")
	} else {
		log.Error("Clearing All BGP EVPN Neighbour Failed\n")
		errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
	}
}

func createEVPN(log *nlog.Entry, client *netconf.NetconfClient, sw *operation.ConfigSwitch, force bool) (string, error) {
	adapter := ad.GetAdapter(sw.Model)
	evpnResponse, _ := adapter.GetEvpnInstance(client)
	evpnOnSwitch := evpnResponse.Name
	if evpnOnSwitch != "" && evpnOnSwitch != sw.Fabric {
		statusMsg := fmt.Sprintf("EVPN %s already configured on switch", evpnOnSwitch)
		log.Info(statusMsg)
		if force {
			//Delete the existing Overlay Gateway
			log.Info("Delete the exising EVPN on the switch")
			adapter.DeleteEvpnInstance(client, evpnOnSwitch)
		} else {
			return statusMsg, errors.New(statusMsg)
		}
	}
	x, err := adapter.CreateEvpnInstance(client, sw.Fabric, sw.DuplicateMacTimer, sw.DuplicateMaxTimerMaxCount)
	return x, err
}

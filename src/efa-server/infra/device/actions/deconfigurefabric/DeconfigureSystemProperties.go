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

//UnconfigureSystemwideProperties is used to unconfigure the systemwide properties of a switching device
func UnconfigureSystemwideProperties(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    sw.Fabric,
		"Operation": "Unconfigure Switch Properties",
		"Switch":    sw.Host,
	})

	adapter := ad.GetAdapter(sw.Model)
	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := client.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Unconfigure Switch Properties Login", Error: err, Host: sw.Host}
		return
	}
	defer client.Close()
	Operation := "UnconfigureSystemL2Mtu"
	_, err := adapter.UnconfigureSystemL2Mtu(client)
	if err != nil {
		log.Errorf("UnConfigure System L2 MTU Failed on the device %s with an error: %s\n", sw.Host, err)
		errs <- actions.OperationError{Operation: "UnConfigure System L2 MTU", Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
	}
	Operation = "UnconfigureSystemL2Mtu"
	_, err = adapter.UnconfigureSystemIPMtu(client)
	if err != nil {
		log.Errorf("Unconfigure System IP MTU Failed on the device %s with an error: %s\n", sw.Host, err)
		errs <- actions.OperationError{Operation: "Unconfigure System IP MTU", Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
	}

	Operation = "UnconfigureMacAndArp"
	_, err = adapter.UnconfigureMacAndArp(client)
	if err != nil {
		log.Errorf("Unconfigure Mac And Arp Parameters Failed on the device %s with an error: %s\n", sw.Host, err)
		errs <- actions.OperationError{Operation: "Unconfigure Mac And Arp Parameters Failed", Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
	}
}

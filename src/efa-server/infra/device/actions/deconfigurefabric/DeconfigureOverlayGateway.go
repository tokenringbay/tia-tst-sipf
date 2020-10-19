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

//UnConfigureOverlayGateway is used to unconfigure "overlay-gateway" from the switch.
func UnConfigureOverlayGateway(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    sw.Fabric,
		"Operation": "Unconfigure Overlay Gateway",
		"Switch":    sw.Host,
	})
	adapter := ad.GetAdapter(sw.Model)
	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	client.Login()
	defer client.Close()

	log.Info("Unconfigure overlay gateway", sw.Fabric)

	Operation := "Unconfigure Overlay Gateway"
	x, err := adapter.DeleteOverlayGateway(client, sw.Fabric)
	if err == nil {
		log.Infof("Completed %s\n", x)
	} else {
		log.Errorf("Error %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
		return
	}

	Operation = "UnConfigure Anycast Gateway"
	log.Info("Started UnConfigure Anycast Gateway")
	x, err = adapter.UnconfigureAnycastGateway(client)
	if err == nil {
		log.Infof("Completed %s\n", x)
	} else {
		log.Errorf("Error %s\n", err)
		errs <- actions.OperationError{Operation: Operation, Error: errors.New(Operation + ":" + err.Error()), Host: sw.Host}
		return
	}
}

package deconfigurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/adapter/interface"
	"efa-server/infra/device/client"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//UnconfigureInterfaces is used to unconfigure the physical and logical interfaces of the switch.
func UnconfigureInterfaces(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    sw.Fabric,
		"Operation": "UnConfigure Interfaces",
		"Switch":    sw.Host,
	})

	adapter := ad.GetAdapter(sw.Model)
	client := &client.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	client.Login()
	defer client.Close()

	for _, intf := range sw.Interfaces {
		ifName := intf.InterfaceName
		ifType := intf.InterfaceType
		var err error
		if ifType == domain.IntfTypeLoopback {
			_, err = unconfigureInterfaceLoopback(adapter, client, ifName, ifType, intf.IP)
		} else {
			if intf.Donor == "" {
				_, err = unconfigureInterface(adapter, client, ifName, ifType, intf.IP)
			} else {
				_, err = unconfigureInterfaceUnnumbered(adapter, client, ifName, ifType)
			}
		}

		if err == nil {
			log.Infof("Interface %s %s with IP address %s", ifType, ifName, intf.IP)
		} else {
			msg := fmt.Sprintf("Interface %s %s with IP address %s:Error %s", ifType, ifName, intf.IP, err)
			log.Errorln(msg)
			errs <- actions.OperationError{Operation: "Configure Interface", Error: errors.New(msg), Host: sw.Host}
			return
		}
	}
	return
}

func unconfigureInterfaceLoopback(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string,
	ipAddress string) (string, error) {
	return adapter.DeleteInterfaceLoopback(client, intfName)
}

func unconfigureInterface(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string,
	ipAddress string) (string, error) {
	return adapter.UnconfigureInterfaceNumbered(client, intfType, intfName, ipAddress)
}

func unconfigureInterfaceUnnumbered(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string) (string, error) {
	return adapter.UnconfigureInterfaceUnnumbered(client, intfType, intfName)
}

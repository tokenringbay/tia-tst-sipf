package configurefabric

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

//ConfigureInterfaces is used to configure the physical and logical interfaces of the switch.
func ConfigureInterfaces(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Configure Interfaces",
	})

	adapter := ad.GetAdapter(sw.Model)
	netconfClient := &client.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := netconfClient.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Interface Login", Error: err, Host: sw.Host}
		return
	}
	defer netconfClient.Close()

	for _, intf := range sw.Interfaces {
		ifName := intf.InterfaceName
		ifType := intf.InterfaceType
		var err error
		if ifType == domain.IntfTypeLoopback {
			_, err = configureInterfaceLoopback(adapter, netconfClient, ifName, ifType, intf.IP, intf.ConfigType)
		} else {
			if intf.Donor == "" {
				_, err = configuredInterface(adapter, netconfClient, ifName, ifType, intf.IP,
					intf.ConfigType, intf.Description)
			} else {
				_, err = configureUnnumberedInterface(adapter, netconfClient, ifName, ifType, intf.Donor, intf.DonorPort,
					intf.ConfigType)
			}
		}

		if err == nil {
			log.Infof("Interface %s %s with IP address %s", ifType, ifName, intf.IP)
		} else {
			msg := fmt.Sprintf("Interface %s %s with IP address %s: %s", ifType, ifName, intf.IP, err)
			log.Errorln(msg)
			errs <- actions.OperationError{Operation: "Configure Interface", Error: errors.New(msg), Host: sw.Host}
		}
	}

}

func configureUnnumberedInterface(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string,
	donorType string, donorName string, configType string) (string, error) {
	if configType == domain.ConfigCreate {
		return adapter.ConfigureInterfaceUnnumbered(client, intfType, intfName, donorType, donorName)
	}
	if configType == domain.ConfigUpdate {
		adapter.UnconfigureInterfaceUnnumbered(client, intfType, intfName)
		return adapter.ConfigureInterfaceUnnumbered(client, intfType, intfName, donorType, donorName)
	}

	if configType == domain.ConfigDelete {
		adapter.UnconfigureInterfaceUnnumbered(client, intfType, intfName)
	}
	return "", nil
}

func configureInterfaceLoopback(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string,
	ipaddress string, configType string) (string, error) {
	if configType == domain.ConfigUpdate || configType == domain.ConfigCreate {
		return adapter.ConfigureInterfaceLoopback(client, intfName, ipaddress)
	}
	if configType == domain.ConfigDelete {
		//will only deconfigure IP address, not delete the loopback
		return adapter.DeleteInterfaceLoopback(client, intfName)
	}
	return "", nil
}

func configuredInterface(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string,
	ipAddress string, configType string,
	description string) (string, error) {
	if configType == domain.ConfigCreate {
		return adapter.ConfigureInterfaceNumbered(client, intfType, intfName, ipAddress, description)
	}
	if configType == domain.ConfigUpdate {
		//Physical Interface requires IP address removed and then re-added
		if intfMap, err := adapter.GetInterface(client, intfType, intfName); err == nil {
			if intfMap["address"] != ipAddress {
				adapter.UnconfigureInterfaceNumbered(client, intfType, intfName, intfMap["address"])
				adapter.UnconfigureInterfaceUnnumbered(client, intfType, intfName)
			}
		}
		return adapter.ConfigureInterfaceNumbered(client, intfType, intfName, ipAddress, description)
	}

	if configType == domain.ConfigDelete {
		adapter.UnconfigureInterfaceNumbered(client, intfType, intfName, ipAddress)
	}
	return "", nil
}

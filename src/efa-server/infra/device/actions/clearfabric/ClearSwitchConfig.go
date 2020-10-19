package clearfabric

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
	"sync"

	"efa-server/domain"
	"efa-server/infra/device/adapter/interface"
	"net"
)

func clearSwitchConfig(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.ClearSwitchDetail,
	errs chan actions.OperationError) {
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

	log.Infoln("Delete Overlay Gateway")
	if err := deleteOverlayGateway(log, adapter, netconfClient); err != nil {
		errs <- actions.OperationError{Operation: "Delete Overlay Gateway Failed", Error: err, Host: sw.Host}
	}

	log.Infoln("Delete AnyCast Gateway Mac")
	if err := deleteAnyCastGatewayMac(log, adapter, netconfClient); err != nil {
		errs <- actions.OperationError{Operation: "Delete  AnyCast Gateway Mac Failed", Error: err, Host: sw.Host}
	}

	log.Infoln("Delete EVPN")
	if err := deleteEVPN(log, adapter, netconfClient); err != nil {
		errs <- actions.OperationError{Operation: "Delete EVPN Failed", Error: err, Host: sw.Host}
	}
	log.Infoln("Delete Router BGP")
	if err := deleteRouterBGP(adapter, netconfClient); err != nil {
		errs <- actions.OperationError{Operation: "Delete Router BGP Failed", Error: err, Host: sw.Host}
	}
	log.Infoln("Delete Router ID")
	_, err := adapter.UnconfigureRouterID(netconfClient)
	if err != nil {
		log.Errorf("UnConfigure Router ID Failed: %s\n", err)
		errs <- actions.OperationError{Operation: "UnConfigure Router ID", Error: err, Host: sw.Host}
	}
	log.Infoln("Delete System L2 MTU")
	_, err = adapter.UnconfigureSystemL2Mtu(netconfClient)
	if err != nil {
		log.Errorf("UnConfigure System L2 MTU Failed: %s\n", err)
		errs <- actions.OperationError{Operation: "UnConfigure System L2 MTU", Error: err, Host: sw.Host}
	}
	log.Infoln("Delete System IP MTU")
	_, err = adapter.UnconfigureSystemIPMtu(netconfClient)
	if err != nil {
		log.Errorf("Unconfigure System IP MTU Failed: %s\n", err)
		errs <- actions.OperationError{Operation: "Unconfigure System IP MTU", Error: err, Host: sw.Host}
	}

	/*
		log.Infoln("Delete Device Host Name")
		_, err = adapter.UnconfigureSwitchHostName(netconfClient)
		if err != nil {
			log.Errorf("Unconfigure Device Host Name Failed: %s\n", err)
			errs <- actions.OperationError{Operation: "Unconfigure Device Host Name Failed", Error: err, Host: sw.Host}
		}
	*/

	log.Infoln("Delete Mac And Arp Parameters")
	_, err = adapter.UnconfigureMacAndArp(netconfClient)
	if err != nil {
		log.Errorf("Unconfigure Delete Mac And Arp Parameters Failed: %s\n", err)
		errs <- actions.OperationError{Operation: "Unconfigure Mac And Arp Parameters Failed", Error: err, Host: sw.Host}
	}

	log.Infoln("Delete IP Routes")
	ipRouteMap, err := adapter.GetIPRoutes(netconfClient)

	_, ipnet, _ := net.ParseCIDR(sw.LoopBackIPRange)
	for loopbackIP, veIP := range ipRouteMap {
		ip, _, _ := net.ParseCIDR(loopbackIP)
		if ipnet.Contains(ip) {
			_, err := adapter.DeConfigureIPRoute(netconfClient, loopbackIP, veIP)
			if err != nil {
				log.Errorf("Unconfigure IP Route Failed: %s\n", err)
				errs <- actions.OperationError{Operation: "Unconfigure IP Route", Error: err, Host: sw.Host}
			}
		}
	}

	deleteInterfaces(log, adapter, netconfClient, &sw, errs)

}

func deleteInterfaces(log *nlog.Entry, adapter interfaces.Switch, client *client.NetconfClient,
	sw *operation.ClearSwitchDetail, errs chan actions.OperationError) {

	for _, intf := range sw.Interfaces {
		ifName := intf.InterfaceName
		ifType := intf.InterfaceType
		var err error
		var x string
		if ifType == domain.IntfTypeLoopback {
			log.Infoln("Delete loopback interface ", ifName)
			x, err = unconfigureLoopbackInterface(adapter, client, ifName, ifType)
		} else {
			//Try to fetch the interface details of the interface before clear
			if intfMap, err := adapter.GetInterface(client, ifType, ifName); err == nil {
				//fmt.Println("Clear", client.Host, ifType, ifName, intfMap)
				log.Infoln("Delete interface ", ifType, ifName, intfMap)
				if intfMap["address"] != "" {
					//Address is populated for Numbered interfaces
					adapter.UnconfigureInterfaceNumbered(client, ifType, ifName, intfMap["address"])
				}
				if intfMap["description"] != "" {
					//remove stale description from the lldp links
					adapter.UnconfigureInterfaceDesc(client, ifType, ifName)
				}
				/*if intfMap["speed"] != "" {
					//remove stale speed from the lldp links
					adapter.UnconfigureInterfaceSpeed(client, ifType, ifName)
				}*/
				if intfMap["donor_type"] == "loopback" {
					adapter.UnconfigureInterfaceUnnumbered(client, ifType, ifName)
				}
			}
		}

		if err != nil {
			msg := fmt.Sprintf("Interface %s %s with IP address %s:Error %s", ifType, ifName, intf.IP, err)
			log.Errorln(msg, x)
			errs <- actions.OperationError{Operation: "UnConfigure Interface", Error: errors.New(msg), Host: sw.Host}
		}
	}
}

func deleteOverlayGateway(log *nlog.Entry, adapter interfaces.Switch, client *client.NetconfClient) error {
	//Check if Overlay Gateway with some other name exists

	ovgResponse, _ := adapter.GetOverlayGateway(client)
	gwOnSwitch := ovgResponse.Name
	if gwOnSwitch != "" {
		statusMsg := fmt.Sprintf("Overlay gateway %s already configured on switch", gwOnSwitch)
		log.Info(statusMsg)

		//Delete the existing Overlay Gateway
		log.Info("Delete the exising Overlay Gateway on the switch")
		_, err := adapter.DeleteOverlayGateway(client, gwOnSwitch)
		return err
	}
	return nil
}

func deleteAnyCastGatewayMac(log *nlog.Entry, adapter interfaces.Switch, client *client.NetconfClient) error {
	_, err := adapter.UnconfigureAnycastGateway(client)
	return err
}

func deleteEVPN(log *nlog.Entry, adapter interfaces.Switch, client *client.NetconfClient) error {
	evpnResponse, _ := adapter.GetEvpnInstance(client)
	evpnOnSwitch := evpnResponse.Name
	if evpnOnSwitch != "" {
		statusMsg := fmt.Sprintf("EVPN %s already configured on switch", evpnOnSwitch)
		log.Info(statusMsg)

		log.Info("Delete the exising EVPN on the switch")
		_, err := adapter.DeleteEvpnInstance(client, evpnOnSwitch)
		return err
	}
	return nil
}

func deleteRouterBGP(adapter interfaces.Switch, client *client.NetconfClient) error {
	if err := adapter.IsRouterBgpPresent(client); err == nil {
		_, err := adapter.UnconfigureRouterBgp(client)
		if err != nil {
			return err
		}
		_, err = adapter.UnConfigureNumberedRouteMap(client)
		if err != nil {
			return err
		}
	}
	return nil
}

func unconfigureLoopbackInterface(adapter interfaces.Switch, client *client.NetconfClient, intfName string, intfType string) (string, error) {
	if intfMap, err := adapter.GetInterface(client, intfType, intfName); err == nil {
		if intfMap["Name"] != "" {
			return adapter.DeleteInterfaceLoopback(client, intfName)
		}
	}
	return "", nil
}

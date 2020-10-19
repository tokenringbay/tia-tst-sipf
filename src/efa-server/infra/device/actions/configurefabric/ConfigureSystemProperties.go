package configurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"efa-server/usecase"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//ConfigureSystemwideProperties is used to configure the systemwide properties of a switching device
func ConfigureSystemwideProperties(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Configure Switch Properties",
	})
	adapter := ad.GetAdapter(sw.Model)
	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := client.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Switch Properties Login", Error: err, Host: sw.Host}
		return
	}
	defer client.Close()
	if _, err := adapter.ConfigureSystemL2Mtu(client, sw.Mtu); err != nil {
		log.Errorf("Configure System L2 MTU Failed on the device %s with an error: %s\n", sw.Host, err)
		errs <- actions.OperationError{Operation: "Configure System L2 MTU", Error: err, Host: sw.Host}
	}
	if _, err := adapter.ConfigureSystemIPMtu(client, sw.IPMtu); err != nil {
		log.Errorf("Configure System IP MTU Failed on the device %s with an error: %s\n", sw.Host, err)
		errs <- actions.OperationError{Operation: "Configure System IP MTU", Error: err, Host: sw.Host}
	}

	if sw.ConfigureOverlayGateway == "Yes" && sw.Role == usecase.LeafRole {
		//There is only creation of anycast gateway mac, no update or delete of anycast gateway mac
		Operation := "Configure Anycast Gateway"
		log.Infof("Started Anycast Gateway IPV4 Mac = %s IPV6 Mac = %s", sw.AnycastMac, sw.IPV6AnycastMac)

		x, err := adapter.ConfigureAnycastGateway(client, sw.AnycastMac, sw.IPV6AnycastMac)
		if err == nil {
			log.Infof("Completed %s\n", x)
		} else {
			log.Errorf("Error %s\n", err)
			errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
		}
	}

}

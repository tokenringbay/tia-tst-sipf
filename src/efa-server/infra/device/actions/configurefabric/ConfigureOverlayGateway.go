package configurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/adapter/interface"
	netconf "efa-server/infra/device/client"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

//ConfigureOverlayGateway is used to configure "overlay-gateway" on the switch.
func ConfigureOverlayGateway(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, force bool, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Configure Overlay Gateway",
	})

	adapter := ad.GetAdapter(sw.Model)
	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := client.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Interface Login", Error: err, Host: sw.Host}
		return
	}
	defer client.Close()

	Operation := "Configure Overlay Gateway"

	if sw.MctSecondaryNode == false {
		log.Infof("Configure Loopback=%s", sw.VtepLoopbackPortNumber)
		//There is only Create operation. No Update or delete of fields wihin Overlay Gateway
		x, err := createOverlayGateway(log, client, adapter, sw.Fabric, "layer2-extension",
			sw.VtepLoopbackPortNumber, strconv.FormatBool(sw.VlanVniAutoMap), force)

		if err == nil {
			log.Infof("Completed %s\n", x)
		} else {
			log.Errorf("Error %s\n", err)
			errs <- actions.OperationError{Operation: Operation, Error: err, Host: sw.Host}
		}
	} else {
		log.Infof("switch %s detected as MCT swcondary Hence overlaygateway is Pushed only through principal", sw.Host)
	}

}

func createOverlayGateway(log *nlog.Entry, client *netconf.NetconfClient, adapter interfaces.Switch, gwName string, gwType string,
	loopbackID string, mapVniAuto string, force bool) (string, error) {
	//Check if Overlay Gateway with some other name exists
	ovgResponse, _ := adapter.GetOverlayGateway(client)
	gwOnSwitch := ovgResponse.Name

	if gwOnSwitch != "" && gwOnSwitch != gwName {
		statusMsg := fmt.Sprintf("Overlay gateway %s already configured on switch", gwOnSwitch)
		log.Info(statusMsg)
		if force {
			//Delete the existing Overlay Gateway
			log.Info("Delete the exising Overlay Gateway on the switch")
			adapter.DeleteOverlayGateway(client, gwOnSwitch)
		} else {
			return statusMsg, errors.New(statusMsg)
		}
	}
	//Create Overlay Gatway
	return adapter.CreateOverlayGateway(client, gwName, gwType,
		loopbackID, mapVniAuto)
}

/*
func isOverlayGatewayConfigPushNeeded(sw *operation.ConfigSwitch,  adapter interfaces.Switch,client *netconf.NetconfClient) (bool, error) {
	// Check the management cluster status.
	//If the node is standalone or principal node, then configure overlay-gateway.
	_, _, principalNode, err := adapter.GetManagementClusterStatus(client)
	return (principalNode == "" || principalNode == sw.Host), err
}
*/

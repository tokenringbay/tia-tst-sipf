package configurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"efa-server/infra/device/actions/deconfigurefabric"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"efa-server/usecase"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//ConfigureSwitch is used to configure both underlay and overlay on a switch.
func ConfigureSwitch(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.ConfigSwitch,
	force bool, persist bool, fabricError chan actions.OperationError) {
	defer fabricGate.Done()
	ctx = context.WithValue(ctx, appcontext.DeviceName, sw.Host)
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Configure Switch",
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go ConfigureSystemwideProperties(ctx, &wg, &sw, force, fabricError)
	wg.Wait()

	wg.Add(1)
	go ConfigureInterfaces(ctx, &wg, &sw, force, fabricError)
	wg.Wait()

	wg.Add(1)
	go ConfigureBGP(ctx, &wg, &sw, force, fabricError)
	wg.Wait()

	log.Infoln("MCT Data plane sending BGP unconfigure ", sw.UnconfigureMCTBGPNeighbors)
	wg.Add(1)
	go deconfigurefabric.UnconfigureDataPlaneCluster(ctx, &wg, &sw.UnconfigureMCTBGPNeighbors, force, fabricError)
	wg.Wait()

	if sw.Role == usecase.LeafRole {
		wg.Add(1)
		go ConfigureEvpn(ctx, &wg, &sw, force, fabricError)
	}
	wg.Wait()

	log.Infoln("MCT Data plane sending BGP configure ", sw.ConfigureMCTBGPNeighbors)
	wg.Add(1)
	go ConfigureDataPlaneCluster(ctx, &wg, &sw.ConfigureMCTBGPNeighbors, force, fabricError)
	wg.Wait()

	/*if persist {
		wg.Add(1)
		go persistConfig(ctx, &wg, &sw, fabricError)

	}*/
	log.Info("Switch Waiting for Child")
	wg.Wait()
	log.Info("Switch Wait Completed")

}

func persistConfig(ctx context.Context, wg *sync.WaitGroup, sw *operation.ConfigSwitch, errs chan actions.OperationError) {
	defer wg.Done()
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Persist Config",
	})

	adapter := ad.GetAdapter(sw.Model)
	client := &netconf.NetconfClient{Host: sw.Host, User: sw.UserName, Password: sw.Password}
	if err := client.Login(); err != nil {
		errs <- actions.OperationError{Operation: "Configure Switch Properties Login", Error: err, Host: sw.Host}
		return
	}
	defer client.Close()
	if _, err := adapter.PersistConfig(client); err != nil {
		msg := fmt.Sprintf("Persist Config Failed %s:Error %s", sw.Host, err)
		log.Errorln(msg)
		errs <- actions.OperationError{Operation: "Configure Switch Properties", Error: errors.New(msg), Host: sw.Host}
	}

}

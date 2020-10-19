package deconfigurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"efa-server/usecase"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//UnconfigureSwitch is used to unconfigure both underlay and overlay on a switch.
func UnconfigureSwitch(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.ConfigSwitch, force bool, fabricError chan actions.OperationError, persist bool) {
	defer fabricGate.Done()

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    sw.Fabric,
		"Operation": "UnConfigure Switch",
		"Switch":    sw.Host,
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go UnconfigureSystemwideProperties(ctx, &wg, &sw, fabricError)
	wg.Wait()

	wg.Add(1)
	go UnconfigureInterfaces(ctx, &wg, &sw, fabricError)
	wg.Wait()

	//Make  UnconfigureDataPlaneCluster available in deconfigure context
	wg.Add(1)
	go UnconfigureDataPlaneCluster(ctx, &wg, &sw.UnconfigureMCTBGPNeighbors, force, fabricError)
	wg.Wait()

	wg.Add(1)
	go UnconfigureBGP(ctx, &wg, &sw, fabricError)

	if sw.Role == usecase.LeafRole {
		wg.Add(1)
		go UnconfigureEvpn(ctx, &wg, &sw, fabricError)
		wg.Add(1)
		go UnConfigureOverlayGateway(ctx, &wg, &sw, force, fabricError)
	}

	wg.Wait()
	if persist {
		wg.Add(1)
		go persistConfig(ctx, &wg, &sw, fabricError)

	}

	wg.Wait()
	log.Info("Waiting for Child")

	log.Info("Completed")
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
		fmt.Println(err)
		msg := fmt.Sprintf("Persist Config Failed %s:Error %s", sw.Host, err)
		log.Errorln(msg)
		errs <- actions.OperationError{Operation: "Configure Switch Properties", Error: errors.New(msg), Host: sw.Host}
	}
	log.Infof("persist config for device %s", sw.Host)
}

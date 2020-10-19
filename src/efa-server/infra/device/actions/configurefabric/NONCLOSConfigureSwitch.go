package configurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"efa-server/infra/device/actions/deconfigurefabric"
	"efa-server/usecase"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//ConfigureNonCLOSSwitch is used to configure both underlay and overlay on a switch.
func ConfigureNonCLOSSwitch(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.ConfigSwitch,
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
	go ConfigureNonClosBGP(ctx, &wg, &sw, force, fabricError)
	wg.Wait()

	log.Infoln("MCT Data plane sending BGP unconfigure ", sw.UnconfigureMCTBGPNeighbors)
	wg.Add(1)
	go deconfigurefabric.UnconfigureDataPlaneCluster(ctx, &wg, &sw.UnconfigureMCTBGPNeighbors, force, fabricError)
	wg.Wait()

	if sw.Role == usecase.RackRole {
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

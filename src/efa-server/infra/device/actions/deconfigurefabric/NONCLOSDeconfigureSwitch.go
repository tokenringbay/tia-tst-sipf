package deconfigurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"efa-server/usecase"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//UnconfigureNonCLOSSwitch is used to unconfigure both underlay and overlay on a switch.
func UnconfigureNonCLOSSwitch(ctx context.Context, fabricGate *sync.WaitGroup, sw operation.ConfigSwitch, force bool, fabricError chan actions.OperationError, persist bool) {
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
	go UnconfigureNonClosBGP(ctx, &wg, &sw, fabricError)

	if sw.Role == usecase.RackRole {
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

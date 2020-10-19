package deconfigurefabric

import (
	nlog "github.com/sirupsen/logrus"

	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"sync"
)

//CleanupDevicesInNonCLOSFabric cleans up NON CLOS configuration from a collection of devices used by Delete Fabric
func CleanupDevicesInNonCLOSFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    config.FabricName,
		"Operation": "Cleanup Devices In Fabric",
	})
	Errors := make([]actions.OperationError, 0)

	log.Info("Started")
	if len(config.MctCluster) > 0 && len(config.MctCluster[domain.MctDelete]) > 0 {
		Errors = DeConfigureMctClusters(ctx, domain.MctDelete, config.MctCluster[domain.MctDelete], force)
		if len(Errors) != 0 {
			log.Errorln("Deconfiguring MCT Cluster Failed - ", Errors)
			return Errors
		}
	}

	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)

	for _, configSwitch := range config.Hosts {
		fabricGate.Add(1)
		go UnconfigureNonCLOSSwitch(ctx, &fabricGate, configSwitch, force, fabricErrors, persist)
	}
	log.Info("Waiting for  cleanup devices to Complete")

	//Go Function to Wait and Close
	go func() {
		fabricGate.Wait()
		close(fabricErrors)

	}()
	log.Info("Wait for cleanup devices Completed")
	for err := range fabricErrors {
		//log.Error(err)
		Errors = append(Errors, err)
	}
	return Errors
}

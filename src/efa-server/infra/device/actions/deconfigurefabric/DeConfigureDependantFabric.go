package deconfigurefabric

import (
	nlog "github.com/sirupsen/logrus"

	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"sync"
)

//CleanupDependantDevicesInFabric cleans up interfaces and BGP Neighbors from the dependant devices
func CleanupDependantDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    config.FabricName,
		"Operation": "Cleanup Devices In Fabric",
	})
	Errors := make([]actions.OperationError, 0)

	log.Info("Started")

	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)

	for _, configSwitch := range config.Hosts {
		fabricGate.Add(1)
		go UnconfigureDependantSwitch(ctx, &fabricGate, configSwitch, force, fabricErrors, persist)
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

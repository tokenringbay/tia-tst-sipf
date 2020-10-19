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

//CleanupDevicesInFabric cleans up IP Fabric configuration from a collection of devices used by Delete Fabric
func CleanupDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
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
		go UnconfigureSwitch(ctx, &fabricGate, configSwitch, force, fabricErrors, persist)
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

//DeConfigureMctClusters "deconfigures" MCT cluster on a collection of devices
func DeConfigureMctClusters(ctx context.Context, opcode uint, MctCluster []operation.ConfigCluster, force bool) []actions.OperationError {
	MctOperation := [4]string{1: "configure", 2: "deconfigure", 3: "update"}
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    MctCluster[0].FabricName,
		"Operation": "De Configure MCT Cluster",
	})
	log.Infoln("MCT Cluster Start opcode - ", MctOperation[opcode])
	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	mctErrors := make(chan actions.OperationError)
	//For each Cluster Invoke
	for iter := range MctCluster {
		configCluster := MctCluster[iter]
		log.Infoln("Printing MCT Cluster Details Before Sending - ", configCluster)
		fabricGate.Add(1)
		switch opcode {
		case domain.MctDelete:
			go UnconfigureManagementCluster(ctx, &fabricGate, &configCluster, force, mctErrors)
		}
	}

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		close(mctErrors)

	}()

	log.Infof("Wait FOR MCT %s opcode to Complete", MctOperation[opcode])

	//Check for errors in the sub-action
	for err := range mctErrors {
		Errors = append(Errors, err)
	}
	return Errors
}

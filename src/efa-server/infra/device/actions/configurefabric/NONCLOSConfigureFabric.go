package configurefabric

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"sync"
)

//ConfigureNonClosFabric Configures IP Fabric excluding MCT Cluster
func ConfigureNonClosFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	clusterStatus := make(map[string]bool)
	log := appcontext.Logger(ctx)

	log.Info("Start")

	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)

	//For each Switch Invoke Configure Switch
	for iter := range config.Hosts {
		configSwitch := config.Hosts[iter]
		fabricGate.Add(1)
		go ConfigureNonCLOSSwitch(ctx, &fabricGate, configSwitch, force, persist, fabricErrors)

	}

	log.Info("Waiting for Switch Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		log.Info("Wait Completed")
		close(fabricErrors)

	}()

	//Check for errors in the sub-action
	for err := range fabricErrors {
		Errors = append(Errors, err)
	}

	if len(Errors) > 0 {
		log.Error("Configure Fabric Failed")
		return Errors
	}

	//Configure MCT Cluster
	if MCTErrors := configureMCTCluster(ctx, config, force, clusterStatus); len(MCTErrors) > 0 {
		log.Errorln("Error Configuring MCT cluster - ", MCTErrors)
		log.Error("Configure Fabric Failed")
		return MCTErrors
	}

	var overlayGate sync.WaitGroup
	overlayErrors := make(chan actions.OperationError, 1)
	//For each Switch Invoke Configure Overlay
	for iter := range config.Hosts {
		sw := config.Hosts[iter]
		markIfSwitchIsMCTSecondary(&sw, clusterStatus)
		if sw.Role == "Rack" && sw.ConfigureOverlayGateway == "Yes" && sw.MctSecondaryNode == false {
			overlayGate.Add(1)
			go ConfigureOverlayGateway(ctx, &overlayGate, &sw, force, overlayErrors)
		}
	}
	log.Info("Waiting for Overlay Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		overlayGate.Wait()
		log.Info("overlay Wait Completed")
		close(overlayErrors)

	}()
	for err := range overlayErrors {
		Errors = append(Errors, err)
	}

	// save the configs on all the devices
	if persist {
		var saveConfig sync.WaitGroup
		saveConfigErrors := make(chan actions.OperationError, 1)
		//For each Switch Invoke Configure Overlay
		for iter := range config.Hosts {
			sw := config.Hosts[iter]
			saveConfig.Add(1)
			go persistConfig(ctx, &saveConfig, &sw, saveConfigErrors)
		}

		//Utility go-routine waiting for actions to complete
		go func() {
			saveConfig.Wait()
			log.Info("saving configuration wait Completed")
			close(saveConfigErrors)

		}()
		for err := range saveConfigErrors {
			Errors = append(Errors, err)
		}
	}

	if len(Errors) > 0 {
		log.Error("Configure Fabric Failed")
		return Errors
	}
	log.Debug("Configure Fabric Completed...")
	return Errors
}

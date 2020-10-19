package clearfabric

import (
	"bytes"
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"errors"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//ClearFabric clears up Configuration from list of devices
func ClearFabric(ctx context.Context, ClearFabricRequest operation.ClearFabricRequest) error {
	log := appcontext.Logger(ctx)

	log.Info("Start")

	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)

	//For each Switch Invoke Configure Switch
	for _, clearSwitchDetail := range ClearFabricRequest.Hosts {
		fabricGate.Add(1)
		go clearSwitchConfig(ctx, &fabricGate, clearSwitchDetail, fabricErrors)
	}

	log.Info("Waiting for Switch Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		close(fabricErrors)

	}()

	log.Info("Wait Completed")

	//Check for errors in the sub-action
	for err := range fabricErrors {
		Errors = append(Errors, err)
	}

	if len(Errors) == 0 {
		return nil
	}

	log.Error("Clear Config Failed")
	var buffer bytes.Buffer
	for i := 0; i < len(Errors); i++ {
		buffer.WriteString(fmt.Sprintf("On the device[%s], the operation[%s] has failed, with the reason:[%s]\n", Errors[i].Host, Errors[i].Operation, Errors[i].Error.Error()))
	}

	return errors.New(buffer.String())
}

//ClearMctClusters clears up the MCT cluster config from list of devices
func ClearMctClusters(ctx context.Context, FabricName string, clusters []operation.ConfigCluster) error {
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"Operation": "Clear Config",
	})

	log.Info("MCT Cluster Clear Start")

	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	mctErrors := make(chan actions.OperationError, 2)

	//Clear MCT Cluster from Switch
	for index := range clusters {
		cluster := clusters[index]
		fabricGate.Add(1)
		go ClearManagementCluster(ctx, &fabricGate, &cluster, false, mctErrors)
	}
	log.Info("Waiting For Switch Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		close(mctErrors)

	}()

	log.Info("Wait Completed")

	//Check for errors in the sub-action
	for err := range mctErrors {
		Errors = append(Errors, err)
	}

	if len(Errors) == 0 {
		return nil
	}

	log.Error("clear Fabric Failed")
	var buffer bytes.Buffer
	for i := 0; i < len(Errors); i++ {
		buffer.WriteString(fmt.Sprintf("	On the device[%s], the operation[%s] has failed, with the reason:[%s]\n", Errors[i].Host, Errors[i].Operation, Errors[i].Error.Error()))
	}

	return errors.New(buffer.String())
}

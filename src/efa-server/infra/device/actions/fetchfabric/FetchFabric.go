package fetchfabric

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

//FetchFabric fetches config from a set of devices of a given fabric.
func FetchFabric(ctx context.Context, FabricRequest operation.FabricFetchRequest) (operation.FabricFetchResponse, error) {
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    FabricRequest.FabricName,
		"Operation": "Configure Fabric",
	})

	log.Info("Start")

	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)
	switchResponses := make(chan operation.ConfigSwitchResponse, len(FabricRequest.Hosts))

	//For each Switch Invoke Fetch Switch
	for _, switchIdentity := range FabricRequest.Hosts {
		fabricGate.Add(1)
		go FetchSwitchConfig(ctx, &fabricGate, switchIdentity, switchResponses, fabricErrors)
	}

	log.Info("Waiting for Switch Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		close(fabricErrors)
		close(switchResponses)

	}()

	log.Info("Wait Completed")

	//Check for errors in the sub-action
	for err := range fabricErrors {
		Errors = append(Errors, err)
	}

	ActionFabricFetchResponse := operation.FabricFetchResponse{FabricName: FabricRequest.FabricName}
	ActionFabricFetchResponse.SwitchResponse = make([]operation.ConfigSwitchResponse, 0)
	for resp := range switchResponses {
		ActionFabricFetchResponse.SwitchResponse = append(ActionFabricFetchResponse.SwitchResponse, resp)
	}

	if len(Errors) == 0 {
		return ActionFabricFetchResponse, nil
	}

	log.Error("Fetch Fabric Failed")
	var buffer bytes.Buffer
	for i := 0; i < len(Errors); i++ {
		buffer.WriteString(fmt.Sprintf("	On the device[%s], the operation[%s] has failed, with the reason:[%s]\n", Errors[i].Host, Errors[i].Operation, Errors[i].Error.Error()))
	}

	return ActionFabricFetchResponse, errors.New(buffer.String())
}

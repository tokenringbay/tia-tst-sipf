package handler

import (
	"net/http"

	"efa-server/infra"
	"efa-server/infra/logging"

	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/infra/constants"
	"efa-server/infra/rest/generated/server/go"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

//DebugClear provides REST handler for handling
//POST Request for Clearing the Config
func DebugClear(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()

	CommandName := "debug clear"
	success := true
	statusMsg := ""

	var DebugClearRequest Restmodel.DebugClearRequest

	alog := logging.AuditLog{Request: &logging.Request{Command: CommandName}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)

	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &DebugClearRequest)
	if err != nil {
		success = false
		return
	}
	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"Devices": DebugClearRequest.IpAddress,
	}
	alog.LogMessageReceived()

	fabricType := domain.CLOSFabricType
	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(constants.DefaultFabric); err == nil {
		if FabricProperties, properr := infra.GetUseCaseInteractor().Db.GetFabricProperties(Fabric.ID); properr == nil {
			fabricType = FabricProperties.FabricType
		}
	}
	ctx = context.WithValue(ctx, appcontext.FabricType, fabricType)

	err = checkForDevicesinFabric(constants.DefaultFabric, DebugClearRequest.IpAddress)

	//Make a call to fetch Fabric Config the Fabric
	if err == nil {
		err = infra.GetUseCaseInteractor().AddDevicesAndClearFabric(ctx, DebugClearRequest.IpAddress, DebugClearRequest.IpAddress,
			DebugClearRequest.Username, DebugClearRequest.Password)
	}

	//Indicating there is an overall Failure
	if err != nil {
		statusMsg = fmt.Sprintf("%s Failed.", CommandName)
		success = false
		http.Error(w, "",
			http.StatusInternalServerError)

		OpenAPIError := swagger.ErrorModel{Message: err.Error()}
		bytess, _ := json.Marshal(&OpenAPIError)
		w.Write(bytess)

	} else {
		//Send  Fabric config Show Response
		statusMsg = fmt.Sprintf("%s Succeeded.", CommandName)
		OpenAPIResp := swagger.DebugClearResponse{Status: "Successful"}

		//Write Success Structure to the Body
		bytess, _ := json.Marshal(&OpenAPIResp)
		w.Write(bytess)
	}

}

func checkForDevicesinFabric(FabricName string, ipAddressList []string) error {
	presentDevices := make([]string, 0)
	for _, ip := range ipAddressList {
		if _, err := infra.GetUseCaseInteractor().Db.GetDevice(FabricName, ip); err == nil {
			presentDevices = append(presentDevices, ip)
		}
	}
	if len(presentDevices) > 0 {
		msg := fmt.Sprintf("Devices %s already present as part of fabric,"+
			"Use deconfigure command to delete the devices from the fabric", strings.Join(presentDevices, ","))
		return errors.New(msg)
	}
	return nil
}

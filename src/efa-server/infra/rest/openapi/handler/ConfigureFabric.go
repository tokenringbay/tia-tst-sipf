package handler

import (
	"net/http"

	"bytes"
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/logging"
	"efa-server/infra/rest/generated/server/go"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"strconv"
)

//ConfigureFabric provides REST handler for handling
//POST Request for configuring the Fabric
func ConfigureFabric(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	success := true
	statusMsg := ""
	var Message string

	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:ConfigureFabric"}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)
	//Extract the Parameters
	vars := mux.Vars(r)
	FabricName := vars["fabric_name"]
	Persist := vars["persist"]
	Force := vars["force"]

	PersistBool, _ := strconv.ParseBool(Persist)

	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"FabricName": FabricName,
		"Persist":    Persist,
		"Force":      Force,
	}
	alog.LogMessageReceived()

	fabricType := domain.CLOSFabricType
	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(constants.DefaultFabric); err == nil {
		if FabricProperties, properr := infra.GetUseCaseInteractor().Db.GetFabricProperties(Fabric.ID); properr == nil {
			fabricType = FabricProperties.FabricType
		}
	}
	ctx = context.WithValue(ctx, appcontext.FabricType, fabricType)
	//Make a call to Configure the Fabric
	response, err := infra.GetUseCaseInteractor().ConfigureFabric(ctx, FabricName, false, PersistBool)

	//Indicating there is an overall Failure
	if err != nil {
		http.Error(w, statusMsg,
			http.StatusInternalServerError)

		//Buffer for writing messages to the Log
		var buffer bytes.Buffer
		StatusModelList := make([]Restmodel.DeviceStatusModel, 0, len(response.Errors))
		for _, ConfigureError := range response.Errors {
			buffer.WriteString(fmt.Sprintf("Configuration of device with ip-address = %s [Failed]\n", ConfigureError.Host))

			StatusModel := Restmodel.DeviceStatusModel{IpAddress: ConfigureError.Host, Status: "Failed"}
			//Format the error Message

			if ConfigureError.Error != nil {
				Message = fmt.Sprintf("Operation[%s] has failed with the reason:%s\n",
					ConfigureError.Operation, ConfigureError.Error.Error())
			} else {
				Message = fmt.Sprintf("Operation[%s] has failed, with unknown reason", ConfigureError.Operation)
			}
			buffer.WriteString(Message)
			StatusModel.Error_ = []Restmodel.ErrorModel{Restmodel.ErrorModel{Message: Message}}

			StatusModelList = append(StatusModelList, StatusModel)

		}

		//Write StatusModel to the Body
		bytess, _ := json.Marshal(&StatusModelList)
		w.Write(bytess)

		success = false
		buffer.WriteString("Configure Fabric Failed\n")
		statusMsg = buffer.String()
	} else {
		//Send Configure Fabric Response
		statusMsg = "Configure Fabric Succeeded"
		OpenAPIResp := swagger.ConfigureFabricResponse{FabricName: FabricName, Status: "Successful"}

		//Write Success Structure to the Body
		bytess, _ := json.Marshal(&OpenAPIResp)
		w.Write(bytess)
	}

}

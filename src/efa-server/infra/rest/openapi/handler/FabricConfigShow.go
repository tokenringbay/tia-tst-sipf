package handler

import (
	"net/http"

	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/logging"
	"efa-server/infra/rest/generated/server/go"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
)

//FabricConfigShow provides REST handler for handling
//GET Request for fetching the config from device
func FabricConfigShow(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	CommandName := "config show"
	success := true
	statusMsg := ""

	alog := logging.AuditLog{Request: &logging.Request{Command: CommandName}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)
	//Extract the Parameters
	vars := mux.Vars(r)
	FabricName := vars["fabricName"]
	Role := vars["role"]

	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"FabricName": FabricName,
	}
	alog.LogMessageReceived()

	//Make a call to fetch Fabric Config the Fabric
	response, err := infra.GetUseCaseInteractor().FetchFabricConfigs(ctx, FabricName, Role)

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
		b, _ := json.MarshalIndent(response, "", "  ")
		OpenAPIResp := swagger.ConfigShowResponse{Response: string(b)}

		//Write Success Structure to the Body
		bytess, _ := json.Marshal(&OpenAPIResp)
		w.Write(bytess)
	}

}

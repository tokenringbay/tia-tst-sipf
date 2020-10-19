package handler

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/logging"
	"efa-server/infra/rest/generated/server/go"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

//ValidateFabric is a REST handler to handle
// "configure fabric : validate" REST POST request
func ValidateFabric(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	success := true
	statusMsg := ""

	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:Validate Fabric"}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)

	vars := mux.Vars(r)
	FabricName := vars["fabric_name"]

	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"FabricName": FabricName,
	}
	alog.LogMessageReceived()

	fabricType := domain.CLOSFabricType
	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(constants.DefaultFabric); err == nil {
		if FabricProperties, properr := infra.GetUseCaseInteractor().Db.GetFabricProperties(Fabric.ID); properr == nil {
			fabricType = FabricProperties.FabricType
		}
	}
	ctx = context.WithValue(ctx, appcontext.FabricType, fabricType)

	ValidateResponse, err := infra.GetUseCaseInteractor().ValidateFabricTopology(ctx, FabricName)
	//Indicating there is generic Failure
	if err != nil {
		success = false
		http.Error(w, statusMsg,
			http.StatusInternalServerError)
	}
	//Send Fabric Validate Response
	OpenAPIResp := swagger.FabricValidateResponse{FabricName: ValidateResponse.FabricName, MissingLinks: ValidateResponse.MissingLinks,
		MissingLeaves: ValidateResponse.NoLeaves, MissingSpines: ValidateResponse.NoSpines, SpineSpineLinks: ValidateResponse.SpineSpineLinks,
		LeafLeafLinks: ValidateResponse.LeafLeafLinks}
	bytess, _ := json.Marshal(&OpenAPIResp)

	//Set the status Messages so that it is audit logged
	statusMsg = "Validate Fabric Success"
	w.Write(bytess)

}

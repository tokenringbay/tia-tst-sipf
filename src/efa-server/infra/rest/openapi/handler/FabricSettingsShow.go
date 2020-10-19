package handler

import (
	"net/http"

	"efa-server/domain"
	"efa-server/infra"
	"efa-server/infra/constants"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"efa/infra/rest/generated/client"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"reflect"
)

func prepareFabricResponse(FabricProperties *domain.FabricProperties, FabricSetting map[string]string) {
	val := reflect.ValueOf(FabricProperties).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		FabricSetting[typeField.Name] = valueField.String()
	}
}

//ShowFabricSettings is a REST handler to handle
// GET request for fabric settings
func ShowFabricSettings(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	//success := true
	statusMsg := ""
	FabricSetting := make(map[string]string, 0)
	//alog := logging.AuditLog{Request: &logging.Request{Command: "Show Fabric Settings"}}
	//alog.LogMessageInit()
	//defer alog.LogMessageEnd(&success, &statusMsg)
	vars := mux.Vars(r)
	FabricName := vars["name"]
	//alog.Request.Params = map[string]interface{}{
	//	"FabricName": FabricName,
	//}
	//alog.LogMessageReceived()
	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(FabricName); err == nil {

		if FabricProperties, properr := infra.GetUseCaseInteractor().Db.GetFabricProperties(Fabric.ID); properr == nil {

			var response Restmodel.FabricdataResponse
			prepareFabricResponse(&FabricProperties, FabricSetting)
			FabricSetting["ID"] = fmt.Sprintf("%d", FabricProperties.ID)
			FabricSetting["FabricID"] = fmt.Sprintf("%d", Fabric.ID)
			response.FabricName = Fabric.Name
			response.FabricId = int32(Fabric.ID)
			response.FabricSettings = FabricSetting
			bytess, _ := json.Marshal(&response)
			//success = true
			w.Write(bytess)

		} else {
			statusMsg = fmt.Sprintf("Unable to retrieve Fabric Properties for %s\n", FabricName)
			http.Error(w, "",
				http.StatusNotFound)
			OpenAPIError := swagger.ErrorModel{Message: statusMsg}
			bytess, _ := json.Marshal(&OpenAPIError)
			w.Write(bytess)
		}
	} else {
		statusMsg = fmt.Sprintf("Unable to retrieve Fabric for %s\n", FabricName)
		http.Error(w, "",
			http.StatusNotFound)
		OpenAPIError := swagger.ErrorModel{Message: statusMsg}
		bytess, _ := json.Marshal(&OpenAPIError)
		w.Write(bytess)
	}
}

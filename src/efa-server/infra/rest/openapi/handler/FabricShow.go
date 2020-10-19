package handler

import (
	"net/http"

	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/device/adapter"
	"efa-server/infra/logging"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"efa/infra/rest/generated/client"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
)

//ShowDevicesInFabric is a REST handler to handle
// devices in a Fabric
func ShowDevicesInFabric(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	success := true
	statusMsg := ""
	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric show"}}
	alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)
	vars := mux.Vars(r)
	FabricName := vars["name"]
	alog.Request.Params = map[string]interface{}{
		"FabricName": FabricName,
	}
	alog.LogMessageReceived()

	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(FabricName); err == nil {
		response := Restmodel.SwitchesdataResponse{}
		if Devices, properr := infra.GetUseCaseInteractor().Db.GetDevicesInFabric(Fabric.ID); properr == nil {
			response.Items = make([]Restmodel.SwitchdataResponse, 0, len(Devices))
			for _, device := range Devices {
				rack, rackerr := infra.GetUseCaseInteractor().Db.GetRackbyIP(FabricName, device.IPAddress)
				rackName := ""
				if rackerr == nil {
					rackName = rack.RackName
				}

				switchData := Restmodel.SwitchdataResponse{IpAddress: device.IPAddress, Role: device.DeviceRole,
					Firmware: device.FirmwareVersion, Model: adapter.TranslateModelString(device.Model), Rack: rackName, Name: device.Name,
					Fabric: &Restmodel.SwitchdataResponseFabric{FabricName: FabricName, FabricId: int32(Fabric.ID)}}
				response.Items = append(response.Items, switchData)

			}

			bytess, _ := json.Marshal(&response)

			success = true
			w.Write(bytess)

		} else {
			statusMsg = fmt.Sprintf("Unable to retrieve Devices for %s\n", FabricName)
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

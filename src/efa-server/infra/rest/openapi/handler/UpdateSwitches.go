package handler

import (
	"efa-server/infra/logging"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"efa-server/infra"
	"efa-server/infra/constants"
)

//UpdateSwitches is a REST handler to handle "device credentials update" REST POST request
func UpdateSwitches(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	success := true
	statusMsg := ""

	var UpdateSwitchParams Restmodel.UpdateSwitchParameters

	alog := logging.AuditLog{Request: &logging.Request{Command: "device credentials update"}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)

	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &UpdateSwitchParams)
	if err != nil {
		success = false
		return
	}
	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"Devices": UpdateSwitchParams.DeviceIpAddress,
	}
	alog.LogMessageReceived()

	UpdateDeviceResponseList, err := infra.GetUseCaseInteractor().UpdateDevices(ctx, UpdateSwitchParams.DeviceIpAddress, UpdateSwitchParams.Username, UpdateSwitchParams.Password)

	//fmt.Println(UpdateDeviceResponseList)

	//Top level structure for Switches Response
	SwitchesUpdatedDataResponse := Restmodel.SwitchesUpdateResponse{}
	SwitchesUpdatedDataResponse.Items = make([]Restmodel.SwitchUpdateResponse, 0, len(UpdateSwitchParams.DeviceIpAddress))

	for _, updateDeviceResponse := range UpdateDeviceResponseList {
		//Populate the Response Model
		response := Restmodel.SwitchUpdateResponse{
			IpAddress:         updateDeviceResponse.IPAddress,
			DeviceCredentials: updateDeviceResponse.Status,
		}
		SwitchesUpdatedDataResponse.Items = append(SwitchesUpdatedDataResponse.Items, response)
	}
	bytes, _ := json.Marshal(&SwitchesUpdatedDataResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)

}

package handler

import (
	"net/http"

	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/logging"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"efa-server/usecase"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//CreateSwitches is a REST handler to handle
// "fabric configure : Add device" REST POST request
func CreateSwitches(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	success := true
	statusMsg := ""

	var NewSwitchesRequest Restmodel.NewSwitches

	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:Add Device"}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)

	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &NewSwitchesRequest)
	if err != nil {
		success = false
		return
	}
	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"FabricName": NewSwitchesRequest.Fabric,
		"Spines":     NewSwitchesRequest.SpineIpAddress,
		"Leaves":     NewSwitchesRequest.LeafIpAddress,
		"Racks":      NewSwitchesRequest.Racks,
		"Force":      NewSwitchesRequest.Force,
	}

	alog.LogMessageReceived()

	fabricType := domain.CLOSFabricType
	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(NewSwitchesRequest.Fabric); err == nil {
		if FabricProperties, properr := infra.GetUseCaseInteractor().Db.GetFabricProperties(Fabric.ID); properr == nil {
			fabricType = FabricProperties.FabricType
		}
	}
	ctx = context.WithValue(ctx, appcontext.FabricType, fabricType)

	AddDeviceResponseList, err := callUseCase(ctx, fabricType, NewSwitchesRequest)

	//Indicating there is an overall Failure, So populate the status model with appropriate errors
	if err != nil {
		success = false
		StatusModelList := make([]Restmodel.DeviceStatusModel, 0, len(AddDeviceResponseList))

		for _, AddDeviceResponse := range AddDeviceResponseList {
			//Populate the initial status as Successful
			StatusModel := Restmodel.DeviceStatusModel{IpAddress: AddDeviceResponse.IPAddress, Role: AddDeviceResponse.Role, Status: "Successful"}
			StatusModel.Error_ = make([]Restmodel.ErrorModel, 0)
			//If there are errors then poluatate status for that device as Failed
			if len(AddDeviceResponse.Errors) > 0 {
				StatusModel.Status = "Failed"
			}
			//Populate each error from the device
			for _, er := range AddDeviceResponse.Errors {
				fmt.Println(er)
				StatusModel.Error_ = append(StatusModel.Error_, Restmodel.ErrorModel{Message: fmt.Sprint(er)})
			}
			StatusModelList = append(StatusModelList, StatusModel)
		}

		http.Error(w, statusMsg,
			http.StatusInternalServerError)
		bytess, _ := json.Marshal(&StatusModelList)
		w.Write(bytess)
	} else {
		//Top level structure for Switches Response
		SwitchesExtendedDataResponse := Restmodel.SwitchesdataResponse{}
		SwitchesExtendedDataResponse.Items = make([]Restmodel.SwitchdataResponse, 0, len(AddDeviceResponseList))

		for _, AddDeviceResponse := range AddDeviceResponseList {

			//Populate the Response Model
			response := Restmodel.SwitchdataResponse{
				IpAddress: AddDeviceResponse.IPAddress,
				Fabric:    &Restmodel.SwitchdataResponseFabric{FabricName: AddDeviceResponse.FabricName, FabricId: int32(AddDeviceResponse.FabricID)},
				Role:      AddDeviceResponse.Role}
			SwitchesExtendedDataResponse.Items = append(SwitchesExtendedDataResponse.Items, response)

		}

		bytess, _ := json.Marshal(&SwitchesExtendedDataResponse)
		w.Write(bytess)
	}

}

func callUseCase(ctx context.Context, fabricType string, NewSwitchesRequest Restmodel.NewSwitches) (AddDeviceResponse []usecase.AddDeviceResponse, err error) {
	if fabricType == domain.CLOSFabricType {
		return infra.GetUseCaseInteractor().AddDevices(ctx, NewSwitchesRequest.Fabric, NewSwitchesRequest.LeafIpAddress, NewSwitchesRequest.SpineIpAddress,
			NewSwitchesRequest.Username, NewSwitchesRequest.Password, NewSwitchesRequest.Force)
	}
	//non-clos
	rackList := make([]usecase.Rack, 0)
	for _, rack := range NewSwitchesRequest.Racks {
		//Always two nodes should be present in the Rack
		if len(rack.RackDevices) == 2 {
			Rack := usecase.Rack{IP1: rack.RackDevices[0], IP2: rack.RackDevices[1]}
			rackList = append(rackList, Rack)
		}
	}
	//fmt.Println("rackList", rackList)
	return infra.GetUseCaseInteractor().AddRacks(ctx, NewSwitchesRequest.Fabric, rackList,
		NewSwitchesRequest.Username, NewSwitchesRequest.Password, NewSwitchesRequest.Force)
}

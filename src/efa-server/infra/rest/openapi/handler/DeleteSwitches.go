package handler

import (
	"context"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/logging"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"efa-server/usecase"
	"encoding/json"
	"fmt"
	"net/http"
	//"io/ioutil"
	//"github.com/gorilla/mux"
	//"strconv"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"io/ioutil"
)

//DeleteSwitches is a REST handler to handle
// "fabric deconfigure : Delete Device" REST POST request
func DeleteSwitches(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	var DelSwitchReq Restmodel.DeleteSwitchesRequest

	success := true
	statusMsg := ""

	alog := logging.AuditLog{Request: &logging.Request{Command: "fabric configure:Delete Device"}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)

	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &DelSwitchReq)
	if err != nil {
		success = false
		return
	}

	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"FabricName": constants.DefaultFabric,
		"Devices":    DelSwitchReq.Switches,
		"Force":      false,
		"persist":    DelSwitchReq.Persist,
		"Racks":      DelSwitchReq.Racks,
		"DevCleanup": DelSwitchReq.DeviceCleanup,
	}
	alog.LogMessageReceived()

	fabricType := domain.CLOSFabricType
	if Fabric, err := infra.GetUseCaseInteractor().Db.GetFabric(constants.DefaultFabric); err == nil {
		if FabricProperties, properr := infra.GetUseCaseInteractor().Db.GetFabricProperties(Fabric.ID); properr == nil {
			fabricType = FabricProperties.FabricType
		}
	}

	ctx = context.WithValue(ctx, appcontext.FabricType, fabricType)
	AddDeviceResponseList, err := callDeleteUseCase(ctx, fabricType, DelSwitchReq)

	//fmt.Println(AddDeviceResponseList)

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

		http.Error(w, statusMsg, http.StatusInternalServerError)
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

func callDeleteUseCase(ctx context.Context, fabricType string, DelSwitchReq Restmodel.DeleteSwitchesRequest) (AddDeviceResponse []usecase.AddDeviceResponse, err error) {
	if fabricType == domain.CLOSFabricType {
		return infra.GetUseCaseInteractor().DeleteDevicesFromFabric(ctx, constants.DefaultFabric, DelSwitchReq.Switches,
			"", "", false, DelSwitchReq.Persist, DelSwitchReq.DeviceCleanup)
	}

	//non-clos
	rackList := make([]usecase.Rack, 0)
	for _, rack := range DelSwitchReq.Racks {
		//Always two nodes should be present in the Rack
		if len(rack.RackDevices) == 2 {
			Rack := usecase.Rack{IP1: rack.RackDevices[0], IP2: rack.RackDevices[1]}
			rackList = append(rackList, Rack)
		}
	}
	//fmt.Println("rackList", rackList)
	return infra.GetUseCaseInteractor().DeleteDevicesFromNonCLOSFabric(ctx, constants.DefaultFabric, rackList,
		"", "", false, DelSwitchReq.Persist, DelSwitchReq.DeviceCleanup)
}

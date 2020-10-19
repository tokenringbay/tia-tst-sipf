package usecase

import (
	"context"
	"efa-server/gateway/appcontext"
	Interactor "efa-server/usecase/interactorinterface"
)

//UpdateDeviceResponse describes the Response when device details are updated.
type UpdateDeviceResponse struct {
	//IP Address of the Device
	IPAddress string
	//Status of updating device information
	Status string
}

//UpdateDevices updates device parameters like credentials
func (sh *DeviceInteractor) UpdateDevices(ctx context.Context, DeviceIPaddressList []string, UserName string, Password string) (UpdateDevicesResponse []UpdateDeviceResponse, err error) {

	//Setup the logger
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "device credentails update")

	LOG := appcontext.Logger(ctx)

	for _, device := range DeviceIPaddressList {
		func() {
			dbDevice, err := sh.Db.GetDeviceInAnyFabric(device)
			if err != nil {
				deviceResponse := UpdateDeviceResponse{IPAddress: device, Status: "Error while retrieving Switch details from Database : " + err.Error()}
				UpdateDevicesResponse = append(UpdateDevicesResponse, deviceResponse)
				LOG.Errorln(deviceResponse.Status)
				return
			}

			sh.EvaluateCredentials(UserName, Password, &dbDevice)

			var DeviceAdapter Interactor.DeviceAdapter
			//Open Connection to the Switch
			if DeviceAdapter, err = sh.DeviceAdapterFactory(ctx, device, dbDevice.UserName, dbDevice.Password); err != nil {
				deviceResponse := UpdateDeviceResponse{IPAddress: device, Status: "Switch connection Failed  : " + err.Error()}
				UpdateDevicesResponse = append(UpdateDevicesResponse, deviceResponse)
				LOG.Errorln(deviceResponse.Status)
				return
			}
			defer DeviceAdapter.CloseConnection(ctx)

			//Start Transaction
			sh.DBMutex.Lock()
			defer sh.DBMutex.Unlock()

			//Save Device
			if err = sh.Db.CreateDevice(&dbDevice); err != nil {
				deviceResponse := UpdateDeviceResponse{IPAddress: device, Status: "Updating Switch credentials Failed : " + err.Error()}
				UpdateDevicesResponse = append(UpdateDevicesResponse, deviceResponse)
				LOG.Errorln(deviceResponse.Status)
				return
			}

			deviceResponse := UpdateDeviceResponse{IPAddress: device, Status: "Successfully Updated Switch Credentials"}
			LOG.Infoln(deviceResponse.Status)
			UpdateDevicesResponse = append(UpdateDevicesResponse, deviceResponse)

		}()
	}
	return UpdateDevicesResponse, nil
}

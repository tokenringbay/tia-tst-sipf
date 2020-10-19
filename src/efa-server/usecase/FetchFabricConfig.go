package usecase

import "efa-server/domain"
import (
	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"errors"
	"fmt"
)

//FetchFabricConfigs is used to fetch the "config" from the devices of the fabric
func (sh *DeviceInteractor) FetchFabricConfigs(ctx context.Context, FabricName string, Role string) (operation.FabricFetchResponse, error) {
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Fetch Config")
	LOG := appcontext.Logger(ctx)
	response := operation.FabricFetchResponse{}

	Fabric, err := sh.Db.GetFabric(FabricName)
	//Check the presence of Fabric
	if err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", FabricName)
		LOG.Errorln(statusMsg)
		return response, errors.New(statusMsg)
	}

	sh.FabricProperties, err = sh.Db.GetFabricProperties(Fabric.ID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch fabric properties for %d", Fabric.ID)
		LOG.Errorln(statusMsg)
		return response, errors.New(statusMsg)
	}

	ActionFabricFetchRequest, err := sh.PrepareActionFabricFetchRequest(ctx, &Fabric, Role)
	if err != nil {
		statusMsg := fmt.Sprintf("Fabric %s ", FabricName)
		LOG.Errorln(statusMsg)
		return response, errors.New(statusMsg)
	}

	//Fetch Fabric Information from Switches
	response, err = sh.FabricAdapter.FetchFabricConfiguration(ctx, ActionFabricFetchRequest)
	if err != nil {
		return response, err
	}

	return response, nil
}

func matchRole(DeviceRole string, RequestedRole string) bool {
	if RequestedRole == "spine" && DeviceRole == SpineRole {
		return true
	}
	if RequestedRole == "leaf" && DeviceRole == LeafRole {
		return true
	}
	if RequestedRole == "all" && (DeviceRole == LeafRole || DeviceRole == SpineRole || DeviceRole == RackRole) {
		return true
	}
	return false
}

//PrepareActionFabricFetchRequest prepares the operation.FabricFetchRequest object
func (sh *DeviceInteractor) PrepareActionFabricFetchRequest(ctx context.Context, Fabric *domain.Fabric, Role string) (operation.FabricFetchRequest, error) {

	LOG := appcontext.Logger(ctx)
	resp := operation.FabricFetchRequest{}
	resp.FabricName = Fabric.Name
	devices, err := sh.Db.GetDevicesInFabric(Fabric.ID)

	if err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch devices from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return resp, errors.New(statusMsg)
	}

	resp.Hosts = make([]operation.SwitchIdentity, 0, len(devices))
	for _, dev := range devices {
		if matchRole(dev.DeviceRole, Role) {

			intfs, err := sh.Db.GetInterfaceSwitchConfigsOnDeviceID(Fabric.ID, dev.ID)

			if err != nil {
				statusMsg := fmt.Sprintf("Failed to fetch interface switch configs for %s:%s", sh.FabricName, dev.IPAddress)
				LOG.Errorln(statusMsg)
				return resp, errors.New(statusMsg)
			}
			// Length of Interfaces is (+3) because of VE and Loop back interfaces.
			host := operation.SwitchIdentity{Host: dev.IPAddress, UserName: dev.UserName, Password: dev.Password,
				Role: dev.DeviceRole, Interfaces: make(map[string][]string), Model: dev.Model}
			for _, intf := range intfs {
				host.Interfaces[intf.IntType] = append(host.Interfaces[intf.IntType], intf.IntName)
			}
			host.Interfaces[domain.IntfTypeVe] = append(host.Interfaces[domain.IntfTypeVe], sh.FabricProperties.ControlVE)
			host.Interfaces[domain.IntfTypeLoopback] = append(host.Interfaces[domain.IntfTypeLoopback], sh.FabricProperties.VTEPLoopBackPortNumber)
			host.Interfaces[domain.IntfTypeLoopback] = append(host.Interfaces[domain.IntfTypeLoopback], sh.FabricProperties.LoopBackPortNumber)

			resp.Hosts = append(resp.Hosts, host)
		}
	}
	return resp, err
}

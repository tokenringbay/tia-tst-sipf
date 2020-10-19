package gateway

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
	ClearFabric "efa-server/infra/device/actions/clearfabric"
	ConfigureFabric "efa-server/infra/device/actions/configurefabric"
	DeconfigureFabric "efa-server/infra/device/actions/deconfigurefabric"
	NONCLOSDeconfigureFabric "efa-server/infra/device/actions/deconfigurefabric"
	FetchFabric "efa-server/infra/device/actions/fetchfabric"
	"efa-server/infra/device/adapter"
	"strings"
)

//FabricAdapter provides  method to Configure Fabric
type FabricAdapter struct {
}

//ConfigureDeConfigureMctClusters is used to "configure" and "deconfigure" MCT cluster
func (ad *FabricAdapter) ConfigureDeConfigureMctClusters(ctx context.Context, operation uint, config []operation.ConfigCluster, force bool) []actions.OperationError {
	err := ConfigureFabric.ConfigureDeConfigureMctClusters(ctx, operation, config, force)
	return err
}

//ConfigureFabric Configures IP Fabric excluding MCT Cluster
func (ad *FabricAdapter) ConfigureFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	err := ConfigureFabric.ConfigureFabric(ctx, config, force, persist)
	return err

}

//FetchFabricConfiguration fetches the Configurations from the Fabric
func (ad *FabricAdapter) FetchFabricConfiguration(ctx context.Context, FabricRequest operation.FabricFetchRequest) (operation.FabricFetchResponse, error) {
	return FetchFabric.FetchFabric(ctx, FabricRequest)

}

//CleanupDevicesInFabric cleans up IP Fabric configuration from a collection of devices used by Delete Fabric
func (ad *FabricAdapter) CleanupDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	Errors := DeconfigureFabric.CleanupDevicesInFabric(ctx, config, force, persist)
	return Errors

}

//CleanupDevicesInNonCLOSFabric cleans up NON CLOS configuration from a collection of devices used by Delete Fabric
func (ad *FabricAdapter) CleanupDevicesInNonCLOSFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	Errors := NONCLOSDeconfigureFabric.CleanupDevicesInNonCLOSFabric(ctx, config, force, persist)
	return Errors

}

//CleanupDependantDevicesInFabric cleans up interfaces and bgp neighbors from dependant Devices
func (ad *FabricAdapter) CleanupDependantDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	Errors := DeconfigureFabric.CleanupDependantDevicesInFabric(ctx, config, force, persist)
	return Errors
}

//ClearConfig clears up Configuration from list of devices
func (ad *FabricAdapter) ClearConfig(ctx context.Context, ClearFabricEquest operation.ClearFabricRequest) error {
	return ClearFabric.ClearFabric(ctx, ClearFabricEquest)
}

//ClearMctClusters is used to "clear" MCT cluster config
func (ad *FabricAdapter) ClearMctClusters(ctx context.Context, FabricName string, clusters []operation.ConfigCluster) error {
	return ClearFabric.ClearMctClusters(ctx, FabricName, clusters)

}

//IsRoutingDevice is used determine if the device is a Routing device
func (ad *FabricAdapter) IsRoutingDevice(ctx context.Context, ModelVersion string) bool {
	data := strings.Split(ModelVersion, "_")
	Model := data[0]
	return adapter.AvalancheType == Model || adapter.OrcaType == Model || adapter.OrcaTType == Model

}

//IsMCTLeavesCompatible is used to determine if the the two devices are compatible
func (ad *FabricAdapter) IsMCTLeavesCompatible(ctx context.Context, DeviceModelVersion string, RemoteDeviceModelVersion string) bool {
	data := strings.Split(DeviceModelVersion, "_")
	DeviceModel := data[0]

	data = strings.Split(RemoteDeviceModelVersion, "_")
	RemoteDeviceModel := data[0]

	//If Devices are Avalanche,Freedom and Cedar then the DeviceModel and RemoveDeviceModel should be same
	if (DeviceModel == adapter.AvalancheType || DeviceModel == adapter.FreedomType || DeviceModel == adapter.CedarType) && (DeviceModel == RemoteDeviceModel) {
		return true
	}

	//Orca and Orcat are compatible
	if (DeviceModel == adapter.OrcaType || DeviceModel == adapter.OrcaTType) && (RemoteDeviceModel == adapter.OrcaType || RemoteDeviceModel == adapter.OrcaTType) {
		return true
	}

	return false
}

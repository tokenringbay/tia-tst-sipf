package mock

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
)

//FabricAdapter represents a mock FabricAdapter
type FabricAdapter struct {
	MockConfigureDeConfigureMctClusters func(ctx context.Context, operation uint, config []operation.ConfigCluster, force bool) []actions.OperationError
	MockConfigureFabric                 func(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	MockFetchFabricConfiguration        func(ctx context.Context, FabricRequest operation.FabricFetchRequest) (operation.FabricFetchResponse, error)
	MockClearConfig                     func(ctx context.Context, ClearFabricEquest operation.ClearFabricRequest) error
	MockCleanupDevicesInFabric          func(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	MockCleanupDevicesInNonCLOSFabric   func(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	MockCleanupDependantDevicesInFabric func(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	MockClearMctClusters                func(ctx context.Context, FabricName string, clusters []operation.ConfigCluster) error
	MockIsRoutingDevice                 func(ctx context.Context, Model string) bool
	MockIsMCTLeavesCompatible           func(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool
}

//ConfigureDeConfigureMctClusters returns mock of ConfigureDeConfigureMctClusters
func (fa *FabricAdapter) ConfigureDeConfigureMctClusters(ctx context.Context, operation uint, config []operation.ConfigCluster, force bool) []actions.OperationError {
	if fa.MockConfigureDeConfigureMctClusters != nil {
		return fa.MockConfigureDeConfigureMctClusters(ctx, operation, config, force)
	}
	return []actions.OperationError{}
}

//ConfigureFabric returns mock of ConfigureFabric
func (fa *FabricAdapter) ConfigureFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	if fa.MockConfigureFabric != nil {
		return fa.MockConfigureFabric(ctx, config, force, persist)
	}
	return []actions.OperationError{}
}

//FetchFabricConfiguration returns mock of FetchFabricConfiguration
func (fa *FabricAdapter) FetchFabricConfiguration(ctx context.Context, FabricRequest operation.FabricFetchRequest) (operation.FabricFetchResponse, error) {
	if fa.MockFetchFabricConfiguration != nil {
		return fa.MockFetchFabricConfiguration(ctx, FabricRequest)
	}
	return operation.FabricFetchResponse{}, nil
}

//ClearConfig returns mock of ClearConfig
func (fa *FabricAdapter) ClearConfig(ctx context.Context, ClearFabricEquest operation.ClearFabricRequest) error {
	if fa.MockClearConfig != nil {
		return fa.MockClearConfig(ctx, ClearFabricEquest)
	}
	return nil
}

//CleanupDevicesInFabric returns mock of CleanupDevicesInFabric
func (fa *FabricAdapter) CleanupDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	if fa.MockCleanupDevicesInFabric != nil {
		return fa.MockCleanupDevicesInFabric(ctx, config, force, persist)
	}
	return []actions.OperationError{}
}

//CleanupDevicesInNonCLOSFabric returns mock of CleanupDevicesInNonCLOSFabric
func (fa *FabricAdapter) CleanupDevicesInNonCLOSFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	if fa.MockCleanupDevicesInNonCLOSFabric != nil {
		return fa.MockCleanupDevicesInNonCLOSFabric(ctx, config, force, persist)
	}
	return []actions.OperationError{}
}

//CleanupDependantDevicesInFabric returns mock of CleanupDevicesInNonCLOSFabric
func (fa *FabricAdapter) CleanupDependantDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {
	if fa.MockCleanupDependantDevicesInFabric != nil {
		return fa.MockCleanupDependantDevicesInFabric(ctx, config, force, persist)
	}
	return []actions.OperationError{}
}

//ClearMctClusters returns mock of ClearMctCluster
func (fa *FabricAdapter) ClearMctClusters(ctx context.Context, FabricName string, clusters []operation.ConfigCluster) error {
	if fa.MockClearMctClusters != nil {
		return fa.MockClearMctClusters(ctx, FabricName, clusters)
	}
	return nil
}

//IsRoutingDevice is used determine if the device is a Routing device
func (fa *FabricAdapter) IsRoutingDevice(ctx context.Context, Model string) bool {
	if fa.MockIsRoutingDevice != nil {
		return fa.MockIsRoutingDevice(ctx, Model)
	}
	return false

}

//IsMCTLeavesCompatible is used to determine if the the two devices are compatible
func (fa *FabricAdapter) IsMCTLeavesCompatible(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool {
	if fa.MockIsMCTLeavesCompatible != nil {
		return fa.MockIsMCTLeavesCompatible(ctx, DeviceModel, RemoteDeviceModel)
	}
	return false
}

package interactorinterface

import (
	"context"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions"
)

//FabricAdapter is an interface to the fabric operations
type FabricAdapter interface {
	ConfigureDeConfigureMctClusters(ctx context.Context, operation uint, config []operation.ConfigCluster, force bool) []actions.OperationError
	ConfigureFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	FetchFabricConfiguration(ctx context.Context, FabricRequest operation.FabricFetchRequest) (operation.FabricFetchResponse, error)
	ClearConfig(ctx context.Context, ClearFabricEquest operation.ClearFabricRequest) error
	CleanupDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	CleanupDevicesInNonCLOSFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	CleanupDependantDevicesInFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError
	ClearMctClusters(ctx context.Context, FabricName string, cluster []operation.ConfigCluster) error
	IsRoutingDevice(ctx context.Context, Model string) bool
	IsMCTLeavesCompatible(ctx context.Context, DeviceModel string, RemoteDeviceModel string) bool
}

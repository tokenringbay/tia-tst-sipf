package interactorinterface

import (
	"efa-server/domain"
)

//DatabaseRepository is an interface to Database operations
type DatabaseRepository interface {
	//Transaction Operations
	OpenTransaction() error
	CommitTransaction() error
	RollBackTransaction() error

	Backup() error

	//Fabric Operations
	GetFabric(FabricName string) (domain.Fabric, error)
	CreateFabric(Fabric *domain.Fabric) error
	DeleteFabric(FabricName string) error

	//FabricProperties
	GetFabricProperties(FabricID uint) (domain.FabricProperties, error)
	CreateFabricProperties(FabricProperties *domain.FabricProperties) error
	UpdateFabricProperties(FabricProperties *domain.FabricProperties) error

	//Device Operations
	GetDevice(FabricName string, IPAddress string) (domain.Device, error)
	GetDeviceInAnyFabric(IPAddress string) (domain.Device, error)
	GetDevicesInFabric(FabricID uint) ([]domain.Device, error)
	GetDevicesCountInFabric(FabricID uint) uint16
	GetDeviceUsingDeviceID(FabricID uint, DeviceID uint) (domain.Device, error)
	GetDevicesInFabricMatching(FabricID uint, device []string) ([]domain.Device, error)
	GetDevicesInFabricNotMatching(FabricID uint, device []string) ([]domain.Device, error)

	GetRack(FabricName string, IP1 string, IP2 string) (domain.Rack, error)
	GetRackbyIP(FabricName string, IP string) (domain.Rack, error)
	GetRackAll(FabricName string) ([]domain.Rack, error)
	GetRacksInFabric(FabricID uint) ([]domain.Rack, error)
	CreateRack(Rack *domain.Rack) error
	DeleteRack(RackID []uint) error
	SaveRack(Rack *domain.Rack) error

	CreateDevice(Device *domain.Device) error
	DeleteDevice(DeviceID []uint) error
	SaveDevice(Device *domain.Device) error

	//Interface Operations
	CreateInterface(Interface *domain.Interface) error
	DeleteInterface(Interface *domain.Interface) error
	DeleteInterfaceUsingID(FabricID uint, InterfaceID uint) error
	ResetDeleteOperationOnInterface(FabricID uint, InterfaceID uint) error
	GetInterfacesonDevice(FabricID uint, DeviceID uint) ([]domain.Interface, error)
	GetInterfaceIDsMarkedForDeletion(FabricID uint) (map[uint]uint, error)
	GetInterface(FabricID uint, DeviceID uint,
		InterfaceType string, InterfaceName string) (domain.Interface, error)
	GetInterfaceOnMac(Mac string, FabricID uint) (domain.Interface, error)
	GetLLDP(FabricID uint, DeviceID uint,
		LocalInterfaceType string, LocalInterfaceName string) (domain.LLDP, error)
	GetLLDPsonDevice(FabricID uint, DeviceID uint) ([]domain.LLDP, error)
	CreateLLDP(Interface *domain.LLDP) error
	DeleteLLDP(Interface *domain.LLDP) error
	GetLLDPOnRemoteMacExcludingMarkedForDeletion(RemoteMac string, FabricID uint) (domain.LLDP, error)
	UpdateLLDPConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	DeleteLLDPMarkedForDelete(FabricID uint) error

	//ASN Pool
	CreateASN(ASN *domain.ASNAllocationPool) error
	DeleteASNPool() error
	DeleteUsedASNPool() error
	DeleteASN(ASN *domain.ASNAllocationPool) error
	GetNextASNForRole(FabricID uint, role string) (domain.ASNAllocationPool, error)
	GetASNCountOnASN(FabricID uint, asn uint64) (int64, error)
	GetASNCountOnRole(FabricID uint, role string) (int64, error)
	GetASNAndCountOnASNAndRole(FabricID uint, asn uint64, role string) (int64, domain.ASNAllocationPool)
	CreateUsedASN(UsedASN *domain.UsedASN) error
	DeleteUsedASN(UsedASN *domain.UsedASN) error
	GetUsedASNOnASNAndDeviceAndRole(FabricID uint, DeviceID uint, asn uint64, role string) (domain.UsedASN, error)
	GetUsedASNCountOnASNAndDevice(FabricID uint, asn uint64, DeviceID uint) (int64, error)
	GetUsedASNCountOnASNAndRole(FabricID uint, asn uint64, role string) (int64, error)

	//IP Pool
	CreateIPEntry(IPEntry *domain.IPAllocationPool) error
	DeleteIPEntry(IPEntry *domain.IPAllocationPool) error
	DeleteIPPool() error
	GetIPEntryAndCountOnIPAddressAndType(FabricID uint, ipaddress string, IPType string) (int64, domain.IPAllocationPool, error)
	GetNextIPEntryOnType(FabricID uint, IPType string) (domain.IPAllocationPool, error)
	GetUsedIPOnDeviceInterfaceIDIPAddresssAndType(FabricID uint, DeviceID uint, ipaddress string, IPType string, InterfaceID uint) (domain.UsedIP, error)
	GetUsedIPOnDeviceInterfaceIDAndType(FabricID uint, DeviceID uint, IPType string, InterfaceID uint) (domain.UsedIP, error)
	CreateUsedIPEntry(UsedIPEntry *domain.UsedIP) error
	DeleteUsedIPEntry(UsedIPEntry *domain.UsedIP) error
	DeleteUsedIPPool() error

	//IP Pair Pool
	CreateIPPairEntry(IPEntry *domain.IPPairAllocationPool) error
	DeleteIPPairEntry(IPEntry *domain.IPPairAllocationPool) error
	DeleteIPPairPool() error
	GetIPPairEntryAndCountOnIPAddressAndType(FabricID uint, ipaddressOne string, ipaddressTwo string, IPType string) (int64, domain.IPPairAllocationPool, error)
	GetIPPairEntryAndCountOnEitherIPAddressAndType(FabricID uint, ipaddress string, IPType string) (int64, domain.IPPairAllocationPool, error)
	GetNextIPPairEntryOnType(FabricID uint, IPType string) (domain.IPPairAllocationPool, error)
	GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint, ipaddressOne string, ipaddressTwo string, IPType string,
		InterfaceOneID uint, InterfaceTwoID uint) (domain.UsedIPPair, error)
	GetUsedIPPairOnDeviceInterfaceIDAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPType string, InterfaceOneID uint, InterfaceTwoID uint) (domain.UsedIPPair, error)
	CreateUsedIPPairEntry(UsedIPEntry *domain.UsedIPPair) error
	DeleteUsedIPPairEntry(UsedIPEntry *domain.UsedIPPair) error
	DeleteUsedIPPairPool() error

	GetLLDPNeighbor(FabricID uint, DeviceOneID uint,
		DeviceTwoID uint, InterfaceOneID uint, InterfaceTwoID uint) (domain.LLDPNeighbor, error)
	CreateLLDPNeighbor(BGPNeighbor *domain.LLDPNeighbor) error
	DeleteLLDPNeighbor(BGPNeighbor *domain.LLDPNeighbor) error
	GetLLDPNeighborsOnDevice(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error)
	GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error)
	GetLLDPNeighborsOnDeviceMarkedForDeletion(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error)
	GetLLDPNeighborsOnEitherDevice(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error)
	GetLLDPNeighborsBetweenTwoDevices(FabricID uint, DeviceIDOne uint, DeviceIDTwo uint) ([]domain.LLDPNeighbor, error)
	GetLLDPNeighborsOnRemoteDeviceID(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.LLDPNeighbor, error)
	GetMctClusters(FabricID uint, DeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error)
	GetMctClustersCount(FabricID uint, DeviceID uint, ConfigType []string) (uint64, error)
	CreateMctClusters(MCTConfig *domain.MctClusterConfig) error
	GetMctMemberPortsConfig(FabricID uint, DeviceID uint, RemoteDeviceID uint, ConfigType []string) ([]domain.MCTMemberPorts, error)
	CreateMctClustersMembers(MCTConfigs []domain.MCTMemberPorts, ClusterID uint16, ConfigType string) error
	DeleteMCTPortConfig(MctPorts []domain.MCTMemberPorts) error
	UpdateMctPortConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	UpdateMctClusterConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	DeleteMctPortsMarkedForDelete(FabricID uint) error
	DeleteMctClustersMarkedForDelete(FabricID uint) error
	MarkMctClusterForDelete(FabricID uint, DeviceID uint) error
	MarkMctClusterMemberPortsForDelete(FabricID uint, DeviceID uint) error
	MarkMctClusterMemberPortsForCreate(FabricID uint, DeviceID uint, RemoteDeviceID uint) error
	MarkMctClusterForDeleteWithBothDevices(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error
	MarkMctClusterMemberPortsForDeleteWithBothDevices(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error
	GetMctClusterConfigWithDeviceIP(IPAddress string) ([]domain.MctClusterConfig, error)
	DeleteMctClustersUsingClusterObject(oldMct domain.MctClusterConfig) error
	DeleteMctClustersWithMgmtIP(IPAddress string) error
	GetMctClustersWithBothDevices(FabricID uint, DeviceID uint, NeighborDeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error)

	//Switch Configs
	CreateSwitchConfig(SwitchConfig *domain.SwitchConfig) error
	UpdateSwitchConfigsASConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	UpdateSwitchConfigsLoopbackConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	UpdateSwitchConfigsVTEPLoopbackConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	GetSwitchConfigOnFabricIDAndDeviceID(FabricID uint, DeviceID uint) (domain.SwitchConfig, error)
	GetSwitchConfigs(FabricName string) ([]domain.SwitchConfig, error)
	GetSwitchConfigOnDeviceIP(FabricName string, DeviceIP string) (domain.SwitchConfig, error)
	CreateInterfaceSwitchConfig(SwitchConfig *domain.InterfaceSwitchConfig) error
	UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(FabricID uint, interfaceID uint, configType string) error
	UpdateInterfaceSwitchConfigsConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	GetInterfaceSwitchConfigOnFabricIDAndInterfaceID(FabricID uint, InterfaceID uint) (domain.InterfaceSwitchConfig, error)
	GetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID(FabricID uint, InterfaceID uint) int64
	GetInterfaceSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.InterfaceSwitchConfig, error)
	GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion(FabricID uint, InterfaceIDs []uint) ([]domain.InterfaceSwitchConfig, error)
	UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs(FabricID uint, InterfaceIDs []uint, configType string) error
	GetBGPSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error)
	GetMCTBGPSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error)
	GetBGPSwitchConfigs(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error)
	GetBGPSwitchConfigsExcludingMarkedForDeletion(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error)
	GetBGPSwitchConfigsOnRemoteDeviceID(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error)
	UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID(FabricID uint, InterfaceIDs []uint, configType string) error
	GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID(FabricID uint, RemoteInterfaceID uint) (domain.RemoteNeighborSwitchConfig, error)
	GetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID(FabricID uint, RemoteInterfaceID uint) int64

	CreateBGPSwitchConfig(BGPSwitchConfig *domain.RemoteNeighborSwitchConfig) error
	UpdateBGPSwitchConfigsConfigType(FabricID uint, QueryconfigTypes []string, configType string) error
	UpdateBGPSwitchConfigsOnInterfaceIDConfigType(FabricID uint, InterfaceID uint, configType string) error

	CreateExecutionLog(ExecutionLog *domain.ExecutionLog) error
	GetExecutionLogList(limit int, status string) ([]domain.ExecutionLog, error)
	GetExecutionLogByUUID(string) (domain.ExecutionLog, error)
	UpdateExecutionLog(ExecutionLog *domain.ExecutionLog) error

	CreateMctClusterConfig(MCTConfig *domain.MCTClusterDetails) error
	DeleteMCTCluster(DeviceID uint) error
	GetMCTCluster(DeviceID uint) (domain.MCTClusterDetails, error)

	CreateRackEvpnConfig(RackEvpnConfig *domain.RackEvpnNeighbors) error
	GetRackEvpnConfig(RackID uint) ([]domain.RackEvpnNeighbors, error)
	GetRackEvpnConfigOnDeviceID(DeviceID uint) ([]domain.RackEvpnNeighbors, error)
	GetRackEvpnConfigOnRemoteDeviceID(DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RackEvpnNeighbors, error)
}

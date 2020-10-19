package mock

import (
	"efa-server/domain"
)

var userName = "admin"

//DatabaseRepository represents a mock DatabaseRepository
type DatabaseRepository struct {
	MockOpenTransaction     func() error
	MockCommitTransaction   func() error
	MockRollBackTransaction func() error
	MockBackup              func() error

	MockGetFabric                     func(FabricName string) (domain.Fabric, error)
	MockCreateFabric                  func(Fabric *domain.Fabric) error
	MockDeleteFabric                  func(FabricName string) error
	MockGetFabricProperties           func(FabricID uint) (domain.FabricProperties, error)
	MockCreateFabricProperties        func(FabricProperties *domain.FabricProperties) error
	MockUpdateFabricProperties        func(FabricProperties *domain.FabricProperties) error
	MockGetDevice                     func(FabricName string, IPAddress string) (domain.Device, error)
	MockGetDevicesInFabric            func(FabricId uint) ([]domain.Device, error)
	MockGetDevicesCountInFabric       func(FabricId uint) uint16
	MockGetDeviceInAnyFabric          func(IPAddress string) (domain.Device, error)
	MockGetDevicesInFabricMatching    func(FabricID uint, device []string) ([]domain.Device, error)
	MockGetDevicesInFabricNotMatching func(FabricID uint, device []string) ([]domain.Device, error)

	MockCreateDevice func(Device *domain.Device) error
	MockDeleteDevice func(DeviceID []uint) error
	MockSaveDevice   func(Device *domain.Device) error

	MockGetRack                           func(FabricName string, IP1 string, IP2 string) (domain.Rack, error)
	MockGetRackbyIP                       func(FabricName string, IP string) (domain.Rack, error)
	MockGetRackAll                        func(FabricName string) ([]domain.Rack, error)
	MockGetRacksInFabric                  func(FabricID uint) ([]domain.Rack, error)
	MockCreateRack                        func(Rack *domain.Rack) error
	MockDeleteRack                        func(RackID []uint) error
	MockSaveRack                          func(Rack *domain.Rack) error
	MockCreateRackEvpnConfig              func(RackEvpnConfig *domain.RackEvpnNeighbors) error
	MockGetRackEvpnConfig                 func(RackID uint) ([]domain.RackEvpnNeighbors, error)
	MockGetRackEvpnConfigOnDeviceID       func(DeviceID uint) ([]domain.RackEvpnNeighbors, error)
	MockGetRackEvpnConfigOnRemoteDeviceID func(DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RackEvpnNeighbors, error)

	MockCreateInterface                  func(Interface *domain.Interface) error
	MockResetDeleteOperationOnInterface  func(FabricID uint, InterfaceID uint) error
	MockDeleteInterface                  func(Interface *domain.Interface) error
	MockDeleteInterfaceUsingID           func(FabricID uint, InterfaceID uint) error
	MockGetInterfacesonDevice            func(FabricID uint, DeviceID uint) ([]domain.Interface, error)
	MockGetInterfaceIDsMarkedForDeletion func(FabricID uint) (map[uint]uint, error)
	MockGetInterface                     func(FabricID uint, DeviceID uint,
		InterfaceType string, InterfaceName string) (domain.Interface, error)
	MockGetInterfaceOnMac func(Mac string, FabricID uint) (domain.Interface, error)
	MockGetLLDP           func(FabricID uint, DeviceID uint,
		LocalInterfaceType string, LocalInterfaceName string) (domain.LLDP, error)
	MockGetLLDPsonDevice                             func(FabricID uint, DeviceID uint) ([]domain.LLDP, error)
	MockCreateLLDP                                   func(Interface *domain.LLDP) error
	MockDeleteLLDP                                   func(Interface *domain.LLDP) error
	MockGetLLDPOnRemoteMacExcludingMarkedForDeletion func(RemoteMac string, FabricID uint) (domain.LLDP, error)
	MockGetLLDPNeighborsBetweenTwoDevices            func(FabricID uint, DeviceOneID uint, DeviceTwoID uint) ([]domain.LLDPNeighbor, error)
	MockGetLLDPNeighborsOnRemoteDeviceID             func(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.LLDPNeighbor, error)

	MockCreateASN                  func(ASN *domain.ASNAllocationPool) error
	MockDeleteASNPool              func() error
	MockDeleteUsedASNPool          func() error
	MockDeleteASN                  func(ASN *domain.ASNAllocationPool) error
	MockGetNextASNForRole          func(FabricID uint, role string) (domain.ASNAllocationPool, error)
	MockGetASNCountOnRole          func(FabricID uint, role string) (int64, error)
	MockGetASNCountOnASN           func(FabricID uint, asn uint64) (int64, error)
	MockGetASNAndCountOnASNAndRole func(FabricID uint, asn uint64, role string) (int64, domain.ASNAllocationPool)

	MockCreateUsedASN                   func(UsedASN *domain.UsedASN) error
	MockDeleteUsedASN                   func(UsedASN *domain.UsedASN) error
	MockGetUsedASNOnASNAndDeviceAndRole func(FabricID uint, DeviceID uint, asn uint64, role string) (domain.UsedASN, error)
	MockGetUsedASNCountOnASNAndDevice   func(FabricID uint, asn uint64, DeviceID uint) (int64, error)
	MockGetUsedASNCountOnASNAndRole     func(FabricID uint, asn uint64, role string) (int64, error)

	MockCreateIPEntry                                 func(IPEntry *domain.IPAllocationPool) error
	MockDeleteIPEntry                                 func(IPEntry *domain.IPAllocationPool) error
	MockDeleteIPPool                                  func() error
	MockGetIPEntryAndCountOnIPAddressAndType          func(FabricID uint, ipaddress string, IPType string) (int64, domain.IPAllocationPool, error)
	MockGetNextIPEntryOnType                          func(FabricID uint, IPType string) (domain.IPAllocationPool, error)
	MockGetUsedIPOnDeviceInterfaceIDIPAddresssAndType func(FabricID uint, DeviceID uint, ipaddress string, IPType string, InterfaceID uint) (domain.UsedIP, error)
	MockGetUsedIPOnDeviceInterfaceIDAndType           func(FabricID uint, DeviceID uint, IPType string, InterfaceId uint) (domain.UsedIP, error)
	MockCreateUsedIPEntry                             func(UsedIPEntry *domain.UsedIP) error
	MockDeleteUsedIPEntry                             func(UsedIPEntry *domain.UsedIP) error
	MockDeleteUsedIPPool                              func() error

	MockCreateIPPairEntry                                 func(IPEntry *domain.IPPairAllocationPool) error
	MockDeleteIPPairEntry                                 func(IPEntry *domain.IPPairAllocationPool) error
	MockDeleteIPPairPool                                  func() error
	MockGetIPPairEntryAndCountOnEitherIPAddressAndType    func(FabricID uint, ipaddress string, IPType string) (int64, domain.IPPairAllocationPool, error)
	MockGetIPPairEntryAndCountOnIPAddressAndType          func(FabricID uint, ipaddressOne string, ipaddressTwo string, IPType string) (int64, domain.IPPairAllocationPool, error)
	MockGetNextIPPairEntryOnType                          func(FabricID uint, IPType string) (domain.IPPairAllocationPool, error)
	MockGetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType func(FabricID uint, DeviceOneID uint, DeviceTwoID uint, ipaddressOne string, ipaddressTwo string, IPType string,
		InterfaceOneId uint, InterfaceTwoId uint) (domain.UsedIPPair, error)
	MockGetUsedIPPairOnDeviceInterfaceIDAndType func(FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPType string, InterfaceOneId uint, InterfaceTwoId uint) (domain.UsedIPPair, error)
	MockCreateUsedIPPairEntry                   func(UsedIPEntry *domain.UsedIPPair) error
	MockDeleteUsedIPPairEntry                   func(UsedIPEntry *domain.UsedIPPair) error
	MockDeleteUsedIPPairPool                    func() error

	MockGetLLDPNeighbor func(FabricID uint, DeviceOneID uint,
		DeviceTwoID uint, InterfaceOneID uint, InterfaceTwoID uint) (domain.LLDPNeighbor, error)
	MockCreateLLDPNeighbor                                 func(BGPNeighbor *domain.LLDPNeighbor) error
	MockDeleteLLDPNeighbor                                 func(BGPNeighbor *domain.LLDPNeighbor) error
	MockGetLLDPNeighborsOnDevice                           func(FabricID uint, DeviceId uint) ([]domain.LLDPNeighbor, error)
	MockGetLLDPNeighborsOnDeviceExcludingMarkedForDeletion func(FabricID uint, DeviceId uint) ([]domain.LLDPNeighbor, error)
	MockGetLLDPNeighborsOnDeviceMarkedForDeletion          func(FabricID uint, DeviceId uint) ([]domain.LLDPNeighbor, error)

	MockGetLLDPNeighborsOnEitherDevice func(FabricID uint, DeviceId uint) ([]domain.LLDPNeighbor, error)

	MockCreateSwitchConfig                                                func(SwitchConfig *domain.SwitchConfig) error
	MockUpdateSwitchConfigsASConfigType                                   func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockUpdateSwitchConfigsLoopbackConfigType                             func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockUpdateSwitchConfigsVTEPLoopbackConfigType                         func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockGetSwitchConfigOnFabricIDAndDeviceID                              func(FabricID uint, DeviceID uint) (domain.SwitchConfig, error)
	MockGetSwitchConfigs                                                  func(FabricName string) ([]domain.SwitchConfig, error)
	MockGetSwitchConfigOnDeviceIP                                         func(FabricName string, DeviceIP string) (domain.SwitchConfig, error)
	MockCreateInterfaceSwitchConfig                                       func(SwitchConfig *domain.InterfaceSwitchConfig) error
	MockUpdateInterfaceSwitchConfigsOnInterfaceIDConfigType               func(FabricID uint, interfaceId uint, configType string) error
	MockUpdateInterfaceSwitchConfigsConfigType                            func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockGetInterfaceSwitchConfigOnFabricIDAndInterfaceID                  func(FabricID uint, InterfaceID uint) (domain.InterfaceSwitchConfig, error)
	MockGetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID             func(FabricID uint, InterfaceID uint) int64
	MockGetInterfaceSwitchConfigsOnDeviceID                               func(FabricID uint, DeviceID uint) ([]domain.InterfaceSwitchConfig, error)
	MockGetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion func(FabricID uint, InterfaceIDs []uint) ([]domain.InterfaceSwitchConfig, error)
	MockUpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs            func(FabricID uint, InterfaceIDs []uint, configType string) error

	MockGetBGPSwitchConfigsOnDeviceID                         func(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error)
	MockGetBGPSwitchConfigsOnRemoteDeviceID                   func(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error)
	MockGetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID      func(FabricID uint, RemoteInterfaceID uint) (domain.RemoteNeighborSwitchConfig, error)
	MockGetMCTBGPSwitchConfigsOnDeviceID                      func(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error)
	MockGetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID func(FabricID uint, RemoteInterfaceID uint) int64
	MockCreateBGPSwitchConfig                                 func(BGPSwitchConfig *domain.RemoteNeighborSwitchConfig) error
	MockUpdateBGPSwitchConfigsOnInterfaceIDConfigType         func(FabricID uint, InterfaceID uint, configType string) error
	MockUpdateBGPSwitchConfigsConfigType                      func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockGetBGPSwitchConfigs                                   func(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error)
	MockGetBGPSwitchConfigsExcludingMarkedForDeletion         func(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error)
	MockUpdateConfigTypeForBGPSwitchConfigsOnIntefaceID       func(FabricID uint, InterfaceIDs []uint, configType string) error
	MockCreateExecutionLog                                    func(ExecutionLog *domain.ExecutionLog) error
	MockGetExecutionLogList                                   func(limit int, status string) ([]domain.ExecutionLog, error)
	MockGetExecutionLogByUUID                                 func(string) (domain.ExecutionLog, error)
	MockUpdateExecutionLog                                    func(ExecutionLog *domain.ExecutionLog) error
	//MCT MOCKS
	MockCreateMctClusterConfig func(MCTConfig *domain.MCTClusterDetails) error
	MockDeleteMCTCluster       func(DeviceID uint) error
	MockGetMCTCluster          func(DeviceID uint) (domain.MCTClusterDetails, error)

	MockGetMctClusters                   func(FabricID uint, DeviceId uint, ConfigType []string) ([]domain.MctClusterConfig, error)
	MockGetMctClustersCount              func(FabricID uint, DeviceId uint, ConfigType []string) (uint64, error)
	MockCreateMctClusters                func(MCTConfig *domain.MctClusterConfig) error
	MockGetMctMemberPortsConfig          func(FabricID uint, DeviceID uint, RemoteDeviceID uint, ConfigType []string) ([]domain.MCTMemberPorts, error)
	MockCreateMctClustersMembers         func(MCTConfigs []domain.MCTMemberPorts, ClusterId uint16, ConfigType string) error
	MockDeleteMCTPortConfig              func(MctPorts []domain.MCTMemberPorts) error
	MockUpdateMctPortConfigType          func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockUpdateMctClusterConfigType       func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockDeleteMctPortsMarkedForDelete    func(FabricID uint) error
	MockDeleteMctClustersMarkedForDelete func(FabricID uint) error
	MockGetDeviceUsingDeviceID           func(FabricId uint, DeviceID uint) (domain.Device, error)
	//
	MockMarkMctClusterForDelete                           func(FabricID uint, DeviceID uint) error
	MockMarkMctClusterMemberPortsForDelete                func(FabricID uint, DeviceID uint) error
	MockMarkMctClusterForDeleteWithBothDevices            func(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error
	MockMarkMctClusterMemberPortsForDeleteWithBothDevices func(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error
	MockGetMctClusterConfigWithDeviceIP                   func(IPAddress string) ([]domain.MctClusterConfig, error)
	MockDeleteMctClustersUsingClusterObject               func(OldMct domain.MctClusterConfig) error
	MockUpdateLLDPConfigType                              func(FabricID uint, QueryconfigTypes []string, configType string) error
	MockDeleteLLDPMarkedForDelete                         func(FabricID uint) error
	MockDeleteMctClustersWithMgmtIP                       func(IPAddress string) error
	MockMarkMctClusterMemberPortsForCreate                func(FabricID uint, DeviceID uint, RemoteDeviceID uint) error
	MockGetMctClustersWithBothDevices                     func(FabricID uint, DeviceID uint, NeighborDeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error)
}

//GetMctClustersWithBothDevices Get MCT cluster for Device and its Neighbor
func (db *DatabaseRepository) GetMctClustersWithBothDevices(FabricID uint, DeviceID uint, NeighborDeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error) {
	if db.MockGetMctClustersWithBothDevices != nil {
		return db.MockGetMctClustersWithBothDevices(FabricID, DeviceID, NeighborDeviceID, ConfigType)
	}
	return []domain.MctClusterConfig{}, nil
}

//GetMctClusters represents a mock  GetMctClusters
func (db *DatabaseRepository) GetMctClusters(FabricID uint, DeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error) {
	if db.MockGetMctClusters != nil {
		return db.MockGetMctClusters(FabricID, DeviceID, ConfigType)
	}
	return []domain.MctClusterConfig{}, nil
}

//GetMctClustersCount represents a mock GetMctClustersCount
func (db *DatabaseRepository) GetMctClustersCount(FabricID uint, DeviceID uint, ConfigType []string) (uint64, error) {

	if db.MockGetMctClustersCount != nil {
		return db.MockGetMctClustersCount(FabricID, DeviceID, ConfigType)
	}
	return 0, nil
}

//CreateMctClusters represents a mock CreateMctClusters
func (db *DatabaseRepository) CreateMctClusters(MCTConfig *domain.MctClusterConfig) error {
	if db.MockCreateMctClusters != nil {
		return db.MockCreateMctClusters(MCTConfig)
	}
	return nil
}

//GetMctMemberPortsConfig represents a mock GetMctMemberPortsConfig
func (db *DatabaseRepository) GetMctMemberPortsConfig(FabricID uint, DeviceID uint, RemoteDeviceID uint, ConfigType []string) ([]domain.MCTMemberPorts, error) {
	if db.MockGetMctMemberPortsConfig != nil {
		return db.MockGetMctMemberPortsConfig(FabricID, DeviceID, RemoteDeviceID, ConfigType)
	}
	return []domain.MCTMemberPorts{}, nil
}

//CreateMctClustersMembers represents a mock CreateMctClustersMembers
func (db *DatabaseRepository) CreateMctClustersMembers(MCTConfigs []domain.MCTMemberPorts, ClusterID uint16, ConfigType string) error {
	if db.MockCreateMctClustersMembers != nil {
		return db.MockCreateMctClustersMembers(MCTConfigs, ClusterID, ConfigType)
	}
	return nil
}

//DeleteMCTPortConfig represents a mock DeleteMCTPortConfig
func (db *DatabaseRepository) DeleteMCTPortConfig(MctPorts []domain.MCTMemberPorts) error {
	if db.MockDeleteMCTPortConfig != nil {
		return db.MockDeleteMCTPortConfig(MctPorts)
	}
	return nil
}

//UpdateMctPortConfigType represents a mock UpdateMctPortConfigType
func (db *DatabaseRepository) UpdateMctPortConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateMctPortConfigType != nil {
		return db.MockUpdateMctPortConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil
}

//UpdateMctClusterConfigType represents a mock UpdateMctClusterConfigType
func (db *DatabaseRepository) UpdateMctClusterConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateMctClusterConfigType != nil {
		return db.MockUpdateMctClusterConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil

}

//DeleteMctPortsMarkedForDelete represents a mock DeleteMctPortsMarkedForDelete
func (db *DatabaseRepository) DeleteMctPortsMarkedForDelete(FabricID uint) error {
	if db.MockDeleteMctPortsMarkedForDelete != nil {
		return db.MockDeleteMctPortsMarkedForDelete(FabricID)
	}
	return nil
}

//DeleteMctClustersMarkedForDelete represents a mock DeleteMctClustersMarkedForDelete
func (db *DatabaseRepository) DeleteMctClustersMarkedForDelete(FabricID uint) error {
	if db.MockDeleteMctClustersMarkedForDelete != nil {
		return db.MockDeleteMctClustersMarkedForDelete(FabricID)
	}
	return nil
}

//CreateMctClusterConfig represetnts a mock CreateMctClusterConfig
func (db *DatabaseRepository) CreateMctClusterConfig(MCTConfig *domain.MCTClusterDetails) error {
	if db.MockCreateMctClusterConfig != nil {
		return db.MockCreateMctClusterConfig(MCTConfig)
	}
	return nil
}

//DeleteMCTCluster represents a mock DeleteMCTCluster
func (db *DatabaseRepository) DeleteMCTCluster(DeviceID uint) error {
	if db.MockDeleteMCTCluster != nil {
		return db.MockDeleteMCTCluster(DeviceID)
	}
	return nil
}

//GetMCTCluster represents a mock GetMCTCluster
func (db *DatabaseRepository) GetMCTCluster(DeviceID uint) (domain.MCTClusterDetails, error) {
	if db.MockGetMCTCluster != nil {
		return db.MockGetMCTCluster(DeviceID)
	}
	return domain.MCTClusterDetails{}, nil
}

//GetFabric represents a mock GetFabric
func (db *DatabaseRepository) GetFabric(FabricName string) (domain.Fabric, error) {
	if db.MockGetFabric != nil {
		return db.MockGetFabric(FabricName)
	}
	return domain.Fabric{Name: FabricName, ID: 1}, nil
}

//CreateFabric represents a mock CreateFabric
func (db *DatabaseRepository) CreateFabric(Fabric *domain.Fabric) error {
	if db.MockCreateFabric != nil {
		return db.MockCreateFabric(Fabric)
	}
	return nil
}

//DeleteFabric represents a mock DeleteFabric
func (db *DatabaseRepository) DeleteFabric(FabricName string) error {
	if db.MockDeleteFabric != nil {
		return db.MockDeleteFabric(FabricName)
	}
	return nil
}

//GetFabricProperties represents a mock GetFabricProperties
func (db *DatabaseRepository) GetFabricProperties(FabricID uint) (domain.FabricProperties, error) {
	if db.MockGetFabricProperties != nil {
		return db.MockGetFabricProperties(FabricID)
	}
	return domain.FabricProperties{}, nil
}

//CreateFabricProperties represents a mock CreateFabricProperties
func (db *DatabaseRepository) CreateFabricProperties(FabricProperties *domain.FabricProperties) error {
	if db.MockCreateFabricProperties != nil {
		return db.MockCreateFabricProperties(FabricProperties)
	}
	return nil
}

//UpdateFabricProperties represents a mock UpdateFabricProperties
func (db *DatabaseRepository) UpdateFabricProperties(FabricProperties *domain.FabricProperties) error {
	if db.MockUpdateFabricProperties != nil {
		return db.MockUpdateFabricProperties(FabricProperties)
	}
	return nil
}

//GetDevice represents a mock GetDevice
func (db *DatabaseRepository) GetDevice(FabricName string, IPAddress string) (domain.Device, error) {
	if db.MockGetDevice != nil {
		return db.MockGetDevice(FabricName, IPAddress)
	}
	return domain.Device{IPAddress: IPAddress, UserName: userName}, nil
}

//GetDeviceInAnyFabric represents a mock GetDeviceInAnyFabric
func (db *DatabaseRepository) GetDeviceInAnyFabric(IPAddress string) (domain.Device, error) {
	if db.MockGetDeviceInAnyFabric != nil {
		return db.MockGetDeviceInAnyFabric(IPAddress)
	}
	return domain.Device{IPAddress: IPAddress, UserName: userName}, nil
}

//GetDevicesInFabric represents a mock GetDevicesInFabric
func (db *DatabaseRepository) GetDevicesInFabric(FabricID uint) ([]domain.Device, error) {
	if db.MockGetDevicesInFabric != nil {
		return db.MockGetDevicesInFabric(FabricID)
	}
	return []domain.Device{}, nil
}

//GetDevicesInFabricMatching represents a mock GetDevicesInFabric
func (db *DatabaseRepository) GetDevicesInFabricMatching(FabricID uint, device []string) ([]domain.Device, error) {
	if db.MockGetDevicesInFabricMatching != nil {
		return db.MockGetDevicesInFabricMatching(FabricID, device)
	}
	return []domain.Device{}, nil
}

//GetDevicesInFabricNotMatching represents a mock GetDevicesInFabric
func (db *DatabaseRepository) GetDevicesInFabricNotMatching(FabricID uint, device []string) ([]domain.Device, error) {
	if db.MockGetDevicesInFabricNotMatching != nil {
		return db.MockGetDevicesInFabricNotMatching(FabricID, device)
	}
	return []domain.Device{}, nil
}

//GetDevicesCountInFabric represents a mock GetDevicesCountInFabric
func (db *DatabaseRepository) GetDevicesCountInFabric(FabricID uint) uint16 {
	if db.MockGetDevicesCountInFabric != nil {
		return db.MockGetDevicesCountInFabric(FabricID)
	}
	return 0
}

//OpenTransaction represents a mock OpenTransaction
func (db *DatabaseRepository) OpenTransaction() error {
	if db.MockOpenTransaction != nil {
		return db.MockOpenTransaction()
	}
	return nil
}

//CommitTransaction represents a mock CommitTransaction
func (db *DatabaseRepository) CommitTransaction() error {
	if db.MockCommitTransaction != nil {
		return db.MockCommitTransaction()
	}
	return nil
}

//RollBackTransaction represents a mock RollBackTransaction
func (db *DatabaseRepository) RollBackTransaction() error {
	if db.MockRollBackTransaction != nil {
		return db.MockRollBackTransaction()
	}
	return nil
}

//Backup represents Database
func (db *DatabaseRepository) Backup() error {
	if db.MockBackup != nil {
		return db.MockBackup()
	}
	return nil
}

//CreateDevice represents a mock CreateDevice
func (db *DatabaseRepository) CreateDevice(Device *domain.Device) error {
	if db.MockCreateDevice != nil {
		return db.MockCreateDevice(Device)
	}
	return nil
}

//DeleteDevice represents a mock DeleteDevice
func (db *DatabaseRepository) DeleteDevice(DeviceID []uint) error {
	if db.MockDeleteDevice != nil {
		return db.MockDeleteDevice(DeviceID)
	}
	return nil
}

//SaveDevice represents a mock SaveDevice
func (db *DatabaseRepository) SaveDevice(Device *domain.Device) error {
	if db.MockSaveDevice != nil {
		return db.MockSaveDevice(Device)
	}
	return nil
}

//GetRack returns an instance of Rack for a given "FabricName, Rack Mgmt IPAddress" i/p
func (db *DatabaseRepository) GetRack(FabricName string, IP1 string, IP2 string) (domain.Rack, error) {

	if db.MockGetRack != nil {
		return db.MockGetRack(FabricName, IP1, IP2)
	}

	return domain.Rack{}, nil
}

//GetRackbyIP returns an instance of Rack for a given "FabricName, Rack Mgmt IPAddress" i/p
func (db *DatabaseRepository) GetRackbyIP(FabricName string, IP string) (domain.Rack, error) {
	if db.MockGetRackbyIP != nil {
		return db.MockGetRackbyIP(FabricName, IP)
	}

	return domain.Rack{}, nil
}

//GetRackAll returns all instance of Rack for a given "FabricName" i/p
func (db *DatabaseRepository) GetRackAll(FabricName string) ([]domain.Rack, error) {
	if db.MockGetRackAll != nil {
		return db.MockGetRackAll(FabricName)
	}

	return []domain.Rack{}, nil
}

//GetRacksInFabric returns an array of devices in a given fabric
func (db *DatabaseRepository) GetRacksInFabric(FabricID uint) ([]domain.Rack, error) {

	if db.MockGetRacksInFabric != nil {
		return db.MockGetRacksInFabric(FabricID)
	}

	return []domain.Rack{}, nil
}

//CreateRack creates an instance of Rack in the database
func (db *DatabaseRepository) CreateRack(Rack *domain.Rack) error {

	if db.MockCreateRack != nil {
		return db.MockCreateRack(Rack)
	}

	return nil
}

//DeleteRack deletes instances of Rack for a given array of RackID
func (db *DatabaseRepository) DeleteRack(RackID []uint) error {
	if db.MockDeleteRack != nil {
		return db.MockDeleteRack(RackID)
	}
	return nil
}

//SaveRack creates an instance of Rack in the database
func (db *DatabaseRepository) SaveRack(Rack *domain.Rack) error {
	if db.MockSaveRack != nil {
		return db.MockSaveRack(Rack)
	}
	return nil
}

//CreateRackEvpnConfig creates an instance of "RackEvpnNeighbors" in the database
func (db *DatabaseRepository) CreateRackEvpnConfig(RackEvpnConfig *domain.RackEvpnNeighbors) error {

	if db.MockCreateRackEvpnConfig != nil {
		return db.MockCreateRackEvpnConfig(RackEvpnConfig)
	}
	return nil

}

//GetRackEvpnConfig returns RackEvpnNeighbors
func (db *DatabaseRepository) GetRackEvpnConfig(RackID uint) ([]domain.RackEvpnNeighbors, error) {
	if db.MockGetRackEvpnConfig != nil {
		return db.MockGetRackEvpnConfig(RackID)
	}
	return []domain.RackEvpnNeighbors{}, nil
}

//GetRackEvpnConfigOnDeviceID returns RackEvpnNeighbors
func (db *DatabaseRepository) GetRackEvpnConfigOnDeviceID(DeviceID uint) ([]domain.RackEvpnNeighbors, error) {
	if db.MockGetRackEvpnConfigOnDeviceID != nil {
		return db.MockGetRackEvpnConfigOnDeviceID(DeviceID)
	}

	return []domain.RackEvpnNeighbors{}, nil
}

//GetRackEvpnConfigOnRemoteDeviceID returns RackEvpnNeighbors
func (db *DatabaseRepository) GetRackEvpnConfigOnRemoteDeviceID(DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RackEvpnNeighbors, error) {
	if db.MockGetRackEvpnConfigOnRemoteDeviceID != nil {
		return db.MockGetRackEvpnConfigOnRemoteDeviceID(DeviceID, RemoteDeviceIDs)
	}

	return []domain.RackEvpnNeighbors{}, nil
}

//GetLLDPNeighborsBetweenTwoDevices returns array of instances of "domain.LLDPNeighbor" between two devices in fabric
func (db *DatabaseRepository) GetLLDPNeighborsBetweenTwoDevices(FabricID uint, DeviceOneID uint, DeviceTwoID uint) ([]domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighborsBetweenTwoDevices != nil {
		return db.MockGetLLDPNeighborsBetweenTwoDevices(FabricID, DeviceOneID, DeviceTwoID)
	}
	return []domain.LLDPNeighbor{}, nil
}

//GetLLDPNeighborsOnRemoteDeviceID returns array of instances of "domain.LLDPNeighbor" between device and remote device
func (db *DatabaseRepository) GetLLDPNeighborsOnRemoteDeviceID(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighborsOnRemoteDeviceID != nil {
		return db.MockGetLLDPNeighborsOnRemoteDeviceID(FabricID, DeviceID, RemoteDeviceIDs)
	}
	return []domain.LLDPNeighbor{}, nil
}

//GetInterface represents a mock GetInterface
func (db *DatabaseRepository) GetInterface(FabricID uint, DeviceID uint,
	InterfaceType string, InterfaceName string) (domain.Interface, error) {
	if db.MockGetInterface != nil {
		return db.MockGetInterface(FabricID, DeviceID, InterfaceType, InterfaceName)
	}
	return domain.Interface{}, nil
}

//CreateInterface represents a mock CreateInterface
func (db *DatabaseRepository) CreateInterface(Interface *domain.Interface) error {
	if db.MockCreateInterface != nil {
		return db.MockCreateInterface(Interface)
	}
	return nil
}

//ResetDeleteOperationOnInterface represents a mock ResetDeleteOperationOnInterface
func (db *DatabaseRepository) ResetDeleteOperationOnInterface(FabricID uint, InterfaceID uint) error {
	if db.MockResetDeleteOperationOnInterface != nil {
		return db.MockResetDeleteOperationOnInterface(FabricID, InterfaceID)
	}
	return nil
}

//DeleteInterface represents a mock DeleteInterface
func (db *DatabaseRepository) DeleteInterface(Interface *domain.Interface) error {
	if db.MockDeleteInterface != nil {
		return db.MockDeleteInterface(Interface)
	}
	return nil
}

//DeleteInterfaceUsingID represents a mock DeleteInterfaceUsingID
func (db *DatabaseRepository) DeleteInterfaceUsingID(FabricID uint, InterfaceID uint) error {
	if db.MockDeleteInterfaceUsingID != nil {
		return db.MockDeleteInterfaceUsingID(FabricID, InterfaceID)
	}
	return nil
}

//GetInterfacesonDevice represents a mock GetInterfacesonDevice
func (db *DatabaseRepository) GetInterfacesonDevice(FabricID uint, DeviceID uint) ([]domain.Interface, error) {
	if db.MockGetInterfacesonDevice != nil {
		return db.MockGetInterfacesonDevice(FabricID, DeviceID)
	}
	return []domain.Interface{}, nil
}

//GetInterfaceIDsMarkedForDeletion represents a mock GetInterfaceIDsMarkedForDeletion
func (db *DatabaseRepository) GetInterfaceIDsMarkedForDeletion(FabricID uint) (map[uint]uint, error) {
	if db.MockGetInterfaceIDsMarkedForDeletion != nil {
		return db.MockGetInterfaceIDsMarkedForDeletion(FabricID)
	}
	return map[uint]uint{}, nil
}

//GetInterfaceOnMac represents a mock GetInterfaceOnMac
func (db *DatabaseRepository) GetInterfaceOnMac(Mac string, FabricID uint) (domain.Interface, error) {
	if db.MockGetInterfaceOnMac != nil {
		return db.MockGetInterfaceOnMac(Mac, FabricID)
	}
	return domain.Interface{}, nil
}

//GetLLDP represents a mock GetLLDP
func (db *DatabaseRepository) GetLLDP(FabricID uint, DeviceID uint,
	LocalInterfaceType string, LocalInterfaceName string) (domain.LLDP, error) {
	if db.MockGetLLDP != nil {
		return db.MockGetLLDP(FabricID, DeviceID, LocalInterfaceType, LocalInterfaceName)
	}
	return domain.LLDP{}, nil
}

//GetLLDPsonDevice represents a mock GetLLDPsonDevice
func (db *DatabaseRepository) GetLLDPsonDevice(FabricID uint, DeviceID uint) ([]domain.LLDP, error) {
	if db.MockGetLLDPsonDevice != nil {
		return db.MockGetLLDPsonDevice(FabricID, DeviceID)
	}
	return []domain.LLDP{}, nil
}

//CreateLLDP represents a mock CreateLLDP
func (db *DatabaseRepository) CreateLLDP(Interface *domain.LLDP) error {
	if db.MockCreateLLDP != nil {
		return db.MockCreateLLDP(Interface)
	}
	return nil
}

//DeleteLLDP represents a mock DeleteLLDP
func (db *DatabaseRepository) DeleteLLDP(Interface *domain.LLDP) error {
	if db.MockDeleteLLDP != nil {
		return db.MockDeleteLLDP(Interface)
	}
	return nil
}

//GetLLDPOnRemoteMacExcludingMarkedForDeletion represents a mock GetLLDPOnRemoteMacExcludingMarkedForDeletion
func (db *DatabaseRepository) GetLLDPOnRemoteMacExcludingMarkedForDeletion(RemoteMac string, FabricID uint) (domain.LLDP, error) {
	if db.MockGetLLDPOnRemoteMacExcludingMarkedForDeletion != nil {
		return db.MockGetLLDPOnRemoteMacExcludingMarkedForDeletion(RemoteMac, FabricID)
	}
	return domain.LLDP{}, nil
}

//CreateASN represents a mock CreateASN
func (db *DatabaseRepository) CreateASN(ASN *domain.ASNAllocationPool) error {
	if db.MockCreateASN != nil {
		return db.MockCreateASN(ASN)
	}
	return nil
}

//DeleteASNPool represents a mock DeleteASNPool
func (db *DatabaseRepository) DeleteASNPool() error {
	if db.MockDeleteASNPool != nil {
		return db.MockDeleteASNPool()
	}
	return nil
}

//DeleteUsedASNPool represents a mock DeleteUsedASNPool
func (db *DatabaseRepository) DeleteUsedASNPool() error {
	if db.MockDeleteUsedASNPool != nil {
		return db.MockDeleteUsedASNPool()
	}
	return nil
}

//DeleteASN represents a mock DeleteASN
func (db *DatabaseRepository) DeleteASN(ASN *domain.ASNAllocationPool) error {
	if db.MockDeleteASN != nil {
		return db.MockDeleteASN(ASN)
	}
	return nil
}

//CreateUsedASN represents a mock CreateUsedASN
func (db *DatabaseRepository) CreateUsedASN(UsedASN *domain.UsedASN) error {
	if db.MockCreateUsedASN != nil {
		return db.MockCreateUsedASN(UsedASN)
	}
	return nil
}

//DeleteUsedASN represents a mock DeleteUsedASN
func (db *DatabaseRepository) DeleteUsedASN(UsedASN *domain.UsedASN) error {
	if db.MockDeleteUsedASN != nil {
		return db.MockDeleteUsedASN(UsedASN)
	}
	return nil
}

//GetASNCountOnASN represents a mock GetASNCountOnASN
func (db *DatabaseRepository) GetASNCountOnASN(FabricID uint, asn uint64) (int64, error) {
	if db.MockGetASNCountOnASN != nil {
		return db.MockGetASNCountOnASN(FabricID, asn)
	}
	return 0, nil
}

//GetASNCountOnRole represents a mock GetASNCountOnRole
func (db *DatabaseRepository) GetASNCountOnRole(FabricID uint, role string) (int64, error) {
	if db.MockGetASNCountOnRole != nil {
		return db.MockGetASNCountOnRole(FabricID, role)
	}
	return 0, nil
}

//GetNextASNForRole represents a mock GetNextASNForRole
func (db *DatabaseRepository) GetNextASNForRole(FabricID uint, role string) (domain.ASNAllocationPool, error) {
	if db.MockGetNextASNForRole != nil {
		return db.MockGetNextASNForRole(FabricID, role)
	}
	return domain.ASNAllocationPool{}, nil
}

//GetASNAndCountOnASNAndRole represents a mock GetASNAndCountOnASNAndRole
func (db *DatabaseRepository) GetASNAndCountOnASNAndRole(FabricID uint, asn uint64, role string) (int64, domain.ASNAllocationPool) {
	if db.MockGetASNAndCountOnASNAndRole != nil {
		return db.MockGetASNAndCountOnASNAndRole(FabricID, asn, role)
	}
	return 0, domain.ASNAllocationPool{}
}

//GetUsedASNOnASNAndDeviceAndRole represents a mock GetUsedASNOnASNAndDeviceAndRole
func (db *DatabaseRepository) GetUsedASNOnASNAndDeviceAndRole(FabricID uint, DeviceID uint, asn uint64, role string) (domain.UsedASN, error) {
	if db.MockGetUsedASNOnASNAndDeviceAndRole != nil {
		return db.MockGetUsedASNOnASNAndDeviceAndRole(FabricID, DeviceID, asn, role)
	}
	return domain.UsedASN{}, nil
}

//GetUsedASNCountOnASNAndDevice represents a mock GetUsedASNCountOnASNAndDevice
func (db *DatabaseRepository) GetUsedASNCountOnASNAndDevice(FabricID uint, asn uint64, DeviceID uint) (int64, error) {
	if db.MockGetUsedASNCountOnASNAndDevice != nil {
		return db.MockGetUsedASNCountOnASNAndDevice(FabricID, asn, DeviceID)
	}
	return 0, nil
}

//GetUsedASNCountOnASNAndRole represents a mock GetUsedASNCountOnASNAndRole
func (db *DatabaseRepository) GetUsedASNCountOnASNAndRole(FabricID uint, asn uint64, role string) (int64, error) {
	if db.MockGetUsedASNCountOnASNAndRole != nil {
		return db.MockGetUsedASNCountOnASNAndRole(FabricID, asn, role)
	}
	return 0, nil
}

//CreateIPEntry represents a mock CreateIPEntry
func (db *DatabaseRepository) CreateIPEntry(IPEntry *domain.IPAllocationPool) error {
	if db.MockCreateIPEntry != nil {
		return db.MockCreateIPEntry(IPEntry)
	}
	return nil
}

//DeleteIPEntry represents a mock DeleteIPEntry
func (db *DatabaseRepository) DeleteIPEntry(IPEntry *domain.IPAllocationPool) error {
	if db.MockDeleteIPEntry != nil {
		return db.MockDeleteIPEntry(IPEntry)
	}
	return nil
}

//DeleteIPPool represents a mock DeleteIPPool
func (db *DatabaseRepository) DeleteIPPool() error {
	if db.MockDeleteIPPool != nil {
		return db.MockDeleteIPPool()
	}
	return nil
}

//GetIPEntryAndCountOnIPAddressAndType represents a mock GetIPEntryAndCountOnIPAddressAndType
func (db *DatabaseRepository) GetIPEntryAndCountOnIPAddressAndType(FabricID uint, ipaddress string, IPType string) (int64, domain.IPAllocationPool, error) {
	if db.MockGetIPEntryAndCountOnIPAddressAndType != nil {
		return db.MockGetIPEntryAndCountOnIPAddressAndType(FabricID, ipaddress, IPType)
	}
	return 0, domain.IPAllocationPool{}, nil
}

//GetNextIPEntryOnType represents a mock GetNextIPEntryOnType
func (db *DatabaseRepository) GetNextIPEntryOnType(FabricID uint, IPType string) (domain.IPAllocationPool, error) {
	if db.MockGetNextIPEntryOnType != nil {
		return db.MockGetNextIPEntryOnType(FabricID, IPType)
	}
	return domain.IPAllocationPool{}, nil
}

//GetUsedIPOnDeviceInterfaceIDIPAddresssAndType represents a mock GetUsedIPOnDeviceInterfaceIDIPAddresssAndType
func (db *DatabaseRepository) GetUsedIPOnDeviceInterfaceIDIPAddresssAndType(FabricID uint, DeviceID uint, IPAddress string, IPType string, InterfaceID uint) (domain.UsedIP, error) {
	if db.MockGetUsedIPOnDeviceInterfaceIDIPAddresssAndType != nil {
		return db.MockGetUsedIPOnDeviceInterfaceIDIPAddresssAndType(FabricID, DeviceID, IPAddress, IPType, InterfaceID)
	}
	return domain.UsedIP{}, nil
}

//GetUsedIPOnDeviceInterfaceIDAndType represents a mock GetUsedIPOnDeviceInterfaceIDAndType
func (db *DatabaseRepository) GetUsedIPOnDeviceInterfaceIDAndType(FabricID uint, DeviceID uint, IPType string, InterfaceID uint) (domain.UsedIP, error) {
	if db.MockGetUsedIPOnDeviceInterfaceIDAndType != nil {
		return db.MockGetUsedIPOnDeviceInterfaceIDAndType(FabricID, DeviceID, IPType, InterfaceID)
	}
	return domain.UsedIP{}, nil
}

//CreateUsedIPEntry represents a mock CreateUsedIPEntry
func (db *DatabaseRepository) CreateUsedIPEntry(UsedIPEntry *domain.UsedIP) error {
	if db.MockCreateUsedIPEntry != nil {
		return db.MockCreateUsedIPEntry(UsedIPEntry)
	}
	return nil
}

//DeleteUsedIPEntry represents a mock DeleteUsedIPEntry
func (db *DatabaseRepository) DeleteUsedIPEntry(UsedIPEntry *domain.UsedIP) error {
	if db.MockDeleteUsedIPEntry != nil {
		return db.MockDeleteUsedIPEntry(UsedIPEntry)
	}
	return nil
}

//DeleteUsedIPPool represents a mock DeleteUsedIPPool
func (db *DatabaseRepository) DeleteUsedIPPool() error {
	if db.MockDeleteUsedIPPool != nil {
		return db.MockDeleteUsedIPPool()
	}
	return nil
}

//CreateIPPairEntry represents a mock CreateIPPairEntry
func (db *DatabaseRepository) CreateIPPairEntry(IPEntry *domain.IPPairAllocationPool) error {
	if db.MockCreateIPPairEntry != nil {
		return db.MockCreateIPPairEntry(IPEntry)
	}
	return nil
}

//DeleteIPPairEntry represents a mock DeleteIPPairEntry
func (db *DatabaseRepository) DeleteIPPairEntry(IPEntry *domain.IPPairAllocationPool) error {
	if db.MockDeleteIPPairEntry != nil {
		return db.MockDeleteIPPairEntry(IPEntry)
	}
	return nil
}

//DeleteIPPairPool represents a mock DeleteIPPairPool
func (db *DatabaseRepository) DeleteIPPairPool() error {
	if db.MockDeleteIPPairPool != nil {
		return db.MockDeleteIPPairPool()
	}
	return nil
}

//GetIPPairEntryAndCountOnIPAddressAndType represents a mock GetIPPairEntryAndCountOnIPAddressAndType
func (db *DatabaseRepository) GetIPPairEntryAndCountOnIPAddressAndType(FabricID uint, ipaddressOne string, ipaddressTwo string, IPType string) (int64, domain.IPPairAllocationPool, error) {
	if db.MockGetIPPairEntryAndCountOnIPAddressAndType != nil {
		return db.MockGetIPPairEntryAndCountOnIPAddressAndType(FabricID, ipaddressOne, ipaddressTwo, IPType)
	}
	return 0, domain.IPPairAllocationPool{}, nil
}

//GetIPPairEntryAndCountOnEitherIPAddressAndType represents a mock GetIPPairEntryAndCountOnEitherIPAddressAndType
func (db *DatabaseRepository) GetIPPairEntryAndCountOnEitherIPAddressAndType(FabricID uint, ipaddress string, IPType string) (int64, domain.IPPairAllocationPool, error) {
	if db.MockGetIPPairEntryAndCountOnEitherIPAddressAndType != nil {
		return db.MockGetIPPairEntryAndCountOnEitherIPAddressAndType(FabricID, ipaddress, IPType)
	}
	return 0, domain.IPPairAllocationPool{}, nil
}

//GetNextIPPairEntryOnType represents a mock GetNextIPPairEntryOnType
func (db *DatabaseRepository) GetNextIPPairEntryOnType(FabricID uint, IPType string) (domain.IPPairAllocationPool, error) {
	if db.MockGetNextIPPairEntryOnType != nil {
		return db.MockGetNextIPPairEntryOnType(FabricID, IPType)
	}
	return domain.IPPairAllocationPool{}, nil
}

//GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType represents a mock GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType
func (db *DatabaseRepository) GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPAddressOne string, IPAddressTwo string, IPType string,
	InterfaceOneID uint, InterfaceTwoID uint) (domain.UsedIPPair, error) {
	if db.MockGetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType != nil {
		return db.MockGetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType(FabricID, DeviceOneID, DeviceTwoID, IPAddressOne, IPAddressTwo, IPType, InterfaceOneID, InterfaceTwoID)
	}
	return domain.UsedIPPair{}, nil
}

//GetUsedIPPairOnDeviceInterfaceIDAndType represents a mock GetUsedIPPairOnDeviceInterfaceIDAndType
func (db *DatabaseRepository) GetUsedIPPairOnDeviceInterfaceIDAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPType string, InterfaceOneID uint, InterfaceTwoID uint) (domain.UsedIPPair, error) {
	if db.MockGetUsedIPPairOnDeviceInterfaceIDAndType != nil {
		return db.MockGetUsedIPPairOnDeviceInterfaceIDAndType(FabricID, DeviceOneID, DeviceTwoID, IPType, InterfaceOneID, InterfaceTwoID)
	}
	return domain.UsedIPPair{}, nil
}

//CreateUsedIPPairEntry represents a mock CreateUsedIPPairEntry
func (db *DatabaseRepository) CreateUsedIPPairEntry(UsedIPEntry *domain.UsedIPPair) error {
	if db.MockCreateUsedIPPairEntry != nil {
		return db.MockCreateUsedIPPairEntry(UsedIPEntry)
	}
	return nil
}

//DeleteUsedIPPairEntry represents a mock DeleteUsedIPPairEntry
func (db *DatabaseRepository) DeleteUsedIPPairEntry(UsedIPEntry *domain.UsedIPPair) error {
	if db.MockDeleteUsedIPPairEntry != nil {
		return db.MockDeleteUsedIPPairEntry(UsedIPEntry)
	}
	return nil
}

//DeleteUsedIPPairPool represents a mock DeleteUsedIPPairPool
func (db *DatabaseRepository) DeleteUsedIPPairPool() error {
	if db.MockDeleteUsedIPPairPool != nil {
		return db.MockDeleteUsedIPPairPool()
	}
	return nil
}

//GetLLDPNeighbor represents a mock GetLLDPNeighbor
func (db *DatabaseRepository) GetLLDPNeighbor(FabricID uint, DeviceOneID uint,
	DeviceTwoID uint, InterfaceOneID uint, InterfaceTwoID uint) (domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighbor != nil {
		return db.MockGetLLDPNeighbor(FabricID, DeviceOneID, DeviceTwoID, InterfaceOneID, InterfaceTwoID)
	}
	return domain.LLDPNeighbor{}, nil
}

//CreateLLDPNeighbor represents a mock CreateLLDPNeighbor
func (db *DatabaseRepository) CreateLLDPNeighbor(BGPNeighbor *domain.LLDPNeighbor) error {
	if db.MockCreateLLDPNeighbor != nil {
		return db.MockCreateLLDPNeighbor(BGPNeighbor)
	}
	return nil
}

//DeleteLLDPNeighbor represents a mock DeleteLLDPNeighbor
func (db *DatabaseRepository) DeleteLLDPNeighbor(BGPNeighbor *domain.LLDPNeighbor) error {
	if db.MockDeleteLLDPNeighbor != nil {
		return db.MockDeleteLLDPNeighbor(BGPNeighbor)
	}
	return nil
}

//GetLLDPNeighborsOnEitherDevice represents a mock GetLLDPNeighborsOnEitherDevice
func (db *DatabaseRepository) GetLLDPNeighborsOnEitherDevice(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighborsOnEitherDevice != nil {
		return db.MockGetLLDPNeighborsOnEitherDevice(FabricID, DeviceID)
	}
	return []domain.LLDPNeighbor{}, nil
}

//GetLLDPNeighborsOnDevice represents a mock GetLLDPNeighborsOnDevice
func (db *DatabaseRepository) GetLLDPNeighborsOnDevice(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighborsOnDevice != nil {
		return db.MockGetLLDPNeighborsOnDevice(FabricID, DeviceID)
	}
	return []domain.LLDPNeighbor{}, nil
}

//GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion represents a mock GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion
func (db *DatabaseRepository) GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighborsOnDeviceExcludingMarkedForDeletion != nil {
		return db.MockGetLLDPNeighborsOnDeviceExcludingMarkedForDeletion(FabricID, DeviceID)
	}
	return []domain.LLDPNeighbor{}, nil
}

//GetLLDPNeighborsOnDeviceMarkedForDeletion represents a mock GetLLDPNeighborsOnDeviceMarkedForDeletion
func (db *DatabaseRepository) GetLLDPNeighborsOnDeviceMarkedForDeletion(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	if db.MockGetLLDPNeighborsOnDeviceMarkedForDeletion != nil {
		return db.MockGetLLDPNeighborsOnDeviceMarkedForDeletion(FabricID, DeviceID)
	}
	return []domain.LLDPNeighbor{}, nil
}

//CreateSwitchConfig represents a mock CreateSwitchConfig
func (db *DatabaseRepository) CreateSwitchConfig(SwitchConfig *domain.SwitchConfig) error {
	if db.MockCreateSwitchConfig != nil {
		return db.MockCreateSwitchConfig(SwitchConfig)
	}
	return nil
}

//UpdateSwitchConfigsASConfigType represents a mock UpdateSwitchConfigsASConfigType
func (db *DatabaseRepository) UpdateSwitchConfigsASConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateSwitchConfigsASConfigType != nil {
		return db.MockUpdateSwitchConfigsASConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil
}

//UpdateSwitchConfigsLoopbackConfigType represents a mock UpdateSwitchConfigsLoopbackConfigType
func (db *DatabaseRepository) UpdateSwitchConfigsLoopbackConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateSwitchConfigsLoopbackConfigType != nil {
		return db.MockUpdateSwitchConfigsLoopbackConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil
}

//UpdateSwitchConfigsVTEPLoopbackConfigType represents a mock UpdateSwitchConfigsVTEPLoopbackConfigType
func (db *DatabaseRepository) UpdateSwitchConfigsVTEPLoopbackConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateSwitchConfigsVTEPLoopbackConfigType != nil {
		return db.MockUpdateSwitchConfigsVTEPLoopbackConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil
}

//GetSwitchConfigOnFabricIDAndDeviceID represents a mock GetSwitchConfigOnFabricIDAndDeviceID
func (db *DatabaseRepository) GetSwitchConfigOnFabricIDAndDeviceID(FabricID uint, DeviceID uint) (domain.SwitchConfig, error) {
	if db.MockGetSwitchConfigOnFabricIDAndDeviceID != nil {
		return db.MockGetSwitchConfigOnFabricIDAndDeviceID(FabricID, DeviceID)
	}
	return domain.SwitchConfig{}, nil
}

//GetSwitchConfigs represents a mock GetSwitchConfigs
func (db *DatabaseRepository) GetSwitchConfigs(FabricName string) ([]domain.SwitchConfig, error) {
	if db.MockGetSwitchConfigs != nil {
		return db.MockGetSwitchConfigs(FabricName)
	}
	return []domain.SwitchConfig{}, nil
}

//GetSwitchConfigOnDeviceIP represents a mock GetSwitchConfigOnDeviceIP
func (db *DatabaseRepository) GetSwitchConfigOnDeviceIP(FabricName string, DeviceIP string) (domain.SwitchConfig, error) {
	if db.MockGetSwitchConfigOnDeviceIP != nil {
		return db.MockGetSwitchConfigOnDeviceIP(FabricName, DeviceIP)
	}
	return domain.SwitchConfig{}, nil
}

//CreateInterfaceSwitchConfig represents a mock CreateInterfaceSwitchConfig
func (db *DatabaseRepository) CreateInterfaceSwitchConfig(SwitchConfig *domain.InterfaceSwitchConfig) error {
	if db.MockCreateInterfaceSwitchConfig != nil {
		return db.MockCreateInterfaceSwitchConfig(SwitchConfig)
	}
	return nil
}

//UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType represents a mock UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType
func (db *DatabaseRepository) UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(FabricID uint, InterfaceID uint, configType string) error {
	if db.MockUpdateInterfaceSwitchConfigsOnInterfaceIDConfigType != nil {
		return db.MockUpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(FabricID, InterfaceID, configType)
	}
	return nil
}

//UpdateInterfaceSwitchConfigsConfigType represents a mock UpdateInterfaceSwitchConfigsConfigType
func (db *DatabaseRepository) UpdateInterfaceSwitchConfigsConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateInterfaceSwitchConfigsConfigType != nil {
		return db.MockUpdateInterfaceSwitchConfigsConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil
}

//GetInterfaceSwitchConfigOnFabricIDAndInterfaceID represents a mock GetInterfaceSwitchConfigOnFabricIDAndInterfaceID
func (db *DatabaseRepository) GetInterfaceSwitchConfigOnFabricIDAndInterfaceID(FabricID uint, InterfaceID uint) (domain.InterfaceSwitchConfig, error) {
	if db.MockGetInterfaceSwitchConfigOnFabricIDAndInterfaceID != nil {
		return db.MockGetInterfaceSwitchConfigOnFabricIDAndInterfaceID(FabricID, InterfaceID)
	}

	return domain.InterfaceSwitchConfig{}, nil
}

//GetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID represents a mock GetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID
func (db *DatabaseRepository) GetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID(FabricID uint, InterfaceID uint) int64 {
	if db.MockGetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID != nil {
		return db.MockGetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID(FabricID, InterfaceID)
	}
	return 0
}

//GetInterfaceSwitchConfigsOnDeviceID represents a mock GetInterfaceSwitchConfigsOnDeviceID
func (db *DatabaseRepository) GetInterfaceSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.InterfaceSwitchConfig, error) {
	if db.MockGetInterfaceSwitchConfigsOnDeviceID != nil {
		return db.MockGetInterfaceSwitchConfigsOnDeviceID(FabricID, DeviceID)
	}
	return []domain.InterfaceSwitchConfig{}, nil
}

//GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion represents a mock GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion
func (db *DatabaseRepository) GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion(FabricID uint, InterfaceIDs []uint) ([]domain.InterfaceSwitchConfig, error) {
	if db.MockGetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion != nil {
		return db.MockGetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion(FabricID, InterfaceIDs)
	}
	return []domain.InterfaceSwitchConfig{}, nil
}

//UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs represents a mock UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs
func (db *DatabaseRepository) UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs(FabricID uint, InterfaceIDs []uint, configType string) error {
	if db.MockUpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs != nil {
		return db.MockUpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs(FabricID, InterfaceIDs, configType)
	}
	return nil
}

//GetBGPSwitchConfigsOnDeviceID represents a mock GetBGPSwitchConfigsOnDeviceID
func (db *DatabaseRepository) GetBGPSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	if db.MockGetBGPSwitchConfigsOnDeviceID != nil {
		return db.MockGetBGPSwitchConfigsOnDeviceID(FabricID, DeviceID)
	}
	return []domain.RemoteNeighborSwitchConfig{}, nil
}

//GetBGPSwitchConfigsOnRemoteDeviceID represents a mock GetBGPSwitchConfigsOnDeviceID
func (db *DatabaseRepository) GetBGPSwitchConfigsOnRemoteDeviceID(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	if db.MockGetBGPSwitchConfigsOnRemoteDeviceID != nil {
		return db.MockGetBGPSwitchConfigsOnRemoteDeviceID(FabricID, DeviceID, RemoteDeviceIDs)
	}
	return []domain.RemoteNeighborSwitchConfig{}, nil
}

//GetMCTBGPSwitchConfigsOnDeviceID represents a mock GetMCTBGPSwitchConfigsOnDeviceID
func (db *DatabaseRepository) GetMCTBGPSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	if db.MockGetMCTBGPSwitchConfigsOnDeviceID != nil {
		return db.MockGetMCTBGPSwitchConfigsOnDeviceID(FabricID, DeviceID)
	}
	return []domain.RemoteNeighborSwitchConfig{}, nil
}

//GetBGPSwitchConfigs represents a mock GetBGPSwitchConfigs
func (db *DatabaseRepository) GetBGPSwitchConfigs(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	if db.MockGetBGPSwitchConfigs != nil {
		return db.MockGetBGPSwitchConfigs(FabricID, InterfaceIDs)
	}
	return []domain.RemoteNeighborSwitchConfig{}, nil
}

//GetBGPSwitchConfigsExcludingMarkedForDeletion represents a mock GetBGPSwitchConfigsExcludingMarkedForDeletion
func (db *DatabaseRepository) GetBGPSwitchConfigsExcludingMarkedForDeletion(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	if db.MockGetBGPSwitchConfigsExcludingMarkedForDeletion != nil {
		return db.MockGetBGPSwitchConfigsExcludingMarkedForDeletion(FabricID, InterfaceIDs)
	}
	return []domain.RemoteNeighborSwitchConfig{}, nil
}

//UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID represents a mock UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID
func (db *DatabaseRepository) UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID(FabricID uint, InterfaceIDs []uint, configType string) error {
	if db.MockUpdateConfigTypeForBGPSwitchConfigsOnIntefaceID != nil {
		return db.MockUpdateConfigTypeForBGPSwitchConfigsOnIntefaceID(FabricID, InterfaceIDs, configType)
	}
	return nil
}

//GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID represents a mock GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID
func (db *DatabaseRepository) GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID(FabricID uint, RemoteInterfaceID uint) (domain.RemoteNeighborSwitchConfig, error) {
	if db.MockGetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID != nil {
		return db.MockGetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID(FabricID, RemoteInterfaceID)
	}
	return domain.RemoteNeighborSwitchConfig{}, nil

}

//GetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID represents a mock GetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID
func (db *DatabaseRepository) GetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID(FabricID uint, RemoteInterfaceID uint) int64 {
	if db.MockGetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID != nil {
		return db.MockGetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID(FabricID, RemoteInterfaceID)
	}
	return 0
}

//CreateBGPSwitchConfig represents a mock CreateBGPSwitchConfig
func (db *DatabaseRepository) CreateBGPSwitchConfig(BGPSwitchConfig *domain.RemoteNeighborSwitchConfig) error {
	if db.MockCreateBGPSwitchConfig != nil {
		return db.MockCreateBGPSwitchConfig(BGPSwitchConfig)
	}
	return nil
}

//UpdateBGPSwitchConfigsOnInterfaceIDConfigType represents a mock UpdateBGPSwitchConfigsOnInterfaceIDConfigType
func (db *DatabaseRepository) UpdateBGPSwitchConfigsOnInterfaceIDConfigType(FabricID uint, InterfaceID uint, configType string) error {
	if db.MockUpdateBGPSwitchConfigsOnInterfaceIDConfigType != nil {
		return db.MockUpdateBGPSwitchConfigsOnInterfaceIDConfigType(FabricID, InterfaceID, configType)
	}
	return nil
}

//UpdateBGPSwitchConfigsConfigType represents a mock UpdateBGPSwitchConfigsConfigType
func (db *DatabaseRepository) UpdateBGPSwitchConfigsConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateBGPSwitchConfigsConfigType != nil {
		return db.MockUpdateBGPSwitchConfigsConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil
}

//CreateExecutionLog represents a mock CreateExecutionLog
func (db *DatabaseRepository) CreateExecutionLog(ExecutionLog *domain.ExecutionLog) error {
	if db.MockCreateExecutionLog != nil {
		return db.MockCreateExecutionLog(ExecutionLog)
	}
	return nil
}

//GetDeviceUsingDeviceID represents a mock GetDeviceUsingDeviceID
func (db *DatabaseRepository) GetDeviceUsingDeviceID(FabricID uint, DeviceID uint) (domain.Device, error) {
	if db.MockGetDeviceUsingDeviceID != nil {
		return db.MockGetDeviceUsingDeviceID(FabricID, DeviceID)
	}
	return domain.Device{}, nil
}

//GetExecutionLogByUUID represents a mock GetExecutionLogByUUID
func (db *DatabaseRepository) GetExecutionLogByUUID(uuid string) (domain.ExecutionLog, error) {
	if db.MockGetExecutionLogByUUID != nil {
		return db.MockGetExecutionLogByUUID(uuid)
	}
	return domain.ExecutionLog{}, nil
}

//GetExecutionLogList represents a mock GetExecutionLogList
func (db *DatabaseRepository) GetExecutionLogList(limit int, status string) ([]domain.ExecutionLog, error) {
	if db.MockGetExecutionLogList != nil {
		return db.MockGetExecutionLogList(limit, status)
	}
	return []domain.ExecutionLog{}, nil
}

//UpdateExecutionLog represents a mock UpdateExecutionLog
func (db *DatabaseRepository) UpdateExecutionLog(ExecutionLog *domain.ExecutionLog) error {

	if db.MockUpdateExecutionLog != nil {
		return db.MockUpdateExecutionLog(ExecutionLog)
	}
	return nil
}

//MarkMctClusterForDelete represents a mock MarkMctClusterForDelete
func (db *DatabaseRepository) MarkMctClusterForDelete(FabricID uint, DeviceID uint) error {
	if db.MockMarkMctClusterForDelete != nil {
		return db.MockMarkMctClusterForDelete(FabricID, DeviceID)
	}
	return nil

}

//MarkMctClusterMemberPortsForDelete represents a mock MarkMctClusterMemberPortsForDelete
func (db *DatabaseRepository) MarkMctClusterMemberPortsForDelete(FabricID uint, DeviceID uint) error {
	if db.MockMarkMctClusterMemberPortsForDelete != nil {
		return db.MockMarkMctClusterMemberPortsForDelete(FabricID, DeviceID)
	}
	return nil

}

//MarkMctClusterForDeleteWithBothDevices represents a mock MarkMctClusterForDeleteWithBothDevices
func (db *DatabaseRepository) MarkMctClusterForDeleteWithBothDevices(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error {
	if db.MockMarkMctClusterForDeleteWithBothDevices != nil {
		return db.MockMarkMctClusterForDeleteWithBothDevices(FabricID, DeviceID, MctNeighborDeviceID)
	}
	return nil

}

//MarkMctClusterMemberPortsForDeleteWithBothDevices represents a mock MarkMctClusterMemberPortsForDeleteWithBothDevices
func (db *DatabaseRepository) MarkMctClusterMemberPortsForDeleteWithBothDevices(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error {
	if db.MockMarkMctClusterMemberPortsForDeleteWithBothDevices != nil {
		return db.MockMarkMctClusterMemberPortsForDeleteWithBothDevices(FabricID, DeviceID, MctNeighborDeviceID)
	}
	return nil

}

//GetMctClusterConfigWithDeviceIP represents a mock GetMctClusterConfigWithDeviceIP
func (db *DatabaseRepository) GetMctClusterConfigWithDeviceIP(IPAddress string) ([]domain.MctClusterConfig, error) {
	if db.MockGetMctClusterConfigWithDeviceIP != nil {
		return db.MockGetMctClusterConfigWithDeviceIP(IPAddress)
	}
	return []domain.MctClusterConfig{}, nil

}

//DeleteMctClustersUsingClusterObject represents a mock DeleteMctClustersUsingClusterObject
func (db *DatabaseRepository) DeleteMctClustersUsingClusterObject(oldMct domain.MctClusterConfig) error {
	if db.MockDeleteMctClustersUsingClusterObject != nil {
		return db.MockDeleteMctClustersUsingClusterObject(oldMct)
	}
	return nil

}

//UpdateLLDPConfigType represents a mock UpdateLLDPConfigType
func (db *DatabaseRepository) UpdateLLDPConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	if db.MockUpdateLLDPConfigType != nil {
		return db.MockUpdateLLDPConfigType(FabricID, QueryconfigTypes, configType)
	}
	return nil

}

//DeleteLLDPMarkedForDelete represents a mock DeleteLLDPMarkedForDelete
func (db *DatabaseRepository) DeleteLLDPMarkedForDelete(FabricID uint) error {
	if db.MockDeleteLLDPMarkedForDelete != nil {
		return db.MockDeleteLLDPMarkedForDelete(FabricID)
	}
	return nil

}

//DeleteMctClustersWithMgmtIP represents a mock DeleteMctClustersWithMgmtIP
func (db *DatabaseRepository) DeleteMctClustersWithMgmtIP(IPAddress string) error {
	if db.MockDeleteMctClustersWithMgmtIP != nil {
		return db.MockDeleteMctClustersWithMgmtIP(IPAddress)
	}
	return nil

}

//MarkMctClusterMemberPortsForCreate represents a mock MarkMctClusterMemberPortsForCreate
func (db *DatabaseRepository) MarkMctClusterMemberPortsForCreate(FabricID uint, DeviceID uint, RemoteDeviceID uint) error {
	if db.MockMarkMctClusterMemberPortsForCreate != nil {
		return db.MockMarkMctClusterMemberPortsForCreate(FabricID, DeviceID, RemoteDeviceID)
	}
	return nil
}

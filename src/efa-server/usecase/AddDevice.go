package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/infra/constants"
	"efa-server/usecase/comparator/interfaceconfig"
	"efa-server/usecase/comparator/lldp"
	"efa-server/usecase/comparator/lldpneighbor"
	"efa-server/usecase/comparator/mctconfig"
	"efa-server/usecase/comparator/rackevpnneighbor"
	Interactor "efa-server/usecase/interactorinterface"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

//LOG is the main Log variable used through out the UseCase package
//var LOG *nlog.Entry

//DEC is the main Decorator variable used through out the Usecase package
var DEC = appcontext.DecorateRuntimeContext

//Constants used to specify the Role of the Device
//TODO - Move to Domain Package
const (
	SpineRole = "Spine"
	LeafRole  = "Leaf"
	RackRole  = "Rack"
)

//MctContext used for MCT Operations
type MctContext struct {
	operation        uint
	DeviceOneAdapter Interactor.DeviceAdapter
	DeviceTwoAdapter Interactor.DeviceAdapter
	Device           *domain.Device
	NeighborDevice   *domain.Device
	LLDP             *domain.LLDP
	PhyInterface     *domain.Interface
}

//DeviceInteractor is the reciever object for the Device
type DeviceInteractor struct {
	Db                   Interactor.DatabaseRepository
	FabricAdapter        Interactor.FabricAdapter
	DeviceAdapterFactory func(ctx context.Context, IPAddress string, UserName string, Password string) (Interactor.DeviceAdapter, error)
	FabricID             uint
	FabricName           string
	FabricProperties     domain.FabricProperties
	AppContext           context.Context
	DBMutex              sync.Mutex
	Refresh              bool
}

//AddDeviceFirstStage does the following
// 1 Fetches Interfaces from the Device
// 2 Enables ethernet Interfaces that are admin down
// 3 Persists the Interfaces
func (sh *DeviceInteractor) AddDeviceFirstStage(ctx context.Context, FabricName string, IPAddress string,
	UserName string, Password string, Role string) error {
	var err error
	ctx = context.WithValue(ctx, appcontext.DeviceName, IPAddress)
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Adding Device in first stage")

	//check for existing Device
	var Device domain.Device
	if Device, err = sh.Db.GetDevice(FabricName, IPAddress); err == nil {
		statusMsg := fmt.Sprintf("Switch %s already exist", IPAddress)
		LOG.Infoln(statusMsg)
		sh.Refresh = true
	}

	Device.IPAddress = IPAddress
	Device.DeviceRole = Role
	sh.EvaluateCredentials(UserName, Password, &Device)

	Device.FabricID = sh.FabricID

	var DeviceAdapter Interactor.DeviceAdapter
	//Open Connection to the Switch
	if DeviceAdapter, err = sh.DeviceAdapterFactory(ctx, IPAddress, Device.UserName, Device.Password); err != nil {
		statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", IPAddress, err.Error())
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	defer DeviceAdapter.CloseConnection(ctx)

	LOG.Infoln("Fetch Interface assets from Device")
	//Fetch Switch Interfaces
	if statusMsg, err := sh.fetchInterfacesAndASN(ctx, &Device, DeviceAdapter); err != nil {
		LOG.Infof("Fetch Interface assets from Device %s failed for %s:%s.\n", IPAddress, FabricName, err.Error())
		return errors.New(statusMsg)
	}

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()

	//Save Device
	if err = sh.Db.CreateDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s create Failed", IPAddress)
		return errors.New(statusMsg)
	}
	//Persist Only the Interfaces
	if statusMsg, err := sh.persistInterfaces(ctx, &Device, DeviceAdapter); err != nil {
		return errors.New(statusMsg)
	}

	return nil
}

//AddDeviceSecondStage does the following
// 1 Fetches LLDP Data from the Device
func (sh *DeviceInteractor) AddDeviceSecondStage(ctx context.Context, FabricName string, IPAddress string,
	UserName string, Password string, Role string) error {
	var err error
	ctx = context.WithValue(ctx, appcontext.DeviceName, IPAddress)
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Adding Device in second stage")
	//check for existing Device
	var Device domain.Device
	if Device, err = sh.Db.GetDevice(FabricName, IPAddress); err != nil {
		statusMsg := fmt.Sprintf("Stage 2 Fabric %s Switch %s does not exist - ERROR %v", FabricName, IPAddress, err)
		LOG.Infoln(statusMsg)
		return errors.New(statusMsg)
	}

	Device.IPAddress = IPAddress
	sh.EvaluateCredentials(UserName, Password, &Device)

	Device.FabricID = sh.FabricID

	var DeviceAdapter Interactor.DeviceAdapter
	//Open Connection to the Switch
	if DeviceAdapter, err = sh.DeviceAdapterFactory(ctx, IPAddress, Device.UserName, Device.Password); err != nil {
		statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", IPAddress, err.Error())
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	defer DeviceAdapter.CloseConnection(ctx)

	LOG.Infoln("Fetch LLDP assets")
	//Fetch LLDP Assets
	if statusMsg, err := sh.fetchLLDPAssets(ctx, &Device, DeviceAdapter); err != nil {
		return errors.New(statusMsg)
	}

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()
	//Persist Switch Assets
	if statusMsg, err := sh.persistLLDP(ctx, &Device, DeviceAdapter); err != nil {
		return errors.New(statusMsg)
	}

	return nil
}

//AddDeviceThirdStage does the following
// 2 Builds LLDP Neighbor relationship
// 3 Generate Device Level Configs to be pushed
func (sh *DeviceInteractor) AddDeviceThirdStage(ctx context.Context, FabricName string, IPAddress string,
	UserName string, Password string, Role string) error {
	var err error
	ctx = context.WithValue(ctx, appcontext.DeviceName, IPAddress)
	LOG := appcontext.Logger(ctx)

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()

	LOG.Infof("Adding Device in third stage")
	//check for existing Device
	var Device domain.Device
	if Device, err = sh.Db.GetDevice(FabricName, IPAddress); err != nil {
		statusMsg := fmt.Sprintf("====> Stage 3 Fabric %s switch %s does not exist ERROR %v", FabricName, IPAddress, err)
		LOG.Infoln(statusMsg)
		return errors.New(statusMsg)
	}

	Device.IPAddress = IPAddress
	sh.EvaluateCredentials(UserName, Password, &Device)
	Device.Interfaces, _ = sh.Db.GetInterfacesonDevice(sh.FabricID, Device.ID)
	Device.LLDPS, _ = sh.Db.GetLLDPsonDevice(sh.FabricID, Device.ID)
	Device.FabricID = sh.FabricID

	//Compute Role & Neighbor Relationships
	if statusMsg, err := sh.computeRoleAndBuildNeighborRelationShip(ctx, &Device, Role); err != nil {
		return errors.New(statusMsg)

	}

	if _, err := sh.buildSwitchConfig(ctx, &Device); err != nil {
		return err
	}

	//Save Device Again
	if err := sh.Db.SaveDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s Save Failed", IPAddress)
		return errors.New(statusMsg)

	}

	return nil
}

//AddDeviceFourthStage does the following
// 1 Generate IneterfaceSwitchConfigs for the Device
// 2 Generate RemoteBGP Neighbor configs for the Device
func (sh *DeviceInteractor) AddDeviceFourthStage(ctx context.Context, FabricName string, IPAddress string,
	UserName string, Password string, Role string) error {
	var err error
	ctx = context.WithValue(ctx, appcontext.DeviceName, IPAddress)
	LOG := appcontext.Logger(ctx)

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()

	LOG.Infof("Adding Device in fourth stage")

	//check for existing Device
	var Device domain.Device
	if Device, err = sh.Db.GetDevice(FabricName, IPAddress); err != nil {
		statusMsg := fmt.Sprintf("Generate Config called and Switch %s does not exist", IPAddress)
		LOG.Infoln(statusMsg)
		return errors.New(statusMsg)
	}

	Device.IPAddress = IPAddress
	sh.EvaluateCredentials(UserName, Password, &Device)

	Device.FabricID = sh.FabricID
	//Fetch the Assets from the database
	Device.Interfaces, _ = sh.Db.GetInterfacesonDevice(sh.FabricID, Device.ID)
	Device.LLDPS, _ = sh.Db.GetLLDPsonDevice(sh.FabricID, Device.ID)

	//Specify a bool to identify the Interface configs are being build for Rack(NON-CLOS)
	if _, err := sh.buildInterfaceConfigs(ctx, &Device, false); err != nil {
		return err
	}

	//Save Device Again
	if err := sh.Db.SaveDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s Save Failed", IPAddress)
		return errors.New(statusMsg)

	}

	return nil
}

// EvaluateCredentials evaluates the credentials needed to be used to connect to switch.
// 1. If User has provided username/password in the CLI, then that is taken.
// 2. If User has not provided, then if application DB contains credentials, then that is used to connect to switch.(after decrpyting it)
// 3. If User has not provided and DB doesn't contain the credentials, then default switch credentials are used.
func (sh *DeviceInteractor) EvaluateCredentials(UserName string, Password string, Device *domain.Device) error {
	if len(Device.UserName) == 0 {
		// Device not exists in DB and being called for first time(addition of device)
		if len(UserName) == 0 {
			// username/password not provided by user, assume "admin"/"password" as credentials.
			Device.UserName = "admin"
			Device.Password = "password"
		} else {
			// Update Device object with user provided credentials.
			Device.UserName = UserName
			Device.Password = Password
		}
	} else {
		// Else case is when device being added again(refresh/update.
		// This can happen even during delete of device too.
		// If user has provided credentials during delete/update, take the new one, else use the existing one from DB
		if len(UserName) > 0 {
			Device.UserName = UserName
			Device.Password = Password
		}
	}

	return nil
}

func (sh *DeviceInteractor) fetchInterfacesAndASN(ctx context.Context, Device *domain.Device, DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	//Fetch Interfaces
	var err error
	LOG := appcontext.Logger(ctx)

	deviceDetail, err := DeviceAdapter.GetDeviceDetail(sh.FabricID, Device.ID, Device.IPAddress)
	if err != nil {
		statusMsg := fmt.Sprintf("Unable to fetch device details(Model and Firmware Version) for %s", Device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	Device.Model = deviceDetail.Model
	Device.FirmwareVersion = deviceDetail.FirmwareVersion

	if err = DeviceAdapter.CheckSupportedFirmware(Device.IPAddress); err != nil {
		statusMsg := fmt.Sprintf("Unsupported firmware version for %s : %s", Device.IPAddress, Device.FirmwareVersion)
		LOG.Infoln(statusMsg, err)
		return statusMsg, errors.New(statusMsg)
	}

	if Device.Name, err = DeviceAdapter.GetSwitchHostName(Device.IPAddress); err != nil {
		statusMsg := fmt.Sprintf("HostName fetch for switch %s Failed", Device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}

	if Device.Interfaces, err = DeviceAdapter.GetInterfaces(sh.FabricID, Device.ID, sh.FabricProperties.ControlVE); err != nil {
		statusMsg := fmt.Sprintf("Interface fetch for switch %s Failed", Device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	//Enable Interfaces on the Device
	if _, err := sh.enableInterfaces(ctx, Device, DeviceAdapter); err != nil {
		statusMsg := fmt.Sprintf("Enable Interfaces on switch %s Failed", Device.IPAddress)
		LOG.Errorln(statusMsg, err)
		return statusMsg, err
	}
	//Fetch ASN
	LOG.Infof("Fetch ASN for Device %s", Device.IPAddress)
	if Device.LocalAs, err = DeviceAdapter.GetASN(sh.FabricID, Device.ID); err != nil {
		statusMsg := fmt.Sprintf("LOCAL  ASN fetch for switch %s Failed", Device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	statusMsg := fmt.Sprintf("Successfully fetched Interfaces")
	return statusMsg, nil
}

func (sh *DeviceInteractor) fetchLLDPAssets(ctx context.Context, Device *domain.Device, DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	var err error
	//Fetch LLDP
	if Device.LLDPS, err = DeviceAdapter.GetLLDPs(sh.FabricID, Device.ID); err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch LLDP for %s", Device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	statusMsg := fmt.Sprintf("Successfully fetched LLDPs")
	return statusMsg, nil
}

//Compute the Device Role for the Switch and Builds its Neighbor Relationship
func (sh *DeviceInteractor) computeRoleAndBuildNeighborRelationShip(ctx context.Context, device *domain.Device, ExpectedRole string) (string, error) {
	LOG := appcontext.Logger(ctx)

	LOG.Infoln("Compute Role and Neighbor relationships")

	deviceMap := make(map[uint]domain.Device, 0)
	//key is deviceID and Neighbor DeviceOneID
	ClusterConfigMap := make(map[string]domain.MctClusterConfig, 0)
	MctMemberPortsMap := make(map[string][]domain.MCTMemberPorts, 0)
	device.DeviceRole = ExpectedRole
	var err error
	if deviceMap, err = sh.prepareMapDeviceIDToDevice(ctx); err != nil {
		statusMsg := fmt.Sprintf("Failed to build Device Role Map")
		LOG.Infoln(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}

	//For Every Interface MAC Find a Match in LLDP Table
	LLDPNeighbors := make([]domain.LLDPNeighbor, 0)
	for index := range device.Interfaces {
		phy := &device.Interfaces[index]
		lldp, err := sh.Db.GetLLDPOnRemoteMacExcludingMarkedForDeletion(phy.Mac, sh.FabricID)
		if err == nil {

			statusMsg := fmt.Sprintf("LLDP Neighbor found local Mac%s remote Mac %s remote device %s",
				phy.Mac, lldp.LocalIntMac, deviceMap[lldp.DeviceID].IPAddress)
			LOG.Infoln(statusMsg)

			//DISCOVER MCT Cluster
			if device.DeviceRole == LeafRole && deviceMap[lldp.DeviceID].DeviceRole == LeafRole && device.ID != lldp.DeviceID {
				RemoteDevice := deviceMap[lldp.DeviceID]
				var mctxt MctContext
				mctxt.Device = device
				mctxt.NeighborDevice = &RemoteDevice
				mctxt.LLDP = &lldp
				mctxt.PhyInterface = phy
				statusMsg, err = sh.discoverMctCluster(ctx, ClusterConfigMap, MctMemberPortsMap, &mctxt)
				if err != nil {
					LOG.Info(statusMsg)
					return statusMsg, err
				}
			}
			neighborOne, neighborTwo, nerr := sh.prepareNeighborDetails(ctx, phy, &lldp,
				device.DeviceRole, deviceMap[lldp.DeviceID].DeviceRole)
			if nerr != nil {
				return "Unable to build Neighbor", nerr
			}
			LLDPNeighbors = append(LLDPNeighbors, neighborOne)
			LLDPNeighbors = append(LLDPNeighbors, neighborTwo)
		}
	}

	if statusMsg, err := sh.persistLLDPNeighbors(ctx, device, LLDPNeighbors); err != nil {
		return statusMsg, err
	}
	if statusMsg, err := sh.markMctClusterForDeleteUsingDeviceID(ctx, device.ID, ClusterConfigMap); err != nil {
		LOG.Errorln(statusMsg, err)
		return statusMsg, err
	}

	for k, val := range ClusterConfigMap {
		RemoteDevice, err := sh.Db.GetDeviceUsingDeviceID(sh.FabricID, val.MCTNeighborDeviceID)
		if err != nil {
			statusMsg := fmt.Sprintf("Unable Fetch Device using ID %d", val.MCTNeighborDeviceID)
			LOG.Errorln(statusMsg)
			return statusMsg, err
		}
		var mctxt MctContext
		mctxt.Device = device
		mctxt.NeighborDevice = &RemoteDevice
		mctxt.DeviceOneAdapter, err = sh.DeviceAdapterFactory(ctx, device.IPAddress,
			device.UserName, device.Password)
		defer mctxt.DeviceOneAdapter.CloseConnection(ctx)
		if err != nil {
			statusMsg := fmt.Sprintf("Switch %s connection Failed", device.IPAddress)
			LOG.Errorln(statusMsg)
			return statusMsg, errors.New(statusMsg)
		}
		mctxt.DeviceTwoAdapter, err = sh.DeviceAdapterFactory(ctx, RemoteDevice.IPAddress,
			RemoteDevice.UserName, RemoteDevice.Password)
		defer mctxt.DeviceTwoAdapter.CloseConnection(ctx)
		if err != nil {
			statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", RemoteDevice.IPAddress, err.Error())
			LOG.Errorln(statusMsg)
			return statusMsg, errors.New(statusMsg)
		}

		if statusMsg, err := sh.persistMCTConfigs(ctx, &val, MctMemberPortsMap[k], ClusterConfigMap, &mctxt); err != nil {
			return statusMsg, err
		}
	}
	statusMsg := fmt.Sprintf("Successfully computed Role %s", device.DeviceRole)
	LOG.Infoln(statusMsg)

	return statusMsg, nil
}

func (sh *DeviceInteractor) markMctClusterForDeleteUsingDeviceID(ctx context.Context, DeviceID uint, ClusterConfigMap map[string]domain.MctClusterConfig) (string, error) {
	LOG := appcontext.Logger(ctx)
	msg := fmt.Sprintf("Delete MCT cluster detected device - %d", DeviceID)
	LOG.Infoln(msg)
	GetKey := func(DeviceOneID uint, DeviceTwoID uint) string {
		return fmt.Sprintf("%d : %d", DeviceOneID, DeviceTwoID)
	}
	OldMcts, merr := sh.Db.GetMctClusters(sh.FabricID, DeviceID, []string{})
	if merr != nil {
		statusMsg := fmt.Sprintf("Unable to Retrieve MctCluster from DB for device %d", DeviceID)
		LOG.Errorln(statusMsg, merr)
		return statusMsg, merr
	}
	for _, OldMct := range OldMcts {
		oldMctKey := GetKey(OldMct.DeviceID, OldMct.MCTNeighborDeviceID)
		if _, ok := ClusterConfigMap[oldMctKey]; ok == false {
			LOG.Infof("marking cluster for delete[%s:%d] Remote [%s:%d] cluster ID - %d ", OldMct.DeviceOneMgmtIP, OldMct.DeviceID, OldMct.DeviceTwoMgmtIP, OldMct.MCTNeighborDeviceID, OldMct.ID)
			if err := sh.Db.MarkMctClusterForDeleteWithBothDevices(OldMct.FabricID, OldMct.DeviceID, OldMct.MCTNeighborDeviceID); err != nil {
				LOG.Errorln("Error deletting stale MCT CLuster ", err)
				return "Error deleting stale MCT Cluster", err
			}
			if err := sh.Db.MarkMctClusterMemberPortsForDeleteWithBothDevices(OldMct.FabricID, OldMct.DeviceID, OldMct.MCTNeighborDeviceID); err != nil {
				LOG.Errorln("Error deletting stale MCT CLuster Ports ", err)
				return "Error deleting stale MCT Cluster ports", err
			}
			if OldMct.PeerTwoIP != "" && OldMct.PeerOneIP != "" {
				err := sh.ReleaseIPPair(sh.AppContext, OldMct.FabricID, OldMct.DeviceID, OldMct.MCTNeighborDeviceID,
					domain.MCTPoolName, OldMct.PeerTwoIP, OldMct.PeerOneIP, OldMct.VEInterfaceOneID, OldMct.VEInterfaceTwoID)
				if err != nil {
					LOG.Errorf("Error : Releasing IP FOR %s %s", OldMct.PeerOneIP, OldMct.PeerTwoIP)
				}
			} else {
				err := sh.Db.DeleteMctClustersUsingClusterObject(OldMct)
				if err != nil {
					return "Error deletting stale MCT CLuster", err
				}
			}
		}
	}
	return "SUCCESS", nil
}
func (sh *DeviceInteractor) ip4toInt(IPv4Address net.IP) int64 {
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(IPv4Address.To4())
	return IPv4Int.Int64()
}
func (sh *DeviceInteractor) evaluateNodeID(dev1, dev2 *domain.Device) (string, string) {
	ip1 := net.ParseIP(dev1.IPAddress)
	ip2 := net.ParseIP(dev2.IPAddress)
	deviceOneIP := sh.ip4toInt(ip1)
	deviceTwoIP := sh.ip4toInt(ip2)

	if deviceOneIP < deviceTwoIP {
		return "1", "2"
	}
	return "2", "1"
}

func (sh *DeviceInteractor) discoverMctCluster(ctx context.Context, ClusterConfigMap map[string]domain.MctClusterConfig,
	MctMemberPortsMap map[string][]domain.MCTMemberPorts, mctxt *MctContext) (string, error) {
	LOG := appcontext.Logger(ctx)

	var err error
	var IntOneSpeed int
	var IntTwoSpeed int
	device := mctxt.Device
	NeighborDevice := mctxt.NeighborDevice
	phy := mctxt.PhyInterface
	lldp := mctxt.LLDP

	if sh.FabricAdapter.IsMCTLeavesCompatible(ctx, device.Model, NeighborDevice.Model) == false {
		statusMsg := fmt.Sprintf("MCT Leaf devices(%s,%s) are not compatible", device.IPAddress, NeighborDevice.IPAddress)
		LOG.Errorln(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}
	GetKey := func(DeviceOneID uint, DeviceTwoID uint) string {
		return fmt.Sprintf("%d : %d", DeviceOneID, DeviceTwoID)
	}

	mctxt.DeviceOneAdapter, err = sh.DeviceAdapterFactory(ctx, device.IPAddress,
		device.UserName, device.Password)
	if err != nil {
		statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", device.IPAddress, err.Error())
		LOG.Errorln(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}
	defer mctxt.DeviceOneAdapter.CloseConnection(ctx)

	mctxt.DeviceTwoAdapter, err = sh.DeviceAdapterFactory(ctx, NeighborDevice.IPAddress,
		NeighborDevice.UserName, NeighborDevice.Password)
	if err != nil {
		statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", NeighborDevice.IPAddress, err.Error())
		LOG.Errorln(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}
	defer mctxt.DeviceTwoAdapter.CloseConnection(ctx)

	key := GetKey(device.ID, lldp.DeviceID)

	GetSpeed := func(speed string) (uint64, error) {
		InterfaceSpeed, err := strconv.ParseUint(speed, 10, 64)
		return InterfaceSpeed, err
	}
	SetSpeed := func(cluster domain.MctClusterConfig, InterfaceSpeed string) (string, error) {
		ClusterSpeed, cerr := GetSpeed(cluster.PeerInterfaceSpeed)
		PeerInterfaceSpeed := strconv.FormatUint(ClusterSpeed, 10)

		if cerr != nil {
			LOG.Info("Error: While Converting Cluster Peer Interface  Speed")
			return PeerInterfaceSpeed, cerr
		}
		phySpeed, cerr := GetSpeed(InterfaceSpeed)
		if cerr != nil {
			LOG.Info("Error: While Converting Physical Interface Speed")
			return PeerInterfaceSpeed, cerr
		}
		if ClusterSpeed > phySpeed {
			//Set Minimum Of all Interfaces
			PeerInterfaceSpeed = strconv.FormatUint(phySpeed, 10)
		}
		return PeerInterfaceSpeed, cerr
	}
	phyTwo, err := sh.Db.GetInterfaceOnMac(lldp.LocalIntMac, sh.FabricID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to Retrieve Other Physical Interface for Neighbor %s", lldp.LocalIntMac)
		LOG.Errorln(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}

	if _, ok := ClusterConfigMap[key]; ok == false {
		IntOneSpeed = 0
		IntOneSpeed, err = mctxt.DeviceOneAdapter.GetInterfaceSpeed(phy.IntType, phy.IntName)
		if IntOneSpeed == 0 || err != nil {
			statusMsg := fmt.Sprintln("[Operation Discover MCT cluster]Interface One Speed  ", phy.IntType, ":", phy.IntName, "-", IntOneSpeed)
			LOG.Error(statusMsg)
			if IntOneSpeed == 0 {
				return statusMsg, errors.New(statusMsg)
			}
			return statusMsg, err

		}
		IntSpeed := fmt.Sprintf("%d", IntOneSpeed)

		LOG.Infof("MCT Cluster Discovered Between Device %s and Device %s ", device.IPAddress, NeighborDevice.IPAddress)
		//Need to find better mechanism to assign node ID for now this is ok since A switch can have only one cluster
		localNodeID, remoteNodeID := sh.evaluateNodeID(device, NeighborDevice)
		LOG.Infof("Local NodeID (%s,%s), Remote NodeID (%s,%s) \n", device.IPAddress, localNodeID, NeighborDevice.IPAddress, remoteNodeID)
		ClusterConfigMap[key], _ = sh.prepareMctClusterDetails(ctx, device, NeighborDevice, IntSpeed, localNodeID, remoteNodeID)

	} else {
		//Cluster is Already Created Check if Interface Speed is changed
		IntOneSpeed = 0
		IntOneSpeed, err = mctxt.DeviceOneAdapter.GetInterfaceSpeed(phy.IntType, phy.IntName)
		if IntOneSpeed == 0 || err != nil {
			statusMsg := fmt.Sprintf("[Operation Discover MCT cluster] Error: Fetching Interface speed[One] %s %s - %d", phyTwo.IntType, phyTwo.IntName, IntOneSpeed)
			LOG.Error(statusMsg)
			if IntOneSpeed == 0 {
				return statusMsg, errors.New(statusMsg)
			}
			return statusMsg, err
		}

		IntSpeed := fmt.Sprintf("%d", IntOneSpeed)
		PeerInterfaceSpeed, err := SetSpeed(ClusterConfigMap[key], IntSpeed)
		if err != nil {
			return "Failed", err
		}
		cMap := ClusterConfigMap[key]
		cMap.PeerInterfaceSpeed = PeerInterfaceSpeed
		ClusterConfigMap[key] = cMap
	}

	MctMemberPortsMap[key], err = sh.prepareMctMemberPortDetails(device.ID, lldp.DeviceID, phy, &phyTwo, IntOneSpeed, IntTwoSpeed, MctMemberPortsMap[key])

	return "success", err
}

func (sh *DeviceInteractor) prepareMapDeviceIDToDevice(ctx context.Context) (map[uint]domain.Device, error) {
	LOG := appcontext.Logger(ctx)
	devices := make([]domain.Device, 0)
	deviceMap := make(map[uint]domain.Device, 0)
	devices, err := sh.Db.GetDevicesInFabric(sh.FabricID)
	if err != nil {
		LOG.Infoln(err)
		statusMsg := fmt.Sprintf("Failed to fetch switches for Fabric %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return deviceMap, err
	}
	for _, device := range devices {
		deviceMap[device.ID] = device
	}
	return deviceMap, nil
}

func (sh *DeviceInteractor) prepareMctClusterDetails(ctx context.Context, DeviceOne, DeviceTwo *domain.Device,
	InterfaceSpeed string, LocalNodeID string, RemoteNodeID string) (domain.MctClusterConfig, error) {
	LOG := appcontext.Logger(ctx)
	var ClusterConfig domain.MctClusterConfig
	FabricProperties, err := sh.Db.GetFabricProperties(sh.FabricID)
	if err != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch fabric Properties  %s", sh.FabricName)
		LOG.Errorln(statusMsg)
	}
	ClusterConfig.DeviceID = DeviceOne.ID
	ClusterConfig.ClusterID = 1
	ClusterConfig.DeviceOneMgmtIP = DeviceOne.IPAddress
	ClusterConfig.DeviceTwoMgmtIP = DeviceTwo.IPAddress
	ClusterConfig.FabricID = sh.FabricID
	ClusterConfig.MCTNeighborDeviceID = DeviceTwo.ID
	ClusterConfig.PeerInterfaceSpeed = InterfaceSpeed
	ClusterConfig.PeerInterfacetype = "Port-channel"
	if sh.FabricAdapter.IsRoutingDevice(ctx, DeviceTwo.Model) {
		ClusterConfig.PeerInterfaceName = FabricProperties.RoutingMctPortChannel
	} else {
		ClusterConfig.PeerInterfaceName = FabricProperties.MctPortChannel
	}
	ClusterConfig.ControlVlan = FabricProperties.ControlVlan
	ClusterConfig.ControlVE = FabricProperties.ControlVE
	ClusterConfig.LocalNodeID = LocalNodeID
	ClusterConfig.RemoteNodeID = RemoteNodeID
	return ClusterConfig, nil
}

func (sh *DeviceInteractor) prepareMctMemberPortDetails(DeviceID, RemoteDeviceID uint,
	phy, phyTwo *domain.Interface, IntOneSpeed int, IntTwoSpeed int,
	MctMemberPorts []domain.MCTMemberPorts) ([]domain.MCTMemberPorts, error) {

	var MemberPort domain.MCTMemberPorts
	MemberPort.DeviceID = DeviceID
	MemberPort.FabricID = sh.FabricID
	MemberPort.RemoteDeviceID = RemoteDeviceID
	MemberPort.InterfaceID = phy.ID
	MemberPort.InterfaceName = phy.IntName
	MemberPort.InterfaceType = phy.IntType
	MemberPort.RemoteInterfaceID = phyTwo.ID
	MemberPort.RemoteInterfaceName = phyTwo.IntName
	MemberPort.RemoteInterfaceType = phyTwo.IntType
	MemberPort.InterfaceSpeed = IntOneSpeed
	MemberPort.RemoteInterfaceSpeed = IntTwoSpeed

	MctMemberPorts = append(MctMemberPorts, MemberPort)
	return MctMemberPorts, nil
}

func (sh *DeviceInteractor) prepareNeighborDetails(ctx context.Context, phy *domain.Interface, lldp *domain.LLDP,
	DeviceOneRole string, DeviceTwoRole string) (domain.LLDPNeighbor, domain.LLDPNeighbor, error) {
	var lldpNeighbor domain.LLDPNeighbor
	var lldpOtherNeighbor domain.LLDPNeighbor
	LOG := appcontext.Logger(ctx)
	//Parse IP from CIDR notation
	intfOneIP, _, intOneIPerr := net.ParseCIDR(phy.IPAddress)

	lldpNeighbor.FabricID = sh.FabricID
	lldpOtherNeighbor.FabricID = sh.FabricID
	lldpNeighbor.DeviceOneID = phy.DeviceID
	lldpNeighbor.DeviceOneRole = DeviceOneRole
	lldpNeighbor.InterfaceOneID = phy.ID
	lldpNeighbor.InterfaceOneName = phy.IntName
	lldpNeighbor.InterfaceOneType = phy.IntType

	//members swapped for other side
	lldpOtherNeighbor.DeviceTwoID = phy.DeviceID
	lldpOtherNeighbor.DeviceTwoRole = DeviceOneRole
	lldpOtherNeighbor.InterfaceTwoID = phy.ID
	lldpOtherNeighbor.InterfaceTwoName = phy.IntName
	lldpOtherNeighbor.InterfaceTwoType = phy.IntType
	if intOneIPerr == nil {
		lldpNeighbor.InterfaceOneIP = intfOneIP.String()
		lldpOtherNeighbor.InterfaceTwoIP = intfOneIP.String()
	}

	phyTwo, err := sh.Db.GetInterfaceOnMac(lldp.LocalIntMac, sh.FabricID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to Retrieve Other Physical Interface for Neighbor %s", lldp.LocalIntMac)
		LOG.Errorln(statusMsg)
		return domain.LLDPNeighbor{}, domain.LLDPNeighbor{}, errors.New(statusMsg)
	}
	intfTwoIP, _, intTwoIPerr := net.ParseCIDR(phyTwo.IPAddress)
	lldpNeighbor.DeviceTwoID = phyTwo.DeviceID
	lldpNeighbor.DeviceTwoRole = DeviceTwoRole
	lldpNeighbor.InterfaceTwoID = phyTwo.ID
	lldpNeighbor.InterfaceTwoName = phyTwo.IntName
	lldpNeighbor.InterfaceTwoType = phyTwo.IntType
	//members swapped for other side
	lldpOtherNeighbor.DeviceOneID = phyTwo.DeviceID
	lldpOtherNeighbor.DeviceOneRole = DeviceTwoRole
	lldpOtherNeighbor.InterfaceOneID = phyTwo.ID
	lldpOtherNeighbor.InterfaceOneName = phyTwo.IntName
	lldpOtherNeighbor.InterfaceOneType = phyTwo.IntType
	if intTwoIPerr == nil {
		lldpNeighbor.InterfaceTwoIP = intfTwoIP.String()
		lldpOtherNeighbor.InterfaceOneIP = intfTwoIP.String()
	}
	return lldpNeighbor, lldpOtherNeighbor, nil

}

func (sh *DeviceInteractor) persistInterfaces(ctx context.Context, Device *domain.Device, DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	LOG := appcontext.Logger(ctx)

	//Persist Interfaces
	//Methods to use as key for Set Operations
	GetInterfaceKey := func(data interface{}) string {
		s, _ := data.(domain.Interface)
		return fmt.Sprintln(s.IntName, s.IntType)
	}
	//Method to compare two Interface objects to find whether they are updated
	InterfaceEqualMethod := func(first interface{}, second interface{}) (bool, domain.Interface, domain.Interface) {
		f, _ := first.(domain.Interface)
		s, _ := second.(domain.Interface)
		if f.IPAddress == s.IPAddress && f.Mac == s.Mac {
			return true, f, s
		}
		return false, f, s
	}

	OldInterfaces, _ := sh.Db.GetInterfacesonDevice(sh.FabricID, Device.ID)
	NewInterfaces := Device.Interfaces

	CreatedInterfaces, DeletedInterfaces, UpdatedInterfaces := interfaceconfig.Compare(GetInterfaceKey, InterfaceEqualMethod,
		OldInterfaces, NewInterfaces)

	LOG.Infoln("Created Interfaces", CreatedInterfaces)
	LOG.Infoln("Deleted Interfaces", DeletedInterfaces)
	LOG.Infoln("Updated Interfaces", UpdatedInterfaces)
	for _, CInf := range CreatedInterfaces {
		CInf.DeviceID = Device.ID
		CInf.ConfigType = domain.ConfigCreate
		if err := sh.Db.CreateInterface(&CInf); err != nil {
			statusMsg := fmt.Sprintf("Failed to create Physical Interface %s %s", CInf.IntType, CInf.IntName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, UIntf := range UpdatedInterfaces {
		UIntf.DeviceID = Device.ID
		UIntf.ConfigType = domain.ConfigUpdate
		if err := sh.Db.CreateInterface(&UIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to update Physical Interface %s %s", UIntf.IntType, UIntf.IntType)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, object := range DeletedInterfaces {
		//mark for Delete
		object.DeviceID = Device.ID
		object.ConfigType = domain.ConfigDelete
		if err := sh.Db.CreateInterface(&object); err != nil {
			statusMsg := fmt.Sprintf("Failed to delete Physical Interface %s %s", object.IntType, object.IntType)
			return statusMsg, errors.New(statusMsg)
		}
	}
	//update Device Interfaces from DB, so that it has the DeviceOneID and Interface ID
	Device.Interfaces, _ = sh.Db.GetInterfacesonDevice(sh.FabricID, Device.ID)
	statusMsg := fmt.Sprintf("Successfully saved Interfaces")
	return statusMsg, nil
}

func (sh *DeviceInteractor) persistSwitchAssets(ctx context.Context, Device *domain.Device, DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	//presist Interfaces
	if statusMsg, err := sh.persistInterfaces(ctx, Device, DeviceAdapter); err != nil {
		return statusMsg, err
	}
	//Persist LLDP
	if statusMsg, err := sh.persistLLDP(ctx, Device, DeviceAdapter); err != nil {
		return statusMsg, err
	}
	statusMsg := fmt.Sprintf("Successfully saved Assets")
	return statusMsg, nil
}
func (sh *DeviceInteractor) persistMCTConfigs(ctx context.Context, NewMct *domain.MctClusterConfig,
	NewMctMemberPorts []domain.MCTMemberPorts,
	ClusterConfigMap map[string]domain.MctClusterConfig,
	mctxt *MctContext) (string, error) {
	LOG := appcontext.Logger(ctx)
	//Lets Create Control Vlan and VE and Allocate IP Addree Here
	//MctCluster Config with Peer IP

	var err error
	var statusMsg string
	err = sh.persistControlVlan(ctx, NewMct)
	if err != nil {
		statusMsg = fmt.Sprintf("Failed to save control vlan %s on Devices %s, %s : %s",
			NewMct.ControlVE, NewMct.DeviceOneMgmtIP, NewMct.DeviceTwoMgmtIP, err.Error())

		return statusMsg, err
	}
	statusMsg, err = sh.persistMCTClusterConfigs(ctx, NewMct, ClusterConfigMap, mctxt)
	if err != nil {
		LOG.Errorln(statusMsg)
		return statusMsg, err
	}
	statusMsg, err = sh.persistMCTPortConfigs(ctx, NewMct, NewMctMemberPorts)
	if err != nil {
		LOG.Errorln(statusMsg)
		return statusMsg, err
	}
	return statusMsg, err
}

func (sh *DeviceInteractor) persistControlVlan(ctx context.Context, NewMct *domain.MctClusterConfig) error {
	LOG := appcontext.Logger(ctx)
	//TODO Make separate intrface table for VE/VLAN
	var err error
	var CIntfOne domain.Interface
	var CIntfTwo domain.Interface
	GetControlVE := func() error {
		GetIntf := func(DeviceID uint) (domain.Interface, error) {
			CIntf, err := sh.Db.GetInterface(NewMct.FabricID, DeviceID, "ve", NewMct.ControlVE)
			if err != nil && err != gorm.ErrRecordNotFound {
				statusMsg := fmt.Sprintf("Unable To Fetch VE Interface For Device %d  %s", DeviceID, NewMct.ControlVE)
				LOG.Errorln(statusMsg)
				return CIntf, err
			}
			if err == gorm.ErrRecordNotFound {
				var Interface domain.Interface
				Interface.DeviceID = DeviceID
				Interface.FabricID = NewMct.FabricID
				Interface.IntType = "ve"
				Interface.IntName = NewMct.ControlVE
				Interface.ConfigType = domain.ConfigCreate

				err = sh.Db.CreateInterface(&Interface)
				if err != nil {
					statusMsg := fmt.Sprintf("Unable To Create VE interface %s For Device %d", NewMct.ControlVE, DeviceID)
					LOG.Errorln(statusMsg)
					return CIntf, err
				}
				CIntf = Interface
			}
			return CIntf, nil

		}
		if CIntfOne, err = GetIntf(NewMct.DeviceID); err != nil {
			return err
		}
		if CIntfTwo, err = GetIntf(NewMct.MCTNeighborDeviceID); err != nil {
			return err
		}
		NewMct.VEInterfaceOneID = CIntfOne.ID
		NewMct.VEInterfaceTwoID = CIntfTwo.ID
		return nil
	}
	if err = GetControlVE(); err != nil {
		return err
	}

	FabricProperties, err := sh.Db.GetFabricProperties(sh.FabricID)
	if err != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch fabric Properties  %s", sh.FabricName)
		LOG.Errorln(statusMsg)
	}
	if CIntfOne.IPAddress, CIntfTwo.IPAddress, err = sh.reserveOrObtainIPPair(sh.AppContext,
		NewMct.DeviceID, NewMct.MCTNeighborDeviceID, CIntfOne.IntType, CIntfOne.IntName,
		CIntfTwo.IntType, CIntfTwo.IntName, FabricProperties.MCTLinkIPRange,
		domain.MCTPoolName, CIntfOne.ID, CIntfTwo.ID); err != nil {
		//If there is UPDATE persist MCT Config will detect
		//If they in different subnet MCT Validation code would detect it.
		if strings.Contains(err.Error(), "not present in the Used IP Table for Fabric") {
			//Kludge - This Error Will Occur When one of the device is already participating in mCT
			//CLuster with other device , And Hence one of the VE interface is already having IP address
			//and other VE Interface doesnt have IP addres
			statusMsg := fmt.Sprintf("Persist Control Known Error While reserving IP Address %s", err.Error())
			LOG.Errorln(statusMsg)
			return nil
		}
		statusMsg := fmt.Sprintf("Persist Control vlan Error While Reserving IP %s", err.Error())
		LOG.Errorln(statusMsg)
		return err
	}
	NewMct.PeerTwoIP = CIntfOne.IPAddress
	NewMct.PeerOneIP = CIntfTwo.IPAddress
	return nil
}
func (sh *DeviceInteractor) persistMCTClusterConfigs(ctx context.Context, NewMct *domain.MctClusterConfig,
	ClusterConfigMap map[string]domain.MctClusterConfig, mctxt *MctContext) (string, error) {
	LOG := appcontext.Logger(ctx)
	var ConfigType []string
	GetKey := func(DeviceOneID uint, DeviceTwoID uint) string {
		return fmt.Sprintf("%d : %d", DeviceOneID, DeviceTwoID)
	}
	setBit := func(n uint64, pos uint) uint64 {
		n |= (1 << pos)
		return n
	}
	//ConfigType = append(ConfigType, domain.ConfigNone)
	OldMcts, merr := sh.Db.GetMctClusters(NewMct.FabricID, NewMct.DeviceID, ConfigType)
	if merr != nil {
		statusMsg := fmt.Sprintf("Unable To Retrieve MctCluster from DB for device %s", NewMct.DeviceOneMgmtIP)
		LOG.Errorln(statusMsg, merr)
		return statusMsg, merr
	}

	if len(OldMcts) == 0 {
		NewMct.ConfigType = domain.ConfigCreate
		NewMct.ClusterID = 1
		err := sh.Db.CreateMctClusters(NewMct)
		if err != nil {
			statusMsg := fmt.Sprintf("Failed To created MCT Cluster for Device %s", NewMct.DeviceOneMgmtIP)
			LOG.Errorln(statusMsg, err)
			return statusMsg, err
		}
		statusMsg := fmt.Sprintf("successfully Create MCT Cluster for device %d", NewMct.DeviceID)
		return statusMsg, nil
	}

	for _, OldMct := range OldMcts {

		if OldMct.MCTNeighborDeviceID == NewMct.MCTNeighborDeviceID &&
			OldMct.DeviceID == NewMct.DeviceID {
			var IsUpdated bool
			//Lets us Not Update NODEID Since They cant be updated in efa context
			NewMct.LocalNodeID = OldMct.LocalNodeID
			NewMct.RemoteNodeID = OldMct.RemoteNodeID
			NewMct.ID = OldMct.ID
			NewMct.ConfigType = OldMct.ConfigType
			NewMct.ClusterID = OldMct.ClusterID
			IsUpdated = false
			if OldMct.PeerOneIP != NewMct.PeerOneIP || OldMct.PeerTwoIP != NewMct.PeerTwoIP {
				//By This time IP Pair Should be in Same subnet hence no validation
				OldMct.ConfigType = domain.ConfigUpdate
				IsUpdated = true
				msg := fmt.Sprintf("Peer IP change is detected in MCT Cluster For Device %s Neighbor Device %s\n",
					NewMct.DeviceOneMgmtIP, NewMct.DeviceTwoMgmtIP)
				LOG.Infoln(msg)
				OldMct.PeerOneIP = NewMct.PeerOneIP
				OldMct.PeerTwoIP = NewMct.PeerTwoIP
				//Update Both Ends
				OldMct.UpdatedAttributes = setBit(OldMct.UpdatedAttributes, domain.BitPositionForPeerOneIP)
				OldMct.UpdatedAttributes = setBit(OldMct.UpdatedAttributes, domain.BitPositionForPeerTwoIP)
			}
			if OldMct.PeerInterfaceSpeed != NewMct.PeerInterfaceSpeed && NewMct.PeerInterfaceSpeed != "0" {
				IsUpdated = true
				OldMct.ConfigType = domain.ConfigUpdate
				msg := fmt.Sprintf("Change in MCT Cluster Speed Detected Between Device %s "+
					"Neighbor Device %s OldSpeed %s New Speed %s\n",
					NewMct.DeviceOneMgmtIP, NewMct.DeviceTwoMgmtIP,
					OldMct.PeerInterfaceSpeed, NewMct.PeerInterfaceSpeed)
				LOG.Infof("%s", msg)
				OldMct.PeerInterfaceSpeed = NewMct.PeerInterfaceSpeed
				OldMct.UpdatedAttributes = setBit(OldMct.UpdatedAttributes, domain.BitPositionForPeerSpeed)

			}
			NewMct.UpdatedAttributes = OldMct.UpdatedAttributes
			LOG.Infoln("NewMct UpdatedAttributes ", NewMct.UpdatedAttributes)
			if OldMct.ConfigType == domain.ConfigUpdate && IsUpdated {
				//Update DB if there is update else skip
				err := sh.Db.CreateMctClusters(&OldMct)
				if err != nil {
					statusMsg := fmt.Sprintf("failedto Mark  MCT cluster for Delete for device %d", OldMct.DeviceID)
					return statusMsg, err
				}
			}
			fetchCluster := func() (string, error) {
				ClusterName := fmt.Sprintf("%s%s%d", sh.FabricName, "-cluster-", OldMct.ClusterID)
				ClusterOne, err := mctxt.DeviceOneAdapter.GetClusterByName(ClusterName)
				if err != nil {
					statusMsg := fmt.Sprintln("Failed to fetch cluster from device - ", OldMct.DeviceOneMgmtIP)
					LOG.Errorln(statusMsg)
					return statusMsg, err

				}
				ClusterTwo, err := mctxt.DeviceTwoAdapter.GetClusterByName(ClusterName)
				if err != nil {
					statusMsg := fmt.Sprintln("Failed to fetch cluster from device - ", OldMct.DeviceOneMgmtIP)
					LOG.Errorln(statusMsg)
					return statusMsg, err

				}
				_, ok1 := ClusterOne["cluster-id"]
				_, ok2 := ClusterTwo["cluster-id"]
				if (OldMct.ConfigType != domain.ConfigCreate) && (ok1 == false || ok2 == false) {
					//Cluster knocked out from Both the devices
					LOG.Infof("MCT Cluster knocked Out from Both the Device %s %s",
						OldMct.DeviceOneMgmtIP, OldMct.DeviceTwoMgmtIP)
					//Mark these cluster to CREATED
					OldMct.ConfigType = domain.ConfigUpdate
					NewMct.ConfigType = domain.ConfigUpdate
					NewMct.UpdatedAttributes = setBit(NewMct.UpdatedAttributes, domain.BitPositionForMctCreate)
					OldMct.UpdatedAttributes = setBit(OldMct.UpdatedAttributes, domain.BitPositionForMctCreate)

					err := sh.Db.CreateMctClusters(&OldMct)
					if err != nil {
						statusMsg := fmt.Sprintf("failed to Mark  MCT cluster for Create for device %d", OldMct.DeviceID)
						LOG.Errorln(statusMsg)
						return statusMsg, err
					}
					err = sh.Db.MarkMctClusterMemberPortsForCreate(OldMct.FabricID, OldMct.DeviceID, OldMct.MCTNeighborDeviceID)
					if err != nil {
						msg := fmt.Sprintf("Marking All the MCT Port for create Failed for Device %s and Neighbor"+
							" Device %s", OldMct.DeviceOneMgmtIP, OldMct.DeviceTwoMgmtIP)
						LOG.Errorln(msg)
						return msg, err
					}
					err = sh.Db.MarkMctClusterMemberPortsForCreate(OldMct.FabricID, OldMct.MCTNeighborDeviceID, OldMct.DeviceID)
					if err != nil {
						msg := fmt.Sprintf("Marking All the MCT Port for create Failed for Device %s and Neighbor"+
							" Device %s", OldMct.DeviceTwoMgmtIP, OldMct.DeviceOneMgmtIP)
						LOG.Errorln(msg)
						return msg, err
					}
				}
				return "", nil

			}
			if msg, err := fetchCluster(); err != nil {
				return msg, err
			}

		} else {
			key := GetKey(NewMct.DeviceID, NewMct.MCTNeighborDeviceID)
			if existingMct, _ := sh.Db.GetMctClustersWithBothDevices(NewMct.FabricID, NewMct.DeviceID, NewMct.MCTNeighborDeviceID, []string{}); len(existingMct) > 0 {
				LOG.Infof("Mct Device Already Exists Nothing to be done for %s and %s", NewMct.DeviceOneMgmtIP, NewMct.DeviceTwoMgmtIP)
				continue
			}
			if _, ok := ClusterConfigMap[key]; ok == true {
				NewMct.ConfigType = domain.ConfigCreate
				NewMct.ClusterID = 1
				err := sh.Db.CreateMctClusters(NewMct)
				if err != nil {
					statusMsg := fmt.Sprintf("Failed To Create MCT cluster For Device %s", NewMct.DeviceOneMgmtIP)
					LOG.Errorln(statusMsg, err)
					return statusMsg, err
				}
				statusMsg := fmt.Sprintf("Successfully Created MCT Cluster For Device %s", NewMct.DeviceOneMgmtIP)
				LOG.Info(statusMsg)

			}

		}
		if NewMct.ID == 0 {
			LOG.Debug("Debug")
		}

	}
	statusMsg := fmt.Sprintf("successfully Create MCT Cluster for device %d", NewMct.DeviceID)
	return statusMsg, nil

}

func (sh *DeviceInteractor) persistMCTPortConfigs(ctx context.Context, NewMct *domain.MctClusterConfig, NewMctMemberPorts []domain.MCTMemberPorts) (string, error) {
	var err error
	LOG := appcontext.Logger(ctx)
	OldMct, _ := sh.Db.GetMctMemberPortsConfig(sh.FabricID, NewMct.DeviceID, NewMct.MCTNeighborDeviceID, []string{})
	//Methods to use as key for Set Operations
	GetMctKey := func(data interface{}) string {
		s, _ := data.(domain.MCTMemberPorts)
		return fmt.Sprintln(s.DeviceID, s.RemoteDeviceID, s.InterfaceName, s.RemoteInterfaceName)
	}
	setBit := func(n uint64, pos uint) uint64 {
		n |= (1 << pos)
		return n
	}
	hasBit := func(n uint64, pos uint) bool {
		val := n & (1 << pos)
		return (val > 0)
	}
	//Method to compare two Interface objects to find whether they are updated
	MctConfigEqualMethod := func(first interface{}, second interface{}) (bool, domain.MCTMemberPorts,
		domain.MCTMemberPorts) {
		f, _ := first.(domain.MCTMemberPorts)
		s, _ := second.(domain.MCTMemberPorts)
		//TODO Check Interface Speed and update set Equal in comparator
		if (f.DeviceID == s.DeviceID && f.RemoteDeviceID == s.RemoteDeviceID) &&
			(f.InterfaceType == s.InterfaceType && f.InterfaceName == s.InterfaceName) &&
			(f.RemoteInterfaceType == s.RemoteInterfaceType && f.RemoteInterfaceName == s.RemoteInterfaceName) {
			return true, f, s
		}
		return false, f, s
	}
	CreatedMCTMemberPorts, DeletedMCTMemberPorts, UpdatedMCTMctMemberPorts := mctconfig.Compare(
		GetMctKey, MctConfigEqualMethod, OldMct, NewMctMemberPorts)

	LOG.Infoln("Created MCTPorts", CreatedMCTMemberPorts)
	LOG.Infoln("Deleted MCTPorts", DeletedMCTMemberPorts)
	LOG.Infoln("Updated MCTPorts", UpdatedMCTMctMemberPorts)
	err = sh.Db.CreateMctClustersMembers(CreatedMCTMemberPorts, NewMct.ID, domain.ConfigCreate)
	if err != nil {
		statusMsg := fmt.Sprintln("Failed to Create MCT Cluster Member Ports", CreatedMCTMemberPorts, err)
		LOG.Errorln(statusMsg, err)
		return statusMsg, errors.New(statusMsg)
	}
	err = sh.Db.CreateMctClustersMembers(UpdatedMCTMctMemberPorts, NewMct.ID, domain.ConfigUpdate)
	if err != nil {
		statusMsg := fmt.Sprintln("Failed to Update MCT Cluster Member ports", UpdatedMCTMctMemberPorts, err)
		LOG.Errorln(statusMsg, err)
		return statusMsg, errors.New(statusMsg)
	}
	err = sh.Db.CreateMctClustersMembers(DeletedMCTMemberPorts, NewMct.ID, domain.ConfigDelete)
	if err != nil {
		statusMsg := fmt.Sprintln("Failed To Delete MCT Cluster Member ports", DeletedMCTMemberPorts, err)
		LOG.Errorln(statusMsg, err)
		return statusMsg, errors.New(statusMsg)
	}
	if NewMct.ConfigType == domain.ConfigNone || (NewMct.ConfigType == domain.ConfigUpdate &&
		(!hasBit(NewMct.UpdatedAttributes, domain.BitPositionForForPortAdd) ||
			!hasBit(NewMct.UpdatedAttributes, domain.BitPositionForForPortDelete))) {
		if (len(CreatedMCTMemberPorts) > 0) || (len(DeletedMCTMemberPorts) > 0) {
			//UPdate the Cluster so that it can detected that Member Ports are updated(add/delete)
			//Check If Add/Delete Port can be sent in one Request
			LOG.Infoln("MCT Cluster Port Update Detected")
			NewMct.ConfigType = domain.ConfigUpdate
			if len(DeletedMCTMemberPorts) > 0 {
				NewMct.UpdatedAttributes = setBit(NewMct.UpdatedAttributes, domain.BitPositionForForPortDelete)
			}
			if len(CreatedMCTMemberPorts) > 0 {
				NewMct.UpdatedAttributes = setBit(NewMct.UpdatedAttributes, domain.BitPositionForForPortAdd)
			}
			LOG.Infoln("MCT CLuster Memberport Update NewMct UpdatedAttributes - ", NewMct.UpdatedAttributes)
			err := sh.Db.CreateMctClusters(NewMct)
			if err != nil {
				statusMsg := fmt.Sprintf("Failed TO create MCT Cluster for device %s", NewMct.DeviceOneMgmtIP)
				LOG.Errorln(statusMsg, err)
				return statusMsg, err
			}
		}
	}
	return "SUCCESS", err
}

//Persist Neighbor relationship
func (sh *DeviceInteractor) persistLLDPNeighbors(ctx context.Context, device *domain.Device, LLDPNeighbors []domain.LLDPNeighbor) (string, error) {
	LOG := appcontext.Logger(ctx)

	OldLLDPNeighbors, _ := sh.Db.GetLLDPNeighborsOnEitherDevice(sh.FabricID, device.ID)

	//Methods to use as key for Set Operations
	GetInterfaceKey := func(data interface{}) string {
		s, _ := data.(domain.LLDPNeighbor)
		return fmt.Sprintln(s.DeviceOneID, s.DeviceTwoID, s.InterfaceOneID, s.InterfaceTwoID)
	}
	//Method to compare two Interface objects to find whether they are updated
	InterfaceEqualMethod := func(first interface{}, second interface{}) (bool, domain.LLDPNeighbor, domain.LLDPNeighbor) {
		f, _ := first.(domain.LLDPNeighbor)
		s, _ := second.(domain.LLDPNeighbor)
		if f.InterfaceOneIP == s.InterfaceOneIP && f.InterfaceTwoIP == s.InterfaceTwoIP {
			return true, f, s
		}
		return false, f, s
	}

	LOG.Infoln("Old LLDPNeighbors", OldLLDPNeighbors)
	LOG.Infoln("New LLDPNeighbors", LLDPNeighbors)
	CreatedLLDPNeighbors, DeletedLLDPNeighbors, UpdatedLLDPNeighbors := lldpneighbor.Compare(GetInterfaceKey, InterfaceEqualMethod,
		OldLLDPNeighbors, LLDPNeighbors)

	LOG.Infoln("Created LLDPNeighbors", CreatedLLDPNeighbors)
	LOG.Infoln("Deleted LLDPNeighbors", DeletedLLDPNeighbors)
	LOG.Infoln("Updated LLDPNeighbors", UpdatedLLDPNeighbors)

	for _, CIntf := range CreatedLLDPNeighbors {
		CIntf.ConfigType = domain.ConfigCreate
		if err := sh.Db.CreateLLDPNeighbor(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to create LLDP Neighbor %s %s", CIntf.InterfaceOneType, CIntf.InterfaceOneName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, CIntf := range UpdatedLLDPNeighbors {
		CIntf.ConfigType = domain.ConfigUpdate
		if err := sh.Db.CreateLLDPNeighbor(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to update LLDP Neighbor %s %s", CIntf.InterfaceOneType, CIntf.InterfaceOneName)
			return statusMsg, errors.New(statusMsg)
		}
	}

	for _, CIntf := range DeletedLLDPNeighbors {
		//mark for Delete
		CIntf.ConfigType = domain.ConfigDelete
		if err := sh.Db.CreateLLDPNeighbor(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to delete LLDP Neighbor %s %s", CIntf.InterfaceOneType, CIntf.InterfaceOneName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	return "", nil
}

func (sh *DeviceInteractor) persistLLDP(ctx context.Context, Device *domain.Device, DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	LOG := appcontext.Logger(ctx)
	//Persist Interfaces
	//Methods to use as key for Set Operations
	GetKey := func(data interface{}) string {
		s, _ := data.(domain.LLDP)
		return fmt.Sprintln(s.LocalIntName, s.LocalIntType)
	}
	//Method to compare two LLDP objects to find whether they are updated
	EqualMethod := func(first interface{}, second interface{}) (bool, domain.LLDP, domain.LLDP) {
		f, _ := first.(domain.LLDP)
		s, _ := second.(domain.LLDP)
		if f.RemoteIntName == s.RemoteIntName && f.RemoteIntType == s.RemoteIntType && f.RemoteIntMac == s.RemoteIntMac {
			return true, f, s
		}
		return false, f, s
	}

	Old, _ := sh.Db.GetLLDPsonDevice(sh.FabricID, Device.ID)
	New := Device.LLDPS

	Created, Deleted, Updated := lldp.Compare(GetKey, EqualMethod,
		Old, New)

	LOG.Infoln("Created LLDPS", Created)
	LOG.Infoln("Deleted LLDPS", Deleted)
	LOG.Infoln("Updated LLDPS", Updated)
	for _, object := range Created {
		object.ConfigType = domain.ConfigCreate
		object.DeviceID = Device.ID
		if err := sh.Db.CreateLLDP(&object); err != nil {
			statusMsg := fmt.Sprintf("Failed to create LLDP for %s %s", object.LocalIntType, object.LocalIntName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, object := range Updated {
		object.ConfigType = domain.ConfigUpdate
		object.DeviceID = Device.ID
		if err := sh.Db.CreateLLDP(&object); err != nil {
			statusMsg := fmt.Sprintf("Failed to update LLDP %s %s", object.LocalIntType, object.LocalIntType)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, object := range Deleted {
		//mark for Delete
		object.ConfigType = domain.ConfigDelete
		object.DeviceID = Device.ID
		if err := sh.Db.CreateLLDP(&object); err != nil {
			statusMsg := fmt.Sprintf("Failed to delete LLDP %s %s", object.LocalIntType, object.LocalIntType)
			return statusMsg, errors.New(statusMsg)
		}
	}

	//update Device Interfaces from DB, so that it has the DeviceOneID and Interface ID
	Device.LLDPS, _ = sh.Db.GetLLDPsonDevice(sh.FabricID, Device.ID)
	statusMsg := fmt.Sprintf("Successfully saved LLDPS")
	return statusMsg, nil
}

func (sh *DeviceInteractor) persistNonClosEvpnNeighborConfigs(ctx context.Context, ThisRack domain.Rack, NeighborList []domain.RackEvpnNeighbors) (string, error) {
	LOG := appcontext.Logger(ctx)

	ExistingNeighborList, _ := sh.Db.GetRackEvpnConfig(ThisRack.ID)
	LOG.Infoln("Old Non-clos Evpn Neighbor Configs", ExistingNeighborList)
	LOG.Infoln("New Non-clos Evpn Neighbor Configs", NeighborList)

	//Methods to use as key for Set Operations
	GetInterfaceKey := func(data interface{}) string {
		s, _ := data.(domain.RackEvpnNeighbors)
		return fmt.Sprintln(s.LocalRackID, s.LocalDeviceID, s.RemoteRackID, s.RemoteDeviceID)
	}

	//Method to compare two Interface objects to find whether they are updated
	InterfaceEqualMethod := func(first interface{}, second interface{}) (bool, domain.RackEvpnNeighbors, domain.RackEvpnNeighbors) {
		f, _ := first.(domain.RackEvpnNeighbors)
		s, _ := second.(domain.RackEvpnNeighbors)
		if f.LocalRackID == s.LocalRackID && f.LocalDeviceID == s.LocalDeviceID &&
			f.RemoteRackID == s.RemoteRackID && f.RemoteDeviceID == s.RemoteDeviceID {
			return true, f, s
		}
		return false, f, s
	}
	CreatedNeighbors, DeletedNeighbors, UpdatedNeighbors := rackevpnneighbor.Compare(GetInterfaceKey, InterfaceEqualMethod,
		ExistingNeighborList, NeighborList)
	LOG.Infoln("Created non-clos EVPN Neighbors", CreatedNeighbors)
	LOG.Infoln("Deleted on-clos EVPN Neighbors", DeletedNeighbors)
	LOG.Infoln("Updated on-clos EVPN Neighbors", UpdatedNeighbors)

	for _, CIntf := range CreatedNeighbors {
		CIntf.ConfigType = domain.ConfigCreate
		if err := sh.Db.CreateRackEvpnConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to create non-clos EVPN Neighbor [LR:%d LD:%d RR:%d RD:%d",
				CIntf.LocalRackID, CIntf.LocalDeviceID, CIntf.RemoteRackID, CIntf.RemoteDeviceID)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, CIntf := range UpdatedNeighbors {
		CIntf.ConfigType = domain.ConfigUpdate
		if err := sh.Db.CreateRackEvpnConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to update non-clos EVPN Neighbor [LR:%d LD:%d RR:%d RD:%d",
				CIntf.LocalRackID, CIntf.LocalDeviceID, CIntf.RemoteRackID, CIntf.RemoteDeviceID)
			return statusMsg, errors.New(statusMsg)
		}
	}

	for _, CIntf := range DeletedNeighbors {
		//mark for Delete
		CIntf.ConfigType = domain.ConfigDelete
		if err := sh.Db.CreateRackEvpnConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to update non-clos EVPN Neighbor [LR:%d LD:%d RR:%d RD:%d",
				CIntf.LocalRackID, CIntf.LocalDeviceID, CIntf.RemoteRackID, CIntf.RemoteDeviceID)
			return statusMsg, errors.New(statusMsg)
		}
	}
	return "", nil
}

//ListDevices lists the devices in a given fabric
func (sh *DeviceInteractor) ListDevices(FabricName string) ([]domain.Device, error) {
	var Devices []domain.Device
	Fabric, err := sh.Db.GetFabric(FabricName)
	if err == nil {
		return sh.Db.GetDevicesInFabric(Fabric.ID)
	}
	return Devices, errors.New("unable to fetch fabric")

}

//enableInterfaces enables interfaces on the Device
func (sh *DeviceInteractor) enableInterfaces(ctx context.Context, Device *domain.Device,
	DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	LOG := appcontext.Logger(ctx)
	InterfaceNames := make([]string, 0, 0)

	for _, Interface := range Device.Interfaces {
		//Enable only ethernet
		if Interface.IntType == domain.IntfTypeEthernet && Interface.ConfigState != "up" {
			InterfaceNames = append(InterfaceNames, Interface.IntName)
		}
	}
	if len(InterfaceNames) != 0 {
		LOG.Infoln("Enable Interfaces", InterfaceNames)
		if status, err := DeviceAdapter.EnableInterfaces(InterfaceNames); err != nil {
			return status, err
		}
		//Sleep for sec for LLDP's to come back
		time.Sleep(constants.LLDPSleep * time.Second)
	}
	return "", nil
}

//FetchMctNeighborVTEPLoopBackIPAndASN Fetches MCT neighbor Devices ASN, VTEP IP and Loopback IP
func (sh *DeviceInteractor) FetchMctNeighborVTEPLoopBackIPAndASN(ctx context.Context, DeviceID uint) (string, string, string, error) {
	VTEPLoopBackIP := ""
	LoopBackIP := ""
	Asn := ""
	LOG := appcontext.Logger(ctx)
	clusters, err := sh.Db.GetMctClusters(sh.FabricID, DeviceID, []string{})
	if err != nil {
		LOG.Errorln("Error Retrieving ")
		return VTEPLoopBackIP, LoopBackIP, Asn, err
	}
	for _, cluster := range clusters {
		switchConfig, err := sh.Db.GetSwitchConfigOnDeviceIP(sh.FabricName, cluster.DeviceTwoMgmtIP)
		if err != nil {
			statusMsg := fmt.Sprintf("Failed to fetch configs for device %s from %s", cluster.DeviceTwoMgmtIP, sh.FabricName)
			LOG.Errorln(statusMsg)
			return VTEPLoopBackIP, LoopBackIP, Asn, errors.New(statusMsg)
		}
		return switchConfig.VTEPLoopbackIP, switchConfig.LoopbackIP, switchConfig.LocalAS, nil
	}
	return VTEPLoopBackIP, LoopBackIP, Asn, nil
}

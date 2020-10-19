package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	Interactor "efa-server/usecase/interactorinterface"
	"errors"
	"fmt"
)

//AddRackFirstStage does the following
//  Adds each pair of device to the Devices table
func (sh *DeviceInteractor) AddRackFirstStage(ctx context.Context, FabricName string, IP1Address string,
	IP2Address string, UserName string, Password string) error {

	IPPair := fmt.Sprintln(IP1Address, ",", IP2Address)
	ctx = context.WithValue(ctx, appcontext.IPPair, IPPair)

	//Add the first IP Address from the Rack
	if err := sh.AddDeviceFromRackInFirstStage(ctx, FabricName, IP1Address, UserName, Password); err != nil {
		return err
	}
	//Add the second IP Address from the Rack
	if err := sh.AddDeviceFromRackInFirstStage(ctx, FabricName, IP2Address, UserName, Password); err != nil {
		return err
	}

	return nil

}

//AddDeviceFromRackInFirstStage does the following
// 1 Fetches Interfaces from the pair of Device
// 2 Enables ethernet Interfaces that are admin down
// 3 Persists the Interfaces
func (sh *DeviceInteractor) AddDeviceFromRackInFirstStage(ctx context.Context, FabricName string, IPAddress string, UserName string, Password string) error {
	var err error
	LOG := appcontext.Logger(ctx)
	Role := "Rack"

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

//AddRackSecondStage does the following
// 1 Fetches LLDP Data from the Device
func (sh *DeviceInteractor) AddRackSecondStage(ctx context.Context, FabricName string, IP1Address string,
	IP2Address string, UserName string, Password string) error {

	IPPair := fmt.Sprintln(IP1Address, ",", IP2Address)
	ctx = context.WithValue(ctx, appcontext.IPPair, IPPair)

	//Add the first IP Address from the Rack
	if err := sh.AddDeviceFromRackInSecondStage(ctx, FabricName, IP1Address, UserName, Password); err != nil {
		return err
	}
	//Add the second IP Address from the Rack
	if err := sh.AddDeviceFromRackInSecondStage(ctx, FabricName, IP2Address, UserName, Password); err != nil {
		return err
	}

	return nil
}

//AddDeviceFromRackInSecondStage does the following
// 1 Fetches LLDP Data from the Device and persist
func (sh *DeviceInteractor) AddDeviceFromRackInSecondStage(ctx context.Context, FabricName string,
	IPAddress string, UserName string, Password string) error {

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

//AddRackThirdStage does the following
// 2 Builds LLDP Neighbor relationship
// 3 Generate Device Level Configs to be pushed
func (sh *DeviceInteractor) AddRackThirdStage(ctx context.Context, FabricName string, IP1Address string,
	IP2Address string, UserName string, Password string) error {

	IPPair := fmt.Sprintln(IP1Address, ",", IP2Address)
	ctx = context.WithValue(ctx, appcontext.IPPair, IPPair)

	//Add the first IP Address from the Rack
	if err := sh.AddDeviceFromRackInThirdStage(ctx, FabricName, IP1Address, UserName, Password); err != nil {
		return err
	}
	//Add the second IP Address from the Rack
	if err := sh.AddDeviceFromRackInThirdStage(ctx, FabricName, IP2Address, UserName, Password); err != nil {
		return err
	}

	return nil
}

//AddDeviceFromRackInThirdStage builds Relation Ship
func (sh *DeviceInteractor) AddDeviceFromRackInThirdStage(ctx context.Context, FabricName string, IPAddress string,
	UserName string, Password string) error {
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
	if statusMsg, err := sh.computeRoleAndBuildNeighborRelationShipForRackDevice(ctx, &Device); err != nil {
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

//Compute the Device Role for the Switch and Builds its Neighbor Relationship
func (sh *DeviceInteractor) computeRoleAndBuildNeighborRelationShipForRackDevice(ctx context.Context, device *domain.Device) (string, error) {
	LOG := appcontext.Logger(ctx)

	LOG.Infoln("Compute Role and Neighbor relationships")

	deviceMap := make(map[uint]domain.Device, 0)
	//key is deviceID and Neighbor DeviceOneID
	ClusterConfigMap := make(map[string]domain.MctClusterConfig, 0)
	MctMemberPortsMap := make(map[string][]domain.MCTMemberPorts, 0)
	device.DeviceRole = RackRole
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

			//Interface 46,47 are used as MCT Members
			if phy.IntName == MCTInterfaceOne || phy.IntName == MCTInterfaceTwo {
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

//AddRackFourthStage does the following
// 2 Builds LLDP Neighbor relationship
// 3 Generate Device Level Configs to be pushed
func (sh *DeviceInteractor) AddRackFourthStage(ctx context.Context, FabricName string, IP1Address string,
	IP2Address string, UserName string, Password string) error {

	IPPair := fmt.Sprintln(IP1Address, ",", IP2Address)
	ctx = context.WithValue(ctx, appcontext.IPPair, IPPair)

	//Add the first IP Address from the Rack
	if err := sh.AddDeviceFromRackInFourthStage(ctx, FabricName, IP1Address, UserName, Password); err != nil {
		return err
	}
	//Add the second IP Address from the Rack
	if err := sh.AddDeviceFromRackInFourthStage(ctx, FabricName, IP2Address, UserName, Password); err != nil {
		return err
	}
	//compute inter-rack evpn neighborships
	if err := sh.computeNonClosEvpnNeighborships(ctx, FabricName, Rack{IP1Address, IP2Address}); err != nil {
		return err
	}
	return nil
}

//AddDeviceFromRackInFourthStage does the following
// 1 Generate IneterfaceSwitchConfigs for the Device
// 2 Generate RemoteBGP Neighbor configs for the Device
func (sh *DeviceInteractor) AddDeviceFromRackInFourthStage(ctx context.Context, FabricName string, IPAddress string,
	UserName string, Password string) error {
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

	if _, err := sh.buildNonCLOSInterfaceConfigs(ctx, &Device, true); err != nil {
		return err
	}

	//Save Device Again
	if err := sh.Db.SaveDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s Save Failed", IPAddress)
		return errors.New(statusMsg)

	}

	return nil
}

func (sh *DeviceInteractor) buildRackDeviceMapTable(ctx context.Context, FabricName string, RackList []Rack) []RackDeviceMapTable {
	var RDMapTable []RackDeviceMapTable

	for _, rack := range RackList {
		var RDMap RackDeviceMapTable

		r, err := sh.Db.GetRack(FabricName, rack.IP1, rack.IP2)
		if err != nil {
			continue
		}
		d1, err := sh.Db.GetDevice(FabricName, rack.IP1)
		if err != nil {
			continue
		}
		d2, err := sh.Db.GetDevice(FabricName, rack.IP2)
		if err != nil {
			continue
		}
		sw1, err := sh.Db.GetSwitchConfigOnFabricIDAndDeviceID(d1.FabricID, d1.ID)
		if err != nil {
			continue
		}
		sw2, err := sh.Db.GetSwitchConfigOnFabricIDAndDeviceID(d2.FabricID, d2.ID)
		if err != nil {
			continue
		}
		RDMap.Rack = r
		RDMap.Devices = append(RDMap.Devices, DeviceSwitchConfigMapTable{d1, sw1})
		RDMap.Devices = append(RDMap.Devices, DeviceSwitchConfigMapTable{d2, sw2})
		RDMapTable = append(RDMapTable, RDMap)
	}
	return RDMapTable
}

func (sh *DeviceInteractor) computeNonClosEvpnNeighborships(ctx context.Context, FabricName string, ThisRack Rack) error {
	var NeighborList []domain.RackEvpnNeighbors
	var RackList []Rack
	var ThisRackRDMap RackDeviceMapTable
	var err error

	LOG := appcontext.Logger(ctx)

	LOG.Infof("\nFabric Name:%s, This Rack:%+v\n", FabricName, ThisRack)
	RackList, err = sh.fetchRegisteredRacks(ctx, FabricName)
	if err != nil {
		LOG.Errorf("Failed to fetch registered racks")
		return err
	}
	LOG.Infof("\nRegistered Racks:%+v\n", RackList)
	RDMapTable := sh.buildRackDeviceMapTable(ctx, FabricName, RackList)
	LOG.Infof("\nRDMapTable:%+v\n", RDMapTable)
	r, err := sh.Db.GetRack(FabricName, ThisRack.IP1, ThisRack.IP2)
	if err != nil {
		LOG.Errorf("failed to get rack: %s, %s", ThisRack.IP1, ThisRack.IP2)
		return err
	}
	for _, rd := range RDMapTable {
		if rd.Rack.ID == r.ID {
			ThisRackRDMap = rd
			break
		}
	}
	sh.computeRackEvpnNeighborships(ctx, FabricName, RDMapTable, ThisRackRDMap, &NeighborList)
	sh.persistNonClosEvpnNeighborConfigs(ctx, ThisRackRDMap.Rack, NeighborList)
	return nil
}

func (sh *DeviceInteractor) computeDeviceRackEvpnNeighborships(ctx context.Context,
	ThisDevice DeviceSwitchConfigMapTable, ThisRack RackDeviceMapTable,
	RemoteRack RackDeviceMapTable, NeighborList *[]domain.RackEvpnNeighbors) {
	for _, RemoteDevice := range RemoteRack.Devices {
		var Neighbor domain.RackEvpnNeighbors

		Neighbor.LocalRackID = ThisRack.Rack.ID
		Neighbor.LocalDeviceID = ThisDevice.Device.ID
		Neighbor.RemoteRackID = RemoteRack.Rack.ID
		Neighbor.RemoteDeviceID = RemoteDevice.Device.ID
		Neighbor.EVPNAddress = RemoteDevice.Config.LoopbackIP
		Neighbor.RemoteAS = RemoteDevice.Config.LocalAS

		*NeighborList = append(*NeighborList, Neighbor)
	}
}

func (sh *DeviceInteractor) computeRackEvpnNeighborships(ctx context.Context, FabricName string, RackTable []RackDeviceMapTable,
	ThisRack RackDeviceMapTable, NeighborList *[]domain.RackEvpnNeighbors) {
	for _, RemoteRack := range RackTable {
		// it is the current rack for which we are computing neighborships
		if RemoteRack.Rack.ID == ThisRack.Rack.ID {
			continue
		}

		sh.computeDeviceRackEvpnNeighborships(ctx, ThisRack.Devices[0], ThisRack, RemoteRack, NeighborList)
		sh.computeDeviceRackEvpnNeighborships(ctx, ThisRack.Devices[1], ThisRack, RemoteRack, NeighborList)
	}
}

func (sh *DeviceInteractor) buildNonCLOSInterfaceConfigs(ctx context.Context, device *domain.Device, rack bool) (string, error) {
	LOG := appcontext.Logger(ctx)
	LOG.Infoln("Build interface configs")
	FabricProperties, _ := sh.Db.GetFabricProperties(sh.FabricID)
	if msg, err := sh.interfaceConfigsBasedOnDeletedLLDPNeighbors(ctx, device, &FabricProperties); err != nil {
		return msg, err
	}

	lldpNeighbors, err := sh.Db.GetLLDPNeighborsOnDevice(sh.FabricID, device.ID)
	//fmt.Println("CRE", lldpNeighbors)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to Fetch LLDP Neighbors for %s", device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	InterfaceConfigs := make([]domain.InterfaceSwitchConfig, 0)
	RemoteInterfaceConfigs := make([]domain.RemoteNeighborSwitchConfig, 0)
	SwitchConfigMap := sh.prepareMapOfSwitchConfigs(ctx)
	//Interface IDs to be used for querying the InterfaceConfigs and BGP Remote Configs
	//for comparison.
	InterfaceIDs := make([]uint, 0)
	MCTInterfaceIDs := make([]uint, 0)
	MCTBGPConfigs := make([]domain.RemoteNeighborSwitchConfig, 0)

	//For each neighbor find the Interface Configs and BGP Neighbor configs
	for _, neighbor := range lldpNeighbors {

		if sh.isMCTNeighbor(neighbor, rack) {
			//MCT Cluster create BGP neighbor of TYPE NSH

			LOG.Infoln("Handle BGP MCT case")
			MCTBGPOneConf, MCTBGPTwoConf, InterfaceOneID, InterrfaceTwoID, err := sh.buildMctConfigForNeighbor(ctx, &neighbor, &FabricProperties, SwitchConfigMap)
			MCTBGPOneConf.Type = domain.MCTBGPType
			MCTBGPTwoConf.Type = domain.MCTBGPType
			if err != nil {
				return "Unable to build MCT BGP configurations", err

			}
			if InterfaceOneID != 0 && InterrfaceTwoID != 0 {
				MCTInterfaceIDs = append(MCTInterfaceIDs, InterfaceOneID)
				MCTInterfaceIDs = append(MCTInterfaceIDs, InterrfaceTwoID)
				if neighbor.ConfigType != domain.ConfigDelete {
					MCTBGPConfigs = append(MCTBGPConfigs, MCTBGPOneConf)
					MCTBGPConfigs = append(MCTBGPConfigs, MCTBGPTwoConf)
				}
			}
		} else if sh.isL3BackupLink(neighbor, rack) {
			if neighbor.ConfigType == domain.ConfigDelete {
				continue
			}
			IntOneConf, IntTwoConf, RemoteIntOneConf, RemoteIntTwoConf, err := sh.buildInterfaceConfigForL3BackUPNeighbor(ctx,
				&neighbor, &FabricProperties, SwitchConfigMap)
			RemoteIntOneConf.Type = domain.MCTL3LBType
			RemoteIntTwoConf.Type = domain.MCTL3LBType
			if err != nil {
				return "Unable to build Interface conf", err
			}
			//Use all the Interfaces for DB fetch
			InterfaceIDs = append(InterfaceIDs, neighbor.InterfaceOneID)
			InterfaceIDs = append(InterfaceIDs, neighbor.InterfaceTwoID)
			InterfaceConfigs = append(InterfaceConfigs, IntOneConf)
			InterfaceConfigs = append(InterfaceConfigs, IntTwoConf)
			RemoteInterfaceConfigs = append(RemoteInterfaceConfigs, RemoteIntOneConf)
			RemoteInterfaceConfigs = append(RemoteInterfaceConfigs, RemoteIntTwoConf)

		} else if sh.isBGPNeighbor(neighbor, rack) {
			if neighbor.ConfigType == domain.ConfigDelete {
				continue
			}
			//pick neighbors only if they are not part of the same rack as bgp neighbor
			if _, err1 := sh.Db.GetRack(sh.FabricName, SwitchConfigMap[neighbor.DeviceOneID].DeviceIP,
				SwitchConfigMap[neighbor.DeviceTwoID].DeviceIP); err1 == nil {
				//devices in the same rack
				continue
			}
			IntOneConf, IntTwoConf, RemoteIntOneConf, RemoteIntTwoConf, err := sh.buildInterfaceConfigForNeighbor(ctx,
				&neighbor, &FabricProperties, SwitchConfigMap)
			RemoteIntOneConf.Type = domain.FabricBGPType
			RemoteIntTwoConf.Type = domain.FabricBGPType
			if err != nil {
				return "Unable to build Interface conf", err
			}
			//Use all the Interfaces for DB fetch
			InterfaceIDs = append(InterfaceIDs, neighbor.InterfaceOneID)
			InterfaceIDs = append(InterfaceIDs, neighbor.InterfaceTwoID)
			InterfaceConfigs = append(InterfaceConfigs, IntOneConf)
			InterfaceConfigs = append(InterfaceConfigs, IntTwoConf)
			RemoteInterfaceConfigs = append(RemoteInterfaceConfigs, RemoteIntOneConf)
			RemoteInterfaceConfigs = append(RemoteInterfaceConfigs, RemoteIntTwoConf)

		}
		if neighbor.ConfigType == domain.ConfigDelete {
			continue
		}
		//Update the neighbor IP Address to DB
		sh.Db.CreateLLDPNeighbor(&neighbor)
	}
	//Persist MCT Remote Interface Configs
	if _, err := sh.persistRemoteInterfaceConfigs(ctx, device, MCTBGPConfigs, MCTInterfaceIDs); err != nil {
		return "Failed to persist Remote Interface Configs", err
	}

	//Persist Interface Configs
	if _, err := sh.persistInterfaceConfigs(ctx, device, InterfaceConfigs, InterfaceIDs, &FabricProperties); err != nil {
		return "Failed to persist Interface Configs", err
	}
	//Persist Remote Interface Configs
	if _, err := sh.persistRemoteInterfaceConfigs(ctx, device, RemoteInterfaceConfigs, InterfaceIDs); err != nil {
		return "Failed to persist Remote Interface Configs", err
	}

	return "", nil
}

func (sh *DeviceInteractor) buildInterfaceConfigForL3BackUPNeighbor(ctx context.Context, neighbor *domain.LLDPNeighbor,
	FabricProperties *domain.FabricProperties, SwitchConfigMap map[uint]domain.SwitchConfig) (domain.InterfaceSwitchConfig, domain.InterfaceSwitchConfig,
	domain.RemoteNeighborSwitchConfig, domain.RemoteNeighborSwitchConfig, error) {

	var err error
	donorType := ""
	donorName := ""
	//For Numbered cases fetch the Interface IP from the Pool
	if FabricProperties.P2PIPType == domain.P2PIpTypeNumbered {
		if neighbor.InterfaceOneIP, neighbor.InterfaceTwoIP, err = sh.reserveOrObtainIPPair(ctx, neighbor.DeviceOneID, neighbor.DeviceTwoID,
			neighbor.InterfaceOneType, neighbor.InterfaceOneName, neighbor.InterfaceTwoType, neighbor.InterfaceTwoName,
			FabricProperties.MCTL3LBIPRange, domain.RackL3LoopBackPoolName, neighbor.InterfaceOneID, neighbor.InterfaceTwoID); err != nil {
			return domain.InterfaceSwitchConfig{}, domain.InterfaceSwitchConfig{},
				domain.RemoteNeighborSwitchConfig{}, domain.RemoteNeighborSwitchConfig{}, err
		}
	}
	//Interface One
	IntOneDescription := fmt.Sprintf("%s to %s L3-bkup-link", SwitchConfigMap[neighbor.DeviceOneID].DeviceIP,
		SwitchConfigMap[neighbor.DeviceTwoID].DeviceIP)
	IntOneConf := sh.prepareInterfaceSwitchConfig(neighbor.DeviceOneID, neighbor.InterfaceOneID,
		neighbor.InterfaceOneIP, neighbor.InterfaceOneName, neighbor.InterfaceOneType,
		donorType, donorName, IntOneDescription)

	remoteAS := sh.fetchASforDevice(ctx, neighbor.DeviceTwoID, SwitchConfigMap[neighbor.DeviceTwoID].DeviceIP)
	RemoteIntOneConf := sh.prepareRemoteInterfaceConfig(neighbor.DeviceOneID, neighbor.InterfaceTwoID,
		neighbor.DeviceTwoID, neighbor.InterfaceTwoIP, remoteAS, "vxlan")

	//Interface Two
	IntTwoDescription := fmt.Sprintf("%s to %s L3-bkup-link", SwitchConfigMap[neighbor.DeviceTwoID].DeviceIP,
		SwitchConfigMap[neighbor.DeviceOneID].DeviceIP)
	IntTwoConf := sh.prepareInterfaceSwitchConfig(neighbor.DeviceTwoID, neighbor.InterfaceTwoID,
		neighbor.InterfaceTwoIP, neighbor.InterfaceTwoName, neighbor.InterfaceTwoType,
		donorType, donorName, IntTwoDescription)

	remoteAS = sh.fetchASforDevice(ctx, neighbor.DeviceOneID, SwitchConfigMap[neighbor.DeviceOneID].DeviceIP)
	RemoteIntTwoConf := sh.prepareRemoteInterfaceConfig(neighbor.DeviceTwoID, neighbor.InterfaceOneID,
		neighbor.DeviceOneID, neighbor.InterfaceOneIP, remoteAS, "vxlan")

	return IntOneConf, IntTwoConf, RemoteIntOneConf, RemoteIntTwoConf, nil
}

func (sh *DeviceInteractor) isMCTNeighbor(neighbor domain.LLDPNeighbor, rack bool) bool {

	if neighbor.InterfaceOneName == MCTInterfaceOne || neighbor.InterfaceOneName == MCTInterfaceTwo || neighbor.InterfaceTwoName == MCTInterfaceOne || neighbor.InterfaceTwoName == MCTInterfaceTwo {
		return true

	}

	return false
}

func (sh *DeviceInteractor) isBGPNeighbor(neighbor domain.LLDPNeighbor, rack bool) bool {

	//If of any Special interfaces then they are not BGP Neighbors
	if neighbor.InterfaceOneName == MCTInterfaceOne || neighbor.InterfaceTwoName == MCTInterfaceTwo || neighbor.InterfaceOneName == MCTInterfaceTwo || neighbor.InterfaceTwoName == MCTInterfaceOne ||
		neighbor.InterfaceOneName == L3LoopBackInterface || neighbor.InterfaceTwoName == L3LoopBackInterface {
		return false
	}

	return true
}

func (sh *DeviceInteractor) isL3BackupLink(neighbor domain.LLDPNeighbor, rack bool) bool {

	//If of any Special interfaces then they are not BGP Neighbors
	if neighbor.InterfaceOneName == L3LoopBackInterface || neighbor.InterfaceTwoName == L3LoopBackInterface {
		return true
	}

	return false
}

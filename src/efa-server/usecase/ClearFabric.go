package usecase

import (
	"bytes"
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/constants"
	Interactor "efa-server/usecase/interactorinterface"
	"errors"
	"fmt"
	"sync"
)

const clearFabricName = "dummy_clear_fabric"

//AddDevicesAndClearFabric does a dummy discovery of Switches and then calls clear fabric to cleanup configurations on the Fabric
func (sh *DeviceInteractor) AddDevicesAndClearFabric(ctx context.Context, DevicesIPList []string, DevicesIPListToClear []string,
	UserName string, Password string) error {
	var buffer bytes.Buffer
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Clear Config")

	ctx = context.WithValue(ctx, appcontext.FabricName, clearFabricName)
	LOG := appcontext.Logger(ctx)
	err := sh.DiscoverDevicesForClear(ctx, DevicesIPList, DevicesIPListToClear, UserName, Password)
	defer sh.deleteFabric()
	//Discovery failed
	if err != nil {
		return err
	}

	clearFabricRequest, mctClusters, err := sh.GenerateFabricConfigsForClear(ctx, DevicesIPListToClear)
	//Configs to be sent for clear
	if err != nil {
		return err
	}
	//Sent Clear for clearing
	if err := sh.clearFabricConfigs(ctx, clearFabricRequest, mctClusters); err != nil {
		return err
	}

	InterfaceMap := make(map[string]bool)
	for _, SwitchIPAddress := range DevicesIPListToClear {
		OldMcts, merr := sh.Db.GetMctClusterConfigWithDeviceIP(SwitchIPAddress)
		if merr != nil {
			//Since we are trying to Clear the devices if they are not present Ignore the error
			statusMsg := fmt.Sprintf("Unable to Retrieve MctCluster from DB for device IP %s", SwitchIPAddress)
			LOG.Infoln(statusMsg, merr)
			continue
		}
		err := sh.Db.DeleteMctClustersWithMgmtIP(SwitchIPAddress)
		if err != nil {
			LOG.Errorln("Error while Removing mct cluster during clear IP Address ", SwitchIPAddress)
			return err
		}

		for _, OldMct := range OldMcts {
			key := fmt.Sprintf("%d : %d", OldMct.VEInterfaceOneID, OldMct.VEInterfaceTwoID)
			key2 := fmt.Sprintf("%d : %d", OldMct.VEInterfaceTwoID, OldMct.VEInterfaceOneID)
			if _, ok := InterfaceMap[key]; ok {
				//Already IP Release to the POOL
				continue
			}
			if _, ok := InterfaceMap[key2]; ok {
				//Already IP Release to the POOL
				continue
			}
			InterfaceMap[key] = true
			InterfaceMap[key2] = true
			err := sh.ReleaseIPPair(sh.AppContext, OldMct.FabricID, OldMct.DeviceID, OldMct.MCTNeighborDeviceID,
				domain.MCTPoolName, OldMct.PeerTwoIP, OldMct.PeerOneIP, OldMct.VEInterfaceOneID, OldMct.VEInterfaceTwoID)
			if err != nil {
				//Since This is Cleanup Any error would rewind database so dont throw error
				LOG.Errorf("Error : Releasing IP FOR %s %s", OldMct.PeerOneIP, OldMct.PeerTwoIP)
				//return err
			}
			err = sh.Db.DeleteInterfaceUsingID(OldMct.FabricID, OldMct.VEInterfaceOneID)
			if err != nil {
				statusMsg := fmt.Sprintf("Error : Deleting VE %s ON Device %d ", OldMct.ControlVE, OldMct.DeviceID)
				LOG.Errorln(statusMsg)
				//return err
			}
			err = sh.Db.DeleteInterfaceUsingID(OldMct.FabricID, OldMct.VEInterfaceTwoID)
			if err != nil {
				statusMsg := fmt.Sprintf("Error : Deleting VE %s ON Device %d ", OldMct.ControlVE, OldMct.MCTNeighborDeviceID)
				LOG.Errorln(statusMsg)
				//return err
			}
		}
	}

	buffer.WriteString(fmt.Sprintln("\nClear Fabric [Success]"))
	return nil

}

//DiscoverDevicesForClear discovers the devices defined in DevicesIPList
func (sh *DeviceInteractor) DiscoverDevicesForClear(ctx context.Context, DevicesIPList []string, DevicesIPListToClear []string,
	UserName string, Password string) error {
	var buffer bytes.Buffer
	var Fabric domain.Fabric
	//Use a dummy Fabric to hold the details for deletion
	Fabric.Name = clearFabricName

	//Setup the logger for clear config
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Clear Config")
	LOG := appcontext.Logger(ctx)

	LOG.Infoln("List of all devices for clear-config", DevicesIPList)
	LOG.Infoln("Clear specific list of devices for clear-config", DevicesIPListToClear)

	//First Create a dummy fabric
	if err := sh.Db.CreateFabric(&Fabric); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s create failed", Fabric.Name)
		return errors.New(statusMsg)
	}
	sh.FabricName = Fabric.Name
	sh.FabricID = Fabric.ID

	var fabricGate sync.WaitGroup
	buffer.WriteString("\n")

	//Concurrent Execution of Adding Spines and Leaves
	ResultChannel := make(chan result, 1)
	var exists bool
	for _, SwitchIPAddress := range DevicesIPList {
		//Check if the IP exists in the user given list of IP address.
		exists = false
		for _, devIP := range DevicesIPListToClear {
			if SwitchIPAddress == devIP {
				exists = true
			}
		}

		fabricGate.Add(1)
		if exists {
			// Use the user given username and password if the IP exists in the user given list to clear/force.
			LOG.Infoln("DiscoverDevicesForClear : Using new credentials for the device ", SwitchIPAddress)
			go sh.addSingleDeviceForClearFabric(ctx, &fabricGate, ResultChannel, &Fabric, SwitchIPAddress, UserName, Password)
		} else {
			// Use the credentials stored in DB. It means the IP address already exists in default fabric.
			// Due to force option, its getting cleared. So use the existing credentials instead of user given.
			dbDevice, err := sh.Db.GetDeviceInAnyFabric(SwitchIPAddress)
			if err == nil {
				LOG.Infoln("DiscoverDevicesForClear : Using existing credentials for the device ", SwitchIPAddress)
				go sh.addSingleDeviceForClearFabric(ctx, &fabricGate, ResultChannel, &Fabric, SwitchIPAddress, dbDevice.UserName, dbDevice.Password)
			} else {
				status := fmt.Sprintf("Could not retrieve credentials from existing device. Error : %s", err.Error())
				return errors.New(status)
			}
		}
	}

	//Wait for the Concurrent execution to complete
	go func() {
		fabricGate.Wait()
		close(ResultChannel)
	}()

	//Get the Responses from the Channel
	AddDeviceFailed := false
	for result := range ResultChannel {
		buffer.WriteString(result.Status)
		if result.Error != nil {
			AddDeviceFailed = true
		}
	}
	if AddDeviceFailed {
		LOG.Errorln("clear-config discovery failed", DevicesIPList, buffer.String())
		return errors.New(buffer.String())
	}
	return nil

}

func (sh *DeviceInteractor) deleteFabric() {
	sh.Db.DeleteFabric(sh.FabricName)
}

func (sh *DeviceInteractor) addSingleDeviceForClearFabric(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan result, Fabric *domain.Fabric, IPAddress string, UserName string, Password string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.addDeviceForClearFabric(ctx, Fabric, IPAddress, UserName, Password)
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Discovery of device with ip-address = %s [Failed]\n", IPAddress))
		buffer.WriteString(err.Error() + "\n")
		ResultChannel <- result{buffer.String(), err}
		return
	}
	buffer.WriteString(fmt.Sprintf("Discovery of device with ip-address = %s [Succeeded]\n", IPAddress))
	ResultChannel <- result{buffer.String(), err}
}

func (sh *DeviceInteractor) addDeviceForClearFabric(ctx context.Context, Fabric *domain.Fabric, IPAddress string, UserName string, Password string) error {
	var Device domain.Device
	var err error
	ctx = context.WithValue(ctx, appcontext.DeviceName, IPAddress)
	LOG := appcontext.Logger(ctx)

	//To achieve rollback, set the boolean to false and return from the function
	RollBack := true
	Device.FabricID = Fabric.ID
	Device.IPAddress = IPAddress

	sh.EvaluateCredentials(UserName, Password, &Device)

	var DeviceAdapter Interactor.DeviceAdapter
	//open Connection to the Switch
	if DeviceAdapter, err = sh.DeviceAdapterFactory(ctx, IPAddress, Device.UserName, Device.Password); err != nil {
		statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", IPAddress, err.Error())
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	defer DeviceAdapter.CloseConnection(ctx)
	LOG.Infoln("Fetch Interfaces")
	//Fetch Switch Assets
	if statusMsg, err := sh.fetchInterfacesAndASN(ctx, &Device, DeviceAdapter); err != nil {
		return errors.New(statusMsg)
	}
	LOG.Infoln("Fetch LLDP")
	if statusMsg, err := sh.fetchLLDPAssets(ctx, &Device, DeviceAdapter); err != nil {
		return errors.New(statusMsg)
	}

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()
	err = sh.Db.OpenTransaction()
	if err != nil {
		return err
	}
	defer sh.CloseTransaction(ctx, &RollBack)

	//Save Device
	if err = sh.Db.CreateDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s create Failed", IPAddress)
		return errors.New(statusMsg)
	}

	//Persist Switch Assets
	if statusMsg, err := sh.persistAssetsForClearFabric(ctx, &Device, DeviceAdapter); err != nil {
		return errors.New(statusMsg)
	}

	//Compute Role & Neighbor Relationships
	if statusMsg, err := sh.prepareLLDPNeighborsForClear(ctx, &Device); err != nil {
		return errors.New(statusMsg)

	}
	//Save Device Again
	if err := sh.Db.SaveDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s Save Failed", IPAddress)
		return errors.New(statusMsg)

	}

	RollBack = false
	return nil

}

//Persist Interface and LLDP assets
func (sh *DeviceInteractor) persistAssetsForClearFabric(ctx context.Context, Device *domain.Device, DeviceAdapter Interactor.DeviceAdapter) (string, error) {
	LOG := appcontext.Logger(ctx)
	Interfaces := Device.Interfaces
	LOG.Infoln("Created Interfaces", Interfaces)
	for _, CInf := range Interfaces {

		CInf.DeviceID = Device.ID
		CInf.ConfigType = domain.ConfigCreate
		CInf.FabricID = sh.FabricID

		if err := sh.Db.CreateInterface(&CInf); err != nil {
			statusMsg := fmt.Sprintf("Failed to create Physical Interface %s %s", CInf.IntType, CInf.IntName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	LLDPS := Device.LLDPS
	LOG.Infoln("Created LLDPS", LLDPS)
	for _, object := range LLDPS {
		object.ConfigType = domain.ConfigCreate
		object.DeviceID = Device.ID
		object.FabricID = sh.FabricID

		if err := sh.Db.CreateLLDP(&object); err != nil {
			statusMsg := fmt.Sprintf("Failed to create LLDP for %s %s", object.LocalIntType, object.LocalIntName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	Device.Interfaces, _ = sh.Db.GetInterfacesonDevice(sh.FabricID, Device.ID)
	return "", nil
}

//Prepare LLDP Neighbors
func (sh *DeviceInteractor) prepareLLDPNeighborsForClear(ctx context.Context, device *domain.Device) (string, error) {
	//For Every Interface MAC Find a Match in LLDP Table
	LOG := appcontext.Logger(ctx)
	deviceMap := make(map[uint]domain.Device, 0)
	var err error
	if deviceMap, err = sh.prepareMapDeviceIDToDevice(ctx); err != nil {
		statusMsg := fmt.Sprintf("Failed to build Device Role Map")
		LOG.Infoln(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}
	for index := range device.Interfaces {
		phy := &device.Interfaces[index]
		lldp, err := sh.Db.GetLLDPOnRemoteMacExcludingMarkedForDeletion(phy.Mac, sh.FabricID)
		if err == nil {

			statusMsg := fmt.Sprintf("Neighbor Found Local Mac %s Remote Mac %s", phy.Mac, lldp.LocalIntMac)
			LOG.Infoln(statusMsg)
			neighborOne, neighborTwo, nerr := sh.prepareNeighborDetails(ctx, phy, &lldp, device.DeviceRole,
				deviceMap[lldp.DeviceID].DeviceRole)
			if nerr != nil {
				return "Unable to build Neighbor", nerr
			}
			neighborOne.ConfigType = domain.ConfigCreate
			if err := sh.Db.CreateLLDPNeighbor(&neighborOne); err != nil {
				statusMsg := fmt.Sprintf("Failed to create LLDP Neighbor %s %s", neighborOne.InterfaceOneType, neighborOne.InterfaceOneName)
				return statusMsg, errors.New(statusMsg)
			}
			neighborOne.ConfigType = domain.ConfigCreate
			if err := sh.Db.CreateLLDPNeighbor(&neighborTwo); err != nil {
				statusMsg := fmt.Sprintf("Failed to create LLDP Neighbor %s %s", neighborTwo.InterfaceOneType, neighborTwo.InterfaceOneName)
				return statusMsg, errors.New(statusMsg)
			}
		}
	}
	return "Prepare LLDP Neighbors Success", nil
}

//GenerateFabricConfigsForClear generates the switch config to be cleared
func (sh *DeviceInteractor) GenerateFabricConfigsForClear(ctx context.Context, DevicesIPListToClear []string) (operation.ClearFabricRequest,
	[]operation.ConfigCluster, error) {
	LOG := appcontext.Logger(ctx)
	LOG.Infoln("Prepare Configurations for", DevicesIPListToClear)
	ClearFabricRequest, err := sh.prepareActionClearFabricRequest(ctx, DevicesIPListToClear)
	LOG.Infoln("Clear Fabric Request", ClearFabricRequest)
	ClearClusters, cerr := sh.prepareMctClearRequest(ctx, DevicesIPListToClear)
	LOG.Infoln("Clear MCT Clusters", ClearClusters)
	if err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", sh.FabricName)
		LOG.Errorln(statusMsg)
		return ClearFabricRequest, ClearClusters, errors.New(statusMsg)
	}

	if cerr != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", sh.FabricName)
		LOG.Errorln(statusMsg)
		return ClearFabricRequest, ClearClusters, errors.New(statusMsg)
	}

	return ClearFabricRequest, ClearClusters, nil
}
func (sh *DeviceInteractor) clearFabricConfigs(ctx context.Context, ClearFabricRequest operation.ClearFabricRequest,
	ClearClusters []operation.ConfigCluster) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infoln("Invoke Clear MCT Cluster")
	if err := sh.FabricAdapter.ClearMctClusters(ctx, sh.FabricName, ClearClusters); err != nil {
		return err
	}

	LOG.Infoln("Invoke Clear Fabric")
	//Clear Fabric Information from Switches
	err := sh.FabricAdapter.ClearConfig(ctx, ClearFabricRequest)

	return err
}
func (sh *DeviceInteractor) filterDevicesOnIPList(devices []domain.Device, DevicesIPListToClear []string) []domain.Device {
	DeviceMap := make(map[string]bool, 0)
	for _, devIP := range DevicesIPListToClear {
		DeviceMap[devIP] = true
	}

	Devices := make([]domain.Device, 0, 0)
	for _, dev := range devices {
		if ok, _ := DeviceMap[dev.IPAddress]; ok == true {
			Devices = append(Devices, dev)
		}
	}

	return Devices
}
func (sh *DeviceInteractor) prepareActionClearFabricRequest(ctx context.Context, DevicesIPListToClear []string) (operation.ClearFabricRequest, error) {
	LOG := appcontext.Logger(ctx)
	resp := operation.ClearFabricRequest{}
	resp.FabricName = sh.FabricName
	devices, err := sh.Db.GetDevicesInFabric(sh.FabricID)

	if err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch devices from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return resp, errors.New(statusMsg)
	}
	//filter devices on the base of IP addresss
	devices = sh.filterDevicesOnIPList(devices, DevicesIPListToClear)
	LOG.Infoln("Filtered List of Devices to be cleared", devices)

	DefaultFabriProperties, _ := sh.getDefaultFabricSettings()

	resp.Hosts = make([]operation.ClearSwitchDetail, 0, len(devices))
	for _, dev := range devices {
		host := operation.ClearSwitchDetail{Host: dev.IPAddress, UserName: dev.UserName, Password: dev.Password,
			Role: dev.DeviceRole, Model: dev.Model, LoopBackIPRange: DefaultFabriProperties.LoopBackIPRange}
		lldpNeighbors, _ := sh.Db.GetLLDPNeighborsOnDevice(sh.FabricID, dev.ID)
		host.Interfaces = make([]operation.ClearInterfaceDetail, 0, len(lldpNeighbors))
		for _, lldpNeighbor := range lldpNeighbors {
			//Only if the IP address is non-empty set if for clear
			//TODO for Donor
			if lldpNeighbor.InterfaceOneIP != "" {
				Interface := operation.ClearInterfaceDetail{InterfaceName: lldpNeighbor.InterfaceOneName,
					InterfaceType: lldpNeighbor.InterfaceOneType, IP: lldpNeighbor.InterfaceOneIP + "/31"}
				host.Interfaces = append(host.Interfaces, Interface)
			} else {
				//Set the interface
				Interface := operation.ClearInterfaceDetail{InterfaceName: lldpNeighbor.InterfaceOneName,
					InterfaceType: lldpNeighbor.InterfaceOneType}
				host.Interfaces = append(host.Interfaces, Interface)
			}
		}
		//Add Loopback Interfaces same as the one in the Default fabric
		Interface := operation.ClearInterfaceDetail{InterfaceName: DefaultFabriProperties.LoopBackPortNumber,
			InterfaceType: domain.IntfTypeLoopback, IP: ""}
		host.Interfaces = append(host.Interfaces, Interface)
		VTEPInterface := operation.ClearInterfaceDetail{InterfaceName: DefaultFabriProperties.VTEPLoopBackPortNumber,
			InterfaceType: domain.IntfTypeLoopback, IP: ""}
		host.Interfaces = append(host.Interfaces, VTEPInterface)
		resp.Hosts = append(resp.Hosts, host)

	}
	return resp, err
}

func (sh *DeviceInteractor) getDefaultFabricSettings() (domain.FabricProperties, error) {
	var FabricProperties domain.FabricProperties
	var Fabric domain.Fabric
	var err error

	if Fabric, err = sh.Db.GetFabric(constants.DefaultFabric); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", constants.DefaultFabric)
		return FabricProperties, errors.New(statusMsg)
	}

	return sh.Db.GetFabricProperties(Fabric.ID)
	/*
		//Run's with Default properties for clear-config
		return sh.getDefaultProperties(), nil*/

}

func (sh *DeviceInteractor) prepareMctClearRequest(ctx context.Context, DevicesIPListToClear []string) ([]operation.ConfigCluster, error) {
	LOG := appcontext.Logger(ctx)
	Clusters := make([]operation.ConfigCluster, 0)

	DefaultFabriProperties, _ := sh.getDefaultFabricSettings()
	devices, err := sh.Db.GetDevicesInFabric(sh.FabricID)
	switchMap := make(map[string]domain.Device)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch devices from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return Clusters, errors.New(statusMsg)
	}
	devices = sh.filterDevicesOnIPList(devices, DevicesIPListToClear)
	LOG.Infoln("Filtered List of Devices to be cleared", devices)
	for _, device := range devices {
		switchMap[device.IPAddress] = device
	}

	for _, device := range devices {

		clusters, err := sh.Db.GetMctClusterConfigWithDeviceIP(device.IPAddress)
		if err != nil {
			statusMsg := fmt.Sprintf("Unable to Retrieve MCT Cluster Info from DB for device %s", device.IPAddress)
			LOG.Errorln(statusMsg)
			return Clusters, err
		}
		PreapareConfigCluster := func(dev domain.Device) {
			var mct operation.ConfigCluster
			var mctMember1 operation.ClusterMemberNode
			mct.FabricName = sh.FabricName
			mctMember1.NodeMgmtIP = dev.IPAddress
			mctMember1.NodeModel = dev.Model
			mctMember1.NodeMgmtUserName = dev.UserName
			mctMember1.NodeMgmtPassword = dev.Password
			mctMember1.NodePeerIntfType = "Port-channel"
			mctMember1.NodePeerIntfName = DefaultFabriProperties.MctPortChannel
			if sh.FabricAdapter.IsRoutingDevice(ctx, dev.Model) {
				mctMember1.NodePeerIntfName = DefaultFabriProperties.RoutingMctPortChannel
			} else {
				mctMember1.NodePeerIntfName = DefaultFabriProperties.MctPortChannel
			}
			mct.ClusterMemberNodes = append(mct.ClusterMemberNodes, mctMember1)
			mct.ClusterControlVlan = DefaultFabriProperties.ControlVlan
			mct.ClusterControlVe = DefaultFabriProperties.ControlVE
			mct.OperationBitMap = 0
			Clusters = append(Clusters, mct)
		}
		PreapareConfigCluster(device)
		for _, cluster := range clusters {
			if _, ok := switchMap[cluster.DeviceTwoMgmtIP]; ok == false {
				//Neighbor of the cluster in not present in Clear list
				//Hence Clear Neighbor Device as well
				dev, err := sh.Db.GetDeviceUsingDeviceID(cluster.FabricID, cluster.MCTNeighborDeviceID)
				if err != nil {
					statusMsg := fmt.Sprintf("Unable to Retrieve Device Info from DB for Device %s", cluster.DeviceTwoMgmtIP)
					LOG.Errorln(statusMsg)
					return Clusters, err
				}
				PreapareConfigCluster(dev)
				switchMap[cluster.DeviceTwoMgmtIP] = dev

			}
		}
	}
	return Clusters, err
}

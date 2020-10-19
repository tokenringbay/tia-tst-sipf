package usecase

import (
	"bytes"
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	Interactor "efa-server/usecase/interactorinterface"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
	"strings"
	"sync"
)

/*DeleteDevicesFromFabric performs:
	Step 1 : Validation - if all devices are part of the fabric.
   	Step 2 : Cleanup the configurations from the switch.
   	Step 3 : Delete the devices from fabric.
*/
func (sh *DeviceInteractor) DeleteDevicesFromFabric(ctx context.Context, FabricName string, DevicesList []string,
	UserName string, Password string, force bool, persist bool, devCleanUp bool) ([]AddDeviceResponse, error) {
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Delete Device")
	LOG := appcontext.Logger(ctx)

	var fabricGate sync.WaitGroup

	AddDeviceResponseList := make([]AddDeviceResponse, 0, len(DevicesList))
	AddDeviceResponseListError := make([]AddDeviceResponse, 0, len(DevicesList))
	FabricProperties, err := sh.GetFabricSettings(ctx, FabricName)
	if err != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch fabric Properties  %s", sh.FabricName)
		LOG.Infoln(statusMsg)
		return AddDeviceResponseList, err
	}
	sh.FabricProperties = FabricProperties
	ResultChannel := make(chan AddDeviceResponse, len(DevicesList))
	// Concurrent validation of Devices
	for _, SwitchIPAddress := range DevicesList {
		fabricGate.Add(1)
		go sh.validateSingleDevice(ctx, &fabricGate, ResultChannel, FabricName, SwitchIPAddress, UserName, Password, devCleanUp)
	}

	//Wait for the Concurrent execution to complete
	go func() {
		fabricGate.Wait()
		close(ResultChannel)
	}()

	//Get the Responses from the Channel
	ValidateDeviceFailed := false
	for result := range ResultChannel {
		if len(result.Errors) > 0 {
			ValidateDeviceFailed = true
			AddDeviceResponseListError = append(AddDeviceResponseListError, result)
		}
		AddDeviceResponseList = append(AddDeviceResponseList, result)
	}
	if ValidateDeviceFailed {
		return AddDeviceResponseListError, errors.New("Device(s) Validation [Failed]")
	}

	// Clean-up the devices only if device cleanup is provided
	//TODO Handle Error From below functions till then print errors on console
	if devCleanUp {
		Errors := sh.cleanupDevicesInFabric(ctx, FabricName, DevicesList, force, persist)
		if len(Errors) != 0 {
			for _, err := range Errors {
				var Response AddDeviceResponse
				Response.FabricName = FabricName
				Response.FabricID = sh.FabricID
				Response.IPAddress = err.Host
				Response.Errors = []error{err.Error}
				AddDeviceResponseListError = append(AddDeviceResponseListError, Response)
			}
			return AddDeviceResponseListError, errors.New("Device cleanup Failed")
		}

		SecondaryCleanupErrors := sh.cleanupRelatedConfigsonDependentSwitches(ctx, DevicesList, force, persist)
		if len(SecondaryCleanupErrors) != 0 {
			for _, err := range SecondaryCleanupErrors {
				var Response AddDeviceResponse
				Response.FabricName = FabricName
				Response.FabricID = sh.FabricID
				Response.IPAddress = err.Host
				Response.Errors = []error{err.Error}
				AddDeviceResponseListError = append(AddDeviceResponseListError, Response)
			}
			return AddDeviceResponseListError, errors.New("Device cleanup Failed")
		}
	}

	Error := sh.DeleteDevices(ctx, FabricName, DevicesList, true)
	if Error.Error != nil {
		var Response AddDeviceResponse
		Response.FabricName = FabricName
		Response.FabricID = sh.FabricID
		Response.IPAddress = Error.Host
		Response.Errors = []error{Error.Error}
		AddDeviceResponseListError = append(AddDeviceResponseListError, Response)
		return AddDeviceResponseListError, errors.New("Delete Device Failed")
	}

	if err := sh.Db.Backup(); err != nil {
		LOG.Printf("Failed to backup DB during Deconfigure %s\n", err)
	}
	return AddDeviceResponseList, nil
}

func (sh *DeviceInteractor) validateSingleDevice(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse,
	FabricName string, IPAddress string, UserName string, Password string, devCleanup bool) {
	defer fabricGate.Done()

	var buffer bytes.Buffer
	err := sh.validateExistingDevice(ctx, FabricName, IPAddress, UserName, Password, devCleanup)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: ""}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Validation of device = %s [Failed]\n", IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Validation of device = %s [Succeeded]\n", IPAddress))
	ResultChannel <- Response
}

//This validates if switch exists in db and part of fabric and rechable
func (sh *DeviceInteractor) validateExistingDevice(ctx context.Context, FabricName string,
	IPAddress string, UserName string, Password string, devCleanup bool) error {
	var Fabric domain.Fabric
	var err error
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Validating Device %s for the fabric %s.\n", IPAddress, FabricName)
	if Fabric, err = sh.Db.GetFabric(FabricName); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", FabricName)
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	//check for existing Device
	var Device domain.Device
	if Device, err = sh.Db.GetDevice(FabricName, IPAddress); err != nil {
		statusMsg := fmt.Sprintf("Switch %s doesn't exist in the fabric", IPAddress)
		LOG.Infoln(statusMsg)
		return errors.New(statusMsg)
		//sh.Refresh = true
	}
	sh.EvaluateCredentials(UserName, Password, &Device)

	Device.IPAddress = IPAddress
	Device.FabricID = Fabric.ID

	if devCleanup {
		var DeviceAdapter Interactor.DeviceAdapter
		//Open Connection to the Switch
		if DeviceAdapter, err = sh.DeviceAdapterFactory(ctx, IPAddress, Device.UserName, Device.Password); err != nil {
			statusMsg := fmt.Sprintf("Switch %s connection Failed : %s", IPAddress, err.Error())
			LOG.Errorln(statusMsg)
			return errors.New(statusMsg)
		}
		defer DeviceAdapter.CloseConnection(ctx)
	}

	//Save Device Again(In case Device password has been changed by user)
	if err := sh.Db.SaveDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s Save Failed", IPAddress)
		return errors.New(statusMsg)
	}

	return nil
}

func (sh *DeviceInteractor) cleanupDevicesInFabric(ctx context.Context, FabricName string, DevicesList []string,
	force bool, persist bool) []actions.OperationError {
	LOG := appcontext.Logger(ctx)
	var config operation.ConfigFabricRequest
	var err error

	Errors := make([]actions.OperationError, 0)
	config.FabricName = FabricName
	config.FabricSettings, err = sh.GetFabricSettings(ctx, FabricName)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to perform Query Topology %s", FabricName)
		LOG.Println(statusMsg)
		for _, device := range DevicesList {
			Errors = append(Errors, actions.OperationError{Operation: "Query Fabric setting", Error: errors.New(statusMsg), Host: device})
		}
		return Errors
	}

	err = sh.getHostsSelective(ctx, &config, DevicesList)
	if err != nil {
		LOG.Println(err)
		for _, device := range DevicesList {
			Errors = append(Errors, actions.OperationError{Operation: "Query Fabric setting", Error: err, Host: device})
		}
		return Errors
	}
	//fmt.Println(config)

	Errors = sh.FabricAdapter.CleanupDevicesInFabric(ctx, config, force, persist)

	return Errors
}

//DeleteDevices deletes a device from the Fabric
func (sh *DeviceInteractor) DeleteDevices(ctx context.Context, FabricName string, DevicesList []string, DeviceCleanupRequired bool) actions.OperationError {

	var Error actions.OperationError

	LOG := appcontext.Logger(ctx)

	//To achieve rollback, set the boolean to false and return from the function
	RollBack := true

	LOG.Infof("Deleting Devices %s from the fabric %s.\n", DevicesList, FabricName)

	DeviceIDList := make([]uint, 0)
	DeviceList := make([]domain.Device, 0)
	for _, switchip := range DevicesList {
		var Device domain.Device
		Device, err := sh.Db.GetDevice(FabricName, switchip)
		if err == nil {
			DeviceIDList = append(DeviceIDList, Device.ID)
			DeviceList = append(DeviceList, Device)
			err = sh.freeAlloctedDeviceResource(ctx, Device)
			if err != nil {
				LOG.Infof("Failed to free up resource device %s, fabric %s.\n", switchip, FabricName)
				return actions.OperationError{Operation: "Free Allocated resource", Error: err, Host: switchip}
			}
		} else {
			LOG.Infof("Switch %s doesn't exist in the fabric", switchip)
			return actions.OperationError{Operation: "GetDevice", Error: err, Host: switchip}
		}
	}

	Error = sh.freeMctResources(ctx, DevicesList)
	if Error.Error != nil {
		LOG.Infof("Failed To Release MCT Resources %s ", Error.Error)
		return Error
	}

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()
	err := sh.Db.OpenTransaction()
	if err != nil {
		return actions.OperationError{Operation: "DB Open Transaction", Error: err, Host: ""}
	}
	defer sh.CloseTransaction(ctx, &RollBack)

	if DeviceCleanupRequired {
		if err = sh.Db.DeleteDevice(DeviceIDList); err != nil {
			statusMsg := fmt.Sprint("Switch delete Failed on", DeviceIDList)
			return actions.OperationError{Operation: "DB Open Transaction", Error: errors.New(statusMsg), Host: ""}
		}
	} else {
		//Delete the device so that it results in cascade delete of other database tables
		if err = sh.Db.DeleteDevice(DeviceIDList); err != nil {
			statusMsg := fmt.Sprint("Switch delete Failed on", DeviceIDList)
			return actions.OperationError{Operation: "DB Open Transaction", Error: errors.New(statusMsg), Host: ""}
		}
		//Re-Create the devices in the database

		for _, dev := range DeviceList {
			//Only copy UserID,Password and IP Address
			device := domain.Device{IPAddress: dev.IPAddress, UserName: dev.UserName, Password: dev.Password, FabricID: dev.FabricID}
			if err = sh.Db.CreateDevice(&device); err != nil {
				statusMsg := fmt.Sprint("Switch Create Failed for", device)
				return actions.OperationError{Operation: "DB Open Transaction", Error: errors.New(statusMsg), Host: dev.IPAddress}
			}
		}
	}

	RollBack = false

	return Error
}

//FlushDevices deletes a device from the Fabric
func (sh *DeviceInteractor) FlushDevices(ctx context.Context, FabricName string, DevicesList []string) actions.OperationError {

	var Error actions.OperationError

	LOG := appcontext.Logger(ctx)

	LOG.Infof("Deleting Devices %s from the fabric %s.\n", DevicesList, FabricName)

	DeviceIDList := make([]uint, 0)
	DeviceList := make([]domain.Device, 0)
	for _, switchip := range DevicesList {
		var Device domain.Device
		Device, err := sh.Db.GetDevice(FabricName, switchip)
		if err == nil {
			DeviceIDList = append(DeviceIDList, Device.ID)
			DeviceList = append(DeviceList, Device)
			err = sh.freeAlloctedDeviceResource(ctx, Device)
			if err != nil {
				LOG.Infof("Failed to free up resource device %s, fabric %s.\n", switchip, FabricName)
				return actions.OperationError{Operation: "Free Allocated resource", Error: err, Host: switchip}
			}
		} else {
			LOG.Infof("Switch %s doesn't exist in the fabric", switchip)
			return actions.OperationError{Operation: "GetDevice", Error: err, Host: switchip}
		}
	}

	Error = sh.freeMctResources(ctx, DevicesList)
	if Error.Error != nil {
		LOG.Infof("Failed To Release MCT Resources %s ", Error.Error)
		return Error
	}

	//Delete the device so that it results in cascade delete of other database tables
	if err := sh.Db.DeleteDevice(DeviceIDList); err != nil {
		statusMsg := fmt.Sprint("Switch delete Failed on", DeviceIDList)
		return actions.OperationError{Operation: "DB Open Transaction", Error: errors.New(statusMsg), Host: ""}
	}
	//Re-Create the devices in the database

	for _, dev := range DeviceList {
		//Reset the Device ID, so that is newly created
		dev.ID = 0
		if err := sh.Db.CreateDevice(&dev); err != nil {
			statusMsg := fmt.Sprint("Switch Create Failed for", dev)
			return actions.OperationError{Operation: "DB Open Transaction", Error: errors.New(statusMsg), Host: dev.IPAddress}
		}
	}

	return Error
}
func (sh *DeviceInteractor) freeAlloctedDeviceResource(ctx context.Context, Device domain.Device) error {
	var switchConfig domain.SwitchConfig
	LOG := appcontext.Logger(ctx)
	switchConfig, err := sh.Db.GetSwitchConfigOnFabricIDAndDeviceID(sh.FabricID, Device.ID)
	if err == gorm.ErrRecordNotFound {
		//If not records are found, there are no SwitchConfigs to clear
		return nil
	}
	if err != nil {
		LOG.Infof("Failed to get switchConfig for device %s\n", Device.IPAddress)
		return err
	}
	// Release ASN
	if switchConfig.LocalAS != "" {
		asn, err := strconv.ParseUint(switchConfig.LocalAS, 10, 64)
		err = sh.ReleaseASN(ctx, switchConfig.FabricID, switchConfig.DeviceID, switchConfig.Role, asn)
		if err != nil {
			LOG.Infof("Failed to Release ASN %s for Device %s\n", switchConfig.LocalAS, Device.IPAddress)
			return err
		}
		LOG.Infof("Released ASN %s for Device %s\n", switchConfig.LocalAS, Device.IPAddress)
	}
	FabricProperties, _ := sh.Db.GetFabricProperties(switchConfig.FabricID)
	// Release Loopback IP
	if switchConfig.LoopbackIP != "" {
		intf, err := sh.Db.GetInterface(switchConfig.FabricID, switchConfig.DeviceID, domain.IntfTypeLoopback, FabricProperties.LoopBackPortNumber)
		if err != nil {
			LOG.Infof("Failed to get loopback interface for Device %s\n", Device.IPAddress)
			return err
		}
		err = sh.ReleaseIP(ctx, switchConfig.FabricID, switchConfig.DeviceID, "Loopback",
			switchConfig.LoopbackIP, intf.ID)
		if err != nil {
			LOG.Infof("Failed to Release Loopback IP %s for Device %s\n", switchConfig.LoopbackIP, Device.IPAddress)
			return err
		}
		LOG.Infof("Released Loopback IP %s for Device %s\n", switchConfig.LoopbackIP, Device.IPAddress)
	}
	// Release VTEP Loopback IP
	if (switchConfig.Role == LeafRole || switchConfig.Role == RackRole) && switchConfig.VTEPLoopbackIP != "" {
		intf, err := sh.Db.GetInterface(switchConfig.FabricID, switchConfig.DeviceID, domain.IntfTypeLoopback, FabricProperties.VTEPLoopBackPortNumber)
		if err != nil {
			LOG.Infof("Failed to get loopback interface for Device %s\n", Device.IPAddress)
			return err
		}
		err = sh.ReleaseIP(ctx, switchConfig.FabricID, switchConfig.DeviceID, "Loopback",
			switchConfig.VTEPLoopbackIP, intf.ID)
		if err != nil {
			LOG.Infof("Failed to Release Loopback IP %s for Device %s\n", switchConfig.VTEPLoopbackIP, Device.IPAddress)
			return err
		}
		LOG.Infof("Released Loopback IP %s for Device %s\n", switchConfig.VTEPLoopbackIP, Device.IPAddress)
	}
	if FabricProperties.P2PIPType == domain.P2PIpTypeNumbered {
		LLDPNeighbors, _ := sh.Db.GetLLDPNeighborsOnDevice(sh.FabricID, switchConfig.DeviceID)

		for _, neighbor := range LLDPNeighbors {
			if strings.Contains(neighbor.InterfaceOneType, domain.IntfTypeEthernet) {
				if neighbor.InterfaceOneName == L3LoopBackInterface || neighbor.InterfaceTwoName == L3LoopBackInterface {
					err = sh.ReleaseIPPair(ctx, sh.FabricID, neighbor.DeviceOneID, neighbor.DeviceTwoID, domain.RackL3LoopBackPoolName,
						neighbor.InterfaceOneIP, neighbor.InterfaceTwoIP, neighbor.InterfaceOneID, neighbor.InterfaceTwoID)
					if err != nil {
						LOG.Infof("Failed to release ip:%s, device:%d", neighbor.InterfaceOneIP, switchConfig.DeviceID)
					}
				} else if neighbor.InterfaceOneIP != "" && neighbor.InterfaceOneIP != "unnumbered" {
					err = sh.ReleaseIPPair(ctx, sh.FabricID, neighbor.DeviceOneID, neighbor.DeviceTwoID, "P2P",
						neighbor.InterfaceOneIP, neighbor.InterfaceTwoIP, neighbor.InterfaceOneID, neighbor.InterfaceTwoID)
					if err != nil {
						LOG.Infof("Failed to release ip:%s, device:%d", neighbor.InterfaceOneIP, switchConfig.DeviceID)
					}
				}

			}
		}
	}

	return nil
}

func (sh *DeviceInteractor) freeMctResources(ctx context.Context, SwitchIPAddressList []string) actions.OperationError {
	LOG := appcontext.Logger(ctx)
	var Error actions.OperationError
	InterfaceMap := make(map[string]bool)
	for _, SwitchIPAddress := range SwitchIPAddressList {

		OldMcts, merr := sh.Db.GetMctClusterConfigWithDeviceIP(SwitchIPAddress)
		if merr != nil {
			statusMsg := fmt.Sprintf("Unable to Retrieve MctCluster from DB for device IP %s", SwitchIPAddress)
			LOG.Infoln(statusMsg, merr)
			return actions.OperationError{Operation: "GetMctClusterConfigWithDeviceIP", Error: merr, Host: SwitchIPAddress}
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
				//Let us Ignore The ERROR For Now
				//TODO We have Handle ReleaseIPPair in a way it is in sync with interface switch config
				//Error is being ignored by other ReleaseIPPair users in such scenarios
				LOG.Infof("Releasing IP FOR %s %s", OldMct.PeerOneIP, OldMct.PeerTwoIP)
			}
			err = sh.Db.DeleteInterfaceUsingID(OldMct.FabricID, OldMct.VEInterfaceOneID)
			if err != nil {
				statusMsg := fmt.Sprintf("Error : Deleting VE %s ON Device %d ", OldMct.ControlVE, OldMct.DeviceID)
				return actions.OperationError{Operation: "Delete MCT Interface using ID", Error: errors.New(statusMsg), Host: SwitchIPAddress}
			}
			err = sh.Db.DeleteInterfaceUsingID(OldMct.FabricID, OldMct.VEInterfaceTwoID)
			if err != nil {
				statusMsg := fmt.Sprintf("Error : Deleting VE %s ON Device %d ", OldMct.ControlVE, OldMct.MCTNeighborDeviceID)
				return actions.OperationError{Operation: "Delete MCT Interface using ID", Error: errors.New(statusMsg), Host: SwitchIPAddress}
			}
		}
	}
	return Error
}

// Create switch map for specific devices present in DeviceList
func (sh *DeviceInteractor) getHostsSelective(ctx context.Context, config *operation.ConfigFabricRequest, DevicesIPAddresses []string) error {
	LOG := appcontext.Logger(ctx)
	config.Hosts = make([]operation.ConfigSwitch, 0)
	var switchConfigs []domain.SwitchConfig
	switchMap := make(map[uint]domain.Device)
	switchConfigMap := make(map[uint]domain.SwitchConfig)
	for _, ipaddress := range DevicesIPAddresses {
		switchConfig, err := sh.Db.GetSwitchConfigOnDeviceIP(sh.FabricName, ipaddress)
		if err != nil {
			statusMsg := fmt.Sprintf("Failed to fetch configs for device %s from %s", ipaddress, sh.FabricName)
			LOG.Errorln(statusMsg)
			return errors.New(statusMsg)
		}
		switchConfigs = append(switchConfigs, switchConfig)
		switchConfigMap[switchConfig.DeviceID] = switchConfig
	}

	for _, ipaddress := range DevicesIPAddresses {
		var dev domain.Device
		var host operation.ConfigSwitch

		dev, err := sh.Db.GetDevice(sh.FabricName, ipaddress)
		switchMap[dev.ID] = dev
		if err != nil {
			statusMsg := fmt.Sprintf("Failed to fetch Device %s from %s", ipaddress, sh.FabricName)
			LOG.Errorln(statusMsg)
			return errors.New(statusMsg)
		}

		sh.populateHostDetails(ctx, host, dev, config, switchConfigMap, domain.ConfigDelete)
	}
	Oper := domain.MctDelete
	Clusters, merr := sh.prepareMctClusters(ctx, switchConfigs, config, switchMap, switchConfigMap, Oper)
	if merr != nil {
		statusMsg := fmt.Sprintf("Failed to fetch Device MCT Configs from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	config.MctCluster = Clusters
	return nil
}

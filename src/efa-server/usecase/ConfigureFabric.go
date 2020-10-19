package usecase

import (
	"bytes"
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"efa-server/infra/util"
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"strconv"
	"strings"
	"sync"
)

type result struct {
	Status string
	Error  error
}

//AddDeviceResponse describes the Response when device is added to the fabric
type AddDeviceResponse struct {
	FabricName string
	FabricID   uint
	//IP Address of the Device
	IPAddress string
	Role      string
	//list of Errors encountered on Adding the device
	Errors []error
}

//ValidateFabricResponse is a response object which defines the success/error of the validation operation
//w.r.t. CLOS IP Fabric topology
type ValidateFabricResponse struct {
	FabricName      string
	FabricID        uint
	NoSpines        bool
	NoLeaves        bool
	MissingLinks    []string
	SpineSpineLinks []string
	LeafLeafLinks   []string
}

//ConfigureFabricResponse is a response object which defines the success/error of "configure fabric" operation
//CLOS IP Fabric topology
type ConfigureFabricResponse struct {
	FabricName string
	FabricID   uint
	Errors     []actions.OperationError
}

type stageFunction func(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse,
	FabricName string, IPAddress string, UserName string, Password string, Role string)

// CreateDevice either updates/creates device with given IPaddress, role, credentials and fabricID
func (sh *DeviceInteractor) CreateDevice(FabricName string, IPAddress string, UserName string, Password string, Role string) (id uint, err error) {
	err = nil
	//check for existing Device
	var Device domain.Device
	if Device, err = sh.Db.GetDevice(FabricName, IPAddress); err == nil {
		sh.Refresh = true
	}

	Device.IPAddress = IPAddress
	Device.DeviceRole = Role
	sh.EvaluateCredentials(UserName, Password, &Device)
	Device.FabricID = sh.FabricID

	//Save Device
	if err = sh.Db.CreateDevice(&Device); err != nil {
		statusMsg := fmt.Sprintf("Switch %s create Failed with error : %s", IPAddress, err.Error())
		err = errors.New(statusMsg)
	}
	return Device.ID, err
}

//AddDevices adds multiple devices (spines and leaff) to the Fabric
func (sh *DeviceInteractor) AddDevices(ctx context.Context, FabricName string, LeafIPaddressList []string,
	SpineIPaddressList []string, UserName string, Password string, force bool) (AddDeviceResponseList []AddDeviceResponse, err error) {

	//Setup the logger
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Add Device")
	ctx = context.WithValue(ctx, appcontext.FabricName, FabricName)

	LOG := appcontext.Logger(ctx)
	ipaddress := append(LeafIPaddressList, SpineIPaddressList...)
	//Fetch the existing devices already registered
	existingSpineList, existingLeafList, err := sh.fetchRegisteredDevices(ctx, FabricName)

	//If force is enabled clear up the configuration on devices specified in the IP address
	existingIPaddress := append(existingLeafList, existingSpineList...)
	LOG.Infoln("Existing Device List", existingIPaddress)
	if force {
		LOG.Infoln("Force option enabled on Devices", ipaddress)

		AddDeviceResponseList, err = sh.forceClearofDevices(ctx, FabricName, existingIPaddress, ipaddress, UserName, Password)
		if err != nil {
			return
		}
	}

	AddDeviceResponseList, err = sh.fetchFabricDetails(ctx, FabricName, ipaddress)
	if err != nil {
		return
	}
	LOG.Infoln("Already registered Spine Devices", existingSpineList)
	LOG.Infoln("Already registered Leaf Devices", existingLeafList)

	AddDeviceResponseList, err = sh.deviceAlreadyRegisteredWithDifferentRole(LeafIPaddressList, SpineIPaddressList, existingLeafList, existingSpineList)
	if err != nil {
		return
	}

	//Combined list to be sent for configuring
	totalLeafList := sh.getUniqueList(ctx, existingLeafList, LeafIPaddressList)
	totalSpineList := sh.getUniqueList(ctx, existingSpineList, SpineIPaddressList)
	LOG.Infoln("Full list of Spine Devices", totalSpineList)
	LOG.Infoln("Full list of Leaf Devices", totalLeafList)

	if len(totalLeafList) == 0 && len(totalSpineList) == 0 {
		statusMsg := "No devices are in the fabric"
		err = errors.New(statusMsg)
		LOG.Errorln(err)
		deviceResponse := AddDeviceResponse{IPAddress: "", Errors: []error{err}}
		AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		return
	}

	stageFunctions := []stageFunction{sh.addSingleDeviceFirstStage,
		sh.addSingleDeviceSecondStage, sh.addSingleDeviceThirdStage, sh.addSingleDeviceFourthStage}

	//To achieve rollback, set the boolean to false and return from the function
	RollBack := false

	dberr := sh.Db.OpenTransaction()
	if dberr != nil {
		err = errors.New("Failed to Open Transaction")
		return
	}
	defer sh.CloseTransaction(ctx, &RollBack)

	// Create or update device with correct credentials.
	for _, IPAddress := range SpineIPaddressList {
		LOG.Infoln("Create Spine Device : ", IPAddress)
		if _, err = sh.CreateDevice(FabricName, IPAddress, UserName, Password, SpineRole); err != nil {
			LOG.Errorln(err.Error())
			return
		}
	}

	for _, IPAddress := range LeafIPaddressList {
		LOG.Infoln("Create Leaf Device", IPAddress)
		if _, err = sh.CreateDevice(FabricName, IPAddress, UserName, Password, LeafRole); err != nil {
			LOG.Errorln(err.Error())
			return
		}
	}

	if force {
		//Flush the configuration of the existing devices and re-build the same as part of the ongoing discovery
		//This operation is done in the same transaction so that discovery failure would roll back the devices.
		Error := sh.FlushDevices(ctx, FabricName, existingIPaddress)
		if Error.Error != nil {
			var Response AddDeviceResponse
			Response.FabricName = FabricName
			Response.FabricID = sh.FabricID
			Response.IPAddress = Error.Host
			Response.Errors = []error{Error.Error}
			AddDeviceResponseList = append(AddDeviceResponseList, Response)
			return AddDeviceResponseList, errors.New("Force Clear failed to clear DB")
		}
	}

	// Now reset credentials to empty so that stageFunctions ignore it and use the one in Device table.
	// TODO : UserName and Password should not be passed to stageFunctions. Not doing it now as its risky during EOR.
	UserName = ""
	Password = ""
	for index, stageFunction := range stageFunctions {
		LOG.Println("Executing Stage ", index+1)
		AddDeviceResponseList, err = sh.executeAddDeviceStage(ctx, FabricName, totalLeafList, totalSpineList, UserName,
			Password, force, stageFunction)
		if err != nil {
			//Operation failed so Rollback the Database
			RollBack = true
			break
		}
	}

	return
}

func (sh *DeviceInteractor) forceClearofDevices(ctx context.Context, FabricName string, existingipaddress []string, ipaddressToClear []string,
	UserName string, Password string) ([]AddDeviceResponse, error) {
	var err error

	ipaddress := sh.getUniqueList(ctx, existingipaddress, ipaddressToClear)
	AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)

	//Run Clear on the Devices
	err = sh.AddDevicesAndClearFabric(ctx, ipaddress, ipaddressToClear,
		UserName, Password)
	if err != nil {
		var Response AddDeviceResponse
		Response.FabricName = FabricName
		Response.FabricID = sh.FabricID
		Response.Errors = []error{err}
		AddDeviceResponseList = append(AddDeviceResponseList, Response)
		return AddDeviceResponseList, errors.New("Force Clear failed")
	}

	//cleanup DB for the Devices
	//Reset the Fabric Name and ID
	if Fabric, err := sh.Db.GetFabric(FabricName); err == nil {
		sh.FabricName = FabricName
		sh.FabricID = Fabric.ID
	}

	return AddDeviceResponseList, nil
}
func (sh *DeviceInteractor) fetchFabricDetails(ctx context.Context, FabricName string, ipaddress []string) ([]AddDeviceResponse, error) {
	var err error
	AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)
	LOG := appcontext.Logger(ctx)
	var Fabric domain.Fabric
	if Fabric, err = sh.Db.GetFabric(FabricName); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", FabricName)
		LOG.Errorln(statusMsg, err)
		for _, ip := range ipaddress {
			deviceResponse := AddDeviceResponse{IPAddress: ip, Errors: []error{errors.New(statusMsg)}}
			AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		}
		return AddDeviceResponseList, errors.New(statusMsg)
	}

	//Set the Fabric ID to be used through out the handler
	sh.FabricID = Fabric.ID
	sh.FabricName = FabricName
	sh.FabricProperties, err = sh.GetFabricSettings(ctx, sh.FabricName)
	if err != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch fabric Properties  %s", sh.FabricName)
		LOG.Infoln(statusMsg)
		for _, ip := range ipaddress {
			deviceResponse := AddDeviceResponse{IPAddress: ip, Errors: []error{err}}
			AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		}
		return AddDeviceResponseList, err
	}
	return AddDeviceResponseList, nil
}

func (sh *DeviceInteractor) deviceAlreadyRegisteredWithDifferentRole(LeafIPAddressList []string, SpineIPAddressList []string,
	existingLeafList []string, existingSpineList []string) ([]AddDeviceResponse, error) {
	var overallError error
	AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)
	existingLeafMap := make(map[string]bool, 0)
	for _, exLeaf := range existingLeafList {
		existingLeafMap[exLeaf] = true
	}

	existingSpineMap := make(map[string]bool, 0)
	for _, exSpine := range existingSpineList {
		existingSpineMap[exSpine] = true
	}

	//Check leaf in the existing Spine Map
	for _, leaf := range LeafIPAddressList {
		if _, present := existingSpineMap[leaf]; present {
			//Device already registered as Spine
			overallError = errors.New(fmt.Sprintln(leaf, "already configured as Spine"))
			deviceResponse := AddDeviceResponse{IPAddress: leaf, Errors: []error{overallError}}
			AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)

		}
	}

	//Check leaf in the existing Spine Map
	for _, spine := range SpineIPAddressList {
		if _, present := existingLeafMap[spine]; present {
			//Device already registered as Leaf
			overallError = errors.New(fmt.Sprintln(spine, "already configured as Leaf"))
			deviceResponse := AddDeviceResponse{IPAddress: spine, Errors: []error{overallError}}
			AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		}
	}
	return AddDeviceResponseList, overallError
}

func (sh *DeviceInteractor) executeAddDeviceStage(ctx context.Context, FabricName string, LeafIPaddressList []string,
	SpineIPaddressList []string, UserName string, Password string, force bool, function stageFunction) ([]AddDeviceResponse, error) {

	var fabricGate sync.WaitGroup

	AddDeviceResponseList := make([]AddDeviceResponse, 0, len(SpineIPaddressList)+len(LeafIPaddressList))

	//Concurrent Execution of Adding Spines and Leaves
	ResultChannel := make(chan AddDeviceResponse, len(SpineIPaddressList)+len(LeafIPaddressList))
	for _, SwitchIPAddress := range SpineIPaddressList {
		fabricGate.Add(1)
		go function(ctx, &fabricGate, ResultChannel, FabricName, SwitchIPAddress, UserName, Password, SpineRole)
	}
	for _, SwitchIPAddress := range LeafIPaddressList {
		fabricGate.Add(1)
		go function(ctx, &fabricGate, ResultChannel, FabricName, SwitchIPAddress, UserName, Password, LeafRole)
	}

	//Wait for the Concurrent execution to complete
	go func() {
		fabricGate.Wait()
		close(ResultChannel)
	}()

	overallOperationFailed := false
	for result := range ResultChannel {
		if len(result.Errors) > 0 {
			overallOperationFailed = true
		}

		AddDeviceResponseList = append(AddDeviceResponseList, result)
	}
	if overallOperationFailed {
		return AddDeviceResponseList, errors.New("Add Device Operation Failed")
	}
	return AddDeviceResponseList, nil
}

func (sh *DeviceInteractor) addSingleDeviceFirstStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, IPAddress string, UserName string, Password string, Role string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddDeviceFirstStage(ctx, FabricName, IPAddress, UserName, Password, Role)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: Role}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", Role, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", Role, IPAddress))
	ResultChannel <- Response
}

func (sh *DeviceInteractor) addSingleDeviceSecondStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, IPAddress string, UserName string, Password string, Role string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddDeviceSecondStage(ctx, FabricName, IPAddress, UserName, Password, Role)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: Role}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", Role, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", Role, IPAddress))
	ResultChannel <- Response
}

func (sh *DeviceInteractor) addSingleDeviceThirdStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, IPAddress string, UserName string, Password string, Role string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddDeviceThirdStage(ctx, FabricName, IPAddress, UserName, Password, Role)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: Role}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", Role, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", Role, IPAddress))
	ResultChannel <- Response
}

func (sh *DeviceInteractor) addSingleDeviceFourthStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, IPAddress string, UserName string, Password string, Role string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddDeviceFourthStage(ctx, FabricName, IPAddress, UserName, Password, Role)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: Role}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", Role, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", Role, IPAddress))
	ResultChannel <- Response
}

//ValidateFabricTopology validates the topology of the CLOS IP Fabric.
//e.g: Spine is not connected to another spine, MCT leaf not connected to another leaf, Spine not connected to leaf etc.
func (sh *DeviceInteractor) ValidateFabricTopology(ctx context.Context, FabricName string) (ValidateFabricResponse, error) {

	if sh.FabricProperties.FabricType == domain.NonCLOSFabricType {
		return sh.ValidateNonClosFabricTopology(ctx, FabricName)
	}
	LOG := appcontext.Logger(ctx)
	FabricValidateResponse := ValidateFabricResponse{FabricName: FabricName}
	devices, err := sh.ListDevices(FabricName)
	if err != nil {
		//Only Generic errors are returned
		return FabricValidateResponse, err
	}

	defer sh.cleanupLLDPLinks(sh.FabricID)
	DeviceMap := make(map[uint]string)
	SpineSet := mapset.NewSet()
	LeafSet := mapset.NewSet()
	DeviceCount := len(devices)
	for _, dev := range devices {
		if dev.DeviceRole == SpineRole {
			DeviceMap[dev.ID] = dev.IPAddress
			SpineSet.Add(dev.ID)
		}
		if dev.DeviceRole == LeafRole {
			DeviceMap[dev.ID] = dev.IPAddress
			LeafSet.Add(dev.ID)
		}
	}
	SpineCount := SpineSet.Cardinality()
	LeafCount := LeafSet.Cardinality()

	//Validate on basis of Count
	if SpineCount == DeviceCount {
		FabricValidateResponse.NoLeaves = true
		return FabricValidateResponse, nil
	}
	if LeafCount == DeviceCount {
		FabricValidateResponse.NoSpines = true
		return FabricValidateResponse, nil
	}

	FabricValidateResponse.MissingLinks = make([]string, 0)
	FabricValidateResponse.SpineSpineLinks = make([]string, 0)
	//Validate MCT Cluster
	for _, leaf := range LeafSet.ToSlice() {
		if str, err := sh.validateMctCluster(ctx, leaf.(uint)); err != nil {
			LOG.Error(err)
			FabricValidateResponse.LeafLeafLinks = append(FabricValidateResponse.LeafLeafLinks, str)
		}
	}
	//Validate Device Connectivity for Spines
	for _, leaf := range LeafSet.ToSlice() {
		if err := sh.validateDeviceConnectivity(LeafRole, SpineRole, leaf.(uint), DeviceMap, SpineSet); err != nil {
			FabricValidateResponse.MissingLinks = append(FabricValidateResponse.MissingLinks, fmt.Sprint(err))
		}
	}
	//Validate Device Connectivity for Leaves
	for _, spine := range SpineSet.ToSlice() {
		if err := sh.validateDeviceConnectivity(SpineRole, LeafRole, spine.(uint), DeviceMap, LeafSet); err != nil {
			FabricValidateResponse.MissingLinks = append(FabricValidateResponse.MissingLinks, fmt.Sprint(err))
		}
	}

	//Validate Spine-Spine connectivity
	for _, spine := range SpineSet.ToSlice() {
		if err := sh.validateSpineToSpineConnectivity(spine.(uint), DeviceMap, SpineSet); err != nil {
			FabricValidateResponse.SpineSpineLinks = append(FabricValidateResponse.SpineSpineLinks, fmt.Sprint(err))
		}
	}

	return FabricValidateResponse, nil
}

func (sh *DeviceInteractor) cleanupLLDPLinks(FabricID uint) {
	sh.Db.DeleteLLDPMarkedForDelete(sh.FabricID)
}
func (sh *DeviceInteractor) validateMctCluster(ctx context.Context, DeviceID uint) (string, error) {
	LOG := appcontext.Logger(ctx)
	clusters, err := sh.Db.GetMctClusters(sh.FabricID, DeviceID, []string{domain.ConfigCreate, domain.ConfigUpdate, domain.ConfigNone})
	if err != nil {
		statusMsg := fmt.Sprintf("Unable Count MCT Clusters Per device %d", DeviceID)
		LOG.Info(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}
	if len(clusters) > 1 {
		var buffer bytes.Buffer
		buffer.WriteString("More than one Leaf Connected to Device ")
		buffer.WriteString(clusters[0].DeviceOneMgmtIP)

		LOG.Error(buffer.String())
		return buffer.String(), errors.New(buffer.String())
	}
	return "", nil
}

func (sh *DeviceInteractor) validateSpineToSpineConnectivity(DeviceID uint, DeviceMap map[uint]string, SpineSet mapset.Set) error {

	neighbors, _ := sh.Db.GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion(sh.FabricID, DeviceID)

	neighborSet := mapset.NewSet()
	for _, neighbor := range neighbors {
		if neighbor.DeviceOneRole == SpineRole && neighbor.DeviceTwoRole == SpineRole {
			neighborSet.Add(neighbor.DeviceTwoID)
		}
	}

	var StatusBuffer bytes.Buffer
	intersectionSet := neighborSet.Intersect(SpineSet)

	if intersectionSet.Cardinality() > 0 {
		for _, spinePeerDev := range intersectionSet.ToSlice() {
			StatusBuffer.WriteString(fmt.Sprintf("%s Device %s connected to %s Device %s", SpineRole,
				DeviceMap[DeviceID], SpineRole, DeviceMap[spinePeerDev.(uint)]))
		}
		return errors.New(StatusBuffer.String())
	}

	return nil
}

func (sh *DeviceInteractor) validateDeviceConnectivity(DeviceRole string, PeerDeviceRole string,
	DeviceID uint, DeviceMap map[uint]string, PeerSet mapset.Set) error {

	neighbors, _ := sh.Db.GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion(sh.FabricID, DeviceID)
	neighborSet := mapset.NewSet()
	for _, neighbor := range neighbors {
		neighborSet.Add(neighbor.DeviceTwoID)
	}
	var StatusBuffer bytes.Buffer
	//list of Peers(devices belonging to opposite role) that are not the neighbors of this device
	DisconnectedPeerSet := PeerSet.Difference(neighborSet)
	if DisconnectedPeerSet.Cardinality() > 0 {
		for _, DisPeerDev := range DisconnectedPeerSet.ToSlice() {
			StatusBuffer.WriteString(fmt.Sprintf("%s Device %s not connected to %s Device %s", DeviceRole,
				DeviceMap[DeviceID], PeerDeviceRole, DeviceMap[DisPeerDev.(uint)]))
		}
		return errors.New(StatusBuffer.String())
	}

	return nil
}

//GetActionRequestObject prepares the Action Request Object to be send to actions for configuring
func (sh *DeviceInteractor) GetActionRequestObject(ctx context.Context, FabricName string, Force bool) (operation.ConfigFabricRequest, error) {
	LOG := appcontext.Logger(ctx)
	var config operation.ConfigFabricRequest
	var err error

	//Set the FabricName
	config.FabricName = FabricName

	//Read the Fabric Settings from the DB for the given Fabric
	if config.FabricSettings, err = sh.GetFabricSettings(ctx, FabricName); err != nil {
		statusMsg := fmt.Sprintf("Failed to perform Query Topology %s", FabricName)
		LOG.Println(statusMsg)
		return config, errors.New(statusMsg)
	}

	//Retrieve Fabric and the Switch Details from the DB
	if err := sh.prepareConfiguration(ctx, &config); err != nil {
		return config, err
	}
	return config, err
}

//ConfigureFabric configures the fabric by reading the configuration from the DB
func (sh *DeviceInteractor) ConfigureFabric(ctx context.Context, FabricName string, force bool, persist bool) (ConfigureFabricResponse, error) {
	//Setup the logger
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Configure Fabric")
	ctx = context.WithValue(ctx, appcontext.FabricName, FabricName)
	LOG := appcontext.Logger(ctx)

	response := ConfigureFabricResponse{FabricName: FabricName}
	config, err := sh.GetActionRequestObject(ctx, FabricName, force)
	if err != nil {
		return response, err
	}

	//Send the Config Object to Actions for configuring the switches
	if Errors := sh.FabricAdapter.ConfigureFabric(ctx, config, force, persist); len(Errors) != 0 {
		response.Errors = Errors
		return response, errors.New("Configuration Failed on Switch")
	}

	if err := sh.CleanupDBAfterConfigureSuccess(); err != nil {
		response.Errors = []actions.OperationError{actions.OperationError{Operation: "Clean up DB Failed", Error: err}}
		return response, err
	}
	//On Success backup the DB
	if err := sh.Db.Backup(); err != nil {
		LOG.Printf("Failed to backup DB during Configure %s\n", err)
	}
	return response, nil
}

//CleanupDBAfterConfigureSuccess cleans the DB after "configure fabric" operation is successful
func (sh *DeviceInteractor) CleanupDBAfterConfigureSuccess() error {

	//Update all Switch Configs that have config_type as CREATE or UPDATE to None
	sh.Db.UpdateSwitchConfigsASConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)
	sh.Db.UpdateSwitchConfigsLoopbackConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)
	sh.Db.UpdateSwitchConfigsVTEPLoopbackConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)

	//Update all Interface Switch Configs that have config_type as CREATE or UPDATE to None
	sh.Db.UpdateInterfaceSwitchConfigsConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)

	//Update all BGP Switch Configs that have config_type as CREATE or UPDATE to None
	sh.Db.UpdateBGPSwitchConfigsConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)

	sh.Db.UpdateLLDPConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)
	sh.Db.DeleteLLDPMarkedForDelete(sh.FabricID)
	//Also cascade delete for delete_config
	sh.Db.UpdateMctPortConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)
	sh.Db.UpdateMctClusterConfigType(sh.FabricID, []string{domain.ConfigCreate, domain.ConfigUpdate},
		domain.ConfigNone)
	sh.Db.DeleteMctPortsMarkedForDelete(sh.FabricID)
	sh.Db.DeleteMctClustersMarkedForDelete(sh.FabricID)
	return nil
}

//Prepares the Config Object for configuring the switches
func (sh *DeviceInteractor) prepareConfiguration(ctx context.Context, config *operation.ConfigFabricRequest) error {
	LOG := appcontext.Logger(ctx)
	//Hosts holds the Switch details
	config.Hosts = make([]operation.ConfigSwitch, 0)

	//Prepare Switches Map
	switchMap := make(map[uint]domain.Device)
	switchConfigMap := make(map[uint]domain.SwitchConfig)
	devices, err := sh.Db.GetDevicesInFabric(sh.FabricID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch devices from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	for _, dev := range devices {
		switchMap[dev.ID] = dev
	}

	//Fetch the Switch Configuration from Database
	switchConfigs, err := sh.Db.GetSwitchConfigs(sh.FabricName)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to fetch Device Configs from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	//For each SwitchConfig prepare the Host(Switch Details)
	for _, sw := range switchConfigs {
		switchConfigMap[sw.DeviceID] = sw
	}
	for _, sw := range switchConfigs {
		var host operation.ConfigSwitch
		dev := switchMap[sw.DeviceID]
		sh.populateHostDetails(ctx, host, dev, config, switchConfigMap, domain.ConfigCreate)
	}
	Oper := domain.MctCreate
	Clusters, merr := sh.prepareMctClusters(ctx, switchConfigs, config, switchMap, switchConfigMap, Oper)
	if merr != nil {
		statusMsg := fmt.Sprintf("Failed to fetch Device MCT Configs from %s", sh.FabricName)
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	config.MctCluster = Clusters
	return nil
}

//Prepare MCT Config struct
func (sh *DeviceInteractor) prepareMctClusters(ctx context.Context, switchConfigs []domain.SwitchConfig, config *operation.ConfigFabricRequest,
	switchMap map[uint]domain.Device, switchConfigMap map[uint]domain.SwitchConfig, Oper int) (map[uint][]operation.ConfigCluster, error) {
	LOG := appcontext.Logger(ctx)
	var err error
	var Clusters []domain.MctClusterConfig
	ConfigClusters := make(map[uint][]operation.ConfigCluster)
	GetClusters := func(ConfigType []string) error {
		DeviceVisited := make(map[uint]bool)
		for _, sw := range switchConfigs {
			Switch := switchMap[sw.DeviceID]
			Clusters, err = sh.Db.GetMctClusters(sh.FabricID, sw.DeviceID, ConfigType)

			if err != nil {
				statusMsg := fmt.Sprintf("Unable to Retrieve MctCluster from DB for device %d", sw.DeviceID)
				LOG.Errorln(statusMsg)
				return err
			}
			//In future Cluster Can neighbor with More than one neighbor
			//Hence this loop
			for _, Cluster := range Clusters {
				var ConfigCluster operation.ConfigCluster
				var ClusterMember1 operation.ClusterMemberNode
				var ClusterMember2 operation.ClusterMemberNode
				if _, ok := DeviceVisited[Cluster.DeviceID]; ok {
					//Already Device Visited Hence Continue
					continue
				}
				DeviceVisited[Cluster.DeviceID] = true
				DeviceVisited[Cluster.MCTNeighborDeviceID] = true
				if _, ok := switchMap[Cluster.MCTNeighborDeviceID]; ok == false {
					//Device not present in list but needed to clear MCT clusters
					switchMap[Cluster.MCTNeighborDeviceID], _ = sh.Db.GetDeviceUsingDeviceID(Cluster.FabricID, Cluster.MCTNeighborDeviceID)
					/*TODO: Mark the MCTNeighborDeviceId Table to be deleted here since
					Neighbor device is not getting deleted.*/

				}

				FillConfigCluster := func(opcode uint, ConfigType []string) error {
					MyPorts, RemoteNodePPorts, merr := sh.getMctConnectedPorts(ctx, Cluster.DeviceID, ConfigType)
					if merr != nil {
						return merr
					}
					setBit := func(n uint64, pos uint) uint64 {
						n |= (1 << pos)
						return n
					}
					clearBit := func(n uint64, pos uint) uint64 {
						return n &^ (1 << pos)
					}

					hasBit := func(n uint64, pos uint) bool {
						val := n & (1 << pos)
						return (val > 0)
					}

					UpdatedAttributes := Cluster.UpdatedAttributes
					LOG.Infoln("UpdatedAttributes - ", UpdatedAttributes)
					if opcode == domain.MctUpdate &&
						!hasBit(UpdatedAttributes, domain.BitPositionForMctCreate) &&
						!hasBit(UpdatedAttributes, domain.BitPositionForForPortAdd) &&
						!hasBit(UpdatedAttributes, domain.BitPositionForForPortDelete) &&
						!hasBit(UpdatedAttributes, domain.BitPositionForPeerSpeed) &&
						!hasBit(UpdatedAttributes, domain.BitPositionForPeerOneIP) &&
						!hasBit(UpdatedAttributes, domain.BitPositionForPeerTwoIP) {
						return nil
					}

					ConfigCluster.FabricName = sh.FabricName
					ConfigCluster.ClusterID = fmt.Sprintf("%d", Cluster.ClusterID)
					ConfigCluster.ClusterName = fmt.Sprintf("%s%s%d", sh.FabricName, "-cluster-", Cluster.ClusterID)
					ConfigCluster.ClusterControlVe = Cluster.ControlVE
					ConfigCluster.ClusterControlVlan = Cluster.ControlVlan
					ClusterMember1 = operation.ClusterMemberNode{}
					ClusterMember1.NodeID = Cluster.LocalNodeID
					ClusterMember1.NodeMgmtIP = Switch.IPAddress
					ClusterMember1.NodeModel = Switch.Model
					ClusterMember1.NodeMgmtUserName = Switch.UserName
					ClusterMember1.NodeMgmtPassword = Switch.Password
					ClusterMember1.NodePrincipalPriority = "0"
					ClusterMember1.NodePeerIntfSpeed = Cluster.PeerInterfaceSpeed
					ClusterMember1.NodePeerIntfName = Cluster.PeerInterfaceName
					ClusterMember1.NodePeerIntfType = Cluster.PeerInterfacetype
					ClusterMember1.RemoteNodePeerLoopbackIP = sw.LoopbackIP
					ClusterMember1.NodePeerLoopbackIP = switchConfigMap[Cluster.MCTNeighborDeviceID].LoopbackIP
					ClusterMember1.NodePeerIP = Cluster.PeerOneIP
					ClusterMember1.BFDEnable = config.FabricSettings.BFDEnable
					ClusterMember1.BFDRx = config.FabricSettings.BFDRx
					ClusterMember1.BFDTx = config.FabricSettings.BFDTx
					ClusterMember1.BFDMultiplier = config.FabricSettings.BFDMultiplier
					ClusterMember1.RemoteNodePeerIP = fmt.Sprintf("%s%s", Cluster.PeerTwoIP, "/31")

					ClusterMember2 = operation.ClusterMemberNode{}
					ClusterMember2.NodeID = Cluster.RemoteNodeID
					ClusterMember2.NodeMgmtIP = switchMap[Cluster.MCTNeighborDeviceID].IPAddress
					ClusterMember2.NodeModel = switchMap[Cluster.MCTNeighborDeviceID].Model
					ClusterMember2.NodeMgmtUserName = switchMap[Cluster.MCTNeighborDeviceID].UserName
					ClusterMember2.NodeMgmtPassword = switchMap[Cluster.MCTNeighborDeviceID].Password
					ClusterMember2.NodePrincipalPriority = "0"
					ClusterMember2.NodePeerIntfSpeed = Cluster.PeerInterfaceSpeed
					ClusterMember2.NodePeerIntfName = Cluster.PeerInterfaceName
					ClusterMember2.NodePeerIntfType = Cluster.PeerInterfacetype
					ClusterMember2.RemoteNodePeerLoopbackIP = switchConfigMap[Cluster.MCTNeighborDeviceID].LoopbackIP
					ClusterMember2.NodePeerLoopbackIP = sw.LoopbackIP
					ClusterMember2.NodePeerIP = Cluster.PeerTwoIP
					ClusterMember2.BFDEnable = config.FabricSettings.BFDEnable
					ClusterMember2.BFDRx = config.FabricSettings.BFDRx
					ClusterMember2.BFDTx = config.FabricSettings.BFDTx
					ClusterMember2.BFDMultiplier = config.FabricSettings.BFDMultiplier
					ClusterMember2.RemoteNodePeerIP = fmt.Sprintf("%s%s", Cluster.PeerOneIP, "/31")

					ConfigCluster.OperationBitMap = Cluster.UpdatedAttributes
					mCreate := hasBit(UpdatedAttributes, domain.BitPositionForMctCreate)
					if opcode == domain.MctUpdate && mCreate {
						var CreatedCluster operation.ConfigCluster
						CreatedClusterMember1 := ClusterMember1
						CreatedClusterMember2 := ClusterMember2
						CreatedCluster = ConfigCluster
						CreatedCluster.OperationBitMap = 0
						CreatedCluster.OperationBitMap = setBit(CreatedCluster.OperationBitMap, domain.BitPositionForMctCreate)
						CreatedClusterMember1.RemoteNodeConnectingPorts = MyPorts
						CreatedClusterMember2.RemoteNodeConnectingPorts = RemoteNodePPorts
						CreatedCluster.ClusterMemberNodes = append(CreatedCluster.ClusterMemberNodes, CreatedClusterMember1)
						CreatedCluster.ClusterMemberNodes = append(CreatedCluster.ClusterMemberNodes, CreatedClusterMember2)
						ConfigClusters[domain.MctCreate] = append(ConfigClusters[domain.MctCreate], CreatedCluster)
						return nil
					}

					if opcode == domain.MctUpdate && hasBit(UpdatedAttributes, domain.BitPositionForForPortDelete) {
						var DeletedConfigCluster operation.ConfigCluster
						DeletedClusterMember1 := ClusterMember1
						DeletedClusterMember2 := ClusterMember2
						DeletedConfigCluster = ConfigCluster
						DeletedConfigCluster.OperationBitMap = UpdatedAttributes
						DeletedConfigCluster.OperationBitMap = clearBit(DeletedConfigCluster.OperationBitMap, domain.BitPositionForForPortAdd)
						MyDeletedPorts, RemoteNodeDeltedPorts, merr := sh.getMctConnectedPorts(ctx, Cluster.DeviceID, []string{domain.ConfigDelete})
						if merr != nil {
							LOG.Errorln("Error While getting Mct Deleted Ports")
							return merr
						}
						DeletedClusterMember1.RemoteNodeConnectingPorts = MyDeletedPorts
						DeletedClusterMember2.RemoteNodeConnectingPorts = RemoteNodeDeltedPorts
						DeletedConfigCluster.ClusterMemberNodes = append(DeletedConfigCluster.ClusterMemberNodes, DeletedClusterMember1)
						DeletedConfigCluster.ClusterMemberNodes = append(DeletedConfigCluster.ClusterMemberNodes, DeletedClusterMember2)
						ConfigClusters[opcode] = append(ConfigClusters[opcode], DeletedConfigCluster)
					}
					LOG.Infoln("UpdatedAttributes - ", UpdatedAttributes)
					if opcode == domain.MctUpdate && hasBit(UpdatedAttributes, domain.BitPositionForForPortAdd) {
						ConfigCluster.OperationBitMap = clearBit(UpdatedAttributes, domain.BitPositionForForPortDelete)
					}

					if opcode == domain.MctUpdate && hasBit(UpdatedAttributes, domain.BitPositionForForPortDelete) {
						//clearing other bits since they would have already sent as part of Delete
						ConfigCluster.OperationBitMap = clearBit(UpdatedAttributes, domain.BitPositionForForPortDelete)
						ConfigCluster.OperationBitMap = clearBit(ConfigCluster.OperationBitMap, domain.BitPositionForPeerSpeed)
						ConfigCluster.OperationBitMap = clearBit(ConfigCluster.OperationBitMap, domain.BitPositionForPeerOneIP)
						ConfigCluster.OperationBitMap = clearBit(ConfigCluster.OperationBitMap, domain.BitPositionForPeerTwoIP)
					}
					if opcode == domain.MctUpdate && ConfigCluster.OperationBitMap == 0 {
						LOG.Infoln("None of the Update bits are  set hence Nothing to be Done")
						return nil
					}
					ClusterMember1.RemoteNodeConnectingPorts = MyPorts
					ClusterMember2.RemoteNodeConnectingPorts = RemoteNodePPorts
					ConfigCluster.ClusterMemberNodes = append(ConfigCluster.ClusterMemberNodes, ClusterMember1)
					ConfigCluster.ClusterMemberNodes = append(ConfigCluster.ClusterMemberNodes, ClusterMember2)
					ConfigClusters[opcode] = append(ConfigClusters[opcode], ConfigCluster)
					return nil
				}

				switch Cluster.ConfigType {
				case domain.ConfigCreate:
					err = FillConfigCluster(domain.MctCreate, ConfigType)
				case domain.ConfigDelete:
					err = FillConfigCluster(domain.MctDelete, ConfigType)
				case domain.ConfigUpdate:
					//Since Add/Delete could not be sent in One Request Create Two Request
					//One For create and Other for Delted Ports
					//In Fill cluster if DeltedPOrt BIT set will update cluster for both add/Delete
					err = FillConfigCluster(domain.MctUpdate, []string{domain.ConfigCreate})

				}
				if err != nil {
					statusMsg := fmt.Sprintf("Filling Cluster Failed")
					LOG.Errorln(statusMsg)
					return err
				}
				if Oper == domain.MctDelete {
					//Handles Deconfigure Cluster
					err = FillConfigCluster(domain.MctDelete, ConfigType)
					if err != nil {
						statusMsg := fmt.Sprintf("Deconfigure Filling Cluster Failed")
						LOG.Errorln(statusMsg)
						return err
					}

				}
			}
		}
		return nil
	}

	switch Oper {
	case domain.MctCreate:
		err := GetClusters([]string{domain.ConfigCreate})
		if err != nil {
			statusMsg := fmt.Sprintf("Error While Getting MCT cluster %s \n", err.Error())
			LOG.Errorln(statusMsg)
			return ConfigClusters, err
		}
		err = GetClusters([]string{domain.ConfigUpdate})
		if err != nil {
			statusMsg := fmt.Sprintf("Error While Getting MCT[UPDATE] cluster %s \n", err.Error())
			LOG.Errorln(statusMsg)
			return ConfigClusters, err
		}
		err = GetClusters([]string{domain.ConfigDelete})
		if err != nil {
			statusMsg := fmt.Sprintf("Error While Getting MCT[Delete] cluster %s \n", err.Error())
			LOG.Errorln(statusMsg)
			return ConfigClusters, err
		}

	case domain.MctDelete:
		err = GetClusters([]string{})
		if err != nil {
			statusMsg := fmt.Sprintf("Error While deleting MCT cluster %s \n", err.Error())
			LOG.Errorln(statusMsg)
			return ConfigClusters, err
		}

	}
	return ConfigClusters, nil
}

func (sh *DeviceInteractor) getMctConnectedPorts(ctx context.Context, DeviceID uint, ConfigType []string) ([]operation.InterNodeLinkPort, []operation.InterNodeLinkPort, error) {
	LOG := appcontext.Logger(ctx)
	var MyPorts []operation.InterNodeLinkPort
	var RemoteNodePorts []operation.InterNodeLinkPort
	Ports, perr := sh.Db.GetMctMemberPortsConfig(sh.FabricID, DeviceID, 0, ConfigType)
	if perr != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch MCT Cluster Connected Ports for Device ID %d", DeviceID)
		LOG.Errorln(statusMsg)
		return MyPorts, RemoteNodePorts, perr
	}
	for _, port := range Ports {
		var MyPort operation.InterNodeLinkPort
		var RemoteNodePort operation.InterNodeLinkPort
		MyPort.IntfType = port.InterfaceType
		MyPort.IntfName = port.InterfaceName
		RemoteNodePort.IntfType = port.RemoteInterfaceType
		RemoteNodePort.IntfName = port.RemoteInterfaceName
		MyPorts = append(MyPorts, MyPort)
		RemoteNodePorts = append(RemoteNodePorts, RemoteNodePort)
	}
	return MyPorts, RemoteNodePorts, nil

}

func (sh *DeviceInteractor) populateHostDetails(ctx context.Context, host operation.ConfigSwitch, Switch domain.Device,
	config *operation.ConfigFabricRequest, switchConfigMap map[uint]domain.SwitchConfig, opcode string) {
	LOG := appcontext.Logger(ctx)
	//TODO Get away from two fields Device,Host
	var err error
	sw := switchConfigMap[Switch.ID]
	host.Device = Switch.IPAddress
	host.Host = Switch.IPAddress
	host.Fabric = config.FabricName
	host.UserName = Switch.UserName
	host.Password = Switch.Password
	sw.UserName = Switch.UserName
	sw.Password = Switch.Password

	host.Chassis = "No"
	host.Role = sw.Role
	host.Model = Switch.Model
	host.Principal = false

	host.ConfigureOverlayGateway = config.FabricSettings.ConfigureOverlayGateway
	host.LoopbackPortNumber = config.FabricSettings.LoopBackPortNumber

	//BGP Fields
	host.MaxPaths = config.FabricSettings.MaxPaths
	host.Mtu = config.FabricSettings.MTU
	host.IPMtu = config.FabricSettings.IPMTU
	host.BgpMultihop = config.FabricSettings.BGPMultiHop
	host.BgpLocalAsn = sw.LocalAS
	host.SpinePeerGroup = config.FabricSettings.SpinePeerGroup
	host.LeafPeerGroup = config.FabricSettings.LeafPeerGroup
	host.SingleSpineAs = false
	host.AllowasIn = config.FabricSettings.AllowASIn
	host.BFDEnable = config.FabricSettings.BFDEnable
	//TODO -- NONCLOS --PeerGroup
	if host.Role == LeafRole || host.Role == RackRole {
		host.Network = sw.VTEPLoopbackIP + "/32"
		host.PeerGroup = host.LeafPeerGroup
		host.PeerGroupDescription = "To Spine"
		host.UnconfigureMCTBGPNeighbors, host.ConfigureMCTBGPNeighbors, err = sh.prepareMCTBGPNeighbors(ctx, &sw,
			host, switchConfigMap, opcode, host.BFDEnable)
		if err != nil {
			LOG.Errorln("Error preparing MCT BGP Neighbors - ", err)
			//TODO Should Skip After Error Need to change signature of function
		}
	}
	if host.Role == SpineRole {
		host.PeerGroup = host.SpinePeerGroup
		host.PeerGroupDescription = "To Leaf"
	}
	host.BgpNeighbors = sh.prepareBGPNeighbors(ctx, &sw)

	if host.Role == RackRole {
		evpnNeighbors := sh.prepareNONCLOSBGPEVPNNeighbors(ctx, &sw)
		LOG.Infoln("EVPN Neighbors", sw.DeviceID, evpnNeighbors)
		host.BgpNeighbors = append(host.BgpNeighbors, evpnNeighbors...)
		host.EvpnPeerGroup = config.FabricSettings.RackPeerOvgGroup
		host.EvpnPeerGroupDescription = "Rack Overlay EBGP Group"
		host.PeerGroup = config.FabricSettings.RackPeerEBGPGroup
		host.PeerGroupDescription = "Rack Underlay EBGP Group"

		host.NonCLOSNetwork = sw.LoopbackIP + "/32"

	}

	//Interface Fields
	host.P2PIPType = config.FabricSettings.P2PIPType

	host.P2pLinkRange = config.FabricSettings.P2PLinkRange
	host.BfdMultiplier = config.FabricSettings.BFDMultiplier
	host.BfdRx = config.FabricSettings.BFDRx
	host.BfdTx = config.FabricSettings.BFDTx

	host.Interfaces = sh.prepareInterfaceConfigs(ctx, &sw, config)

	//ovg fields
	host.VtepLoopbackPortNumber = config.FabricSettings.VTEPLoopBackPortNumber
	host.VlanVniAutoMap = formatYesNo(config.FabricSettings.VNIAutoMap)
	host.AnycastMac = config.FabricSettings.AnyCastMac
	host.IPV6AnycastMac = config.FabricSettings.IPV6AnyCastMac

	//EVPN Fields
	host.ArpAgingTimeout = config.FabricSettings.ArpAgingTimeout
	host.MacAgingTimeout = config.FabricSettings.MacAgingTimeout
	host.MacAgingConversationalTimeout = config.FabricSettings.MacAgingConversationalTimeout
	host.MacMoveLimit = config.FabricSettings.MacMoveLimit
	host.DuplicateMacTimer = config.FabricSettings.DuplicateMacTimer
	host.DuplicateMaxTimerMaxCount = config.FabricSettings.DuplicateMaxTimerMaxCount

	if len(host.Interfaces) > 0 {
		config.Hosts = append(config.Hosts, host)
	}

}

//Prepare BGP Peer Object
func (sh *DeviceInteractor) prepareBGPNeighbors(ctx context.Context, switchConfig *domain.SwitchConfig) []operation.ConfigBgpNeighbor {
	bgpConfigs := sh.GetBGPSwitchConfigs(ctx, sh.FabricID, switchConfig.DeviceID)
	Peers := make([]operation.ConfigBgpNeighbor, 0)
	for _, bgp := range bgpConfigs {
		var Peer operation.ConfigBgpNeighbor
		Peer.NeighborAddress = bgp.RemoteIPAddress
		Peer.RemoteAs, _ = strconv.ParseInt(bgp.RemoteAS, 10, 64)
		Peer.ConfigType = bgp.ConfigType
		Peer.NeighborType = bgp.Type
		Peers = append(Peers, Peer)
	}
	return Peers
}

//Prepare BGP Peer Object
func (sh *DeviceInteractor) prepareNONCLOSBGPEVPNNeighbors(ctx context.Context, switchConfig *domain.SwitchConfig) []operation.ConfigBgpNeighbor {
	evpnConfigs := sh.GetEVPNNeighborConfig(ctx, sh.FabricName, switchConfig.DeviceID)
	Peers := make([]operation.ConfigBgpNeighbor, 0)
	for _, evpnConfig := range evpnConfigs {
		var Peer operation.ConfigBgpNeighbor
		Peer.NeighborAddress = evpnConfig.EVPNAddress
		Peer.RemoteAs, _ = strconv.ParseInt(evpnConfig.RemoteAS, 10, 64)
		Peer.ConfigType = evpnConfig.ConfigType
		Peer.NeighborType = domain.EVPNENIGHBORType
		Peers = append(Peers, Peer)
	}
	return Peers
}

//Prepare BGP Peer Object
func (sh *DeviceInteractor) prepareMCTBGPNeighbors(ctx context.Context, switchConfig *domain.SwitchConfig, host operation.ConfigSwitch, switchConfigMap map[uint]domain.SwitchConfig, opcode string, bfdEnabled string) (
	operation.ConfigDataPlaneCluster, operation.ConfigDataPlaneCluster, error) {
	LOG := appcontext.Logger(ctx)
	var unconfigureMCTBGPNeighbors operation.ConfigDataPlaneCluster
	var configureMCTBGPNeighbors operation.ConfigDataPlaneCluster
	bgpConfigs, err := sh.GetMCTBGPSwitchConfigs(ctx, sh.FabricID, switchConfig.DeviceID)
	if err != nil {
		LOG.Errorln("Error Retrieving MCT BGP Neighbors for Device - ", switchConfig.DeviceIP, " - ", err)
		return unconfigureMCTBGPNeighbors, configureMCTBGPNeighbors, err
	}
	unconfigureMCTBGPNeighbors.FabricName = sh.FabricName
	configureMCTBGPNeighbors.FabricName = sh.FabricName
	unconfigureMCTBGPNeighbors.OperationBitMap = 0
	configureMCTBGPNeighbors.OperationBitMap = 0

	for _, bgp := range bgpConfigs {
		var memberNode operation.DataPlaneClusterMemberNode
		memberNode.NodeMgmtIP = switchConfig.DeviceIP
		memberNode.NodeMgmtUserName = switchConfig.UserName
		memberNode.NodeMgmtPassword = switchConfig.Password
		memberNode.NodeModel = host.Model
		memberNode.NodePeerIP = bgp.RemoteIPAddress
		memberNode.NodePeerLoopbackIP = switchConfigMap[bgp.RemoteDeviceID].LoopbackIP
		memberNode.NodeLoopBackNumber = host.LoopbackPortNumber
		memberNode.NodePeerASN = bgp.RemoteAS
		memberNode.NodePeerEncapType = bgp.EncapsulationType
		if bfdEnabled == "Yes" {
			memberNode.NodePeerBFDEnabled = "Yes"
		} else {
			memberNode.NodePeerBFDEnabled = "No"
		}
		if bgp.ConfigType == domain.ConfigDelete || opcode == domain.ConfigDelete {
			unconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes = append(unconfigureMCTBGPNeighbors.DataPlaneClusterMemberNodes, memberNode)
		}
		if (bgp.ConfigType == domain.ConfigCreate || bgp.ConfigType == domain.ConfigUpdate) && opcode == domain.ConfigCreate {
			configureMCTBGPNeighbors.DataPlaneClusterMemberNodes = append(configureMCTBGPNeighbors.DataPlaneClusterMemberNodes, memberNode)
		}
	}
	return unconfigureMCTBGPNeighbors, configureMCTBGPNeighbors, nil
}

//Prepare Interface Configs
func (sh *DeviceInteractor) prepareInterfaceConfigs(ctx context.Context, switchConfig *domain.SwitchConfig, config *operation.ConfigFabricRequest) []operation.ConfigInterface {

	InterfaceConfigs := sh.GetInterfaceSwitchConfigs(ctx, sh.FabricID, switchConfig.DeviceID)

	Interfaces := make([]operation.ConfigInterface, 0)
	for _, intf := range InterfaceConfigs {
		var Interface operation.ConfigInterface
		Interface.InterfaceName = intf.IntName
		Interface.InterfaceType = intf.IntType

		Interface.IP = intf.IPAddress
		Interface.ConfigType = intf.ConfigType
		Interface.Description = intf.Description

		//TODO based on Fabric Settings
		if strings.Contains(Interface.InterfaceType, domain.IntfTypeEthernet) {
			if intf.DonorType == "" {
				// Numbered interface
				Interface.IP = Interface.IP + "/31"
			} else {
				// un-numbered interface. IPAddress will contain the donor ip
				Interface.Donor = domain.IntfTypeLoopback
				Interface.DonorPort = config.FabricSettings.LoopBackPortNumber
			}
		} else { //loopback
			Interface.IP = Interface.IP + "/32"
		}
		Interfaces = append(Interfaces, Interface)
	}

	//Add two additional Interfaces
	{
		var Interface operation.ConfigInterface
		Interface.InterfaceName = config.FabricSettings.LoopBackPortNumber
		Interface.InterfaceType = domain.IntfTypeLoopback
		Interface.IP = switchConfig.LoopbackIP + "/32"
		Interface.ConfigType = switchConfig.LoopbackIPConfigType
		Interfaces = append(Interfaces, Interface)
	}
	if switchConfig.Role == LeafRole || switchConfig.Role == RackRole {
		var Interface operation.ConfigInterface
		Interface.InterfaceName = config.FabricSettings.VTEPLoopBackPortNumber
		Interface.InterfaceType = domain.IntfTypeLoopback
		Interface.IP = switchConfig.VTEPLoopbackIP + "/32"
		Interface.ConfigType = switchConfig.VTEPLoopbackIPConfigType
		Interfaces = append(Interfaces, Interface)
	}
	return Interfaces
}

//GetFabricSettings returns the fabric settings
func (sh *DeviceInteractor) GetFabricSettings(ctx context.Context, FabricName string) (domain.FabricProperties, error) {
	LOG := appcontext.Logger(ctx)
	var FabricProperties domain.FabricProperties
	var Fabric domain.Fabric
	var err error

	if Fabric, err = sh.Db.GetFabric(FabricName); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", FabricName)
		LOG.Errorln(statusMsg)
		return FabricProperties, errors.New(statusMsg)
	}
	sh.FabricID = Fabric.ID
	sh.FabricName = FabricName

	return sh.Db.GetFabricProperties(Fabric.ID)

}

func formatYesNo(data string) bool {
	if data == "Yes" {
		return true
	}
	return false
}

func (sh *DeviceInteractor) fetchRegisteredDevices(ctx context.Context, FabricName string) ([]string, []string, error) {
	spineList := make([]string, 0, 0)
	leafList := make([]string, 0, 0)
	fabric, err := sh.Db.GetFabric(FabricName)
	if err != nil {
		return spineList, leafList, err
	}
	devices, err := sh.Db.GetDevicesInFabric(fabric.ID)
	if err != nil {
		return spineList, leafList, err
	}
	for _, device := range devices {
		if device.DeviceRole == SpineRole {
			spineList = append(spineList, device.IPAddress)
		}
		if device.DeviceRole == LeafRole {
			leafList = append(leafList, device.IPAddress)
		}
	}
	return spineList, leafList, nil
}

func (sh *DeviceInteractor) getUniqueList(ctx context.Context, first []string, second []string) []string {
	GetIPAddressKey := func(data interface{}) string {
		s, _ := data.(string)
		return fmt.Sprintln(s)
	}
	List := func(s util.Interface) []string {
		m := s.GetMap()
		s.List()
		list := make([]string, 0, 0)

		for _, elem := range m {
			intf, _ := elem.(string)
			list = append(list, intf)
		}

		return list
	}
	set := util.NewSet(GetIPAddressKey)
	for _, elem := range first {
		set.Add(elem)
	}
	for _, elem := range second {
		set.Add(elem)
	}

	return List(set)
}

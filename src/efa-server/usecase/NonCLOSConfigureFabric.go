package usecase

import (
	"efa-server/gateway/appcontext"
)
import "context"
import (
	"bytes"
	"efa-server/domain"
	"efa-server/infra/constants"
	"efa-server/infra/util"
	"errors"
	"fmt"
	"sync"
)

const (
	//MCTInterfaceOne default port for MCT in NON-CLOS
	MCTInterfaceOne = "0/46"
	//MCTInterfaceTwo default port for MCT in NON-CLOS
	MCTInterfaceTwo = "0/47"
	//L3LoopBackInterface default port for L3 Loopback in NON-CLOS
	L3LoopBackInterface = "0/48"
)

//Rack for holding Rack IP Address
type Rack struct {
	IP1 string
	IP2 string
}

//DeviceSwitchConfigMapTable Device and Switch Config Details
type DeviceSwitchConfigMapTable struct {
	Device domain.Device
	Config domain.SwitchConfig
}

//RackDeviceMapTable holds Rack and Devices that belong to the Rack
type RackDeviceMapTable struct {
	Rack    domain.Rack
	Devices []DeviceSwitchConfigMapTable
}

type rackstageFunction func(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse,
	FabricName string, Rack Rack, UserName string, Password string)

//AddRacks adds multiple Racks (pair of leaves) to the NON CLOS Fabric
func (sh *DeviceInteractor) AddRacks(ctx context.Context, FabricName string, RackList []Rack,
	UserName string, Password string, force bool) (addDeviceResponse []AddDeviceResponse, err error) {

	//Setup the logger
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Add Racks")
	ctx = context.WithValue(ctx, appcontext.FabricName, FabricName)

	LOG := appcontext.Logger(ctx)

	//Fetch the existing Pairs already registered
	existingRack, err := sh.fetchRegisteredRacks(ctx, FabricName)

	//If force is enabled clear up the configuration on devices specified in the IP address
	if force {
		LOG.Infoln("Force option enabled on Racks", RackList)
		LOG.Infoln("Existing Device Racks", existingRack)
		addDeviceResponse, err = sh.forceClearofRacks(ctx, FabricName, existingRack, RackList, UserName, Password)
		if err != nil {
			return
		}
	}

	ipAddress := make([]string, 0, 0)

	for _, rack := range RackList {
		ipAddress = append(ipAddress, rack.IP1)
		ipAddress = append(ipAddress, rack.IP2)
	}
	addDeviceResponse, err = sh.fetchFabricDetails(ctx, FabricName, ipAddress)
	if err != nil {
		return
	}
	LOG.Infoln("Already registered Pairs", existingRack)

	addDeviceResponse, err = sh.pairsAlreadyRegisteredWithDifferentRack(RackList, existingRack)
	if err != nil {
		return
	}

	LOG.Infoln("Existing Rack", existingRack)
	LOG.Infoln("RackList", RackList)

	//Combined list to be sent for configuring
	totalPairList := sh.getUniqueRackList(ctx, existingRack, RackList)
	LOG.Infoln("totalPairList", totalPairList)

	if len(totalPairList) == 0 {
		AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)
		statusMsg := "No devices are in the fabric"
		err = errors.New(statusMsg)
		LOG.Errorln(err)
		deviceResponse := AddDeviceResponse{IPAddress: "", Errors: []error{err}}
		AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		return AddDeviceResponseList, err
	}

	if len(totalPairList) > 4 {
		errMsg := errors.New("Maximum Configurable Racks cannot be more than 4")
		LOG.Errorln(errMsg)

		AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)
		statusMsg := fmt.Sprint("Maximum Configurable Racks cannot be more than 4. Existing Racks: ", existingRack)

		deviceResponse := AddDeviceResponse{IPAddress: fmt.Sprint(RackList), Role: RackRole, Errors: []error{errors.New(statusMsg)}}
		AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		return AddDeviceResponseList, errors.New("Maximum Configurable Racks cannot be more than 4")
	}

	stageFunctions := []rackstageFunction{sh.addSingleRackFirstStage, sh.addSingleRackSecondStage,
		sh.addSingleRackThirdStage, sh.addSingleRackFourthStage}

	//To achieve rollback, set the boolean to false and return from the function
	RollBack := false

	dberr := sh.Db.OpenTransaction()
	if dberr != nil {
		err = errors.New("Failed to Open Transaction")
		return
	}
	defer sh.CloseTransaction(ctx, &RollBack)

	// Create or update device with correct credentials.
	for _, Pair := range RackList {
		LOG.Infoln("Create Spine Device : ", Pair.IP1)
		var id1, id2 uint
		if id1, err = sh.CreateDevice(FabricName, Pair.IP1, UserName, Password, RackRole); err != nil {
			LOG.Errorln(err.Error())
			return
		}
		if id2, err = sh.CreateDevice(FabricName, Pair.IP2, UserName, Password, RackRole); err != nil {
			LOG.Errorln(err.Error())
			return
		}
		if _, err = sh.CreateRack(FabricName, Pair.IP1, id1, Pair.IP2, id2); err != nil {
			LOG.Errorln(err.Error())
			return
		}
	}

	if force {
		existingipaddress := make([]string, 0)
		for _, exR := range existingRack {
			existingipaddress = append(existingipaddress, exR.IP1)
			existingipaddress = append(existingipaddress, exR.IP2)
		}
		// DeviceCleanupRequired argument in DeleteDevices :
		// Device should not be cleaned here(due to force option) because the same information is used for configure again.
		// If we cleanup device, then credentials will be lost.
		Error := sh.FlushDevices(ctx, FabricName, existingipaddress)
		if Error.Error != nil {
			AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)
			var Response AddDeviceResponse
			Response.FabricName = FabricName
			Response.FabricID = sh.FabricID
			Response.IPAddress = Error.Host
			Response.Errors = []error{Error.Error}
			AddDeviceResponseList = append(AddDeviceResponseList, Response)
			return AddDeviceResponseList, errors.New("Force Clear failed to clear DB")
		}

		//Re-create the racks for Force Clear
		for _, exR := range existingRack {
			ip1, err1 := sh.Db.GetDevice(FabricName, exR.IP1)
			ip2, err2 := sh.Db.GetDevice(FabricName, exR.IP2)
			if err1 == nil && err2 == nil {
				sh.CreateRack(FabricName, exR.IP1, ip1.ID, exR.IP2, ip2.ID)
			}
		}
	}

	UserName = ""
	Password = ""
	for index, stageFunction := range stageFunctions {
		LOG.Println("Executing Stage ", index+1)
		addDeviceResponse, err = sh.executeAddRackStage(ctx, FabricName, totalPairList, UserName,
			Password, force, stageFunction)
		if err != nil {
			//Operation failed so Rollback the Database
			RollBack = true
			break
		}
	}

	return
}

func (sh *DeviceInteractor) fetchRegisteredRacks(ctx context.Context, FabricName string) ([]Rack, error) {
	RackList := make([]Rack, 0, 0)

	fabric, err := sh.Db.GetFabric(FabricName)
	if err != nil {
		return RackList, err
	}
	Racks, err := sh.Db.GetRacksInFabric(fabric.ID)
	if err != nil {
		return RackList, err
	}
	for _, dbrack := range Racks {

		rack := Rack{IP1: dbrack.DeviceOneIP, IP2: dbrack.DeviceTwoIP}
		RackList = append(RackList, rack)
	}
	return RackList, nil
}

//NonClosGetNextRackNumber to get the next available rack number
func (sh *DeviceInteractor) NonClosGetNextRackNumber(FabricName string) (uint, error) {
	var err error
	var racks []domain.Rack
	var rackNo uint
	var rackNumber = uint(1)
	var rackFormat = constants.RackNameSuffix + "%x"

	if racks, err = sh.Db.GetRackAll(FabricName); err == nil {
		for _, rack := range racks {
			fmt.Sscanf(rack.RackName, rackFormat, &rackNo)
			if rackNo == rackNumber {
				rackNumber = rackNumber + 1
			} else {
				break
			}
		}
	}
	return rackNumber, err
}

//CreateRack for creating a Rack if it does not exist
func (sh *DeviceInteractor) CreateRack(FabricName string, IP1 string, ID1 uint, IP2 string, ID2 uint) (id uint, err error) {
	err = nil
	//check for existing Rack
	var Rack domain.Rack
	var rackNumber uint
	if Rack, err = sh.Db.GetRack(FabricName, IP1, IP2); err == nil {
		sh.Refresh = true
	} else {
		rackNumber, _ = sh.NonClosGetNextRackNumber(FabricName)
		Rack.RackName = constants.RackNameSuffix + fmt.Sprint(rackNumber)
	}
	//LOG.Println("Rack Already present", Rack)

	Rack.DeviceOneIP = IP1
	Rack.DeviceOneID = ID1
	Rack.DeviceTwoIP = IP2
	Rack.DeviceTwoID = ID2
	Rack.FabricID = sh.FabricID

	//Save Rack
	if err = sh.Db.CreateRack(&Rack); err != nil {
		statusMsg := fmt.Sprintf("Rack %s %s create Failed with error : %s", IP1, IP2, err.Error())
		err = errors.New(statusMsg)
	}
	return Rack.ID, err
}

func existsInDifferentRack(Rlist []Rack, FirstIP string, SecondIP string) bool {
	for _, rack := range Rlist {
		if rack.IP1 == FirstIP && rack.IP2 != SecondIP {
			return true
		}
		if rack.IP2 == FirstIP && rack.IP1 != SecondIP {
			return true
		}
	}
	return false
}
func (sh *DeviceInteractor) pairsAlreadyRegisteredWithDifferentRack(RackList []Rack,
	existingRackList []Rack) ([]AddDeviceResponse, error) {
	var overallError error
	AddDeviceResponseList := make([]AddDeviceResponse, 0, 0)

	//Check leaf in the existing Spine Map
	for _, rack := range RackList {
		if existsInDifferentRack(existingRackList, rack.IP1, rack.IP2) {
			//Device already registered with different Rack
			overallError = errors.New(fmt.Sprintln(rack.IP1, "already present in another Rack"))
			deviceResponse := AddDeviceResponse{IPAddress: rack.IP1, Errors: []error{overallError}}
			AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		}
		if existsInDifferentRack(existingRackList, rack.IP2, rack.IP1) {
			//Device already registered with different Rack
			overallError = errors.New(fmt.Sprintln(rack.IP2, "already present in another Rack"))
			deviceResponse := AddDeviceResponse{IPAddress: rack.IP2, Errors: []error{overallError}}
			AddDeviceResponseList = append(AddDeviceResponseList, deviceResponse)
		}

	}

	return AddDeviceResponseList, overallError
}

func (sh *DeviceInteractor) executeAddRackStage(ctx context.Context, FabricName string, RackList []Rack,
	UserName string, Password string, force bool, function rackstageFunction) ([]AddDeviceResponse, error) {

	var fabricGate sync.WaitGroup

	AddDeviceResponseList := make([]AddDeviceResponse, 0, len(RackList))

	//Concurrent Execution of Adding Spines and Leaves
	ResultChannel := make(chan AddDeviceResponse, len(RackList))
	for _, rack := range RackList {
		fabricGate.Add(1)
		go function(ctx, &fabricGate, ResultChannel, FabricName, rack, UserName, Password)
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
		return AddDeviceResponseList, errors.New("Add Rack Operation Failed")
	}
	return AddDeviceResponseList, nil
}

func (sh *DeviceInteractor) addSingleRackFirstStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, Rack Rack, UserName string, Password string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddRackFirstStage(ctx, FabricName, Rack.IP1, Rack.IP2, UserName, Password)
	IPAddress := fmt.Sprintln(Rack.IP1, ",", Rack.IP2)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: RackRole}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", RackRole, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", RackRole, IPAddress))
	ResultChannel <- Response
}
func (sh *DeviceInteractor) addSingleRackSecondStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, Rack Rack, UserName string, Password string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddRackSecondStage(ctx, FabricName, Rack.IP1, Rack.IP2, UserName, Password)
	IPAddress := fmt.Sprintln(Rack.IP1, ",", Rack.IP2)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: RackRole}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", RackRole, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", RackRole, IPAddress))
	ResultChannel <- Response
}
func (sh *DeviceInteractor) addSingleRackThirdStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, Rack Rack, UserName string, Password string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddRackThirdStage(ctx, FabricName, Rack.IP1, Rack.IP2, UserName, Password)
	IPAddress := fmt.Sprintln(Rack.IP1, ",", Rack.IP2)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: RackRole}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", RackRole, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", RackRole, IPAddress))
	ResultChannel <- Response
}

func (sh *DeviceInteractor) addSingleRackFourthStage(ctx context.Context, fabricGate *sync.WaitGroup, ResultChannel chan AddDeviceResponse, FabricName string, Rack Rack, UserName string, Password string) {
	defer fabricGate.Done()

	var buffer bytes.Buffer

	err := sh.AddRackFourthStage(ctx, FabricName, Rack.IP1, Rack.IP2, UserName, Password)
	IPAddress := fmt.Sprintln(Rack.IP1, ",", Rack.IP2)
	Response := AddDeviceResponse{IPAddress: IPAddress, FabricName: FabricName, FabricID: sh.FabricID, Role: RackRole}
	if err != nil {
		buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Failed]\n", RackRole, IPAddress))
		buffer.WriteString(err.Error() + "\n")
		Response.Errors = []error{err}
		ResultChannel <- Response
		return
	}
	buffer.WriteString(fmt.Sprintf("Addition of %s device with ip-address = %s [Succeeded]\n", RackRole, IPAddress))
	ResultChannel <- Response
}

func (sh *DeviceInteractor) getUniqueRackList(ctx context.Context, existingRList []Rack, newRList []Rack) []Rack {
	GetIPAddressKey := func(data interface{}) string {
		s, _ := data.(Rack)
		return fmt.Sprintln(s.IP1, s.IP2)
	}
	List := func(s util.Interface) []Rack {
		m := s.GetMap()
		s.List()
		list := make([]Rack, 0, 0)

		for _, elem := range m {
			intf, _ := elem.(Rack)
			list = append(list, intf)
		}

		return list
	}
	set := util.NewSet(GetIPAddressKey)
	for _, elem := range existingRList {
		set.Add(elem)
	}
	for _, elem := range newRList {
		set.Add(elem)
	}

	return List(set)
}

//ValidateNonClosFabricTopology validate the non-clos fabric topology
func (sh *DeviceInteractor) ValidateNonClosFabricTopology(ctx context.Context, FabricName string) (ValidateFabricResponse, error) {

	//LOG := appcontext.Logger(ctx)
	FabricValidateResponse := ValidateFabricResponse{FabricName: FabricName}

	existingRack, err := sh.fetchRegisteredRacks(ctx, FabricName)
	if err != nil {
		//Only Generic errors are returned
		return FabricValidateResponse, err
	}
	defer sh.cleanupLLDPLinks(sh.FabricID)

	FabricValidateResponse.MissingLinks = make([]string, 0)
	RDMapTable := sh.buildRackDeviceMapTable(ctx, FabricName, existingRack)

	for _, RDMap := range RDMapTable {
		if err := sh.ValidateNonClosRack(ctx, RDMap, RDMapTable); err != nil {
			FabricValidateResponse.MissingLinks = append(FabricValidateResponse.MissingLinks, fmt.Sprint(err))
		}
	}
	return FabricValidateResponse, nil
}

//ValidateNonClosRack validate a rack in a non-clos fabric topology
func (sh *DeviceInteractor) ValidateNonClosRack(ctx context.Context, ThisRDMap RackDeviceMapTable,
	RDMapTable []RackDeviceMapTable) error {
	var isMctConnectedPortOne bool
	var isMctConnectedPortTwo bool
	var isBackuplinkConnected bool
	var StatusBuffer bytes.Buffer

	//get all LLDP neighbors between the devices on this Rack
	LLDPNeighbors, err := sh.Db.GetLLDPNeighborsBetweenTwoDevices(sh.FabricID, ThisRDMap.Devices[0].Device.ID,
		ThisRDMap.Devices[1].Device.ID)
	if err != nil {
		StatusBuffer.WriteString(fmt.Sprintf("device %s not connected to device %s",
			ThisRDMap.Devices[0].Device.IPAddress, ThisRDMap.Devices[1].Device.IPAddress))
		return errors.New(StatusBuffer.String())
	}
	isMctConnectedPortOne = false
	isMctConnectedPortTwo = false
	isBackuplinkConnected = false
	for _, LLDPNeighbor := range LLDPNeighbors {
		if LLDPNeighbor.InterfaceOneName == MCTInterfaceOne || LLDPNeighbor.InterfaceTwoName == MCTInterfaceOne {
			if LLDPNeighbor.InterfaceOneName == MCTInterfaceOne && LLDPNeighbor.InterfaceTwoName == MCTInterfaceOne {
				isMctConnectedPortOne = true
			} else {
				isMctConnectedPortOne = false
			}
		}
		if LLDPNeighbor.InterfaceOneName == MCTInterfaceTwo || LLDPNeighbor.InterfaceTwoName == MCTInterfaceTwo {
			if LLDPNeighbor.InterfaceOneName == MCTInterfaceTwo && LLDPNeighbor.InterfaceTwoName == MCTInterfaceTwo {
				isMctConnectedPortTwo = true
			} else {
				isMctConnectedPortTwo = false
			}
		}
		if LLDPNeighbor.InterfaceOneName == L3LoopBackInterface || LLDPNeighbor.InterfaceTwoName == L3LoopBackInterface {
			if LLDPNeighbor.InterfaceOneName == L3LoopBackInterface && LLDPNeighbor.InterfaceTwoName == L3LoopBackInterface {
				isBackuplinkConnected = true
			} else {
				isBackuplinkConnected = false
			}
		}
	}
	if isMctConnectedPortOne == false {
		StatusBuffer.WriteString(fmt.Sprintf("Device %s is not connected to device %s on Mct port %s",
			ThisRDMap.Devices[0].Device.IPAddress, ThisRDMap.Devices[1].Device.IPAddress, MCTInterfaceOne))
		return errors.New(StatusBuffer.String())
	}
	if isMctConnectedPortTwo == false {
		StatusBuffer.WriteString(fmt.Sprintf("Device %s is not connected to device %s on Mct port %s",
			ThisRDMap.Devices[0].Device.IPAddress, ThisRDMap.Devices[1].Device.IPAddress, MCTInterfaceTwo))
		return errors.New(StatusBuffer.String())
	}
	if isBackuplinkConnected == false {
		StatusBuffer.WriteString(fmt.Sprintf("Device %s is not connected to device %s on L3 backup port %s",
			ThisRDMap.Devices[0].Device.IPAddress, ThisRDMap.Devices[1].Device.IPAddress, L3LoopBackInterface))
		return errors.New(StatusBuffer.String())
	}

	if len(RDMapTable) == 1 {
		// single rack, no further vallidation is required
		return nil
	}
	isRackConnected := false
	for _, RDMap := range RDMapTable {
		if ThisRDMap.Rack.ID == RDMap.Rack.ID {
			continue
		}
		// ThisRack has to be connected to at least one more rack in the fabric
		if sh.ValidateRacksConnected(ctx, ThisRDMap, RDMap) {
			isRackConnected = true
			break
		}
	}
	if isRackConnected == false {
		StatusBuffer.WriteString(fmt.Sprintf("rack[devices :%s, %s] is not connected to any other rack",
			ThisRDMap.Devices[0].Device.IPAddress, ThisRDMap.Devices[1].Device.IPAddress))
		return errors.New(StatusBuffer.String())
	}
	return nil
}

//ValidateRacksConnected validate if a rack is connected to at least one rack in a non-clos fabric topology
func (sh *DeviceInteractor) ValidateRacksConnected(ctx context.Context, RackOne RackDeviceMapTable,
	RackTwo RackDeviceMapTable) bool {
	lldneighborsetone, _ := sh.Db.GetLLDPNeighborsBetweenTwoDevices(sh.FabricID, RackOne.Devices[0].Device.ID,
		RackTwo.Devices[0].Device.ID)
	lldneighborsettwo, _ := sh.Db.GetLLDPNeighborsBetweenTwoDevices(sh.FabricID, RackOne.Devices[0].Device.ID,
		RackTwo.Devices[1].Device.ID)
	lldneighborsetthree, _ := sh.Db.GetLLDPNeighborsBetweenTwoDevices(sh.FabricID, RackOne.Devices[1].Device.ID,
		RackTwo.Devices[0].Device.ID)
	lldneighborsetfour, _ := sh.Db.GetLLDPNeighborsBetweenTwoDevices(sh.FabricID, RackOne.Devices[1].Device.ID,
		RackTwo.Devices[1].Device.ID)
	//each rack should be connected to
	if len(lldneighborsetone) == 0 && len(lldneighborsettwo) == 0 &&
		len(lldneighborsetthree) == 0 && len(lldneighborsetfour) == 0 {
		return false
	}
	return true
}

func (sh *DeviceInteractor) forceClearofRacks(ctx context.Context, FabricName string, existingRacks []Rack, racksToClear []Rack,
	UserName string, Password string) ([]AddDeviceResponse, error) {
	var err error
	existingipaddress := make([]string, 0)
	ipaddressToClear := make([]string, 0)

	for _, exR := range existingRacks {
		existingipaddress = append(existingipaddress, exR.IP1)
		existingipaddress = append(existingipaddress, exR.IP2)
	}
	for _, ipR := range racksToClear {
		ipaddressToClear = append(ipaddressToClear, ipR.IP1)
		ipaddressToClear = append(ipaddressToClear, ipR.IP2)
	}

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

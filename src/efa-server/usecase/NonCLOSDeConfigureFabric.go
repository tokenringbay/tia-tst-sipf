package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

/*DeleteDevicesFromNonCLOSFabric performs:
	Step 1 : Validation - if all devices are part of the fabric.
   	Step 2 : Cleanup the configurations from the switch.
   	Step 3 : Delete the devices from fabric.
*/
func (sh *DeviceInteractor) DeleteDevicesFromNonCLOSFabric(ctx context.Context, FabricName string, RackList []Rack,
	UserName string, Password string, force bool, persist bool, devCleanUp bool) ([]AddDeviceResponse, error) {
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Delete Device")
	LOG := appcontext.Logger(ctx)

	var fabricGate sync.WaitGroup
	ips := ""
	ipList := make([]string, 0)
	for _, rack := range RackList {
		ips += " " + rack.IP1 + "," + rack.IP2
		ipList = append(ipList, rack.IP1, rack.IP2)
	}

	AddDeviceResponseList := make([]AddDeviceResponse, 0, len(ipList))
	AddDeviceResponseListError := make([]AddDeviceResponse, 0, len(ipList))

	// fetch existing racks and compare with RackList to check if the user provided pairs
	// are existing
	existingRacks, err := sh.fetchRegisteredRacks(ctx, FabricName)

	if err != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch Rack Details  %s", sh.FabricName)
		LOG.Infoln(statusMsg)
	}

	if len(existingRacks) == 0 {
		statusMsg := fmt.Sprintf("Rack(s) doesnt exist in the Fabric %s", ips)
		LOG.Errorln(statusMsg)
		Response := AddDeviceResponse{}
		Response.Role = "Rack"
		Response.IPAddress = ips
		Response.Errors = []error{errors.New(statusMsg)}
		AddDeviceResponseListError = append(AddDeviceResponseListError, Response)
		return AddDeviceResponseListError, errors.New("Rack(s) Validation [Failed]")
	}

	invIP := ""
	for _, rack := range RackList {
		found := false
		tmpIP := rack.IP1 + "," + rack.IP2
		for _, existingRack := range existingRacks {
			if strings.Contains(tmpIP, existingRack.IP1) && strings.Contains(tmpIP, existingRack.IP2) {
				found = true
				break
			}
		}
		if !found {
			invIP += " " + tmpIP + "\n"
		}
	}

	if invIP != "" {
		statusMsg := fmt.Sprintf("Invalid Rack IP Pairs Found %s", invIP)
		LOG.Errorln(statusMsg)
		Response := AddDeviceResponse{}
		Response.Role = "Rack"
		Response.IPAddress = invIP
		Response.Errors = []error{errors.New(statusMsg)}
		AddDeviceResponseListError = append(AddDeviceResponseListError, Response)
		return AddDeviceResponseListError, errors.New("Rack(s) Validation [Failed]")
	}

	/*FabricProperties, err := sh.GetFabricSettings(ctx, FabricName)
	if err != nil {
		statusMsg := fmt.Sprintf("Unable to Fetch fabric Propertis  %s", sh.FabricName)
		LOG.Infoln(statusMsg)
		return AddDeviceResponseList, err
	}
	sh.FabricProperties = FabricProperties*/
	ResultChannel := make(chan AddDeviceResponse, len(ipList))
	// Concurrent validation of Devices
	for _, SwitchIPAddress := range ipList {
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
		Errors := sh.cleanupDevicesInNonCLOSFabric(ctx, FabricName, ipList, force, persist)
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

		SecondaryCleanupErrors := sh.cleanupRelatedConfigsonDependentSwitches(ctx, ipList, force, persist)
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

	Error := sh.DeleteDevices(ctx, FabricName, ipList, true)
	if Error.Error != nil {
		var Response AddDeviceResponse
		Response.FabricName = FabricName
		Response.FabricID = sh.FabricID
		Response.IPAddress = Error.Host
		Response.Errors = []error{Error.Error}
		AddDeviceResponseListError = append(AddDeviceResponseListError, Response)
		return AddDeviceResponseListError, errors.New("Delete Device Failed")
	}

	//On Success backup the DB
	if bkerr := sh.Db.Backup(); bkerr != nil {
		LOG.Printf("Failed to backup DB during Deconfigure %s\n", bkerr)
	}
	return AddDeviceResponseList, nil
}

func (sh *DeviceInteractor) cleanupRelatedConfigsonDependentSwitches(ctx context.Context, DeletedDevices []string, force bool,
	persist bool) []actions.OperationError {
	LOG := appcontext.Logger(ctx)
	Errors := make([]actions.OperationError, 0)
	DeleteDevicesList, err := sh.Db.GetDevicesInFabricMatching(sh.FabricID, DeletedDevices)
	DeletedDevicesIDList := make([]uint, 0)
	for _, dev := range DeleteDevicesList {
		DeletedDevicesIDList = append(DeletedDevicesIDList, dev.ID)
	}

	RemainingDevicesList, err := sh.Db.GetDevicesInFabricNotMatching(sh.FabricID, DeletedDevices)
	LOG.Println("RemainingDevicesList", RemainingDevicesList, err)

	//No extra devices remaining other than deleted devices
	if len(RemainingDevicesList) == 0 {
		return Errors
	}

	var config operation.ConfigFabricRequest
	config.Hosts = make([]operation.ConfigSwitch, 0)
	for _, rDev := range RemainingDevicesList {

		configSwitch := operation.ConfigSwitch{}
		configSwitch.Device = rDev.IPAddress
		configSwitch.Host = rDev.IPAddress
		configSwitch.UserName = rDev.UserName
		configSwitch.Password = rDev.Password
		configSwitch.Model = rDev.Model

		bgpConfigs, err := sh.Db.GetBGPSwitchConfigsOnRemoteDeviceID(sh.FabricID, rDev.ID, DeletedDevicesIDList)
		LOG.Println("BGPConfigs to be deleted for Device ", rDev.IPAddress, bgpConfigs, err)
		configSwitch.BgpNeighbors = make([]operation.ConfigBgpNeighbor, 0)
		for _, bgp := range bgpConfigs {
			neighbor := operation.ConfigBgpNeighbor{}
			neighbor.NeighborAddress = bgp.RemoteIPAddress
			neighbor.RemoteAs, _ = strconv.ParseInt(bgp.RemoteAS, 10, 64)
			neighbor.ConfigType = domain.ConfigDelete
			neighbor.NeighborType = bgp.Type
			configSwitch.BgpNeighbors = append(configSwitch.BgpNeighbors, neighbor)
		}

		evpnConfigs, err := sh.Db.GetRackEvpnConfigOnRemoteDeviceID(rDev.ID, DeletedDevicesIDList)
		LOG.Println("evpnConfigs to be deleted for Device ", rDev.IPAddress, evpnConfigs, err)
		for _, bgp := range evpnConfigs {
			neighbor := operation.ConfigBgpNeighbor{}
			neighbor.NeighborAddress = bgp.EVPNAddress
			neighbor.RemoteAs, _ = strconv.ParseInt(bgp.RemoteAS, 10, 64)
			neighbor.ConfigType = domain.ConfigDelete
			neighbor.NeighborType = domain.EVPNENIGHBORType
			configSwitch.BgpNeighbors = append(configSwitch.BgpNeighbors, neighbor)
		}

		lldpneighbors, err := sh.Db.GetLLDPNeighborsOnRemoteDeviceID(sh.FabricID, rDev.ID, DeletedDevicesIDList)
		LOG.Println("lldpneighbors to be deleted for Device ", rDev.IPAddress, lldpneighbors, err)
		configSwitch.Interfaces = make([]operation.ConfigInterface, 0)
		for _, lldp := range lldpneighbors {
			Interface := operation.ConfigInterface{}
			Interface.InterfaceName = lldp.InterfaceOneName
			Interface.InterfaceType = lldp.InterfaceOneType
			configSwitch.Interfaces = append(configSwitch.Interfaces, Interface)
		}

		config.Hosts = append(config.Hosts, configSwitch)

	}

	Errors = sh.FabricAdapter.CleanupDependantDevicesInFabric(ctx, config, force, persist)

	return Errors

}

func (sh *DeviceInteractor) cleanupDevicesInNonCLOSFabric(ctx context.Context, FabricName string, DevicesList []string,
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

	Errors = sh.FabricAdapter.CleanupDevicesInNonCLOSFabric(ctx, config, force, persist)

	return Errors
}

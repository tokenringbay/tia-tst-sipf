package fabric

import (
	"context"
	"efa/infra/cli/commands/fabric/settings"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	openAPI "efa/infra/rest/generated/client"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	deleteIPAddress string
	deleteRack      []string
	delpersist      bool
	nodevCleanUp    bool
)

//DeconfigureSwitchCommand provides command to delete Switches in the Fabric
var DeconfigureSwitchCommand = &cobra.Command{
	Use:   "deconfigure",
	Short: "Deconfigure IP Fabric from the device",
	RunE:  utils.TimedRunE(runDelete),
}

func init() {

	fabricName := constants.DefaultFabric
	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)
	response, _, _ := api.FabricApi.GetFabric(context.Background(), fabricName)
	FabricProperties, err := settings.CreateFromMap(response.FabricSettings)
	if err != nil {
		fmt.Println(err)
	}

	if FabricProperties.FabricType == utils.CLOSFabricType {
		DeconfigureSwitchCommand.Flags().StringVar(&deleteIPAddress, "device", "", "Comma separated list of IP Address/Hostnames of devices")
	} else {
		DeconfigureSwitchCommand.Flags().StringArrayVar(&deleteRack, "rack", []string{}, "Comma separated addresses/host-names for non-clos fabric")
	}
	DeconfigureSwitchCommand.Flags().BoolVar(&nodevCleanUp, "no-device-cleanup", false, "Do not cleanup the configurations on the devices")
	DeconfigureSwitchCommand.Flags().BoolVar(&delpersist, "persist", false, "Persist the configuration on the devices")
}

func runDelete(cmd *cobra.Command, args []string) error {
	devCleanUp := !nodevCleanUp
	if len(args) != 0 {
		fmt.Println("Additional arguments passed to the command or space present in the list of device ip address.")
		return nil
	}

	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)
	response, _, err := api.FabricApi.GetFabric(context.Background(), constants.DefaultFabric)
	if err != nil {
		handleConfigShowErrorResponse(err)
		return nil
	}

	cfg = openAPI.NewConfiguration()
	api = openAPI.NewAPIClient(cfg)
	DelSwitchReq := openAPI.DeleteSwitchesRequest{}
	if response.FabricSettings["FabricType"] == utils.NonCLOSFabricType {
		if len(deleteRack) == 0 {
			return errors.New("Required at least one rack address pair to be deleted")
		}
		if len(deleteIPAddress) > 0 {
			return errors.New("Device address should be provided only for CLOS fabric")
		}
		// Input Racks cannot be more than 4
		if len(deleteRack) > 4 {
			return errors.New("Only 4 Rack Pairs are supported")
		}

		//Extract the RackIP Address and set it in the CLI
		delRacks := make([]openAPI.Rack, len(deleteRack))

		for _, rack := range deleteRack {
			rackDevices := strings.Split(rack, ",")
			if !utils.IsValidIPs(rackDevices) {
				return errors.New("Some of the rack IP's are invalid")
			}
			if len(rackDevices) == 1 || len(rackDevices) > 2 {
				return errors.New("Rack must contain pair of IP Addresses")
			}
			Rack := openAPI.Rack{RackDevices: rackDevices}
			delRacks = append(delRacks, Rack)
		}

		// Check for redundant IPs in/across the racks
		if validateRackIps(deleteRack) {
			return errors.New("Rack IPs must be unique")
		}

		DelSwitchReq = openAPI.DeleteSwitchesRequest{Racks: delRacks, DeviceCleanup: devCleanUp, Persist: delpersist}
	} else {
		if len(deleteIPAddress) == 0 {
			return errors.New("Required at least one device ip address to be deleted")
		}
		if len(deleteRack) > 0 {
			return errors.New("Rack address should be provided only for NON-CLOS fabric")
		}

		var devices []string
		devices = strings.Split(deleteIPAddress, ",")
		if !utils.IsValidIPs(devices) {
			return errors.New("Some of the device IP's are invalid")
		}
		DelSwitchReq = openAPI.DeleteSwitchesRequest{Switches: devices, DeviceCleanup: devCleanUp, Persist: delpersist}
	}

	SwitchesdataResponse, _, err := api.SwitchesApi.DeleteSwitches(context.Background(),
		map[string]interface{}{"switches": DelSwitchReq})
	if err != nil {
		handleDeleteSwitchesErrorResponse(err)
		return nil
	}
	handleDeleteSwitchesResponse(&SwitchesdataResponse)

	if response.FabricSettings["FabricType"] == utils.NonCLOSFabricType {
		//If device Cleanup
		if devCleanUp {
			//Validation Routine called for NonCLOSFabricType
			//Second Send Request for Validating the fabric
			FabricValidateResponse, _, err := api.FabricValidationApi.ValidateFabric(context.Background(), constants.DefaultFabric)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			//Stop further processing if Validate devices has any error
			if err = handleValidateResponse(&FabricValidateResponse, "Warning"); err != nil {
				return nil
			}
		}
	}

	return nil
}

func handleDeleteSwitchesErrorResponse(errorObject error) {
	//OpenAPI Generated code sends the message as an error string, so parsing output from string object
	//Body Contains the Error Obect in JSON
	fmt.Println("Delete Device(s) [Failed]")
	if utils.IsServerConnectionError(errorObject) {
		return
	}
	errorMessageList := strings.Split(errorObject.Error(), "Body:")
	if len(errorMessageList) == 2 {
		var StatusModelList []openAPI.DeviceStatusModel
		err := json.Unmarshal([]byte(errorMessageList[1]), &StatusModelList)
		if err == nil {
			for _, errorResponse := range StatusModelList {

				if len(errorResponse.Error_) > 0 {
					fmt.Printf("\tDeletion of %s device(s) with ip-address = %s [Failed]", errorResponse.Role, errorResponse.IpAddress)
					fmt.Println("\n\tErrors:")
					for _, errorResponse := range errorResponse.Error_ {
						fmt.Println("\t" + errorResponse.Message)
					}
				} else {
					fmt.Printf("\tDeletion of %s device(s) with ip-address = %s [Succeeded]\n", errorResponse.Role, errorResponse.IpAddress)
				}
			}
		}
	} else {
		//Generic Error, Just print it
		fmt.Println("\t" + errorObject.Error())
	}
}

func handleDeleteSwitchesResponse(SwitchesdataResponse *openAPI.SwitchesdataResponse) {
	//Fetch the  Responses for Add Switches  and Display
	fmt.Println("Delete Device(s) [Success]")
	for _, switchResponse := range SwitchesdataResponse.Items {
		fmt.Printf("\tDeletion of %s device with ip-address = %s [Succeeded]\n", switchResponse.Role, switchResponse.IpAddress)
	}
}

func validateRackIps(deleteRack []string) bool {
	invIP := ""
	for _, rack := range deleteRack {
		found := false
		ipPair := strings.Split(rack, ",")
		if ipPair[0] == ipPair[1] {
			found = true
			return found
		}
		if strings.Contains(invIP, ipPair[0]) || strings.Contains(invIP, ipPair[1]) {
			found = true
			return found
		}
		invIP += ipPair[0] + "," + ipPair[1]
	}
	return false
}

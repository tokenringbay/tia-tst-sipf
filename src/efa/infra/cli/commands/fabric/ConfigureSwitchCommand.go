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
	"time"
)

var (
	spineIPaddress string
	leafIPaddress  string
	rackIPaddress  []string
	username       string
	password       string
	force          bool
	persist        bool
)

//ConfigureSwitchCommand provides command to add/update devices in fabric
var ConfigureSwitchCommand = &cobra.Command{
	Use:   "configure",
	Short: "Configure IP Fabric on the device",
	RunE:  utils.TimedRunE(runAddSwitch),
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
		ConfigureSwitchCommand.Flags().StringVar(&spineIPaddress, "spine", "", "Comma separated list of spine IP Address/Hostnames")
		ConfigureSwitchCommand.Flags().StringVar(&leafIPaddress, "leaf", "", "Comma separated list of spine IP Address/Hostnames")
	} else {
		ConfigureSwitchCommand.Flags().StringArrayVar(&rackIPaddress, "rack", []string{}, "Comma separated address/host-names for non-clos fabric")
	}
	ConfigureSwitchCommand.Flags().StringVar(&username, "username", "", "Username for the list of devices")
	ConfigureSwitchCommand.Flags().StringVar(&password, "password", "", "Password for the list of devices")
	ConfigureSwitchCommand.Flags().BoolVar(&force, "force", false, "Force the configuration on the devices")
	ConfigureSwitchCommand.Flags().BoolVar(&persist, "persist", false, "Persist the configuration on the devices")
}

//runAddSwitch is implemented using three Rest CALLs.
//First calls CreateSwitches which adds all the switches to the fabric
//Next calls ValidateFabric which validates the Fabric Topology
//Next calls Configure Fabric which Configures Fabric
func runAddSwitch(cmd *cobra.Command, args []string) error {

	if len(args) != 0 {
		fmt.Println("Additional arguments passed to the command or space present in the list of device ip address.")
		return nil
	}

	if username == "root" {
		fmt.Println("\"root\" user cannot be used to manage switches.")
		return nil
	}

	//Using the Default Fabric Name
	NewSwitches := openAPI.NewSwitches{Fabric: constants.DefaultFabric}

	if (len(username) == 0 && len(password) != 0) || (len(username) != 0 && len(password) == 0) {
		return errors.New("Required both flags \"username\" and \"password\"")
	}

	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)
	response, _, err := api.FabricApi.GetFabric(context.Background(), constants.DefaultFabric)
	if err != nil {
		handleConfigShowErrorResponse(err)
		return nil
	}

	if response.FabricSettings["FabricType"] == utils.NonCLOSFabricType {
		if len(spineIPaddress) > 0 || len(leafIPaddress) > 0 {
			return errors.New("Spine and Leaf address should be provided only for CLOS fabric")
		}
		if len(rackIPaddress) == 0 && len(username) != 0 {
			return errors.New("Device Credentials to be provided with Device IP address")
		}
		// Input Racks cannot be more than 4
		if len(rackIPaddress) > 4 {
			return errors.New("Only 4 Rack Pairs are supported")
		}
		//Extract the RackIP Address and set it in the CLI
		NewSwitches.Racks = make([]openAPI.Rack, len(rackIPaddress))
		for _, rack := range rackIPaddress {
			rackDevices := strings.Split(rack, ",")
			if !utils.IsValidIPs(rackDevices) {
				return errors.New("Some of the rack IP's are invalid")
			}
			if len(rackDevices) == 1 || len(rackDevices) > 2 {
				return errors.New("Rack must contain pair of IP Addresses")
			}
			Rack := openAPI.Rack{RackDevices: rackDevices}
			NewSwitches.Racks = append(NewSwitches.Racks, Rack)
		}

		// Check for redundant IPs in/across the racks
		if validateRackIps(rackIPaddress) {
			return errors.New("Rack IPs must be unique")
		}
	} else {
		if len(rackIPaddress) > 0 {
			return errors.New("Rack address should be provided only for NON-CLOS fabric")
		}
		if len(spineIPaddress) > 0 {
			NewSwitches.SpineIpAddress = strings.Split(spineIPaddress, ",")
		}
		if len(leafIPaddress) > 0 {
			NewSwitches.LeafIpAddress = strings.Split(leafIPaddress, ",")
		}
		if len(leafIPaddress) == 0 && len(spineIPaddress) == 0 && len(username) != 0 {
			return errors.New("Device Credentials to be provided with Device IP address")
		}
		if !utils.IsValidIPs(NewSwitches.SpineIpAddress) {
			return errors.New("Some of the spine IP's are invalid")
		}
		if !utils.IsValidIPs(NewSwitches.LeafIpAddress) {
			return errors.New("Some of the leaf IP's are invalid")
		}
	}

	NewSwitches.Username = username
	NewSwitches.Password = password
	NewSwitches.Force = force

	//First Add Switches to the Fabric
	SwitchesdataResponse, _, err := api.SwitchesApi.CreateSwitches(context.Background(),
		map[string]interface{}{"switches": NewSwitches})
	//Stop further processing if Add devices has any error
	if err != nil {
		//Handle error for Create Switches
		handleAddSwitchesErrorResponse(err)
		return nil
	}
	//Handle Success Response Create Switches
	handleAddSwitchesResponse(&SwitchesdataResponse)
	fmt.Println("")

	//Second Send Request for Validating the fabric
	FabricValidateResponse, _, err := api.FabricValidationApi.ValidateFabric(context.Background(), NewSwitches.Fabric)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	//Stop further processing if Validate devices has any error
	if err = handleValidateResponse(&FabricValidateResponse, "Failed"); err != nil {
		return nil
	}

	fmt.Println("")
	//Third send Request for  Configure the Fabric
	ConfigureFabricResponse, _, err := api.ConfigureFabricApi.ConfigureFabric(context.Background(), NewSwitches.Fabric,
		map[string]interface{}{"persist": persist, "force": force})
	if err != nil {
		//Handle Configure Error Response
		handleConfigureErrorResponse(err)
		return nil
	}

	handleConfigureResponse(&ConfigureFabricResponse)

	return nil
}
func timeElapsed(start time.Time) {
	elapsed := time.Since(start)
	fmt.Println("Took ", elapsed)
}

func handleAddSwitchesErrorResponse(errorObject error) {
	//OpenAPI Generated code sends the message as an error string, so parsing output from string object
	//Body Contains the Error Obect in JSON
	fmt.Println("Add Device(s) [Failed]")
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
					//For clear config
					if errorResponse.IpAddress == "" {
						for _, errorResponse := range errorResponse.Error_ {
							fmt.Printf("\t%s\n", errorResponse.Message)
						}
					} else {
						fmt.Printf("\tAddition of %s device with ip-address = %s [Failed]\n", errorResponse.Role, errorResponse.IpAddress)
						for _, errorResponse := range errorResponse.Error_ {
							fmt.Println("\t" + errorResponse.Message)
						}
					}
				} else {
					fmt.Printf("\tAddition of %s device with ip-address = %s [Succeeded]\n", errorResponse.Role, errorResponse.IpAddress)
				}
			}
		}
	} else {
		//Generic Error, Just print it
		fmt.Println("\t" + errorObject.Error())
	}

}
func handleAddSwitchesResponse(SwitchesdataResponse *openAPI.SwitchesdataResponse) {
	//Fetch the  Responses for Add Switches  and Display
	fmt.Println("Add Device(s) [Success]")
	for _, switchResponse := range SwitchesdataResponse.Items {
		fmt.Printf("\tAddition of %s device with ip-address = %s [Succeeded]\n", switchResponse.Role, switchResponse.IpAddress)
	}

}

func handleValidateResponse(FabricValidateResponse *openAPI.FabricValidateResponse, errorType string) error {
	if len(FabricValidateResponse.MissingLinks) > 0 || len(FabricValidateResponse.SpineSpineLinks) > 0 ||
		FabricValidateResponse.MissingLeaves || FabricValidateResponse.MissingSpines ||
		len(FabricValidateResponse.LeafLeafLinks) > 0 {
		fmt.Printf("Validate Fabric [%s]\n", errorType)
		if len(FabricValidateResponse.MissingLinks) > 0 {
			fmt.Println("\t" + "Missing Links")
			for _, links := range FabricValidateResponse.MissingLinks {
				fmt.Println("\t" + links)
			}
		}
		if len(FabricValidateResponse.SpineSpineLinks) > 0 {
			fmt.Println("\t" + "Spine to Spine Links")
			for _, links := range FabricValidateResponse.SpineSpineLinks {
				fmt.Println("\t" + links)
			}
		}
		if len(FabricValidateResponse.LeafLeafLinks) > 0 {
			fmt.Println("\t" + "Leaf to Leaf Links")
			for _, links := range FabricValidateResponse.LeafLeafLinks {
				fmt.Println("\t" + links)
			}
		}
		if FabricValidateResponse.MissingSpines {
			fmt.Println("\tNo Spine Devices")
		}
		if FabricValidateResponse.MissingLeaves {
			fmt.Println("\tNo Leaf Devices")
		}
		return errors.New("Fabric Validation Failed")
	}

	fmt.Println("Validate Fabric [Success]")

	return nil
}

func handleConfigureErrorResponse(errorObject error) {
	//Generated code sends the message as an error string, so parsing output from string object
	fmt.Println("Configure Fabric [Failed]")
	errorMessageList := strings.Split(errorObject.Error(), "Body:")
	if len(errorMessageList) == 2 {
		var StatusModelList []openAPI.DeviceStatusModel
		err := json.Unmarshal([]byte(errorMessageList[1]), &StatusModelList)
		if err == nil {
			for _, errorResponse := range StatusModelList {
				fmt.Printf("\tConfiguration of device with ip-address = %s [Failed]\n", errorResponse.IpAddress)
				for _, errorResponse := range errorResponse.Error_ {
					fmt.Println("\t" + errorResponse.Message)

				}
			}
		}
	} else {
		//Generic Error
		fmt.Println("\t" + errorObject.Error())
	}

}

func handleConfigureResponse(ConfigureFabricResponse *openAPI.ConfigureFabricResponse) error {
	fmt.Println("Configure Fabric [Success]")
	return nil
}

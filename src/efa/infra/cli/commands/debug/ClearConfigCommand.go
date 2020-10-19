package debug

import (
	"fmt"

	openAPI "efa/infra/rest/generated/client"
	"github.com/spf13/cobra"

	"efa/infra/cli/utils"
	"strings"

	"context"
	"errors"

	"encoding/json"
)

var (
	ipaddress   string
	rackaddress []string
	username    string
	password    string
)

//ClearConfigCommand provides command to clear the configuration on a collection of devices
var ClearConfigCommand = &cobra.Command{
	Use:   "clear-config",
	Short: "Clear configuration from device",
	RunE:  utils.TimedRunE(runClearFabric),
}

func init() {
	ClearConfigCommand.Flags().StringVar(&ipaddress, "device", "", "Comma separated list of IP Address/Hostnames of devices")
	//ClearConfigCommand.Flags().StringArrayVar(&rackaddress, "rack", []string{}, "Comma separated addresses/host-names for non-clos fabric")
	ClearConfigCommand.Flags().StringVar(&username, "username", "", "Username for the list of devices")
	ClearConfigCommand.Flags().StringVar(&password, "password", "", "Password for the list of devices")
}

func runClearFabric(cmd *cobra.Command, args []string) error {

	if len(ipaddress) == 0 {
		fmt.Println("Device IP Address/Hostnames has to be specified.")
		return nil
	}
	if len(args) != 0 {
		fmt.Println("Additional arguments passed to the command or space present in the list of device ip address.")
		return nil
	}

	if username == "root" {
		fmt.Println("\"root\" user cannot be used to manage switches.")
		return nil
	}

	ClearRequest := openAPI.DebugClearRequest{}

	if (len(username) == 0 && len(password) != 0) || (len(username) != 0 && len(password) == 0) {
		return errors.New("Require both flags \"username\" and \"password\"")
	}

	if len(ipaddress) > 0 {
		ClearRequest.IpAddress = strings.Split(ipaddress, ",")
	}

	if !utils.IsValidIPs(ClearRequest.IpAddress) {
		return errors.New("Some of the device IP's are invalid")
	}
	ClearRequest.Username = username
	ClearRequest.Password = password

	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)
	_, _, err := api.ClearConfigApi.ClearConfig(context.Background(), map[string]interface{}{"switches": ClearRequest})
	if err != nil {
		handleClearResponse(err)
	} else {
		fmt.Println("Clear Config [Success]")
	}

	return nil
}

func handleClearResponse(errorObject error) {
	//OpenAPI Generated code sends the message as an error string, so parsing output from string object
	//Body Contains the Error Obect in JSON
	fmt.Println("Clear Config [Failed]")
	if utils.IsServerConnectionError(errorObject) {
		return
	}
	errorMessageList := strings.Split(errorObject.Error(), "Body:")

	if len(errorMessageList) == 2 {
		var ErrorModel openAPI.ErrorModel
		err := json.Unmarshal([]byte(errorMessageList[1]), &ErrorModel)
		if err == nil {
			fmt.Println(ErrorModel.Message)
		}
	} else {
		//Generic Error, Just print it
		fmt.Println("\t" + errorObject.Error())
	}

}

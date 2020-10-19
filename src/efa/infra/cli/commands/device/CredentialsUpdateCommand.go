package device

import (
	"context"
	"efa/infra/cli/utils"
	openAPI "efa/infra/rest/generated/client"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	devices  string
	username string
	password string
)

//CredentialsUpdateCommand provides command to update device information/credentials
var CredentialsUpdateCommand = &cobra.Command{
	Use:   "update",
	Short: "Update Credentials of the devices",
	RunE:  utils.TimedRunE(updateSwitchCredentials),
}

func init() {
	CredentialsUpdateCommand.Flags().StringVar(&devices, "device", "", "Comma seperated list of Device IP Address/Hostnames")
	CredentialsUpdateCommand.Flags().StringVar(&username, "username", "", "Username for the list of devices")
	CredentialsUpdateCommand.Flags().StringVar(&password, "password", "", "Password for the list of devices")
	CredentialsUpdateCommand.MarkFlagRequired("device")
	CredentialsUpdateCommand.MarkFlagRequired("username")
	CredentialsUpdateCommand.MarkFlagRequired("password")
}

func updateSwitchCredentials(cmd *cobra.Command, args []string) error {

	if len(args) != 0 {
		fmt.Println("Additional arguments passed to the command or space present in the list of spine/leaf ip address.")
		return nil
	}

	if username == "root" {
		fmt.Println("\"root\" user cannot be used to manage switches.")
		return nil
	}

	//Using the Default Fabric Name
	UpdateSwitchesParams := openAPI.UpdateSwitchParameters{}

	if len(devices) > 0 {
		UpdateSwitchesParams.DeviceIpAddress = strings.Split(devices, ",")
	}

	UpdateSwitchesParams.Username = username
	UpdateSwitchesParams.Password = password

	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)

	//First Add Switches to the Fabric
	SwitchesUpdateResponse, _, err := api.SwitchesApi.UpdateSwitches(context.Background(),
		map[string]interface{}{"switches": UpdateSwitchesParams})
	//Stop further processing if Add devices has any error
	if err != nil {
		if utils.IsServerConnectionError(err) {
			return nil
		}
		//Handle error for Update Switches
		fmt.Println("errored updating the switches", SwitchesUpdateResponse)
		return nil
	}

	for _, switchResponse := range SwitchesUpdateResponse.Items {
		fmt.Println(switchResponse.IpAddress + " : " + switchResponse.DeviceCredentials)
	}

	return nil
}

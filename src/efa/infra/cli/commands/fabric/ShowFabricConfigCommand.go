package fabric

import (
	"context"
	"efa/infra/cli/commands/fabric/settings"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	openAPI "efa/infra/rest/generated/client"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	role string
)

//ShowFabricConfigCommand provides command to fetch configirations for a Fabric
var ShowFabricConfigCommand = &cobra.Command{
	Use:   "show-config",
	Short: "Display IP Fabric config",
	RunE:  utils.TimedRunE(runFabricConfigShow),
}

func prettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func init() {
	role = "all"
	fabricName := constants.DefaultFabric
	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)

	response, _, _ := api.FabricApi.GetFabric(context.Background(), fabricName)
	FabricProperties, err := settings.CreateFromMap(response.FabricSettings)
	if err != nil {
		fmt.Println(err)
	}

	if FabricProperties.FabricType == utils.CLOSFabricType {
		ShowFabricConfigCommand.Flags().StringVar(&role, "device-role", "all", "Filter the config based on device-role(spine/leaf/all)")
	}
}

func runFabricConfigShow(cmd *cobra.Command, args []string) error {
	if len(role) > 0 {
		if role != "all" && role != "spine" && role != "leaf" {
			fmt.Println("Incorrect device-role specified, valid device-role is spine|leaf|all")
			return nil
		}
	}

	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)
	ConfigShowResponse, _, err := api.ConfigShowApi.ConfigShow(context.Background(), constants.DefaultFabric, role)
	if err != nil {
		handleConfigShowErrorResponse(err)
		return nil
	}
	fmt.Println(ConfigShowResponse)
	return nil
}

func handleConfigShowErrorResponse(errorObject error) {
	//OpenAPI Generated code sends the message as an error string, so parsing output from string object
	//Body Contains the Error Obect in JSON
	fmt.Println("Config Show [Failed]")
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

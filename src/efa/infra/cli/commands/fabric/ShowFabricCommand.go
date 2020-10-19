package fabric

import (
	"context"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	openAPI "efa/infra/rest/generated/client"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

//ShowFabricCommand provides command to fetch devices in a Fabric
var ShowFabricCommand = &cobra.Command{
	Use:   "show",
	Short: "Display devices in IP Fabric",
	RunE:  utils.TimedRunE(runFabricShow),
}

func init() {

}

func runFabricShow(cmd *cobra.Command, args []string) error {

	cfg := openAPI.NewConfiguration()
	api := openAPI.NewAPIClient(cfg)

	fabricResponse, _, err := api.FabricApi.GetFabric(context.Background(), constants.DefaultFabric)
	if err != nil {
		handleShowErrorResponse(err)
		return nil
	}

	ShowResponse, _, err := api.SwitchesApi.GetSwitches(context.Background(), constants.DefaultFabric)
	if err != nil {
		handleShowErrorResponse(err)
		return nil
	}

	//Render using Tables
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	if fabricResponse.FabricSettings["FabricType"] == utils.CLOSFabricType {
		table.SetHeader([]string{"IP Address", "Role", "Model", "Firmware"})

		for _, switchRespsonse := range ShowResponse.Items {
			row := []string{switchRespsonse.IpAddress, switchRespsonse.Role,
				switchRespsonse.Model, switchRespsonse.Firmware}
			table.Append(row)
		}
	} else {
		table.SetHeader([]string{"IP Address", "Rack", "Model", "Firmware"})
		for _, switchRespsonse := range ShowResponse.Items {
			row := []string{switchRespsonse.IpAddress, switchRespsonse.Rack,
				switchRespsonse.Model, switchRespsonse.Firmware}
			table.Append(row)
		}
	}
	table.Render()

	return nil
}

func handleShowErrorResponse(errorObject error) {
	//OpenAPI Generated code sends the message as an error string, so parsing output from string object
	//Body Contains the Error Obect in JSON
	fmt.Println("Show [Failed]")
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

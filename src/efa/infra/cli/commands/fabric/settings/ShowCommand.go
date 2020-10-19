package settings

import (
	"context"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	openAPIClient "efa/infra/rest/generated/client"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	fabricName string
	advanced   bool
)

//ShowCommand provides means to display Fabric Properties
var ShowCommand = &cobra.Command{
	Use:   "show",
	Short: "Display IP Fabric setting",
	RunE:  utils.TimedRunE(runFabricShow),
}

func init() {
	ShowCommand.Flags().BoolVar(&advanced, "advanced", false, "List advanced Fabric parameters.")
}

func runFabricShow(cmd *cobra.Command, args []string) error {

	fabricName = constants.DefaultFabric
	var err error
	cfg := openAPIClient.NewConfiguration()
	api := openAPIClient.NewAPIClient(cfg)

	response, _, err := api.FabricApi.GetFabric(context.Background(), fabricName)
	if err != nil {
		handleConfigShowErrorResponse(err)
		return nil
	}
	FabricProperties, err := CreateFromMap(response.FabricSettings)
	if err != nil {
		handleConfigShowErrorResponse(err)
		return nil
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Name", "Value"})
	table.SetRowLine(true)

	table.Append([]string{"Fabric Name ", fabricName})
	table.Append([]string{"Link IP Range", FabricProperties.P2PLinkRange})
	table.Append([]string{"Loopback IP Range", FabricProperties.LoopBackIPRange})
	table.Append([]string{"Loopback Port Number", FabricProperties.LoopBackPortNumber})
	table.Append([]string{"VTEP Loopback Port Number", FabricProperties.VTEPLoopBackPortNumber})
	if FabricProperties.FabricType == utils.CLOSFabricType {
		table.Append([]string{"Spine ASN Block", FabricProperties.SpineASNBlock})
		table.Append([]string{"Leaf ASN Block", FabricProperties.LeafASNBlock})
		table.Append([]string{"P2P IP Type", FabricProperties.P2PIPType})
	} else {
		table.Append([]string{"Rack ASN Block", FabricProperties.RackASNBlock})
		table.Append([]string{"L3 Backup IP Range", FabricProperties.MCTL3LBIPRange})
	}

	table.Append([]string{"Fabric Type", FabricProperties.FabricType})

	if advanced == true {
		table.Append([]string{"Any cast MAC", FabricProperties.AnyCastMac})
		table.Append([]string{"IPV6 Any cast MAC", FabricProperties.IPV6AnyCastMac})
		table.Append([]string{"ARP Aging Timeout", FabricProperties.ArpAgingTimeout})
		table.Append([]string{"MAC Aging Timeout", FabricProperties.MacAgingTimeout})
		table.Append([]string{"MAC Aging Conversational Timeout", FabricProperties.MacAgingConversationalTimeout})
		table.Append([]string{"MAC Move Limit", FabricProperties.MacMoveLimit})
		table.Append([]string{"Duplicate MAC Timer", FabricProperties.DuplicateMacTimer})
		//TODO Need to change Name DuplicateMaxTimerMaxCount to DuplicateMacTimerMaxCount
		table.Append([]string{"Duplicate MAC Timer MAX Count", FabricProperties.DuplicateMaxTimerMaxCount})
		table.Append([]string{"BFD Enable", FabricProperties.BFDEnable})
		if FabricProperties.BFDEnable == "Yes" {
			table.Append([]string{"BFD Tx", FabricProperties.BFDTx})
			table.Append([]string{"BFD Rx", FabricProperties.BFDRx})
			table.Append([]string{"BFD Multiplier", FabricProperties.BFDMultiplier})
		}
		table.Append([]string{"BGP MultiHop", FabricProperties.BGPMultiHop})
		table.Append([]string{"MaxPaths", FabricProperties.MaxPaths})
		table.Append([]string{"AllowAsIn", FabricProperties.AllowASIn})
		table.Append([]string{"MTU", FabricProperties.MTU})
		table.Append([]string{"IPMTU", FabricProperties.IPMTU})
		if FabricProperties.FabricType == utils.CLOSFabricType {
			table.Append([]string{"Configure Overlay Gateway", FabricProperties.ConfigureOverlayGateway})
			table.Append([]string{"Leaf PeerGroup", FabricProperties.LeafPeerGroup})
			table.Append([]string{"Spine PeerGroup", FabricProperties.SpinePeerGroup})
		} else {
			table.Append([]string{"Rack Peer EBGP Group", FabricProperties.RackPeerEBGPGroup})
			table.Append([]string{"Rack Peer Overlay Evpn Group", FabricProperties.RackPeerOvgGroup})
		}
		table.Append([]string{"MCT Link IP Range", FabricProperties.MCTLinkIPRange})
		table.Append([]string{"MCT PortChannel", FabricProperties.MctPortChannel})
		table.Append([]string{"Routing MCT PortChannel", FabricProperties.RoutingMctPortChannel})
		table.Append([]string{"Control Vlan", FabricProperties.ControlVlan})
		table.Append([]string{"Control VE", FabricProperties.ControlVE})
		//Unused May be we should completly Remove them may be they are copied from Legacy EWC Code
		table.Append([]string{"VNI Auto Map", FabricProperties.VNIAutoMap})
	}

	table.Render()

	return nil
}
func handleConfigShowErrorResponse(errorObject error) {
	//OpenAPI Generated code sends the message as an error string, so parsing output from string object
	//Body Contains the Error Obect in JSON
	fmt.Println("setting Show [Failed]")
	if utils.IsServerConnectionError(errorObject) {
		return
	}
	errorMessageList := strings.Split(errorObject.Error(), "Body:")
	if len(errorMessageList) == 2 {
		var ErrorModel openAPIClient.ErrorModel
		err := json.Unmarshal([]byte(errorMessageList[1]), &ErrorModel)
		if err == nil {
			fmt.Println(ErrorModel.Message)
		}
	} else {
		//Generic Error, Just print it
		fmt.Println("\t" + errorObject.Error())
	}

}

//CreateFromMap returns FabricProperties as a map
func CreateFromMap(m map[string]string) (FabricProperties, error) {
	var result FabricProperties
	err := mapstructure.Decode(m, &result)
	return result, err
}

// GetFabricShowOut featch all fabric setting and returns in a Table data structure
func GetFabricShowOut(table *tablewriter.Table) error {
	fabricName = constants.DefaultFabric
	var err error
	cfg := openAPIClient.NewConfiguration()
	api := openAPIClient.NewAPIClient(cfg)

	response, _, err := api.FabricApi.GetFabric(context.Background(), fabricName)
	FabricProperties, err := CreateFromMap(response.FabricSettings)
	if err != nil {
		handleConfigShowErrorResponse(err)
		return err
	}

	table.Append([]string{"Fabric Name ", fabricName})
	table.Append([]string{"Link IP Range", FabricProperties.P2PLinkRange})
	table.Append([]string{"Loopback IP Range", FabricProperties.LoopBackIPRange})
	table.Append([]string{"Loopback Port Number", FabricProperties.LoopBackPortNumber})
	table.Append([]string{"VTEP Loopback Port Number", FabricProperties.VTEPLoopBackPortNumber})
	table.Append([]string{"Spine ASN Block", FabricProperties.SpineASNBlock})
	table.Append([]string{"LEAF ASN Block", FabricProperties.LeafASNBlock})
	table.Append([]string{"P2P IP Type", FabricProperties.P2PIPType})

	table.Append([]string{"Any cast MAC", FabricProperties.AnyCastMac})
	table.Append([]string{"IPV6 Any cast MAC", FabricProperties.IPV6AnyCastMac})
	table.Append([]string{"ARP Aging Timeout", FabricProperties.ArpAgingTimeout})
	table.Append([]string{"MAC Aging Timeout", FabricProperties.MacAgingTimeout})
	table.Append([]string{"MAC Aging Conversational Timeout", FabricProperties.MacAgingConversationalTimeout})
	table.Append([]string{"MAC Move Limit", FabricProperties.MacMoveLimit})
	table.Append([]string{"Duplicate MAC Timer", FabricProperties.DuplicateMacTimer})
	//TODO Need to change Name DuplicateMaxTimerMaxCount to DuplicateMacTimerMaxCount
	table.Append([]string{"Duplicate MAC Timer MAX Count", FabricProperties.DuplicateMaxTimerMaxCount})
	table.Append([]string{"Configure Overlay Gateway", FabricProperties.ConfigureOverlayGateway})
	table.Append([]string{"BFD Enable", FabricProperties.BFDEnable})
	table.Append([]string{"BFD Tx", FabricProperties.BFDTx})
	table.Append([]string{"BFD Rx", FabricProperties.BFDRx})
	table.Append([]string{"BFD Multiplier", FabricProperties.BFDMultiplier})
	table.Append([]string{"BGP MultiHop", FabricProperties.BGPMultiHop})
	table.Append([]string{"MaxPaths", FabricProperties.MaxPaths})
	table.Append([]string{"AllowAsIn", FabricProperties.AllowASIn})
	table.Append([]string{"MTU", FabricProperties.MTU})
	table.Append([]string{"IPMTU", FabricProperties.IPMTU})
	table.Append([]string{"Leaf PeerGroup", FabricProperties.LeafPeerGroup})
	table.Append([]string{"Spine PeerGroup", FabricProperties.SpinePeerGroup})
	table.Append([]string{"MCT Link IP Range", FabricProperties.MCTLinkIPRange})
	table.Append([]string{"Mct PortChannel", FabricProperties.MctPortChannel})
	table.Append([]string{"Routing Mct PortChannel", FabricProperties.RoutingMctPortChannel})
	table.Append([]string{"Control Vlan", FabricProperties.ControlVlan})
	table.Append([]string{"Control VE", FabricProperties.ControlVE})
	//Unused May be we should completly Remove them may be they are copied from Legacy EWC Code
	table.Append([]string{"VNI Auto Map", FabricProperties.VNIAutoMap})

	return nil
}

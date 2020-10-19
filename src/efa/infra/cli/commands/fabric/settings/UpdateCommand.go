package settings

import (
	"context"
	"efa/infra/cli/utils"
	"efa/infra/constants"
	openAPIClient "efa/infra/rest/generated/client"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

var (
	fabricUpdateRequest FabricProperties
)

//FabricProperties as structure
type FabricProperties struct {
	//Switch Level Fields
	ConfigureOverlayGateway string `json:"configure_overlay_gateway"`
	MTU                     string `json:"mtu"`
	IPMTU                   string `json:"ip_mtu"`

	//BGP Fields
	SpineASNBlock  string `json:"spine_asn_block"`
	LeafASNBlock   string `json:"leaf_asn_block"`
	RackASNBlock   string `json:"rack_asn_block"`
	BGPMultiHop    string `json:"bgp_multihop"`
	MaxPaths       string `json:"max_paths"`
	AllowASIn      string `json:"allow_as_in"`
	LeafPeerGroup  string `json:"leaf_peer_group"`
	SpinePeerGroup string `json:"spine_peer_group"`

	//Interface Fields
	P2PLinkRange    string `json:"p2p_link_range"`
	P2PIPType       string `json:"p2p_ip_type"`
	LoopBackIPRange string `json:"loopback_ip_range"`

	LoopBackPortNumber string `json:"loopback_port_number"`
	MCTLinkIPRange     string `json:"mct_link_ip_range"`
	MCTL3LBIPRange     string `json:"mct_l3_lb_ip_range"`
	BFDEnable          string `json:"bfd_enable"`
	BFDTx              string `json:"bfd_tx"`
	BFDRx              string `json:"bfd_rx"`
	BFDMultiplier      string `json:"bfd_multiplier"`

	//OVG Fields
	VTEPLoopBackPortNumber string `json:"vtep_loopback_port_number"`
	VNIAutoMap             string `json:"vni_auto_map"`
	AnyCastMac             string `json:"any_cast_mac"`
	IPV6AnyCastMac         string `json:"ipv6_any_cast_mac"`

	//EVPN Fields
	ArpAgingTimeout               string `json:"arp_aging_timeout"`
	MacAgingTimeout               string `json:"mac_aging_timeout"`
	MacAgingConversationalTimeout string `json:"mac_legacy_aging_timeout"`
	MacMoveLimit                  string `json:"mac_move_limit"`
	DuplicateMacTimer             string `json:"duplicate_mac_timer"`
	DuplicateMaxTimerMaxCount     string `json:"duplicate_mac_timer_max_count"`

	//MCT Configs
	ControlVlan           string `json:"control_vlan"`
	ControlVE             string `json:"control_ve"`
	MctPortChannel        string `json:"mct_port_channel"`
	RoutingMctPortChannel string `json:"routing_mct_port_channel"`

	// Fabric Type
	FabricType string `json:"fabric_type"`

	// NON ClOS Fields
	RackPeerEBGPGroup string `json:"rack_peer_ebgp_group"`
	RackPeerOvgGroup  string `json:"rack_overlay_evpn_group"`
}

//UpdateCommand provides command for updating Fabric Properties
var UpdateCommand = &cobra.Command{
	Use:   "update",
	Short: "Update fabric settings.",
	RunE:  utils.TimedRunE(runFabricUpdate),
}

func init() {

	fabricName = constants.DefaultFabric
	cfg := openAPIClient.NewConfiguration()
	api := openAPIClient.NewAPIClient(cfg)

	response, _, _ := api.FabricApi.GetFabric(context.Background(), fabricName)
	FabricProperties, err := CreateFromMap(response.FabricSettings)

	if err != nil {
		fmt.Println(err)
	}

	UpdateCommand.Flags().SortFlags = false
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.P2PLinkRange, "p2p-link-range", "", "Range Of IP Address.")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.LoopBackIPRange, "loopback-ip-range", "", "Range Of IP Address")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.LoopBackPortNumber, "loopback-port-number", "", "Loopback Port Number <NUMBER: 1-255>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.VTEPLoopBackPortNumber, "vtep-loopback-port-number", "", "VTEP Loopback Port Number <NUMBER: 1-255>")
	if FabricProperties.FabricType == utils.CLOSFabricType {
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.LeafPeerGroup, "leaf-peer-group", "", "Leaf Peer Group Name <WORD: 1-63>")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.SpinePeerGroup, "spine-peer-group", "", "Spine Peer Group Name <WORD: 1-63>")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.SpineASNBlock, "spine-asn-block", "", "Spine ASN Range Separated -;Or Single AS"+"")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.LeafASNBlock, "leaf-asn-block", "", "Leaf ASN Range Separated -")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.ConfigureOverlayGateway, "configure-overlay-gateway", "", "ConfigureOverlayGateway Enabled Yes/No")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.P2PIPType, "p2p-ip-type", "", "IP Type numbered/unnumbered")
	} else {
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MCTL3LBIPRange, "l3-backup-ip-range", "", "Range Of IP Address")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.RackASNBlock, "rack-asn-block", "", "Rack ASN Range Separated -")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.RackPeerEBGPGroup, "rack-peer-ebgp-group", "", "Rack Peer eBgp Group Name <WORD: 1-63>")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.RackPeerOvgGroup, "rack-peer-overlay-evpn-group", "", "Rack Peer Overlay Evpn Group Name <WORD: 1-63>")
	}

	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.AnyCastMac, "anycast-mac-address", "", "IPV4 ANY CAST MAC address.mac address HHHH.HHHH.HHHH")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.IPV6AnyCastMac, "ipv6-anycast-mac-address", "", "IPV6 ANY CAST MAC address.mac address HHHH.HHHH.HHHH")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.ArpAgingTimeout, "arp-aging-timeout", "", "Determines how long an ARP entry stays in cache <NUMBER: 60-100000>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MacAgingTimeout, "mac-aging-timeout", "", "MAC Aging Timeout <NUMBER: 0|60-86400>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MacAgingConversationalTimeout, "mac-aging-conversation-timeout", "", "MAC Conversational Aging time in seconds<NUMBER: 0|60-100000>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MacMoveLimit, "mac-move-limit", "", "MAC move detect limit <NUMBER: 5-500>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.DuplicateMacTimer, "duplicate-mac-timer", "", "Duplicate Mac Timer")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.DuplicateMaxTimerMaxCount, "duplicate-mac-timer-max-count", "", "Duplicate Mac Timer Max Count")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.BFDEnable, "bfd-enable", "", "BFD enabled <STRING Yes/No>")
	if FabricProperties.BFDEnable == "Yes" {
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.BFDTx, "bfd-tx", "", "BFD desired min transmit interval in milliseconds <NUMBER: 50-30000>")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.BFDRx, "bfd-rx", "", "BFD desired min receive interval in milliseconds <NUMBER: 50-30000>")
		UpdateCommand.Flags().StringVar(&fabricUpdateRequest.BFDMultiplier, "bfd-multiplier", "", "BFD detection time multiplier <NUMBER: 3-50> ")
	}
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.BGPMultiHop, "bgp-multihop", "", "Allow EBGP neighbors not on directly connected networks <Number:1-255> ")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MaxPaths, "max-paths", "", "Forward packets over multiple paths<Number:1-64>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.AllowASIn, "allow-as-in", "", "Disables the AS_PATH check of the routes learned from the AS<Number:1-10> ")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MTU, "mtu", "", "The MTU size in bytes <Number:1548-9216>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.IPMTU, "ip-mtu", "", "For SLX IPV4/IPV6 MTU size in bytes <Number:1300-9194>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MCTLinkIPRange, "mctlink-ip-range", "", "Range Of IP Address")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.MctPortChannel, "mct-port-channel", "", "Portchannel interface number <NUMBER: 1-1024>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.RoutingMctPortChannel, "routing-mct-port-channel", "", "Portchannel interface number <NUMBER: 1-64>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.ControlVlan, "control-vlan", "", "vlan number <NUMBER: 1-4090>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.ControlVE, "control-ve", "", "vlan number <NUMBER: 1-4090>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.VNIAutoMap, "vni-auto-map", "", "VNI Auto Map <STRING Yes/No>")
	UpdateCommand.Flags().StringVar(&fabricUpdateRequest.FabricType, "fabric-type", "", "Fabric Type <STRING clos/non-clos>")
}

//PrepareFabricSettingsRequest prepares the Fabric Setting Request
func (FabricUpdate *FabricProperties) PrepareFabricSettingsRequest(FabricSetting *openAPIClient.FabricSettings) {
	val := reflect.ValueOf(FabricUpdate).Elem()
	for i := 0; i < val.NumField(); i++ {
		var FabricParameter openAPIClient.FabricParameter
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		FabricParameter.Key = typeField.Name
		FabricParameter.Value = valueField.String()
		FabricSetting.Keyval = append(FabricSetting.Keyval, FabricParameter)
	}
}

func runFabricUpdate(cmd *cobra.Command, args []string) error {
	if fabricUpdateRequest == (FabricProperties{}) {
		fmt.Println("No Fabric Property Update is Requested")
		return nil
	}
	var FabricSetting openAPIClient.FabricSettings
	FabricSetting.Name = constants.DefaultFabric
	fabricUpdateRequest.PrepareFabricSettingsRequest(&FabricSetting)
	cfg := openAPIClient.NewConfiguration()
	api := openAPIClient.NewAPIClient(cfg)
	data := make(map[string]interface{})
	data["fabricSettings"] = FabricSetting

	FabricUpdateResponse, _, err := api.FabricApi.UpdateFabric(context.Background(), data)
	if err != nil {
		if utils.IsServerConnectionError(err) {
			return nil
		}
		var FabricdataErrorResp openAPIClient.FabricdataErrorResponse
		status := strings.Split(err.Error(), "Body:")
		jerr := json.Unmarshal([]byte(status[1]), &FabricdataErrorResp)
		if jerr != nil {
			fmt.Println("Error While Decoding Server Response")
			return nil
		}
		//fmt.Printf("%s\n", status[0])
		fmt.Printf("%s Fabric settings Update Failed\n", FabricdataErrorResp.FabricName)
		fmt.Printf("Reason: \n")
		for _, val := range FabricdataErrorResp.FabricSettings {
			fmt.Printf("\t%s\n", val)
		}
	} else {
		fmt.Printf("%s Fabric Update Successful\n", FabricSetting.Name)
		fmt.Printf("FabricId: %d\n", FabricUpdateResponse.FabricId)
	}
	return nil
}

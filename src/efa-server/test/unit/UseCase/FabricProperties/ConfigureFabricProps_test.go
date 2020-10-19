package fabricproperties

import (
	"context"
	"efa-server/domain"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"efa-server/infra/rest/openapi/handler"
	"efa-server/test/unit/mock"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var MockFabricName = "test_fabric"
var dbExtension = "fp"
var FabricUpdateRequestNegative = domain.FabricProperties{
	MTU:                           "20000",
	ConfigureOverlayGateway:       "Test",
	IPMTU:                         "20000",
	SpineASNBlock:                 "23-",
	LeafASNBlock:                  "33@45",
	BGPMultiHop:                   "20000",
	MaxPaths:                      "73",
	AllowASIn:                     "24",
	LeafPeerGroup:                 "Ab!!",
	SpinePeerGroup:                "Ab.2",
	P2PLinkRange:                  "2.2.2/23",
	P2PIPType:                     "Test",
	LoopBackIPRange:               "3.3.3./24",
	LoopBackPortNumber:            "320",
	BFDTx:                         "30001",
	BFDRx:                         "30001",
	BFDMultiplier:                 "51",
	VTEPLoopBackPortNumber:        "256",
	VNIAutoMap:                    "Test",
	AnyCastMac:                    "000.111.222",
	IPV6AnyCastMac:                "junk",
	ArpAgingTimeout:               "100001",
	MacAgingTimeout:               "100001",
	MacAgingConversationalTimeout: "100001",
	MacMoveLimit:                  "501",
	DuplicateMacTimer:             "30",
	DuplicateMaxTimerMaxCount:     "40",
	FabricType:                    "non-closs",
}

var FabricUpdateRequestPositive = domain.FabricProperties{
	ConfigureOverlayGateway:       "Yes",
	MTU:                           "1600",
	IPMTU:                         "1600",
	SpineASNBlock:                 "5000-5500",
	LeafASNBlock:                  "6000",
	RackASNBlock:                  "4200000000-4200065534",
	BGPMultiHop:                   "200",
	MaxPaths:                      "25",
	AllowASIn:                     "9",
	LeafPeerGroup:                 "spine-test",
	SpinePeerGroup:                "leaf-test",
	P2PLinkRange:                  "2.2.2.0/23",
	P2PIPType:                     "numbered",
	LoopBackIPRange:               "3.3.3.0/24",
	LoopBackPortNumber:            "254",
	BFDEnable:                     "true",
	BFDTx:                         "300",
	BFDRx:                         "300",
	BFDMultiplier:                 "8",
	VTEPLoopBackPortNumber:        "252",
	VNIAutoMap:                    "Yes",
	AnyCastMac:                    "0000.1111.2222",
	IPV6AnyCastMac:                "0000.3333.4444",
	ArpAgingTimeout:               "220",
	MacAgingTimeout:               "5000",
	MacAgingConversationalTimeout: "5000",
	MacMoveLimit:                  "50",
	DuplicateMacTimer:             "5",
	DuplicateMaxTimerMaxCount:     "3",
	ControlVlan:                   domain.DefaultControlVlan,
	ControlVE:                     domain.DefaultControlVE,
	MctPortChannel:                domain.DefaultMctPortChannel,
	RoutingMctPortChannel:         domain.RoutingDefaultMctPortChannel,
	MCTLinkIPRange:                "10.0.0.2/24",
	MCTL3LBIPRange:                "10.30.30.2/24",
	FabricType:                    domain.CLOSFabricType,
	RackPeerEBGPGroup:             "underlay-ebgp-group",
	RackPeerOvgGroup:              "overlay-ebgp-group",
}

func TestConfigureInvalid(t *testing.T) {
	expected := map[string]string{
		"bfd-rx":                                "30001 is not Valid BFDRx , Valid BGP BFDRx is 50-30000",
		"bfd-tx":                                "30001 is not Valid BFDTx , Valid BGP BFDTx is 50-30000",
		"mac-aging-conversation-timeout":        "100001 is not Valid MACAgingConversationalTimeOut, Valid MACAgingConversationalTimeOut is 0|60-100000",
		"mtu":                                   "20000 is not Valid MTU , Valid MTU is 1548-9216",
		"ipv6-anycast-mac":                      "junk is Not Valid AnyCast Mac Address, Valid AnyCast MAC Address in the form HHHH.HHHH.HHHH.[0200.dea1.0001]",
		"spine-asn-block":                       "23- is not valid ASN Range , Valid ASN Range is 1-4294967295",
		"vtep-loopback-port-number":             "256 is not Valid VTEP loopback portnumber,Valid range is 1-255",
		"anycast-mac":                           "000.111.222 is Not Valid AnyCast Mac Address, Valid AnyCast MAC Address in the form HHHH.HHHH.HHHH.[0200.dea1.0001]",
		"bgp-maxpaths":                          "73 is not Valid BgpMaxPaths , Valid BGP MaxPaths is 1-64",
		"bgp-multihop":                          "20000 is not Valid BgpMultiHop , Valid BGP MultiHop is 0-255",
		"vni-auto-map":                          "test is not valid . Valid Value for vni-auto-map is yes/no",
		"configure-overlay-gateway":             "test is not valid . Valid Value for configure-overlay-gateway is yes/no",
		"bfd-multiplier":                        "51 is not Valid BFDMultiplier , Valid BGP BFDMultiplier is 3-50",
		"leaf-peer-group-name":                  "Ab!! is not Valid Peer Group Name, Valid Peer Group Name is <WORD: 1-63>",
		"loopback-ip-range":                     "3.3.3./24 is not Valid IP Address , Valid IP in the format w.x.y.z/m",
		"spine-peer-group-name":                 "Ab.2 is not Valid Peer Group Name, Valid Peer Group Name is <WORD: 1-63>",
		"loopback-port-number":                  "320 is not Valid loopback portnumber,Valid range is 1-255",
		"mac-aging-timeout":                     "100001 is not Valid MACAgingTimeOut, Valid MACAgingTimeOut is 0|60-86400",
		"mac-move-limit":                        "501 is not Valid MacMoveLimit, Valid MacMoveLimit is 5-500",
		"p2p-ip-type":                           "test is not valid . Valid Value for p2p-type is numbered/unnumbered",
		"allow-as-in":                           "24 is not Valid AllowAsIn , Valid BGP AllowAsIn is 0-10",
		"arp-aging-timeout":                     "100001 is not Valid ARPAgingTimeOut, Valid Arp AgingTimeOut is 60-100000",
		"ip-mtu":                                "20000 is not Valid MTU , Valid MTU is 1300-9194",
		"leaf-asn-block":                        "33@45 is not valid ASN Range , Valid ASN Range is 1-4294967295",
		"p2p-link-range":                        "2.2.2/23 is not Valid IP Address , Valid IP in the format w.x.y.z/m",
		"duplicate-mac-timer-max-count-timeout": "40 is not Valid DuplicateMacTimerMaxCount, Valid ValidateDuplicateMacTimerMaxCount is 3-10",
		"fabric-type":                           "non-closs is not Valid fabric type, Valid  fabric-type is <clos/non-clos>",
	}

	database.Setup(constants.TESTDBLocation + dbExtension)
	defer cleanupDB(database.GetWorkingInstance())
	ret := handler.ValidateFabricProperties(MockFabricName, &FabricUpdateRequestNegative)
	for x, y := range ret {
		fmt.Println(x, y)
	}

	assert.Equal(t, expected, ret)
}

func TestConfigureValid(t *testing.T) {
	var ret string
	MockSpineDeviceAdapter := mock.DeviceAdapter{}

	database.Setup(constants.TESTDBLocation + dbExtension)
	defer cleanupDB(database.GetWorkingInstance())

	UseCaseInteractor := infra.GetUseCaseInteractor()
	UseCaseInteractor.DeviceAdapterFactory = mock.GetDeviceAdapterFactory(MockSpineDeviceAdapter)
	UseCaseInteractor.AddFabric(context.Background(), MockFabricName)

	ret, _, _ = UseCaseInteractor.UpdateFabricProperties(context.Background(), MockFabricName, &FabricUpdateRequestPositive)
	assert.Equal(t, ret, "Fabric Properties updated")
	Fabric, _ := UseCaseInteractor.Db.GetFabric(MockFabricName)
	FabricProps, _ := UseCaseInteractor.Db.GetFabricProperties(Fabric.ID)
	//To Saatisfy relect.DeepEqual
	FabricUpdateRequestPositive.ID = Fabric.ID
	FabricUpdateRequestPositive.FabricID = Fabric.ID
	assert.Equal(t, FabricUpdateRequestPositive, FabricProps)
}

func cleanupDB(Database *database.Database) {
	Database.Close()
	err := os.Remove(constants.TESTDBLocation + dbExtension)
	fmt.Println(err)
}

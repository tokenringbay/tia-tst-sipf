package variant

import (
	ad "efa-server/infra/device/adapter"
	"efa-server/test/functional"
	"testing"

	"efa-server/test/functional/netconf/variant/avalanche"
	"efa-server/test/functional/netconf/variant/cedar"
	"efa-server/test/functional/netconf/variant/freedom"
	"efa-server/test/functional/netconf/variant/orca"
	"os"
	"reflect"
)

type platform struct {
	IP     string
	Model  string
	handle interface{}
}

var platforms = map[string]platform{
	"Freedom": {functional.NetConfFreedomIP, ad.FreedomType, &freedom.Variant{}},
	"Cedar":   {functional.NetConfCedarIP, ad.CedarType, &cedar.Variant{}},
}

func init() {
	if os.Getenv("SKIP_AV") != "1" {
		platforms["Avalanche"] = platform{functional.NetConfAvalancheIP, ad.AvalancheType, &avalanche.Variant{}}
	}
	if os.Getenv("SKIP_OR") != "1" {
		platforms["Orca"] = platform{functional.NetConfOrcaIP, ad.OrcaType, &orca.Variant{}}
	}
}

//TestConfigureMacAndArp testing Mac and ARP
func TestConfigureMacAndArp(t *testing.T) {

	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		handle := p.handle
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			method := reflect.ValueOf(handle).MethodByName("TConfigureMacAndArp")
			if method.IsValid() {
				method.Call([]reflect.Value{
					reflect.ValueOf(t),
					reflect.ValueOf(Host),
					reflect.ValueOf(Model),
				})
			}
		})

	}
}

//TestConfigureAnyCastGateway testing AnyCast Gateway
func TestConfigureAnyCastGateway(t *testing.T) {

	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		handle := p.handle
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			method := reflect.ValueOf(handle).MethodByName("TConfigureAnyCastGateway")
			if method.IsValid() {
				method.Call([]reflect.Value{
					reflect.ValueOf(t),
					reflect.ValueOf(Host),
					reflect.ValueOf(Model),
				})
			}
		})

	}
}

//TestMCTClusterConfigureMCTCluster testing MCT Cluster
func TestMCTClusterConfigureMCTCluster(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		handle := p.handle
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			method := reflect.ValueOf(handle).MethodByName("TMCTClusterConfigureMCTCluster")
			if method.IsValid() {
				method.Call([]reflect.Value{
					reflect.ValueOf(t),
					reflect.ValueOf(Host),
					reflect.ValueOf(Model),
				})
			}
		})

	}
}

//TestConfigurePortChannelAsSwitchPort testing switchport channel
func TestConfigurePortChannelAsSwitchPort(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		handle := p.handle
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			method := reflect.ValueOf(handle).MethodByName("TConfigurePortChannelAsSwitchPort")
			if method.IsValid() {
				method.Call([]reflect.Value{
					reflect.ValueOf(t),
					reflect.ValueOf(Host),
					reflect.ValueOf(Model),
				})
			}
		})

	}
}

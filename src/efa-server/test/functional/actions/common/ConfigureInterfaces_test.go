package common

import (
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/actions/deconfigurefabric"

	"testing"

	"efa-server/domain"

	"efa-server/domain/operation"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	"fmt"
	"github.com/stretchr/testify/assert"
)

var interface1Type = "ethernet"
var interface1Name = "0/4"
var interface1Ip = "4.4.4.4/31"
var interface1Oldip = "5.4.4.4/31"

var interface2Type = domain.IntfTypeLoopback
var interface2Name = "20"
var interface2Ip = "4.4.4.20/32"
var interface2Oldip = "5.4.4.20/32"

//Test EVPN Create and verify using NetConf
func TestInterfaces_CreateConfig(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model
		t.Run(name, func(t *testing.T) {
			//Open up the Client and set other attributes
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()

			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				IP: interface1Ip, ConfigType: domain.ConfigCreate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Ip, ConfigType: domain.ConfigCreate, Description: "testing"})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}

			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)
			//Setup for cleanup
			defer cleanInterfaces(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//TODO add support for Unnumbered Interfaces ie one with Donor Ports

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"address": interface1Ip, "description": "testing"}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"Name": interface2Name, "address": interface2Ip}, interfaceMap)

		})
	}
}

//Test EVPN Create and verify using NetConf
func TestInterfaces_UpdateConfig(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			fmt.Println(Host, UserName, Password)
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			//First Create Interfaces with Old IP
			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				IP: interface1Oldip, ConfigType: domain.ConfigCreate})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Oldip, ConfigType: domain.ConfigCreate})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Now call with Update Config
			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			Interfaces = make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				IP: interface1Ip, ConfigType: domain.ConfigUpdate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Ip, ConfigType: domain.ConfigUpdate, Description: "testing"})

			sw = operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Setup for cleanup
			defer cleanInterfaces(client, Model)

			//TODO add support for Unnumbered Interfaces ie one with Donor Ports

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"address": interface1Ip, "description": "testing"}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"address": interface2Ip, "Name": interface2Name}, interfaceMap)

		})
	}
}

//Test EVPN Create and verify using NetConf
func TestInterfaces_DeleteConfig(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			//First Create Interfaces with Old IP
			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				IP: interface1Oldip, ConfigType: domain.ConfigCreate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Oldip, ConfigType: domain.ConfigCreate, Description: "testing"})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Now call with Update Config
			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			Interfaces = make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				IP: interface1Oldip, ConfigType: domain.ConfigDelete, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Oldip, ConfigType: domain.ConfigDelete, Description: "testing"})

			sw = operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Setup for cleanup
			defer cleanInterfaces(client, Model)

			//TODO add support for Unnumbered Interfaces ie one with Donor Ports

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{}, interfaceMap)

		})
	}
}

func cleanInterfaces(client *netconf.NetconfClient, Model string) (string, error) {
	detail, _ := ad.GetDeviceDetail(client)

	adapter := ad.GetAdapter(detail.Model)
	if intMap, err := adapter.GetInterface(client, interface1Type, interface1Name); err == nil {
		if intMap["address"] != "" {
			adapter.UnconfigureInterfaceNumbered(client, interface1Type, interface1Name, intMap["address"])

		}
		if intMap["donor_type"] != "" {
			adapter.UnconfigureInterfaceUnnumbered(client, interface1Type, interface1Name)
		}
	}
	return adapter.DeleteInterfaceLoopback(client, interface2Name)

}

//Test EVPN Create and verify using NetConf
func TestInterfaces_Delete(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				IP: interface1Ip, ConfigType: domain.ConfigCreate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Ip, ConfigType: domain.ConfigCreate, Description: "testing"})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}

			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"address": interface1Ip, "description": "testing"}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"Name": interface2Name, "address": interface2Ip}, interfaceMap)

			fabricGate.Add(1)
			deconfigurefabric.UnconfigureInterfaces(ctx, fabricGate, &sw, fabricErrors)
			interfaceMap, err = adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{}, interfaceMap)
			//assert.Equal(t, map[string]string{"address": "", "min-rx": "", "multiplier": "", "min-tx": ""}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{}, interfaceMap)
			//assert.Equal(t, map[string]string{"Name":"","address": ""}, interfaceMap)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

		})
	}
}

func TestUnnumberedInterface_CreateConfig(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)

			// Donor interface
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				Donor: interface2Type, DonorPort: interface2Name, IP: "", ConfigType: domain.ConfigCreate, Description: "testing"})

			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Ip, ConfigType: domain.ConfigCreate, Description: "testing"})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}

			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)
			//Setup for cleanup
			defer cleanInterfaces(client, Model)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"donor_type": interface2Type, "donor_name": interface2Name}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"Name": interface2Name, "address": interface2Ip}, interfaceMap)

		})
	}
}

func TestUnnumberedInterface_UpdateConfig(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			//First Create Interfaces with Old IP
			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				Donor: interface2Type, DonorPort: interface2Name, IP: "", ConfigType: domain.ConfigCreate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Oldip, ConfigType: domain.ConfigCreate})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

			//Now call with Update Config
			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			Interfaces = make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				Donor: interface2Type, DonorPort: interface2Name, IP: "", ConfigType: domain.ConfigUpdate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Ip, ConfigType: domain.ConfigUpdate, Description: "testing"})

			sw = operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}
			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

			//Setup for cleanup
			defer cleanInterfaces(client, Model)

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"donor_type": interface2Type, "donor_name": interface2Name}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"address": interface2Ip, "Name": interface2Name}, interfaceMap)

		})
	}
}

//Test EVPN Create and verify using NetConf
func TestUnnumberedInterface_Delete(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		Model := p.Model

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			//Open up the Client and set other attributes
			client := &netconf.NetconfClient{Host: Host, User: UserName, Password: Password}
			client.Login()
			defer client.Close()
			ctx, fabricGate, fabricErrors, Errors := initializeTest()
			detail, err := ad.GetDeviceDetail(client)
			assert.Contains(t, detail.Model, Model)
			adapter := ad.GetAdapter(detail.Model)
			//cleanup EVPN before testing
			cleanInterfaces(client, Model)
			//defer client.Close()

			Interfaces := make([]operation.ConfigInterface, 0)
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface1Name, InterfaceType: interface1Type,
				Donor: interface2Type, DonorPort: interface2Name, IP: "", ConfigType: domain.ConfigCreate, Description: "testing"})
			Interfaces = append(Interfaces, operation.ConfigInterface{InterfaceName: interface2Name, InterfaceType: interface2Type,
				IP: interface2Ip, ConfigType: domain.ConfigCreate, Description: "testing"})

			sw := operation.ConfigSwitch{Fabric: FabricName, Host: Host, UserName: UserName, Password: Password, Interfaces: Interfaces,
				BfdRx: BFDRx, BfdTx: BFDTx, BfdMultiplier: BFDMultiplier, Model: detail.Model}

			//Call the Actions
			configurefabric.ConfigureInterfaces(ctx, fabricGate, &sw, false, fabricErrors)
			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			assert.Empty(t, Errors)

			//Fetch EVPN details using NetConf
			interfaceMap, err := adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"donor_type": interface2Type, "donor_name": interface2Name}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{"Name": interface2Name, "address": interface2Ip}, interfaceMap)

			ctx, fabricGate, fabricErrors, Errors = initializeTest()
			deconfigurefabric.UnconfigureInterfaces(ctx, fabricGate, &sw, fabricErrors)
			interfaceMap, err = adapter.GetInterface(client, interface1Type, interface1Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{}, interfaceMap)
			//assert.Equal(t, map[string]string{"address": "", "min-rx": "", "multiplier": "", "min-tx": ""}, interfaceMap)

			interfaceMap, err = adapter.GetInterface(client, interface2Type, interface2Name)
			assert.Nil(t, err)
			assert.Equal(t, map[string]string{}, interfaceMap)
			//assert.Equal(t, map[string]string{"Name":"","address": ""}, interfaceMap)

			Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
			//Check that action throws no Errors
			assert.Empty(t, Errors)

		})
	}
}

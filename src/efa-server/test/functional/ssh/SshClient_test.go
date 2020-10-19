package ssh

import (
	//"fmt"
	ad "efa-server/infra/device/adapter"
	ssh "efa-server/infra/device/client"
	"efa-server/test/functional"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"strconv"
	"strings"
	"testing"
)

type platform struct {
	IP    string
	Model string
}

var platforms = map[string]platform{
	"Avalanche": {functional.NetConfAvalancheIP, ad.AvalancheType},
	"Freedom":   {functional.NetConfFreedomIP, ad.FreedomType},
	"Cedar":     {functional.NetConfCedarIP, ad.CedarType},
}
var Password = functional.DeviceAdminPassword

func TestShowVersion(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		//Model := p.Model
		t.Run(name, func(t *testing.T) {
			//client := &device.SSHClient{Host,UserName,Password)
			client := &ssh.SSHClient{
				Host:     Host,
				User:     "admin",
				Password: Password,
			}
			err := client.Login()

			if err != nil {
				log.Fatal(err)
			}
			output := client.ExecuteOperationalCommand("show version ")
			assert.Contains(t, output, "Firmware name", "")
			client.Close()
		})
	}
}

func TestInterfaceVlan(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		//Model := p.Model
		t.Run(name, func(t *testing.T) {
			//client := &device.SSHClient{Host,UserName,Password)
			client := &ssh.SSHClient{
				Host:     Host,
				User:     "admin",
				Password: Password,
			}
			err := client.Login()

			if err != nil {
				log.Fatal(err)
			}
			output := client.ExecuteConfigCommand("vlan 100")
			output = client.ExecuteOperationalCommand("show running-config vlan ")
			assert.Contains(t, output, "vlan 100", "")
			client.Close()
		})
	}
}

func TestMgmtClusterNodeId(t *testing.T) {
	for name, p := range platforms {
		Host := p.IP
		//Model := p.Model
		t.Run(name, func(t *testing.T) {
			//client := &device.SSHClient{Host,UserName,Password)
			client := &ssh.SSHClient{
				Host:     Host,
				User:     "admin",
				Password: Password,
			}
			err := client.Login()

			if err != nil {
				log.Fatal(err)
			}
			// Get the current node-id from the switch, store it in currentNodeID
			output := client.ExecuteOperationalCommand("show cluster management")

			currentNodeID := extractNodeIDFromOutput(output)

			assert.NotEqual(t, 0, currentNodeID)

			// Modify the switch node-id to 123.
			output = client.ExecuteOperationalCommand("cluster management node-id 123")
			output = client.ExecuteOperationalCommand("show cluster management")

			nodeID := extractNodeIDFromOutput(output)
			assert.Equal(t, 123, nodeID)

			// Reset the switch nodeid to currentNodeID
			output = client.ExecuteOperationalCommand(string("cluster management node-id ") + fmt.Sprint(currentNodeID))
			output = client.ExecuteOperationalCommand("show cluster management")

			nodeID = extractNodeIDFromOutput(output)
			assert.Equal(t, currentNodeID, nodeID)

			client.Close()
		})
	}
}

func extractNodeIDFromOutput(output string) int {
	nodeID := 0
	outputs := strings.Split(output, "\n")
	for _, element := range outputs {
		if strings.Contains(element, "*") {
			// Current node-id row
			columns := strings.Split(element, " ")
			nodeIDStr := columns[0]
			if nID, err := strconv.Atoi(nodeIDStr); err == nil {
				nodeID = nID
			}
		}
	}
	return nodeID
}

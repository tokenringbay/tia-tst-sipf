package base

import (
	"efa-server/infra/device/client"
	"efa-server/infra/device/models"
	"fmt"
	"github.com/beevik/etree"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

func getLldpRequest(lastIfindex string) string {
	var requestInterface string
	if lastIfindex == "" {
		requestInterface = `<get-lldp-neighbor-detail xmlns="urn:brocade.com:mgmt:brocade-lldp-ext"></get-lldp-neighbor-detail>`
	} else {
		requestInterfaceFmt := `<get-lldp-neighbor-detail xmlns="urn:brocade.com:mgmt:brocade-lldp-ext">
		                      <last-rcvd-ifindex>%s</last-rcvd-ifindex>
			                 </get-lldp-neighbor-detail>`
		requestInterface = fmt.Sprintf(requestInterfaceFmt, lastIfindex)

	}
	return requestInterface
}

//GetLLDPNeighbors is used to get the operational state of the LLDP neighbours, from the switching device
func (base *SLXBase) GetLLDPNeighbors(client *client.NetconfClient) ([]models.InterfaceLLDPResponse, error) {

	LLLDPS := make([]models.InterfaceLLDPResponse, 0)
	hasMore := "true"
	lastIfindex := ""

	for hasMore == "true" {
		request := getLldpRequest(lastIfindex)

		resp, err := client.ExecuteRPC(request)

		if err != nil {
			log.Infoln(err)
			return LLLDPS, err
		}

		doc := etree.NewDocument()
		if err := doc.ReadFromBytes([]byte(resp)); err != nil {
			return LLLDPS, err
		}
		if elems := doc.FindElements("//lldp-neighbor-detail"); elems != nil {
			var lldp models.InterfaceLLDPResponse
			for _, elem := range elems {
				if de := elem.FindElement(".//local-interface-ifindex"); de != nil {
					lastIfindex = de.Text()
				}

				if de := elem.FindElement(".//local-interface-name"); de != nil {
					localInterface := de.Text()
					split := strings.Split(localInterface, " ")
					lldp.LocalInterfaceType = split[0]
					lldp.LocalInterfaceName = split[1]
					if strings.Contains(lldp.LocalInterfaceType, "Te") {
						lldp.LocalInterfaceType = "TenGigabitEthernet"
					}
					if strings.Contains(lldp.LocalInterfaceType, "Fo") {
						lldp.LocalInterfaceType = "FortyGigabitEthernet"
					}
					if strings.Contains(lldp.LocalInterfaceType, "Hu") {
						lldp.LocalInterfaceType = "HundredGigabitEthernet"
					}
					if strings.Contains(lldp.LocalInterfaceType, "Eth") {
						lldp.LocalInterfaceType = "Ethernet"
					}
				}
				if de := elem.FindElement(".//local-interface-mac"); de != nil {
					lldp.LocalInterfaceMac = de.Text()
					mac, _ := net.ParseMAC(lldp.LocalInterfaceMac)
					lldp.LocalInterfaceMac = mac.String()
				}
				if de := elem.FindElement(".//remote-interface-name"); de != nil {
					remoteInterface := de.Text()
					if !strings.Contains(remoteInterface, " ") {
						continue
					}
					split := strings.Split(remoteInterface, " ")
					lldp.RemoteInterfaceType = split[0]
					lldp.RemoteInterfaceName = split[1]
				}
				if de := elem.FindElement(".//remote-interface-mac"); de != nil {
					lldp.RemoteInterfaceMac = de.Text()
					mac, _ := net.ParseMAC(lldp.RemoteInterfaceMac)
					lldp.RemoteInterfaceMac = mac.String()
				}
				LLLDPS = append(LLLDPS, lldp)
			}

			if de := doc.FindElement("//has-more"); de != nil {

				hasMore = de.Text()
			}
		}
	}

	return LLLDPS, nil
}

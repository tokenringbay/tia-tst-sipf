package base

import (
	"efa-server/infra/device/client"
	"fmt"
	"github.com/beevik/etree"
)

//GetIPRoutes is used to get the running-config of "overlay-gateway" from the switching device
func (base *SLXAvalancheBase) GetIPRoutes(client *client.NetconfClient) (map[string]string, error) {

	resp, err := client.GetConfig("ip")
	doc := etree.NewDocument()

	iprouteMap := make(map[string]string)
	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		fmt.Println(err)
		return iprouteMap, err
	}

	if elems := doc.FindElements("//route/static-route-nh"); elems != nil {

		for _, elem := range elems {
			var StaticRouteDest, StaticRouteNextHop string
			if de := elem.FindElement(".//static-route-dest"); de != nil {
				StaticRouteDest = de.Text()
			}
			if de := elem.FindElement(".//static-route-next-hop"); de != nil {
				StaticRouteNextHop = de.Text()
			}
			iprouteMap[StaticRouteDest] = StaticRouteNextHop
		}
	}
	return iprouteMap, err

}

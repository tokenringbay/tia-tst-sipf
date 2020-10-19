package cedar

import (
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	"efa-server/test/functional"
	"github.com/stretchr/testify/assert"
	"testing"
)

//UserName used for testing
var UserName = "admin"

//Password used for testing
var Password = functional.DeviceAdminPassword

//FabricName used for testing
var FabricName = "test_fabric"

//vteploopbackPortnumber used for testing
var vteploopbackPortnumber = "1"

//Variant is the reciever object for testing
type Variant struct {
}

//TConfigureMacAndArp testing Mac and ARP
func (v *Variant) TConfigureMacAndArp(t *testing.T, Host string, Model string) {
	client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
	client.Login()
	defer client.Close()

	detail, _ := ad.GetDeviceDetail(client)
	adapter := ad.GetAdapter(detail.Model)

	macAgingTimeout := "100"
	macMoveLimit := "102"
	macAgingConversationalTimeout := "101"
	arpConversationalValue := "4000"

	msg, err := adapter.ConfigureMacAndArp(client, arpConversationalValue, macAgingTimeout, macAgingConversationalTimeout,
		macMoveLimit)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	responses, err := adapter.GetMac(client)
	assert.Equal(t, map[string]string{"mac-aging-timeout": macAgingTimeout, "mac-move-limit": macMoveLimit,
		"mac-conversational-timeout": macAgingConversationalTimeout, "mac-learning-mode": "conversational"}, responses)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	responses, err = adapter.GetArp(client)
	assert.Equal(t, map[string]string{"arp-aging-mode-conversational": "true", "arp-conversational-timeout": arpConversationalValue}, responses)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	msg, err = adapter.UnconfigureMacAndArp(client)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

}

//TConfigureAnyCastGateway testing AnyCast Gateway
func (v *Variant) TConfigureAnyCastGateway(t *testing.T, Host string, Model string) {
	client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
	client.Login()
	defer client.Close()

	ipv4anyCastGatewayMac := "0201.0101.0101"
	ipv6anyCastGatewayMac := "0201.0101.0102"

	detail, _ := ad.GetDeviceDetail(client)
	adapter := ad.GetAdapter(detail.Model)

	msg, err := adapter.ConfigureAnycastGateway(client, ipv4anyCastGatewayMac, ipv6anyCastGatewayMac)

	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	responses, err := adapter.GetAnycastGateway(client)
	assert.Equal(t, map[string]string{"ip-anycast-gateway-mac": "0201.0101.0101"}, responses)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	msg, err = adapter.UnconfigureAnycastGateway(client)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	responses, err = adapter.GetAnycastGateway(client)
	assert.Equal(t, map[string]string{}, responses)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

}

//TMCTClusterConfigureMCTCluster testing MCT Cluster
func (v *Variant) TMCTClusterConfigureMCTCluster(t *testing.T, Host string, Model string) {
	clusterName := "cluster-1"
	clusterID := "1"
	portChannel := "11"
	speed := "40000"
	peerType := "Port-channel"
	peerIP := "5.4.5.3"
	peerLoopbackIP := "6.6.6.6"
	BFDMultiplier := "3"
	BFDRx := "300"
	BFDTx := "300"
	client := &client.NetconfClient{Host: Host, User: UserName, Password: Password}
	client.Login()
	defer client.Close()
	detail, _ := ad.GetDeviceDetail(client)
	adapter := ad.GetAdapter(detail.Model)

	controlVlan := "4089"
	controlVe := "4088"
	ipAddress := "5.4.5.2/31"

	//cleanup existing cluster
	exclusterName, exclusterID, _, _, _, _, _ := adapter.GetCluster(client)

	if exclusterName != "" {
		adapter.DeleteCluster(client, exclusterName, exclusterID)
	}

	adapter.DeleteInterfaceVe(client, controlVe)

	//pre-requisite
	premsg, err := adapter.CreateClusterControlVlan(client, controlVlan, controlVe, "Control VLAN")
	assert.Equal(t, "<ok/>", premsg, "")
	assert.Nil(t, err, "")

	premsg, err = adapter.DeleteInterfacePo(client, portChannel)

	premsg, err = adapter.CreateInterfacePo(client, portChannel, speed, "mct pc", controlVlan)
	assert.Equal(t, "<ok/>", premsg, "")
	assert.Nil(t, err, "")
	premsg, err = adapter.ConfigureInterfaceVe(client, controlVe, ipAddress, BFDRx, BFDTx, BFDMultiplier)
	assert.Equal(t, "<ok/>", premsg, "")
	assert.Nil(t, err, "")

	//create
	msg, err := adapter.CreateCluster(client, clusterName, clusterID, peerType, portChannel, peerIP)
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	msg, err = adapter.ConfigureCluster(client, clusterName, clusterID, peerType, portChannel, peerIP, peerLoopbackIP,
		controlVlan, controlVe, "")
	assert.Equal(t, "<ok/>", msg, "")
	assert.Nil(t, err, "")

	//Get
	ResultMap, err := adapter.GetClusterByName(client, clusterName)
	assert.Nil(t, err, "")
	assert.Equal(t, map[string]string{"cluster-id": clusterID, "peer-type": peerType, "peer-name": portChannel, "peer-ip": peerIP,
		"df-load-balance": "true", "deploy": "true", "cluster-control-vlan": controlVlan}, ResultMap, "")

	//Delete
	DeleteMsg, err := adapter.DeleteCluster(client, clusterName, clusterID)
	assert.Equal(t, "<ok/>", DeleteMsg, "")
	assert.Nil(t, err, "")

	//Get
	ResultMap, err = adapter.GetClusterByName(client, clusterName)
	assert.Nil(t, err, "")
	assert.Equal(t, map[string]string{}, ResultMap, "")

	//Cleanup of Pre-requisite
	DeleteMsg, err = adapter.DeleteClusterControlVe(client, controlVlan, controlVe)
	assert.Equal(t, "<ok/>", DeleteMsg, "")
	assert.Nil(t, err, "")
	DeleteMsg, err = adapter.DeleteClusterControlVlan(client, controlVlan)
	assert.Equal(t, "<ok/>", DeleteMsg, "")
	assert.Nil(t, err, "")
	DeleteMsg, err = adapter.DeleteInterfacePo(client, portChannel)
	assert.Equal(t, "<ok/>", DeleteMsg, "")
	assert.Nil(t, err, "")
	DeleteMsg, err = adapter.DeleteInterfaceVe(client, controlVe)
	assert.Equal(t, "<ok/>", DeleteMsg, "")
	assert.Nil(t, err, "")

}

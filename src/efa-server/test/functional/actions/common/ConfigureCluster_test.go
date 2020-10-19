package common

import (
	"bytes"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/infra/device/actions/clearfabric"
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/adapter"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/adapter/platform/slx"
	netconf "efa-server/infra/device/client"
	"efa-server/test/functional"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
	"text/template"
	"time"
)

var MCTClusterName = "MyMCTCluster"
var ForcedMCTClusterName = "ForcedMCTCluster"
var MCTClusterID = "121"
var MCTClusterControlVlan = "100"
var MCTClusterControlVlan2 = "200"
var MCTClusterControlVlanDesc = "MCTClusterControlVlan"
var MCTClusterControlVe = "1000"
var MCTClusterControlVe2 = "300"
var MACAgingTimeout = "2400"

var MCTClusterMemberCount = 2

var MCTNode1Id = "5"
var MCTNode2Id = "7"

var MCTNode1PrincPrio = "0"
var MCTNode2PrincPrio = "0"

var MCTPeerIntfType = "Port-channel"
var MCTPeerIntfName = "20"
var MCTPeerIntfName2 = "21"
var MCTPeerIntfName3 = "22"
var MCTPeerIntfDesc = "MCTPeerInterface"
var MCTPeerIntfSpeed = "10000"
var MCTPeerIntfSpeed2 = "1000"

var MCTBgpASN = "100"

var MCTInterNodeLinkIntfType = "Eth"

var MCTNode1PeerIP = "21.0.0.177"
var MCTNode1PeerIP2 = "21.0.1.177"
var MCTNode2PeerIP = "21.0.0.136"
var MCTNode2PeerIP2 = "21.0.1.136"

var MCTNode1RemotePeerIP = "21.0.0.136/24"
var MCTNode1RemotePeerIP2 = "21.0.1.136/24"
var MCTNode2RemotePeerIP = "21.0.0.177/24"
var MCTNode2RemotePeerIP2 = "21.0.1.177/24"

type ClusterInfo struct {
	Node1IP                        string
	Node2IP                        string
	NodeModel                      string
	Node1RemoteNodeConnectingPorts []operation.InterNodeLinkPort
	Node2RemoteNodeConnectingPorts []operation.InterNodeLinkPort
}

var clusterPlatforms = map[string]ClusterInfo{}

func init() {
	if os.Getenv("SKIP_AV") != "1" {
		clusterPlatforms["Avalanche"] = ClusterInfo{
			Node1IP:   functional.ActionsAVAMCTNode1Ip,
			Node2IP:   functional.ActionsAVAMCTNode2Ip,
			NodeModel: adapter.AvalancheType,
			Node1RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{
				operation.InterNodeLinkPort{IntfType: MCTInterNodeLinkIntfType, IntfName: "0/19"},
				operation.InterNodeLinkPort{IntfType: MCTInterNodeLinkIntfType, IntfName: "0/20"},
			},
			Node2RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{
				operation.InterNodeLinkPort{IntfType: MCTInterNodeLinkIntfType, IntfName: "0/19"},
				operation.InterNodeLinkPort{IntfType: MCTInterNodeLinkIntfType, IntfName: "0/20"},
			}}

	}
	if os.Getenv("SKIP_OR") != "1" {
		clusterPlatforms["Orca"] = ClusterInfo{
			Node1IP:   functional.ActionsORCAMCTNode1Ip,
			Node2IP:   functional.ActionsORCAMCTNode2Ip,
			NodeModel: adapter.OrcaType,
			Node1RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{
				operation.InterNodeLinkPort{IntfType: MCTInterNodeLinkIntfType, IntfName: "0/50"},
			},
			Node2RemoteNodeConnectingPorts: []operation.InterNodeLinkPort{
				operation.InterNodeLinkPort{IntfType: MCTInterNodeLinkIntfType, IntfName: "0/50"},
			}}
	}
}

type TestConfigCluster struct {
	FabricName                string
	OperationBitMap           uint64
	ClusterMembers            []*ClusterMember
	DuplicateMacTimer         string
	DuplicateMacTimerMaxCount string
	LegacyMacTimeout          string

	ClusterName        string
	ClusterID          string
	ClusterControlVlan string
	ClusterControlVe   string
}

type ClusterMember struct {
	Host               string
	UserName           string
	Password           string
	Model              string
	LocalASN           string
	NodePeerASN        string
	NodePeerIP         string
	NodePeerEncapType  string
	NodePeerBFDEnabled string
	//Client                    *netconf.NetconfClient
	NodeID                    string
	RemoteNodePeerIP          string
	NodePeerIfType            string
	NodePeerIfName            string
	NodePeerIfSpeed           string
	NodePrincipalPriority     string
	RemoteNodeConnectingPorts []operation.InterNodeLinkPort
}

func getStringFromTemplate(templateName string, dataMap map[string]interface{}) (string, error) {
	t := template.Must(template.New("cluster").Parse(templateName))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, dataMap)
	return tpl.String(), err

}

func configureDataPlaneCluster(t *testing.T, c *TestConfigCluster) {

	cluster := operation.ConfigDataPlaneCluster{FabricName: c.FabricName, OperationBitMap: c.OperationBitMap}
	cluster.DataPlaneClusterMemberNodes = make([]operation.DataPlaneClusterMemberNode, 0)
	for _, m := range c.ClusterMembers {
		mctMember := operation.DataPlaneClusterMemberNode{NodeMgmtIP: m.Host, NodeMgmtUserName: m.UserName, NodeModel: m.Model, NodeMgmtPassword: m.Password,
			NodePeerIP: m.NodePeerIP, NodePeerASN: m.NodePeerASN, NodePeerEncapType: m.NodePeerEncapType, NodePeerBFDEnabled: m.NodePeerBFDEnabled}
		cluster.DataPlaneClusterMemberNodes = append(cluster.DataPlaneClusterMemberNodes, mctMember)
	}

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go configurefabric.ConfigureDataPlaneCluster(ctx, fabricGate, &cluster, false, fabricErrors)
	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)

	assert.Empty(t, Errors)

	for _, member := range c.ClusterMembers {
		m := member
		Client := &netconf.NetconfClient{Host: m.Host, User: m.UserName, Password: m.Password}
		Client.Login()
		defer Client.Close()

		ad := adapter.GetAdapter(member.Model)
		//Fetch BGP details using NetConf
		bgpResponse, err := ad.GetRouterBgp(Client)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(bgpResponse.Neighbors))
		if 1 == len(bgpResponse.Neighbors) {
			assert.Equal(t, member.NodePeerIP, bgpResponse.Neighbors[0].RemoteIP)
			assert.Equal(t, member.NodePeerASN, bgpResponse.Neighbors[0].RemoteAS)
		}

		assert.NotNil(t, bgpResponse.L2VPN)
		assert.Equal(t, 1, len(bgpResponse.L2VPN.Neighbors))
		if 1 == len(bgpResponse.L2VPN.Neighbors) {
			assert.Equal(t, member.NodePeerIP, bgpResponse.L2VPN.Neighbors[0].IPAddress)
			if member.Model == adapter.AvalancheType {
				assert.Equal(t, slx.BGPEncapTypeForRoutingAvalanche, bgpResponse.L2VPN.Neighbors[0].Encapsulation)
			} else if member.Model == adapter.OrcaType || member.Model == adapter.OrcaTType {
				assert.Equal(t, slx.BGPEncapTypeForRoutingOrca, bgpResponse.L2VPN.Neighbors[0].Encapsulation)
			}
			assert.Equal(t, "true", bgpResponse.L2VPN.Neighbors[0].Activate)
		}
	}
}

func configureEvpn(t *testing.T, wg *sync.WaitGroup, sw operation.ConfigSwitch, client *netconf.NetconfClient) {
	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)
	defer wg.Done()

	//Call the Actions
	go configurefabric.ConfigureEvpn(ctx, fabricGate, &sw, false, fabricErrors)
	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	assert.Empty(t, Errors)

	adapter := adapter.GetAdapter(sw.Model)

	evpnResponse, err := adapter.GetEvpnInstance(client)
	assert.Nil(t, err)

	assert.Equal(t, operation.ConfigEVPNRespone{Name: sw.Fabric, DuplicageMacTimerValue: duplicateMacTimer,
		MaxCount: "5", TargetCommunity: "auto", RouteTargetBoth: "true", IgnoreAs: "true"}, evpnResponse)

}

func configureBGP(t *testing.T, wg *sync.WaitGroup, m *ClusterMember) {
	defer wg.Done()

	// create empty router bgp with local-as because ConfigureDataPlaneCluster doesn't add 'local-as' config.
	var bgpRouterCreate = `
		<config>
		       <routing-system xmlns="urn:brocade.com:mgmt:brocade-common-def">
		         <router>
		            <router-bgp xmlns="urn:brocade.com:mgmt:brocade-bgp">
		               <router-bgp-attributes>
		                  <local-as>{{.local_as}}</local-as>
						</router-bgp-attributes>
		            </router-bgp>
		         </router>
		      </routing-system>
		</config>  
		`
	var bgpMap = map[string]interface{}{"local_as": m.LocalASN}
	config, templateError := getStringFromTemplate(bgpRouterCreate, bgpMap)
	assert.Nil(t, templateError)

	Client := &netconf.NetconfClient{Host: m.Host, User: m.UserName, Password: m.Password}
	Client.Login()
	defer Client.Close()

	_, err := Client.EditConfig(config)
	assert.Nil(t, err)

	adapter := adapter.GetAdapter(m.Model)
	localASN, err := adapter.GetLocalAsn(Client)
	assert.Nil(t, err)
	assert.Equal(t, m.LocalASN, localASN)
}

func configureClusterPreRequisites(t *testing.T, c *TestConfigCluster) {

	var wg sync.WaitGroup
	for _, m := range c.ClusterMembers {
		wg.Add(1)
		go configureBGP(t, &wg, m)
	}
	wg.Wait()

	configureDataPlaneCluster(t, c)

	var evpnWg sync.WaitGroup
	for _, m := range c.ClusterMembers {

		Client := &netconf.NetconfClient{Host: m.Host, User: m.UserName, Password: m.Password}
		Client.Login()
		defer Client.Close()

		sw := operation.ConfigSwitch{Fabric: c.FabricName, Host: m.Host, UserName: m.UserName, Password: m.Password,
			DuplicateMacTimer: c.DuplicateMacTimer, DuplicateMaxTimerMaxCount: c.DuplicateMacTimerMaxCount,
			MacAgingTimeout: c.LegacyMacTimeout, Model: m.Model}
		evpnWg.Add(1)
		go configureEvpn(t, &evpnWg, sw, Client)
	}
	evpnWg.Wait()
}

func clearBGPAndEVPN(t *testing.T, wg *sync.WaitGroup, client *netconf.NetconfClient, Model string, eviName string) {
	defer wg.Done()

	adapter := adapter.GetAdapter(Model)

	_, err := adapter.UnconfigureRouterBgp(client)
	assert.Nil(t, err)

	err = adapter.IsRouterBgpPresent(client)
	assert.Equal(t, "No Router BGP", err.Error())

	_, err = adapter.DeleteEvpnInstance(client, eviName)
	assert.Nil(t, err)

	evpnRespone, _ := adapter.GetEvpnInstance(client)
	assert.Empty(t, evpnRespone.Name)
}

func clearClusterPreRequisites(t *testing.T, cluster *TestConfigCluster) {
	var wg sync.WaitGroup
	for _, m := range cluster.ClusterMembers {

		Client := &netconf.NetconfClient{Host: m.Host, User: m.UserName, Password: m.Password}
		Client.Login()
		defer Client.Close()

		wg.Add(1)
		go clearBGPAndEVPN(t, &wg, Client, m.Model, cluster.FabricName)
	}
	wg.Wait()
}

func clearCluster(t *testing.T, c *TestConfigCluster) {
	cluster := operation.ConfigCluster{FabricName: c.FabricName, ClusterName: c.ClusterName, ClusterID: c.ClusterID,
		ClusterControlVlan: c.ClusterControlVlan, ClusterControlVe: c.ClusterControlVe, OperationBitMap: 20}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	for _, m := range c.ClusterMembers {
		clusterMember := operation.ClusterMemberNode{NodeMgmtIP: m.Host, NodeModel: m.Model, NodeMgmtUserName: m.UserName,
			NodeMgmtPassword: m.Password, NodePeerIntfName: m.NodePeerIfName, RemoteNodePeerIP: m.RemoteNodePeerIP}
		cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, clusterMember)
	}

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	clearfabric.ClearManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	time.Sleep(10 * time.Second)

	for _, mctMember := range c.ClusterMembers {
		m := mctMember
		Client := &netconf.NetconfClient{Host: m.Host, User: m.UserName, Password: m.Password}
		Client.Login()
		defer Client.Close()

		adapter := adapter.GetAdapter(mctMember.Model)

		fmt.Println("MCt1", Client.Host)

		vlanMap, err2 := adapter.GetClusterControlVlan(Client, c.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{}, vlanMap)

		clusterMap, err1 := adapter.GetClusterByName(Client, c.ClusterName)
		assert.Nil(t, err1)
		assert.Equal(t, map[string]string{}, clusterMap)

		veMap, err3 := adapter.GetInterfaceVe(Client, c.ClusterControlVe)
		assert.Nil(t, err3)
		assert.Equal(t, map[string]string{}, veMap)

		poMap, err4 := adapter.GetInterfacePo(Client, mctMember.NodePeerIfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{}, poMap)
	}

	clearClusterPreRequisites(t, c)
}

func configureManagementCluster(t *testing.T, c *TestConfigCluster) {
	mgmtcluster := operation.ConfigCluster{FabricName: c.FabricName, ClusterName: c.ClusterName, ClusterID: c.ClusterID, ClusterControlVlan: c.ClusterControlVlan,
		ClusterControlVe: c.ClusterControlVe, OperationBitMap: c.OperationBitMap}
	mgmtcluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)

	for _, m := range c.ClusterMembers {
		clusterMember := operation.ClusterMemberNode{NodeMgmtIP: m.Host, NodeMgmtUserName: m.UserName,
			NodeMgmtPassword: m.Password, NodeID: m.NodeID, NodeModel: m.Model,
			NodePeerIP: m.NodePeerIP, RemoteNodePeerIP: m.RemoteNodePeerIP,
			NodePeerIntfType: m.NodePeerIfType, NodePeerIntfName: m.NodePeerIfName,
			NodePeerIntfSpeed: m.NodePeerIfSpeed, NodePrincipalPriority: m.NodePrincipalPriority}
		clusterMember.RemoteNodeConnectingPorts = m.RemoteNodeConnectingPorts

		mgmtcluster.ClusterMemberNodes = append(mgmtcluster.ClusterMemberNodes, clusterMember)
	}

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go configurefabric.ConfigureManagementCluster(ctx, fabricGate, &mgmtcluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)
}

func TestConfigureManagementCluster(t *testing.T) {

	for name, p := range clusterPlatforms {
		MCTNode1Ip := p.Node1IP
		MCTNode2Ip := p.Node2IP
		Model := p.NodeModel

		Node1RemoteNodeConnectingPorts := p.Node1RemoteNodeConnectingPorts
		Node2RemoteNodeConnectingPorts := p.Node2RemoteNodeConnectingPorts

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			configureClusterMop(MCTNode1Ip, Model, MCTNode2Ip, t, Node1RemoteNodeConnectingPorts, Node2RemoteNodeConnectingPorts)
		})
	}
}

func configureClusterMop(MCTNode1Ip string, Model string, MCTNode2Ip string, t *testing.T, Node1RemoteNodeConnectingPorts []operation.InterNodeLinkPort,
	Node2RemoteNodeConnectingPorts []operation.InterNodeLinkPort) {
	log.SetOutput(os.Stdout)
	client := &netconf.NetconfClient{Host: MCTNode1Ip, User: UserName, Password: Password}
	client.Login()
	defer client.Close()
	detail, _ := ad.GetDeviceDetail(client)

	configCluster := &TestConfigCluster{FabricName: FabricName, OperationBitMap: 5, DuplicateMacTimer: duplicateMacTimer,
		DuplicateMacTimerMaxCount: duplicateMacTimerMaxCount, LegacyMacTimeout: MACAgingTimeout,
		ClusterName: MCTClusterName, ClusterID: MCTClusterID, ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe}
	configCluster.ClusterMembers = make([]*ClusterMember, 0)
	member1 := &ClusterMember{Host: MCTNode1Ip, UserName: UserName, Password: Password, Model: detail.Model,
		LocalASN: MCTBgpASN, NodePeerASN: MCTBgpASN, NodePeerIP: MCTNode1PeerIP, NodePeerEncapType: domain.BGPEncapTypeForCluster,
		NodePeerBFDEnabled: "false", NodeID: MCTNode1Id, RemoteNodePeerIP: MCTNode1RemotePeerIP, NodePeerIfType: MCTPeerIntfType,
		NodePeerIfName: MCTPeerIntfName, NodePeerIfSpeed: MCTPeerIntfSpeed, NodePrincipalPriority: "1"}
	member2 := &ClusterMember{Host: MCTNode2Ip, UserName: UserName, Password: Password, Model: detail.Model,
		LocalASN: MCTBgpASN, NodePeerASN: MCTBgpASN, NodePeerIP: MCTNode2PeerIP, NodePeerEncapType: domain.BGPEncapTypeForCluster,
		NodePeerBFDEnabled: "false", NodeID: MCTNode2Id, RemoteNodePeerIP: MCTNode2RemotePeerIP, NodePeerIfType: MCTPeerIntfType,
		NodePeerIfName: MCTPeerIntfName, NodePeerIfSpeed: MCTPeerIntfSpeed, NodePrincipalPriority: "1"}
	configCluster.ClusterMembers = append(configCluster.ClusterMembers, member1)
	configCluster.ClusterMembers = append(configCluster.ClusterMembers, member2)
	/*for _, m := range configCluster.ClusterMembers {
		m.Client = &netconf.NetconfClient{Host: m.Host, User: m.UserName, Password: m.Password}
		m.Client.Login()
	}
	defer member1.Client.Close()
	defer member2.Client.Close()*/
	clearCluster(t, configCluster)
	configureClusterPreRequisites(t, configCluster)
	member1.RemoteNodeConnectingPorts = Node1RemoteNodeConnectingPorts
	member2.RemoteNodeConnectingPorts = Node2RemoteNodeConnectingPorts
	configureManagementCluster(t, configCluster)
	clearCluster(t, configCluster)
}

package freedom

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/actions/clearfabric"
	"efa-server/infra/device/actions/configurefabric"
	"efa-server/infra/device/actions/deconfigurefabric"
	ad "efa-server/infra/device/adapter"
	deviceclient "efa-server/infra/device/client"
	"efa-server/test/functional"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var FREEDOM = "BR-SLX9140"
var MCTClusterName = "MyMCTCluster"
var ForcedMCTClusterName = "ForcedMCTCluster"
var MCTClusterID = "121"
var MCTClusterControlVlan = "100"
var MCTClusterControlVlan2 = "200"
var MCTClusterControlVlanDesc = "MCTClusterControlVlan"
var MCTClusterControlVe = "1000"
var MCTClusterControlVe2 = "300"

var MCTClusterMemberCount = 2

var MCTNode1Id = "1"
var MCTNode2Id = "2"

var MCTNode1PrincPrio = "0"
var MCTNode2PrincPrio = "0"
var MCTNode1Ip = functional.ActionsMCTNode1Ip
var MCTNode2Ip = functional.ActionsMCTNode2Ip

var InterNodeLinkPort1 = "0/46"
var InterNodeLinkPort2 = "0/47"

var MCTPeerIntfType = "Port-channel"
var MCTPeerIntfName = "64"
var MCTPeerIntfName2 = "1001"
var MCTPeerIntfName3 = "1002"
var MCTPeerIntfDesc = "MCTPeerInterface"
var MCTPeerIntfSpeed = "40000"

var MCTInterNodeLinkIntfType = "Eth"

var MCTNode1PeerIP = "21.0.0.177"
var MCTNode1PeerIP2 = "21.0.1.177"
var MCTNode2PeerIP = "21.0.0.136"
var MCTNode2PeerIP2 = "21.0.1.136"

var MCTNode1RemotePeerIP = "21.0.0.136/24"
var MCTNode1RemotePeerIP2 = "21.0.1.136/24"
var MCTNode2RemotePeerIP = "21.0.0.177/24"
var MCTNode2RemotePeerIP2 = "21.0.1.177/24"

//TestClusterCollection so that we can run in sequence
func TestClusterCollection(t *testing.T) {
	cleanupCluster()
	time.Sleep(1 * time.Minute)

	t.Run("TA_ConfigureCluster", TConfigureCluster)
	t.Run("TB_ConfigureClusterIdempotency", TConfigureClusterIdempotency)
	t.Run("TC_GetManagementClusterStatusOfANode", TGetManagementClusterStatusOfANode)
	t.Run("TD_GetManagementClusterConfigOfANode", TGetManagementClusterConfigOfANode)
	t.Run("TE_UpdateClusterPeerIP", TUpdateClusterPeerIP)
	t.Run("TF_UpdateRemoveClusterPeerIntfMemberPort", TUpdateRemoveClusterPeerIntfMemberPort)
	t.Run("TG_UpdateAddClusterPeerIntfMemberPort", TUpdateAddClusterPeerIntfMemberPort)
	//Strategy for force is to cleanup all the devices and then re-run.
	//t.Run("TH_ConfigureClusterForcibly", TConfigureClusterForcibly)
	t.Run("TI_UnconfigureCluster", TUnconfigureCluster)
	t.Run("TJ_ClearCluster", TClearCluster)

}

func TConfigureCluster(t *testing.T) {

	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: MCTClusterName, ClusterID: MCTClusterID, ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 5}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	mctMember1.NodeID = MCTNode1Id
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	fmt.Println("Ritesh", IntfSpeed)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)

	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	mctMember2.NodeID = MCTNode2Id
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()
	detail, _ = ad.GetDeviceDetail(deviceclient2)
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	fmt.Println("Ritesh", IntfSpeed)
	mctMember2.NodeModel = detail.Model
	deviceclient2.Close()

	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go configurefabric.ConfigureManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	fmt.Println(Errors)
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		/*nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		assert.Nil(t, err1)
		assert.Equal(t, map[string] string{"cluster-priority":mctMember.NodePrincipalPriority, "node-id":mctMember.NodeID}, nodeIdMap)*/

		controlVlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{"control-ve": cluster.ClusterControlVe, "description": MCTClusterControlVlanDesc}, controlVlanMap)

		controlVeMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Equal(t, map[string]string{"ip-address": mctMember.RemoteNodePeerIP}, controlVeMap)
		assert.Nil(t, err3)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{"speed": mctMember.NodePeerIntfSpeed, "description": MCTPeerIntfDesc}, poMap)

		phyMap1, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err5)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap1)

		phyMap2, err6 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		assert.Nil(t, err6)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap2)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Equal(t, map[string]string{"cluster-id": cluster.ClusterID, "cluster-control-vlan": cluster.ClusterControlVlan, "peer-type": mctMember.NodePeerIntfType, "peer-name": mctMember.NodePeerIntfName, "peer-ip": mctMember.NodePeerIP, "df-load-balance": "true", "deploy": "true"}, clusterMap)
		assert.Nil(t, err7)
	}
}

func TConfigureClusterIdempotency(t *testing.T) {
	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: MCTClusterName, ClusterID: MCTClusterID, ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 5}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	mctMember1.NodeID = MCTNode1Id
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	mctMember2.NodeID = MCTNode2Id
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()

	detail, _ = ad.GetDeviceDetail(deviceclient2)
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	mctMember2.NodeModel = detail.Model
	deviceclient2.Close()

	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go configurefabric.ConfigureManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		/*nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		assert.Nil(t, err1)
		assert.Equal(t, map[string] string{"cluster-priority":mctMember.NodePrincipalPriority, "node-id":mctMember.NodeID}, nodeIdMap)*/

		controlVlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{"control-ve": cluster.ClusterControlVe, "description": MCTClusterControlVlanDesc}, controlVlanMap)

		controlVeMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Equal(t, map[string]string{"ip-address": mctMember.RemoteNodePeerIP}, controlVeMap)
		assert.Nil(t, err3)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{"speed": mctMember.NodePeerIntfSpeed, "description": MCTPeerIntfDesc}, poMap)

		phyMap1, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err5)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap1)

		phyMap2, err6 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		assert.Nil(t, err6)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap2)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Equal(t, map[string]string{"cluster-id": cluster.ClusterID, "cluster-control-vlan": cluster.ClusterControlVlan, "peer-type": mctMember.NodePeerIntfType, "peer-name": mctMember.NodePeerIntfName, "peer-ip": mctMember.NodePeerIP, "df-load-balance": "true", "deploy": "true"}, clusterMap)
		assert.Nil(t, err7)
	}
}

func TGetManagementClusterStatusOfANode(t *testing.T) {
	var mctMember1 operation.ClusterMemberNode
	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)

	mctMember1.NodeModel = detail.Model
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient1.Close()

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(1)
	assert.Empty(t, Errors)

	mgmtClusterStatusChannel := make(chan operation.MgmtClusterStatus, 1)
	//rawXmlOp, mgmtClusterStatus, err :=
	go configurefabric.GetManagementClusterStatusOfANode(ctx, fabricGate, &mctMember1, fabricErrors, mgmtClusterStatusChannel)
	//fmt.Println("Raw o/p of show-cluster-management\n", rawXmlOp, err)
	//fmt.Println("mgmtClusterStatus", mgmtClusterStatus)

	mgmtClusterStatus := <-mgmtClusterStatusChannel
	fmt.Println(mgmtClusterStatus.TotalMemberNodeCount)

	var expectedClusterStatus operation.MgmtClusterStatus
	expectedClusterStatus.PrincipalNodeMac = "60:9C:9F:B1:18:00"
	expectedClusterStatus.TotalMemberNodeCount = "2"
	expectedClusterStatus.DisconnectedMemberNodeCount = "0"

	var expectedNode1, expectedNode2 operation.MgmtClusterMemberNode

	expectedNode1.NodeSerial = "EXH3327M021"
	expectedNode1.NodeMac = "60:9C:9F:87:3B:00"
	expectedNode1.NodeMgmtIP = MCTNode2Ip
	expectedNode1.NodeInternalIP = "21.0.0.136"
	expectedNode1.NodeID = "5"
	expectedNode1.NodeCondition = "Good"
	expectedNode1.NodeStatus = "Secondary Connected To Cluster"
	expectedNode1.NodeIsPrincipal = "false"
	expectedNode1.NodeIsLOcal = "true"
	expectedNode1.NodeSwitchType = "BR-SLX9140"
	expectedNode1.NodeFwVersion = "v17s.1.02"
	expectedClusterStatus.MemberNodes = append(expectedClusterStatus.MemberNodes, expectedNode1)

	expectedNode2.NodeSerial = "EXH3319M01D"
	expectedNode2.NodeMac = "60:9C:9F:B1:18:00"
	expectedNode2.NodeMgmtIP = MCTNode1Ip
	expectedNode2.NodeInternalIP = "21.0.0.177"
	expectedNode2.NodeID = "7"
	expectedNode2.NodeCondition = "Good"
	expectedNode2.NodeStatus = "Primary"
	expectedNode2.NodeIsPrincipal = "true"
	expectedNode2.NodeIsLOcal = "false"
	expectedNode2.NodeSwitchType = "BR-SLX9140"
	expectedNode2.NodeFwVersion = "v17s.1.02"
	expectedClusterStatus.MemberNodes = append(expectedClusterStatus.MemberNodes, expectedNode2)

	//assert.Equal(t, mgmtClusterStatus, expectedClusterStatus)
	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

}

func TGetManagementClusterConfigOfANode(t *testing.T) {
	var mctMember1 operation.ClusterMemberNode
	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)

	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(1)
	assert.Empty(t, Errors)

	mgmtClusterConfigChannel := make(chan operation.ConfigCluster, 1)
	go configurefabric.GetManagementClusterConfigOfANode(ctx, fabricGate, &mctMember1, fabricErrors, mgmtClusterConfigChannel)

	mgmtClusterStatus := <-mgmtClusterConfigChannel
	fmt.Println(mgmtClusterStatus)
}

func TUpdateClusterPeerIP(t *testing.T) {

	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: MCTClusterName, ClusterID: MCTClusterID, ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 3}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	mctMember1.NodeID = MCTNode1Id
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()
	detail, _ := ad.GetDeviceDetail(deviceclient1)
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)

	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP2
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP2
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	mctMember2.NodeID = MCTNode2Id
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()

	detail, _ = ad.GetDeviceDetail(deviceclient2)
	mctMember2.NodeModel = detail.Model
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient2.Close()

	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP2
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP2
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go configurefabric.UpdateManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		//nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		//assert.Nil(t, err1)
		//assert.Equal(t, map[string] string{"cluster-priority":mctMember.NodePrincipalPriority, "node-id":mctMember.NodeID}, nodeIdMap)

		controlVlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{"control-ve": cluster.ClusterControlVe, "description": configurefabric.MgmtClusterControlVlanDesc}, controlVlanMap)

		controlVeMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Equal(t, map[string]string{"ip-address": mctMember.RemoteNodePeerIP}, controlVeMap)
		assert.Nil(t, err3)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{"speed": mctMember.NodePeerIntfSpeed, "description": configurefabric.MgmtClusterPeerIntfDesc}, poMap)

		phyMap1, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err5)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap1)

		phyMap2, err6 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		assert.Nil(t, err6)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap2)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Equal(t, map[string]string{"cluster-id": cluster.ClusterID, "cluster-control-vlan": cluster.ClusterControlVlan, "peer-type": mctMember.NodePeerIntfType, "peer-name": mctMember.NodePeerIntfName, "peer-ip": mctMember.NodePeerIP, "df-load-balance": "true", "deploy": "true"}, clusterMap)
		assert.Nil(t, err7)

	}
}

// Test case to remove a peer-interface from PO and update the PO speed.
func TUpdateRemoveClusterPeerIntfMemberPort(t *testing.T) {
	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: MCTClusterName, ClusterID: MCTClusterID,
		ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 8 + 16}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	mctMember1.NodeID = MCTNode1Id
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)

	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP2
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP2
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	//mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed2
	mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	//mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts,interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	mctMember2.NodeID = MCTNode2Id
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()

	detail, _ = ad.GetDeviceDetail(deviceclient2)
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	mctMember2.NodeModel = detail.Model
	deviceclient2.Close()

	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP2
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP2
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	//mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed2
	mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	//mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts,interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	configurefabric.UpdateManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		//mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		//nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		//assert.Nil(t, err1)
		//assert.Equal(t, map[string] string{"cluster-priority":mctMember.NodePrincipalPriority, "node-id":mctMember.NodeID}, nodeIdMap)

		controlVlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{"control-ve": cluster.ClusterControlVe, "description": MCTClusterControlVlanDesc}, controlVlanMap)

		controlVeMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Equal(t, map[string]string{"ip-address": mctMember.RemoteNodePeerIP}, controlVeMap)
		assert.Nil(t, err3)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{"speed": mctMember.NodePeerIntfSpeed, "description": MCTPeerIntfDesc}, poMap)

		phyMap1, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err5)
		assert.Equal(t, map[string]string{}, phyMap1)

		//phyMap2, err6 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		//assert.Nil(t, err6)
		//assert.Equal(t, map[string]string{"port-channel-mode":"active", "port-channel-type":"standard", "description":"clusterPeerIntfMember", "port-channel":mctMember.NodePeerIntfName}, phyMap2)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Equal(t, map[string]string{"cluster-id": cluster.ClusterID, "cluster-control-vlan": cluster.ClusterControlVlan, "peer-type": mctMember.NodePeerIntfType, "peer-name": mctMember.NodePeerIntfName, "peer-ip": mctMember.NodePeerIP, "df-load-balance": "true", "deploy": "true"}, clusterMap)
		assert.Nil(t, err7)

	}
}

func TUpdateAddClusterPeerIntfMemberPort(t *testing.T) {

	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: MCTClusterName, ClusterID: MCTClusterID,
		ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 4 + 16}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	mctMember1.NodeID = MCTNode1Id
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)

	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP2
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP2
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	mctMember2.NodeID = MCTNode2Id
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()

	detail, _ = ad.GetDeviceDetail(deviceclient2)
	mctMember2.NodeModel = detail.Model
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient2.Close()

	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP2
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP2
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	configurefabric.UpdateManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		//nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		//assert.Nil(t, err1)
		//assert.Equal(t, map[string] string{"cluster-priority":mctMember.NodePrincipalPriority, "node-id":mctMember.NodeID}, nodeIdMap)

		controlVlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{"control-ve": cluster.ClusterControlVe, "description": MCTClusterControlVlanDesc}, controlVlanMap)

		controlVeMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Equal(t, map[string]string{"ip-address": mctMember.RemoteNodePeerIP}, controlVeMap)
		assert.Nil(t, err3)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{"speed": mctMember.NodePeerIntfSpeed, "description": MCTPeerIntfDesc}, poMap)

		phyMap1, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err5)
		fmt.Println(phyMap1)
		//assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap1)

		phyMap2, err6 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		assert.Nil(t, err6)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap2)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Equal(t, map[string]string{"cluster-id": cluster.ClusterID, "cluster-control-vlan": cluster.ClusterControlVlan, "peer-type": mctMember.NodePeerIntfType, "peer-name": mctMember.NodePeerIntfName, "peer-ip": mctMember.NodePeerIP, "df-load-balance": "true", "deploy": "true"}, clusterMap)
		assert.Nil(t, err7)

	}
}

func TConfigureClusterForcibly(t *testing.T) {
	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: ForcedMCTClusterName, ClusterID: MCTClusterID, ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 15}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)

	mctMember1.NodeModel = detail.Model
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient1.Close()

	mctMember1.NodeID = MCTNode1Id
	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	mctMember1.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()
	detail, _ = ad.GetDeviceDetail(deviceclient2)
	mctMember2.NodeModel = detail.Model
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient2.Close()

	mctMember2.NodeID = MCTNode2Id
	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	mctMember2.NodePeerIntfSpeed = MCTPeerIntfSpeed
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(2 * MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go configurefabric.ConfigureManagementCluster(ctx, fabricGate, &cluster, true, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		//nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		//assert.Nil(t, err1)
		//assert.Equal(t, map[string] string{"cluster-priority":mctMember.NodePrincipalPriority, "node-id":mctMember.NodeID}, nodeIdMap)

		controlVlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{"control-ve": cluster.ClusterControlVe, "description": MCTClusterControlVlanDesc}, controlVlanMap)

		controlVeMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Equal(t, map[string]string{"ip-address": mctMember.RemoteNodePeerIP}, controlVeMap)
		assert.Nil(t, err3)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{"speed": mctMember.NodePeerIntfSpeed, "description": MCTPeerIntfDesc}, poMap)

		phyMap1, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err5)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap1)

		phyMap2, err6 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		assert.Nil(t, err6)
		assert.Equal(t, map[string]string{"port-channel-mode": "active", "port-channel-type": "standard", "description": "clusterPeerIntfMember", "port-channel": mctMember.NodePeerIntfName}, phyMap2)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Equal(t, map[string]string{"cluster-id": cluster.ClusterID, "cluster-control-vlan": cluster.ClusterControlVlan, "peer-type": mctMember.NodePeerIntfType, "peer-name": mctMember.NodePeerIntfName, "peer-ip": mctMember.NodePeerIP, "df-load-balance": "true", "deploy": "true"}, clusterMap)
		assert.Nil(t, err7)

	}
}

func TUnconfigureCluster(t *testing.T) {
	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: MCTClusterName, ClusterID: MCTClusterID, ClusterControlVlan: MCTClusterControlVlan, ClusterControlVe: MCTClusterControlVe, OperationBitMap: 20}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	mctMember1.NodeID = MCTNode1Id
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)

	mctMember1.NodeModel = detail.Model
	deviceclient1.Close()

	mctMember1.NodePrincipalPriority = MCTNode1PrincPrio
	mctMember1.NodePeerIP = MCTNode1PeerIP
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP
	mctMember1.NodePeerIntfType = MCTPeerIntfType
	mctMember1.NodePeerIntfName = MCTPeerIntfName
	mctMember1.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	var interNodeLink1, interNodeLink2 operation.InterNodeLinkPort
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember1.RemoteNodeConnectingPorts = append(mctMember1.RemoteNodeConnectingPorts, interNodeLink2)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	mctMember2.NodeID = MCTNode2Id
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()

	detail, _ = ad.GetDeviceDetail(deviceclient2)
	mctMember2.NodeModel = detail.Model
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient2.Close()

	mctMember2.NodePrincipalPriority = MCTNode2PrincPrio
	mctMember2.NodePeerIP = MCTNode2PeerIP
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP
	mctMember2.NodePeerIntfType = MCTPeerIntfType
	mctMember2.NodePeerIntfName = MCTPeerIntfName
	mctMember2.RemoteNodeConnectingPorts = make([]operation.InterNodeLinkPort, 0)
	interNodeLink1.IntfType = MCTInterNodeLinkIntfType
	interNodeLink1.IntfName = InterNodeLinkPort1
	interNodeLink2.IntfType = MCTInterNodeLinkIntfType
	interNodeLink2.IntfName = InterNodeLinkPort2
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink1)
	mctMember2.RemoteNodeConnectingPorts = append(mctMember2.RemoteNodeConnectingPorts, interNodeLink2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	go deconfigurefabric.UnconfigureManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {

		mctNodeInterNodeLinkPort1 := mctMember.RemoteNodeConnectingPorts[0]
		mctNodeInterNodeLinkPort2 := mctMember.RemoteNodeConnectingPorts[1]

		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		//nodeIdMap, err1 := adapter.GetNodeIdAndPrincipalPriority(deviceclient)
		//assert.Nil(t, err1)
		//assert.Equal(t, map[string]string{"node-id":mctMember.NodeID}, nodeIdMap)

		vlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{}, vlanMap)

		veMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Nil(t, err3)
		assert.Equal(t, map[string]string{}, veMap)

		phyMap, err4 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort1.IntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{}, phyMap)

		phyMap2, err5 := adapter.GetInterfacePoMember(deviceclient, mctNodeInterNodeLinkPort2.IntfName)
		assert.Nil(t, err5)
		assert.Equal(t, map[string]string{}, phyMap2)

		poMap, err6 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err6)
		assert.Equal(t, map[string]string{}, poMap)

		clusterMap, err7 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Nil(t, err7)
		assert.Equal(t, map[string]string{}, clusterMap)
	}

}

func cleanupCluster() {
	devices := []string{MCTNode1Ip, MCTNode2Ip}
	for _, device := range devices {
		go cleanupClusterOnANode(device)
	}

}
func cleanupClusterOnANode(device string) {
	{
		fmt.Println("Cleanup on host:", device)
		deviceclient := &deviceclient.NetconfClient{Host: device, User: UserName, Password: Password}
		deviceclient.Login()
		defer deviceclient.Close()

		detail, _ := ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)
		clusterName, clusterID, clusterVlan, _, clusterPeerIfName, _, clusErr := adapter.GetCluster(deviceclient)
		fmt.Println("Cleanup on host:", deviceclient.Host)
		if clusErr != nil {
			fmt.Println("Cleanup on err:", clusErr)
			return
		}
		fmt.Println(clusterName, clusterID, clusterVlan, clusterPeerIfName, clusErr)

		if clusterName != "" {
			adapter.DeleteCluster(deviceclient, clusterName, clusterID)
		}

		if clusterVlan != "" {
			ResultMap, _ := adapter.GetClusterControlVlan(deviceclient, clusterVlan)

			clusterVe := ResultMap["control-ve"]

			if clusterVe != "" {

				adapter.DeleteClusterControlVe(deviceclient, clusterVlan, clusterVe)

				adapter.DeleteInterfaceVe(deviceclient, clusterVe)

			}

			adapter.DeleteClusterControlVlan(deviceclient, clusterVlan)

		}
		if clusterPeerIfName != "" {

			adapter.DeleteInterfacePo(deviceclient, clusterPeerIfName)

		}
	}

}
func TClearCluster(t *testing.T) {
	cluster := operation.ConfigCluster{FabricName: FabricName, ClusterName: ForcedMCTClusterName, ClusterID: MCTClusterID,
		ClusterControlVlan: MCTClusterControlVlan2, ClusterControlVe: MCTClusterControlVe2, OperationBitMap: 20}

	cluster.ClusterMemberNodes = make([]operation.ClusterMemberNode, 0)
	var mctMember1, mctMember2 operation.ClusterMemberNode

	mctMember1.NodeMgmtIP = MCTNode1Ip
	mctMember1.NodeMgmtUserName = UserName
	mctMember1.NodeMgmtPassword = Password
	//Set the Model+Version based on the Switch details
	deviceclient1 := &deviceclient.NetconfClient{Host: mctMember1.NodeMgmtIP, User: mctMember1.NodeMgmtUserName, Password: mctMember1.NodeMgmtPassword}
	deviceclient1.Login()

	detail, _ := ad.GetDeviceDetail(deviceclient1)

	mctMember1.NodeModel = detail.Model
	IntfSpeed, _ := ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient1, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient1.Close()

	mctMember1.NodePeerIntfName = MCTPeerIntfName2
	mctMember1.RemoteNodePeerIP = MCTNode1RemotePeerIP2
	configureMCTClusterDefaults(cluster, mctMember1)

	mctMember2.NodeMgmtIP = MCTNode2Ip
	mctMember2.NodeMgmtUserName = UserName
	mctMember2.NodeMgmtPassword = Password
	//Set the Model+Version based on the Switch details
	deviceclient2 := &deviceclient.NetconfClient{Host: mctMember2.NodeMgmtIP, User: mctMember2.NodeMgmtUserName, Password: mctMember2.NodeMgmtPassword}
	deviceclient2.Login()

	detail, _ = ad.GetDeviceDetail(deviceclient2)
	mctMember2.NodeModel = detail.Model
	IntfSpeed, _ = ad.GetAdapter(detail.Model).GetInterfaceSpeed(deviceclient2, MCTInterNodeLinkIntfType, InterNodeLinkPort1)
	MCTPeerIntfSpeed = fmt.Sprint(IntfSpeed)
	deviceclient2.Close()

	mctMember2.NodePeerIntfName = MCTPeerIntfName3
	mctMember2.RemoteNodePeerIP = MCTNode2RemotePeerIP2
	configureMCTClusterDefaults(cluster, mctMember2)

	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember1)
	cluster.ClusterMemberNodes = append(cluster.ClusterMemberNodes, mctMember2)

	ctx, fabricGate, fabricErrors, Errors := initializeTestWithoutNetconfClient(MCTClusterMemberCount)
	assert.Empty(t, Errors)

	clearfabric.ClearManagementCluster(ctx, fabricGate, &cluster, false, fabricErrors)

	Errors = closeActionChannel(fabricGate, fabricErrors, Errors)
	//Check that action throws no Errors
	assert.Empty(t, Errors)

	for _, mctMember := range cluster.ClusterMemberNodes {
		deviceclient := &deviceclient.NetconfClient{Host: mctMember.NodeMgmtIP, User: mctMember.NodeMgmtUserName, Password: mctMember.NodeMgmtPassword}
		deviceclient.Login()
		defer deviceclient.Close()
		detail, _ = ad.GetDeviceDetail(deviceclient)
		adapter := ad.GetAdapter(detail.Model)

		clusterMap, err1 := adapter.GetClusterByName(deviceclient, cluster.ClusterName)
		assert.Nil(t, err1)
		assert.Equal(t, map[string]string{}, clusterMap)

		vlanMap, err2 := adapter.GetClusterControlVlan(deviceclient, cluster.ClusterControlVlan)
		assert.Nil(t, err2)
		assert.Equal(t, map[string]string{}, vlanMap)

		veMap, err3 := adapter.GetInterfaceVe(deviceclient, cluster.ClusterControlVe)
		assert.Nil(t, err3)
		assert.Equal(t, map[string]string{}, veMap)

		poMap, err4 := adapter.GetInterfacePo(deviceclient, mctMember.NodePeerIntfName)
		assert.Nil(t, err4)
		assert.Equal(t, map[string]string{}, poMap)
	}
}

func configureMCTClusterDefaults(cluster operation.ConfigCluster, mctNode operation.ClusterMemberNode) {

	deviceclient := &deviceclient.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	deviceclient.Login()
	defer deviceclient.Close()
	dummyVlan := "400"
	BFDMultiplier := "3"
	BFDRx := "300"
	BFDTx := "300"
	detail, _ := ad.GetDeviceDetail(deviceclient)
	adapter := ad.GetAdapter(detail.Model)

	/* Create vlan */
	adapter.CreateClusterControlVlan(deviceclient, cluster.ClusterControlVlan, cluster.ClusterControlVe, configurefabric.MgmtClusterControlVlanDesc)
	/* Create ve */
	adapter.ConfigureInterfaceVe(deviceclient, cluster.ClusterControlVe, mctNode.RemoteNodePeerIP, BFDRx, BFDTx, BFDMultiplier)
	/* Create port-channel */
	//dummyVlan need not be created as this is vlan is added to PO only in Avalanche or Orca.
	adapter.CreateInterfacePo(deviceclient, mctNode.NodePeerIntfName, mctNode.NodePeerIntfSpeed, configurefabric.MgmtClusterPeerIntfDesc, dummyVlan)
}

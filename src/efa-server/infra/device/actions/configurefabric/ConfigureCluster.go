/*
Contains Actions to configure the cluster on Extreme netowking equipment.
*/

package configurefabric

import (
	nlog "github.com/sirupsen/logrus"

	"context"
	"errors"
	"fmt"
	"sync"

	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"efa-server/infra/device/actions/clearfabric"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"

	"efa-server/infra/device/adapter/interface"
	"net"
)

//MgmtClusterControlVlanDesc implies the description of cluster control VLAN.
var MgmtClusterControlVlanDesc = "MCTClusterControlVlan"

//MgmtClusterPeerIntfDesc implies the description of cluster peer interface.
var MgmtClusterPeerIntfDesc = "MCTPeerInterface"

//MgmtClusterPeerIPUpdate implies the bit to indicate an update of cluster peer ip.
var MgmtClusterPeerIPUpdate uint64 = 1 << 0

//MgmtClusterRemotePeerIPUpdate implies the bit to indicate an update of cluster remote peer ip.
var MgmtClusterRemotePeerIPUpdate uint64 = 1 << 1

//MgmtClusterPeerIntfMemberAdd implies addition of a member port to the peer interface (port-channel)
var MgmtClusterPeerIntfMemberAdd uint64 = 1 << 2

//MgmtClusterPeerIntfMemberDel implies deletion of a member port from the peer interface (port-channel)
var MgmtClusterPeerIntfMemberDel uint64 = 1 << 3

//MgmtClusterPeerIntfSpeedUpdate implies the updation of the port-channel speed
var MgmtClusterPeerIntfSpeedUpdate uint64 = 1 << 4

//ConfigureManagementCluster is used to configure "cluster <name> <id>" and all its sub-config on the switch.
func ConfigureManagementCluster(ctx context.Context, wg *sync.WaitGroup, cluster *operation.ConfigCluster,
	force bool, clusterConfigErrors chan actions.OperationError) {
	defer wg.Done()
	/* When the --force option is set,
	all the cluster config needs to be cleaned up for the intended config to succeed. */
	if force == true {
		var clusterConfigWaitGroup sync.WaitGroup
		for iter := range cluster.ClusterMemberNodes {
			mctNode := cluster.ClusterMemberNodes[iter]
			clusterConfigWaitGroup.Add(1)
			go clearfabric.CleanupManagementClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, clusterConfigErrors)
		}
		clusterConfigWaitGroup.Wait()

		/* Check the status of management cluster from all the nodes of management cluster
		Each node should have only 1 member (itself), in the management cluster. */
		var clusterOperWaitGroup sync.WaitGroup
		for iter := range cluster.ClusterMemberNodes {
			mctNode := cluster.ClusterMemberNodes[iter]
			clusterOperWaitGroup.Add(1)
			var intendedCluster operation.ConfigCluster
			intendedCluster.ClusterMemberNodes = append(intendedCluster.ClusterMemberNodes, mctNode)
			go actions.PollManagementClusterStatusOnANode(ctx, &clusterOperWaitGroup, &intendedCluster, &mctNode, clusterConfigErrors)
		}
		clusterOperWaitGroup.Wait()
	}
	//Temporary channel to capture the errors during Configuration
	mctErrors := make(chan actions.OperationError, 2*2)
	/* Config push to all the nodes of the management cluster */
	var clusterConfigWaitGroup sync.WaitGroup
	for iter := range cluster.ClusterMemberNodes {
		mctNode := cluster.ClusterMemberNodes[iter]
		clusterConfigWaitGroup.Add(1)
		go configureManagementClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, mctErrors)
	}
	if isErrorPresent(&clusterConfigWaitGroup, mctErrors, clusterConfigErrors) {
		return
	}

	/* Check the status of management cluster from all the nodes of management cluster */
	var clusterOperWaitGroup sync.WaitGroup
	for iter := range cluster.ClusterMemberNodes {
		mctNode := cluster.ClusterMemberNodes[iter]
		clusterOperWaitGroup.Add(1)
		go actions.PollManagementClusterStatusOnANode(ctx, &clusterOperWaitGroup, cluster, &mctNode, clusterConfigErrors)
	}
	clusterOperWaitGroup.Wait()
}

func isErrorPresent(sourceWaitGroup *sync.WaitGroup, sourceChannel, destinationChannel chan actions.OperationError) bool {
	go func() {
		sourceWaitGroup.Wait()
		close(sourceChannel)

	}()
	//Check for errors in the sub-action
	configErrors := false
	for err := range sourceChannel {
		destinationChannel <- err
		configErrors = true
	}
	if configErrors {
		return true
	}
	return false
}

//UpdateManagementCluster is used to update "cluster <name> <id>" and all its sub-config on the switch.
func UpdateManagementCluster(ctx context.Context, wg *sync.WaitGroup, cluster *operation.ConfigCluster,
	force bool, clusterConfigErrors chan actions.OperationError) {
	defer wg.Done()

	/* Config push to all the nodes of the management cluster */
	var clusterConfigWaitGroup sync.WaitGroup
	for iter := range cluster.ClusterMemberNodes {
		mctNode := cluster.ClusterMemberNodes[iter]
		clusterConfigWaitGroup.Add(1)
		go updateManagementClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, clusterConfigErrors)
	}
	clusterConfigWaitGroup.Wait()

	/* Check the status of management cluster from all the nodes of management cluster */
	// TODO : Revisit whether the below polling is needed.
	var clusterOperWaitGroup sync.WaitGroup
	for iter := range cluster.ClusterMemberNodes {
		mctNode := cluster.ClusterMemberNodes[iter]
		clusterOperWaitGroup.Add(1)
		go actions.PollManagementClusterStatusOnANode(ctx, &clusterOperWaitGroup, cluster, &mctNode, clusterConfigErrors)
	}
	clusterOperWaitGroup.Wait()
}

func configureManagementClusterOnANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup, cluster *operation.ConfigCluster,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan actions.OperationError) {

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    cluster.FabricName,
		"Operation": "Configure Management Cluster",
		"Switch":    mctNode.NodeMgmtIP,
	})

	defer clusterConfigWaitGroup.Done()

	/*Netconf client*/
	client := &netconf.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr := client.Login()
	if loginErr != nil {
		log.Infof("NETCONF Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer client.Close()

	/*SSH client*/
	adapter := ad.GetAdapter(mctNode.NodeModel)
	sshClient := &netconf.SSHClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr = sshClient.Login()
	if loginErr != nil {
		log.Infof("SSH Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Exec Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer sshClient.Close()

	/* Configure Node Id.*/
	log.Infof("Configuring node-id<%s> for the node<%s>", mctNode.NodeID, mctNode.NodeMgmtIP)
	err := adapter.ConfigureNodeID(sshClient, mctNode.NodeID)
	if err != nil {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Node Id", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	/* Configure Principal Priority for a given node*/
	/*	if (mctNode.NodePrincipalPriority) != operation.MgmtClusterNodePrinciPrioDefault {
		log.Infof("Configuring principal priority<%s> for the node<%s>", mctNode.NodePrincipalPriority, mctNode.NodeMgmtIP)
		x, err := adapter.ConfigureNodePrincipalPriority(client, mctNode.NodeID, "true", mctNode.NodePrincipalPriority)
		if (x != "<ok/>") || (err != nil) {
			clusterConfigErrors <- actions.OperationError{Operation: "Configure Principal Priority", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
	}*/

	/* Create control VLAN and associate the control VLAN with the control VE */
	log.Infof("Configuring control VLAN<%s> and its association with the control VE<%s>", cluster.ClusterControlVlan, cluster.ClusterControlVe)
	x, err := adapter.CreateClusterControlVlan(client, cluster.ClusterControlVlan, cluster.ClusterControlVe, MgmtClusterControlVlanDesc)
	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "Create Cluster Control VLAN", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	/* Create port-channel interface (with the ISL members) to be used as MCT peer interface */
	log.Infof("Configuring cluster peer interface<Type %s/Name %s/Vlan %s>", mctNode.NodePeerIntfType, mctNode.NodePeerIntfName, cluster.ClusterControlVlan)
	x, err = configureClusterPeerInterface(adapter, client, mctNode.NodePeerIntfType, mctNode.NodePeerIntfName,
		mctNode.NodePeerIntfSpeed, mctNode.RemoteNodeConnectingPorts, cluster.ClusterControlVlan)
	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Cluster Peer Interface", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	/* Create control VE and configure the IP address on the same. */
	log.Infof("Configuring control VE<%s> and its associated IP address<%s>", cluster.ClusterControlVe, mctNode.RemoteNodePeerIP)
	x, err = adapter.ConfigureInterfaceVe(client, cluster.ClusterControlVe, mctNode.RemoteNodePeerIP, mctNode.BFDRx, mctNode.BFDTx,
		mctNode.BFDMultiplier)
	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure IRB on Cluster Control VLAN", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	x, err = adapter.ConfigureIPRoute(client, mctNode.NodePeerLoopbackIP+"/32", mctNode.NodePeerIP)
	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure IP Route Failed", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	/* Create MCT Cluster */
	remoteNodePeerIP, _, err := net.ParseCIDR(mctNode.RemoteNodePeerIP)
	if err != nil {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Cluster", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	log.Infof("Configuring Cluster Name:<%s>, Id<%s>, PeerIfType<%s>, PeerIfName<%s>, PeerIp<%s> SourceIp<%s>",
		cluster.ClusterName, cluster.ClusterID, mctNode.NodePeerIntfType, mctNode.NodePeerIntfName, mctNode.NodePeerIP, remoteNodePeerIP.String())

	x, err = configureCluster(adapter, client, cluster.ClusterName, cluster.ClusterID, mctNode.NodePeerIntfType,
		mctNode.NodePeerIntfName, mctNode.NodePeerIP, mctNode.NodePeerLoopbackIP, cluster.ClusterControlVlan, cluster.ClusterControlVe, remoteNodePeerIP.String())

	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Cluster", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	return
}

func updateManagementClusterOnANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup, cluster *operation.ConfigCluster,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan actions.OperationError) {

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    cluster.FabricName,
		"Operation": "Update Management Cluster",
		"Switch":    mctNode.NodeMgmtIP,
	})

	defer clusterConfigWaitGroup.Done()

	if !(isBitSet(cluster.OperationBitMap, MgmtClusterPeerIPUpdate) ||
		isBitSet(cluster.OperationBitMap, MgmtClusterRemotePeerIPUpdate) ||
		isBitSet(cluster.OperationBitMap, MgmtClusterPeerIntfMemberAdd) ||
		isBitSet(cluster.OperationBitMap, MgmtClusterPeerIntfMemberDel) ||
		isBitSet(cluster.OperationBitMap, MgmtClusterPeerIntfSpeedUpdate)) {
		log.Infof("Unsupported update operations %d on %s", cluster.OperationBitMap, mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Update management cluster properties",
			Error: errors.New("Unsupported Operation"), Host: mctNode.NodeMgmtIP}
		return
	}

	/*Netconf client*/
	adapter := ad.GetAdapter(mctNode.NodeModel)
	client := &netconf.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr := client.Login()
	if loginErr != nil {
		log.Infof("NETCONF Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer client.Close()

	clusterName, clusterID, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP, err := adapter.GetCluster(client)

	log.Println(clusterName, clusterID, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP, err)

	if isBitSet(cluster.OperationBitMap, MgmtClusterRemotePeerIPUpdate) {
		ResultMap, err := adapter.GetClusterControlVlan(client, clusterVlan)
		if err != nil {
			clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Get Control VLAN", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
		clusterVe := ResultMap["control-ve"]

		ResultMap, err = adapter.GetInterfaceVe(client, clusterVe)
		if err != nil {
			clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Get Control VE" +
				"", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
		veIPAddr := ResultMap["ip-address"]
		x, err2 := adapter.UnconfigureInterfaceVeIP(client, clusterVe, veIPAddr)
		if err2 != nil {
			clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Delete VE IP", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
		x, err2 = adapter.ConfigureInterfaceVeIP(client, clusterVe, mctNode.RemoteNodePeerIP)
		if err2 != nil {
			clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Set VE IP", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
		fmt.Println(x)
	}

	if isBitSet(cluster.OperationBitMap, MgmtClusterPeerIPUpdate) {
		/* Undeploy the existing cluster */
		x, err := adapter.UndeployCluster(client, clusterName, clusterID)
		if (x != "<ok/>") || (err != nil) {
			clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Undeploy", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
		if clusterPeerIP != "" {
			/* Remove the existing peer-ip */
			x, err = adapter.UnconfigureClusterPeerIP(client, clusterName, clusterID, clusterPeerIP)
			if (x != "<ok/>") || (err != nil) {
				clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Delete Peer IP", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}

		/* Configure the new peer ip and deploy the cluster */
		x, err = adapter.ConfigureClusterPeerIP(client, clusterName, clusterID, mctNode.NodePeerIP)
		if (x != "<ok/>") || (err != nil) {
			clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Set Peer IP", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
	}

	if isBitSet(cluster.OperationBitMap, MgmtClusterPeerIntfMemberAdd) {
		memberPorts := mctNode.RemoteNodeConnectingPorts
		for _, InterNodeLinkPort := range memberPorts {
			x, err := adapter.AddInterfaceToPo(client, InterNodeLinkPort.IntfName, "clusterPeerIntfMember",
				mctNode.NodePeerIntfName, "active", "standard", mctNode.NodePeerIntfSpeed)
			if (x != "<ok/>") || (err != nil) {
				clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster peer interface add member:", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}
	}

	if isBitSet(cluster.OperationBitMap, MgmtClusterPeerIntfMemberDel) {
		memberPorts := mctNode.RemoteNodeConnectingPorts
		for _, InterNodeLinkPort := range memberPorts {
			x, err := adapter.DeleteInterfaceFromPo(client, InterNodeLinkPort.IntfName, mctNode.NodePeerIntfName)
			if (x != "<ok/>") || (err != nil) {
				clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster peer interface delete member:", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}
	}

	/* Update port-channel speed when the members get added/deleted to the same */
	if isBitSet(cluster.OperationBitMap, MgmtClusterPeerIntfSpeedUpdate) {
		x, err := adapter.ConfigureInterfacePoSpeed(client, mctNode.NodePeerIntfName, mctNode.NodePeerIntfSpeed)
		if (x != "<ok/>") || (err != nil) {
			clusterConfigErrors <- actions.OperationError{Operation: "Configure Cluster Peer Interface Speed", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
	}
	return
}

func configureClusterPeerInterface(adapter interfaces.Switch, client *netconf.NetconfClient, peerIntfType string, peerIntfName string,
	peerIntfSpeed string, memberPorts []operation.InterNodeLinkPort, controlVlan string) (string, error) {
	if peerIntfType != "Port-channel" {
		return "", errors.New("Cluster peer-interface needs to be a port-channel")
	}
	x, err := adapter.CreateInterfacePo(client, peerIntfName, peerIntfSpeed, MgmtClusterPeerIntfDesc, controlVlan)
	if (x != "<ok/>") || (err != nil) {
		return x, err
	}

	for _, InterNodeLinkPort := range memberPorts {
		x, err = adapter.AddInterfaceToPo(client, InterNodeLinkPort.IntfName, "clusterPeerIntfMember", peerIntfName, "active", "standard", peerIntfSpeed)
		if (x != "<ok/>") || (err != nil) {
			return x, err
		}
	}
	return "<ok/>", nil
}

//ConfigureCluster is used to configure the cluster
func configureCluster(adapter interfaces.Switch, client *netconf.NetconfClient, clusterName string, clusterID string, peerIfType string,
	peerIfName string, peerIP string, peerLoopbackIP string, controlVlan string, controlVe string, sourceIP string) (string, error) {
	x, err := adapter.CreateCluster(client, clusterName, clusterID, peerIfType, peerIfName, peerIP)
	if (x != "<ok/>") || (err != nil) {
		return x, err
	}
	x, err = adapter.ConfigureCluster(client, clusterName, clusterID, peerIfType, peerIfName, peerIP, peerLoopbackIP, controlVlan, controlVe, sourceIP)
	if (x != "<ok/>") || (err != nil) {
		return x, err
	}

	return "<ok/>", nil
}

// GetManagementClusterStatusOfANode is used to get the management cluster status of a node.
func GetManagementClusterStatusOfANode(ctx context.Context, clusterOperWaitGroup *sync.WaitGroup,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan actions.OperationError, mgmtClusterStatusChan chan operation.MgmtClusterStatus) {

	defer clusterOperWaitGroup.Done()

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		//TODO Fabric Name should be accessed as parameter in one of the structures
		//"Fabric":    "",
		"Operation": "Get Management Cluster Status",
		"Switch":    mctNode.NodeMgmtIP,
	})
	adapter := ad.GetAdapter(mctNode.NodeModel)
	client := &netconf.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr := client.Login()
	if loginErr != nil {
		log.Infof("NETCONF Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer client.Close()

	rawXMLOutput, mgmtClusterStatus, principalNode, err := adapter.GetManagementClusterStatus(client)
	log.Info("Raw o/p of show-cluster-management", rawXMLOutput)
	log.Infof("Principal Node IP obtained on <%s> is <%s>", mctNode.NodeMgmtIP, principalNode)

	if err != nil {
		clusterConfigErrors <- actions.OperationError{Operation: "Get Cluster Management", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	mgmtClusterStatusChan <- operation.MgmtClusterStatus{TotalMemberNodeCount: mgmtClusterStatus.TotalMemberNodeCount,
		DisconnectedMemberNodeCount: mgmtClusterStatus.DisconnectedMemberNodeCount, PrincipalNodeMac: mgmtClusterStatus.PrincipalNodeMac, MemberNodes: mgmtClusterStatus.MemberNodes}
}

//GetManagementClusterConfigOfANode is used to get the management cluster config of a node.
func GetManagementClusterConfigOfANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan actions.OperationError, mgmtClusterConfigChan chan operation.ConfigCluster) {
	defer clusterConfigWaitGroup.Done()

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		//TODO Fabric Name should be accessed as parameter in one of the structures
		"Fabric":    "",
		"Operation": "Get Management Cluster Config",
		"Switch":    mctNode.NodeMgmtIP,
	})

	adapter := ad.GetAdapter(mctNode.NodeModel)
	client := &netconf.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr := client.Login()
	if loginErr != nil {
		log.Infof("NETCONF Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer client.Close()

	var clusterConfig operation.ConfigCluster
	clusterName, clusterID, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP, err := adapter.GetCluster(client)

	clusterConfig.ClusterName = clusterName
	clusterConfig.ClusterID = clusterID
	clusterConfig.ClusterControlVlan = clusterVlan

	ResultMap, err := adapter.GetClusterControlVlan(client, clusterVlan)
	if err != nil {
		clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Get Control VLAN", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	clusterConfig.ClusterControlVe = ResultMap["control-ve"]

	ResultMap, err = adapter.GetInterfaceVe(client, clusterConfig.ClusterControlVe)
	if err != nil {
		clusterConfigErrors <- actions.OperationError{Operation: "MCT Cluster Undeploy", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	var memberNode operation.ClusterMemberNode
	memberNode.NodeMgmtIP = mctNode.NodeMgmtIP
	memberNode.NodeMgmtUserName = mctNode.NodeMgmtUserName
	memberNode.NodeMgmtPassword = mctNode.NodeMgmtPassword

	memberNode.RemoteNodePeerIP = ResultMap["ip-address"]
	memberNode.NodePeerIntfType = clusterPeerIfType
	memberNode.NodePeerIntfName = clusterPeerIfName
	memberNode.NodePeerIP = clusterPeerIP

	ResultMap, err = adapter.GetInterfacePo(client, clusterPeerIfName)
	memberNode.NodePeerIntfSpeed = ResultMap["speed"]

	clusterConfig.ClusterMemberNodes = append(clusterConfig.ClusterMemberNodes, memberNode)

	//TODO : 1. Node-id, Node principal-priority 2. Peer Interface member ports.

	mgmtClusterConfigChan <- clusterConfig
}

func isBitSet(operationBmp uint64, operationBit uint64) bool {
	return ((operationBmp & operationBit) == operationBit)
}

//ConfigureDataPlaneCluster is used to configure MCT cluster data plane.
func ConfigureDataPlaneCluster(ctx context.Context, wg *sync.WaitGroup, cluster *operation.ConfigDataPlaneCluster,
	force bool, clusterConfigErrors chan actions.OperationError) {
	defer wg.Done()

	/* Config push to all the nodes of the management cluster */
	var clusterConfigWaitGroup sync.WaitGroup
	for iter := range cluster.DataPlaneClusterMemberNodes {
		mctNode := cluster.DataPlaneClusterMemberNodes[iter]
		clusterConfigWaitGroup.Add(1)
		go configureDataPlaneClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, clusterConfigErrors)
	}
	clusterConfigWaitGroup.Wait()
}

func configureDataPlaneClusterOnANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup, cluster *operation.ConfigDataPlaneCluster,
	mctNode *operation.DataPlaneClusterMemberNode, clusterConfigErrors chan actions.OperationError) {

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    cluster.FabricName,
		"Operation": "Configure Data Plane Cluster",
		"Switch":    mctNode.NodeMgmtIP,
	})

	defer clusterConfigWaitGroup.Done()

	adapter := ad.GetAdapter(mctNode.NodeModel)
	/*Netconf client*/
	client := &netconf.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr := client.Login()
	if loginErr != nil {
		log.Infof("NETCONF Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer client.Close()

	encapType, _ := adapter.GetRouterBgpL2EVPNNeighborEncapType(client)
	x, err := adapter.ConfigureRouterBgpL2EVPNNeighbor(client, mctNode.NodePeerIP, mctNode.NodePeerLoopbackIP,
		mctNode.NodeLoopBackNumber, mctNode.NodePeerASN, encapType, mctNode.NodePeerBFDEnabled)

	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Data Plane Cluster", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	return
}

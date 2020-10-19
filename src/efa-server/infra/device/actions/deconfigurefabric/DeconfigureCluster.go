package deconfigurefabric

import (
	nlog "github.com/sirupsen/logrus"

	"context"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	"efa-server/infra/device/adapter/interface"
	netconf "efa-server/infra/device/client"
	"sync"
)

//UnconfigureManagementCluster is used to unconfigure "cluster <name> <id>" and all its sub-config on the switch.
func UnconfigureManagementCluster(ctx context.Context, wg *sync.WaitGroup, cluster *operation.ConfigCluster,
	force bool, clusterConfigErrors chan actions.OperationError) {
	defer wg.Done()
	var clusterConfigWaitGroup sync.WaitGroup
	for iter := range cluster.ClusterMemberNodes {
		mctNode := cluster.ClusterMemberNodes[iter]
		clusterConfigWaitGroup.Add(1)
		go unconfigureManagementClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, clusterConfigErrors)
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

func unconfigureManagementClusterOnANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup, cluster *operation.ConfigCluster,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan actions.OperationError) {
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    cluster.FabricName,
		"Operation": "Unconfigure Management Cluster",
		"Switch":    mctNode.NodeMgmtIP,
	})

	defer clusterConfigWaitGroup.Done()

	/* Netconf Client */
	adapter := ad.GetAdapter(mctNode.NodeModel)
	client := &netconf.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	loginErr := client.Login()
	if loginErr != nil {
		log.Infof("NETCONF Login to the host<%s> failed", mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Configure Interface Login", Error: loginErr, Host: mctNode.NodeMgmtIP}
		return
	}
	defer client.Close()

	log.Infof("Deletion of cluster with name<%s> and Id<%s>", cluster.ClusterName, cluster.ClusterID)
	x, err := adapter.DeleteCluster(client, cluster.ClusterName, cluster.ClusterID)
	if (x != "<ok/>") || (err != nil) {
		log.Infof("Deletion of the cluster name<%s> Id<%s> failed on host<%s>", cluster.ClusterName, cluster.ClusterID, mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster config", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	log.Infof("Deletion of the association between control VLAN<%s> and control VE<%s>", cluster.ClusterControlVlan, cluster.ClusterControlVe)
	x, err = adapter.DeleteClusterControlVe(client, cluster.ClusterControlVlan, cluster.ClusterControlVe)
	if (x != "<ok/>") || (err != nil) {
		log.Infof("Deletion of the association between control VLAN<%s> and control VE<%s> failed on host<%s>",
			cluster.ClusterControlVlan, cluster.ClusterControlVe, mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control vlan to ve association", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	log.Infof("Deletion of the control VE<%s>", cluster.ClusterControlVe)
	x, err = adapter.DeleteInterfaceVe(client, cluster.ClusterControlVe)
	if (x != "<ok/>") || (err != nil) {
		log.Infof("Deletion of the control VE<%s> failed on host<%s>", cluster.ClusterControlVe, mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control ve", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	x, err = adapter.DeConfigureIPRoute(client, mctNode.NodePeerLoopbackIP+"/32", mctNode.NodePeerIP)
	if (x != "<ok/>") || (err != nil) {
		clusterConfigErrors <- actions.OperationError{Operation: "DeConfigure IP Route Failed", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	log.Infof("Deletion of the control VLAN<%s>", cluster.ClusterControlVlan)
	x, err = adapter.DeleteClusterControlVlan(client, cluster.ClusterControlVlan)
	if (x != "<ok/>") || (err != nil) {
		log.Infof("Deletion of the control VLAN<%s> failed on host<%s>", cluster.ClusterControlVlan, mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control vlan", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	log.Infof("Unconfigure cluster peer interface if-type<%s> if-name<%s>", mctNode.NodePeerIntfType, mctNode.NodePeerIntfName)
	x, err = unconfigureClusterPeerInterface(log, adapter, client, mctNode.NodePeerIntfName, mctNode.RemoteNodeConnectingPorts)
	if (x != "<ok/>") || (err != nil) {
		log.Infof("Unconfigure cluster peer interface if-type<%s> if-name<%s> failed on host<%s>", mctNode.NodePeerIntfType, mctNode.NodePeerIntfName, mctNode.NodeMgmtIP)
		clusterConfigErrors <- actions.OperationError{Operation: "Unconfigure of cluster peer interface", Error: err, Host: mctNode.NodeMgmtIP}
		return
	}

	log.Infof("Deletion of principal priority for the node<%s>", mctNode.NodeID)
	if (mctNode.NodePrincipalPriority) != operation.MgmtClusterNodePrinciPrioDefault {
		x, err = adapter.UnconfigureNodePrincipalPriority(client, mctNode.NodeID)
		if (x != "<ok/>") || (err != nil) {
			log.Infof("Deletion of principal priority for the host<%s> failed", mctNode.NodeID)
			clusterConfigErrors <- actions.OperationError{Operation: "Deletion of principal priority", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
	}

	return
}

func unconfigureClusterPeerInterface(log *nlog.Entry, adapter interfaces.Switch, client *netconf.NetconfClient, portChannelName string,
	memberPorts []operation.InterNodeLinkPort) (string, error) {
	/* Remove all the memfber ports from the port-channel interface. */
	for _, InterNodeLinkPort := range memberPorts {
		x, err := adapter.DeleteInterfaceFromPo(client, InterNodeLinkPort.IntfName, portChannelName)
		if (x != "<ok/>") || (err != nil) {
			return x, err
		}
	}

	log.Infof("Deletion of the port-channel<%s>", portChannelName)
	x, err := adapter.DeleteInterfacePo(client, portChannelName)
	if (x != "<ok/>") || (err != nil) {
		return x, err
	}
	return "<ok/>", nil
}

//UnconfigureDataPlaneCluster is used to unconfigure MCT cluster data plane.
func UnconfigureDataPlaneCluster(ctx context.Context, wg *sync.WaitGroup, cluster *operation.ConfigDataPlaneCluster,
	force bool, clusterConfigErrors chan actions.OperationError) {
	defer wg.Done()

	/* Config push to all the nodes of the management cluster */
	var clusterConfigWaitGroup sync.WaitGroup
	for iter := range cluster.DataPlaneClusterMemberNodes {
		mctNode := cluster.DataPlaneClusterMemberNodes[iter]
		clusterConfigWaitGroup.Add(1)
		go unconfigureDataPlaneClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, clusterConfigErrors)
	}
	clusterConfigWaitGroup.Wait()
}

func unconfigureDataPlaneClusterOnANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup, cluster *operation.ConfigDataPlaneCluster,
	mctNode *operation.DataPlaneClusterMemberNode, clusterConfigErrors chan actions.OperationError) {

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    cluster.FabricName,
		"Operation": "Unconfigure Data Plane Cluster",
		"Switch":    mctNode.NodeMgmtIP,
	})

	defer clusterConfigWaitGroup.Done()

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

	bgpResponse, _ := adapter.GetRouterBgp(client)
	bgpNeighborResponseMap := make(map[string]bool, 0)
	for _, swNeigh := range bgpResponse.Neighbors {
		bgpNeighborResponseMap[swNeigh.RemoteIP] = true
	}

	if _, ok := bgpNeighborResponseMap[mctNode.NodePeerIP]; ok {
		x, err := adapter.UnconfigureRouterBgpL2EVPNNeighbor(client, mctNode.NodePeerIP, mctNode.NodePeerLoopbackIP)

		if (x != "<ok/>") || (err != nil) {
			clusterConfigErrors <- actions.OperationError{Operation: "Unconfigure Data Plane Cluster", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
	}
	if _, ok := bgpNeighborResponseMap[mctNode.NodePeerLoopbackIP]; ok {
		x, err := adapter.UnconfigureRouterBgpL2EVPNNeighbor(client, mctNode.NodePeerIP, mctNode.NodePeerLoopbackIP)

		if (x != "<ok/>") || (err != nil) {
			clusterConfigErrors <- actions.OperationError{Operation: "Unconfigure Data Plane Cluster", Error: err, Host: mctNode.NodeMgmtIP}
			return
		}
	}

	return
}

package clearfabric

import (
	nlog "github.com/sirupsen/logrus"

	"context"
	"sync"

	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
)

//ClearManagementCluster is used to clear the configured "cluster <name> <id>" and all its sub-config on the switch.
func ClearManagementCluster(ctx context.Context, wg *sync.WaitGroup, cluster *operation.ConfigCluster,
	force bool, clusterConfigErrors chan actions.OperationError) {
	defer wg.Done()
	var clusterConfigWaitGroup sync.WaitGroup
	for iter := range cluster.ClusterMemberNodes {
		mctNode := cluster.ClusterMemberNodes[iter]
		clusterConfigWaitGroup.Add(1)
		go CleanupManagementClusterOnANode(ctx, &clusterConfigWaitGroup, cluster, &mctNode, clusterConfigErrors)
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

//CleanupManagementClusterOnANode is used to clear the configured "cluster <name> <id>" and all its sub-config on a single switching device.
func CleanupManagementClusterOnANode(ctx context.Context, clusterConfigWaitGroup *sync.WaitGroup, cluster *operation.ConfigCluster,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan actions.OperationError) {
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    cluster.FabricName,
		"Operation": "Cleanup Management Cluster Config",
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

	/* Clear the existing cluster config */
	{
		clusterName, clusterID, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP, clusErr := adapter.GetCluster(client)
		if clusErr != nil {
			clusterConfigErrors <- actions.OperationError{Operation: "Get Cluster Config", Error: clusErr, Host: mctNode.NodeMgmtIP}
			return
		}

		log.Info("cleanupManagementClusterOnANode", clusterID, clusterName, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP, mctNode.NodeMgmtIP)
		if clusterName != "" {
			x, err := adapter.DeleteCluster(client, clusterName, clusterID)
			if (x != "<ok/>") || (err != nil) {
				log.Infof("Deletion of the cluster name<%s> Id<%s> failed on host<%s>", clusterName, clusterID, mctNode.NodeMgmtIP)
				clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster config", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}

		if clusterVlan != "" {
			ResultMap, err := adapter.GetClusterControlVlan(client, clusterVlan)

			if err != nil {
				clusterConfigErrors <- actions.OperationError{Operation: "Get MCT Control VLAN", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
			clusterVe := ResultMap["control-ve"]

			if clusterVe != "" {
				log.Infof("Deletion of the association between control VLAN<%s> and control VE<%s>", clusterVlan, clusterVe)
				x, err := adapter.DeleteClusterControlVe(client, clusterVlan, clusterVe)
				if (x != "<ok/>") || (err != nil) {
					log.Infof("Deletion of the association between control VLAN<%s> and control VE<%s> failed on host<%s>", clusterVlan, clusterVe, mctNode.NodeMgmtIP)
					clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control vlan to ve association", Error: err, Host: mctNode.NodeMgmtIP}
					return
				}

				log.Infof("Deletion of the control VE<%s>", clusterVe)
				x, err = adapter.DeleteInterfaceVe(client, clusterVe)
				if (x != "<ok/>") || (err != nil) {
					log.Infof("Deletion of the control VE<%s> failed on host<%s>", clusterVe, mctNode.NodeMgmtIP)
					clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control ve", Error: err, Host: mctNode.NodeMgmtIP}
					return
				}
			}

			log.Infof("Deletion of the control VLAN<%s>", clusterVlan)
			x, err := adapter.DeleteClusterControlVlan(client, clusterVlan)
			if (x != "<ok/>") || (err != nil) {
				log.Infof("Deletion of the control VLAN<%s> failed on host<%s>", clusterVlan, mctNode.NodeMgmtIP)
				clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control vlan", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}
		if clusterPeerIfType != "Ve" {
			// If the Peer IF Type is VE, it means the cluster config is from AVALANCHE/ORCA. So you cannot delete PO.
			if clusterPeerIfName != "" {
				log.Infof("Deletion of the port-channel<%s>", clusterPeerIfName)
				x, err := adapter.DeleteInterfacePo(client, clusterPeerIfName)
				if (x != "<ok/>") || (err != nil) {
					clusterConfigErrors <- actions.OperationError{Operation: "Deletion of peer interface", Error: err, Host: mctNode.NodeMgmtIP}
					return
				}
			}
		}
	}

	/* Clear the input "default" cluster config */
	{
		if cluster.ClusterName != "" {
			x, err := adapter.DeleteCluster(client, cluster.ClusterName, cluster.ClusterID)
			if (x != "<ok/>") || (err != nil) {
				log.Infof("Deletion of the cluster name<%s> Id<%s> failed on host<%s>", cluster.ClusterName, cluster.ClusterID, mctNode.NodeMgmtIP)
				clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster config", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}
		clusterControlVlan := cluster.ClusterControlVlan
		if clusterControlVlan != "" {
			ResultMap, err := adapter.GetClusterControlVlan(client, clusterControlVlan)

			if err != nil {
				clusterConfigErrors <- actions.OperationError{Operation: "Get MCT Control VLAN", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
			clusterControlVe := ResultMap["control-ve"]

			if clusterControlVe != "" {
				log.Infof("Deletion of the association between control VLAN<%s> and control VE<%s>", clusterControlVlan, clusterControlVe)
				x, err := adapter.DeleteClusterControlVe(client, clusterControlVlan, clusterControlVe)
				if (x != "<ok/>") || (err != nil) {
					log.Infof("Deletion of the association between control VLAN<%s> and control VE<%s> failed on host<%s>", clusterControlVlan, clusterControlVe, mctNode.NodeMgmtIP)
					clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control vlan to ve association", Error: err, Host: mctNode.NodeMgmtIP}
					return
				}

				log.Infof("Deletion of the control VE<%s>", clusterControlVe)
				x, err = adapter.DeleteInterfaceVe(client, clusterControlVe)
				if (x != "<ok/>") || (err != nil) {
					log.Infof("Deletion of the control VE<%s> failed on host<%s>", clusterControlVe, mctNode.NodeMgmtIP)
					clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control ve", Error: err, Host: mctNode.NodeMgmtIP}
					return
				}
			}

			log.Infof("Deletion of the control VLAN<%s>", clusterControlVlan)
			x, err := adapter.DeleteClusterControlVlan(client, clusterControlVlan)
			if (x != "<ok/>") || (err != nil) {
				log.Infof("Deletion of the control VLAN<%s> failed on host<%s>", clusterControlVlan, mctNode.NodeMgmtIP)
				clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control vlan", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}

		clusterControlVe := cluster.ClusterControlVe
		if clusterControlVe != "" {
			log.Infof("Deletion of the control VE<%s>", clusterControlVe)
			x, err := adapter.DeleteInterfaceVe(client, clusterControlVe)
			if (x != "<ok/>") || (err != nil) {
				log.Infof("Deletion of the control VE<%s> failed on host<%s>", clusterControlVe, mctNode.NodeMgmtIP)
				clusterConfigErrors <- actions.OperationError{Operation: "Deletion of cluster control ve", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}

		clusterPeerIntfName := mctNode.NodePeerIntfName
		if clusterPeerIntfName != "" {
			log.Infof("Deletion of the port-channel<%s>", clusterPeerIntfName)
			x, err := adapter.DeleteInterfacePo(client, clusterPeerIntfName)
			if (x != "<ok/>") || (err != nil) {
				clusterConfigErrors <- actions.OperationError{Operation: "Deletion of peer interface", Error: err, Host: mctNode.NodeMgmtIP}
				return
			}
		}
	}

	return
}

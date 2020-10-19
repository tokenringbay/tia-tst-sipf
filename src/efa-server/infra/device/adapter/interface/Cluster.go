package interfaces

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
)

//Cluster provides collection of methods supported for Cluster
type Cluster interface {
	//GetManagementClusterStatus is used to get the Cluster Management Status
	GetManagementClusterStatus(client *client.NetconfClient) (string, operation.MgmtClusterStatus, string, error)
	//ConfigureNodeID is used to execute "cluster management node-id <node-id>" on the switching device
	ConfigureNodeID(sshClient *client.SSHClient, nodeID string) error

	//ConfigureNodePrincipalPriority is used to configure "management cluster principal-priority" on the switching device
	ConfigureNodePrincipalPriority(client *client.NetconfClient, nodeID string,
		configureClusterPriority string, clusterPriority string) (string, error)

	//UnconfigureNodePrincipalPriority is used to unconfigure "management cluster principal-priority" from the switching device
	UnconfigureNodePrincipalPriority(client *client.NetconfClient, nodeID string) (string, error)

	//GetNodeIDAndPrincipalPriority is used to get "running-config" of node-id and principal-priority from the switching device
	GetNodeIDAndPrincipalPriority(client *client.NetconfClient) (map[string]string, error)

	//CreateClusterControlVlan is used to create the "management cluster control VLAN" and "associate the control VLAN with control VE", on the switching device
	CreateClusterControlVlan(client *client.NetconfClient, controlVlan string,
		controlVe string, description string) (string, error)

	//GetClusterControlVlan is used to get the "management cluster control VLAN" from the switching device
	GetClusterControlVlan(client *client.NetconfClient, controlVlan string) (map[string]string, error)

	//DeleteClusterControlVe is used to "dissociate management cluster control VLAN and control VE", from the switching device
	DeleteClusterControlVe(client *client.NetconfClient, controlVlan string,
		controlVe string) (string, error)

	//DeleteClusterControlVlan is used to delete "management cluster control VLAN" from the switching device
	DeleteClusterControlVlan(client *client.NetconfClient, controlVlan string) (string, error)

	//CreateCluster is used to create "cluster <id> <name>" on the switching device
	CreateCluster(client *client.NetconfClient, clusterName string, clusterID string,
		peerType string, peerName string, peerIP string) (string, error)

	//ConfigureCluster is used to configure management cluster control VLAN ,Peer IP and Peer Interface on the "cluster <id> <name>"
	ConfigureCluster(client *client.NetconfClient, clusterName string,
		clusterID string, peerType string, peerName string, peerIP string, peerLoopbackIP string,
		clusterControlVlan string, clusterControlVE string, sourceIP string) (string, error)

	//ConfigureClusterPeerIP is used to configure Peer IP on "cluster <id> <name>", on the switching device
	ConfigureClusterPeerIP(client *client.NetconfClient, clusterName string,
		clusterID string, peerIP string) (string, error)

	//GetCluster is used to get running-config of "cluster <id> <name>", from the switching device
	GetCluster(client *client.NetconfClient) (string, string, string, string, string, string, error)

	//GetClusterByName is used to get running-config of "cluster <id> <name>" for a given name, from the switching device
	GetClusterByName(client *client.NetconfClient, name string) (map[string]string, error)

	//UndeployCluster is used to "undeploy" "cluster <id> <name> on the switching device
	UndeployCluster(client *client.NetconfClient, clusterName string, clusterID string) (string, error)

	//UnconfigureClusterPeerIP is used to unconfigure cluster Peer IP from "cluster <id> <name>" on the switching device
	UnconfigureClusterPeerIP(client *client.NetconfClient, clusterName string, clusterID string, peerIP string) (string, error)

	//DeleteCluster is used to delete "cluster <id> <name>" from the switching device
	DeleteCluster(client *client.NetconfClient, clusterName string, clusterID string) (string, error)
}

package base

import (
	"efa-server/infra/device/adapter/platform/slx/routing/base"
	"efa-server/infra/device/client"
)

//SLXOrcaBase structure for cedar base SLX
type SLXOrcaBase struct {
	base.SLXRoutingBase
}

//ConfigureCluster is used to configure the cluster
func (base *SLXOrcaBase) ConfigureCluster(client *client.NetconfClient, clusterName string,
	clusterID string, peerType string, peerName string, peerIP string, peerLoopbackIP string,
	clusterControlVlan string, clusterControlVE string, sourceIP string) (string, error) {
	//Configure Peer Interface and Peer IP
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID,
		"peer_if_type": "Ve", "peer_if_name": clusterControlVE, "peer_ip": peerIP,
		"source_ip": sourceIP}

	config, templateError := base.GetStringFromTemplate(mctClusterAddPeer, mctCluster)
	//fmt.Println(config)

	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	return resp, err
}

//DeleteCluster is used to delete "cluster <id> <name>" from the switching device
func (base *SLXOrcaBase) DeleteCluster(client *client.NetconfClient, clusterName string, clusterID string) (string, error) {
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID}

	config, templateError := base.GetStringFromTemplate(mctClusterDelete, mctCluster)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	return resp, err
}

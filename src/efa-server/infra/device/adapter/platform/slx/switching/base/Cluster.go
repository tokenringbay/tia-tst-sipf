package base

import (
	"efa-server/infra/device/client"
)

//ConfigureCluster is used to configure the cluster
func (base *SLXSwitchingBase) ConfigureCluster(client *client.NetconfClient, clusterName string,
	clusterID string, peerType string, peerName string, peerIP string, peerLoopbackIP string,
	clusterControlVlan string, clusterControlVE string, sourceIP string) (string, error) {

	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID,
		"cluster_control_vlan": clusterControlVlan}
	//First Configure Control VLAN
	config, templateError := base.GetStringFromTemplate(mctClusterAddControlVlan, mctCluster)

	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	//In case of error return from here
	if err != nil {
		return resp, err
	}

	//Configure Peer Interface and Peer IP
	mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID,
		"peer_if_type": peerType, "peer_if_name": peerName, "peer_ip": peerIP}

	config, templateError = base.GetStringFromTemplate(mctClusterAddPeer, mctCluster)
	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)
	return resp, err
}

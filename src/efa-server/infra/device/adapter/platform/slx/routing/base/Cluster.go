package base

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
	"fmt"
	"github.com/beevik/etree"
)

//GetManagementClusterStatus is used to get the operational state of the management cluster, from the switching device
func (base *SLXRoutingBase) GetManagementClusterStatus(client *client.NetconfClient) (string, operation.MgmtClusterStatus, string, error) {

	var MgmtClusterStatus operation.MgmtClusterStatus
	MgmtClusterStatus.TotalMemberNodeCount = "0"
	request := `<show-cluster-management xmlns="urn:brocade.com:mgmt:brocade-cluster"></show-cluster-management>`

	RawResponse, err := client.ExecuteRPC(request)

	if err != nil {
		fmt.Println("Error from show-cluster-management RPC execution: ", err)
		return RawResponse, MgmtClusterStatus, "", err
	}

	//fmt.Println("Response of show-cluster-management: ", RawResponse)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(RawResponse)); err != nil {
		fmt.Println("Error in reading the show-cluster-management RPC o/p")
		return RawResponse, MgmtClusterStatus, "", err
	}

	de := doc.FindElement("//principal-switch-mac")
	if de != nil {
		MgmtClusterStatus.PrincipalNodeMac = de.Text()
	}

	de = doc.FindElement("//total-nodes-in-cluster")
	if de != nil {
		MgmtClusterStatus.TotalMemberNodeCount = de.Text()
	}

	de = doc.FindElement("//nodes-disconnected-from-cluster")
	if de != nil {
		MgmtClusterStatus.DisconnectedMemberNodeCount = de.Text()
	}

	principalNode := ""

	if elems := doc.FindElements("//cluster-node-info"); elems != nil {
		var clusterMember operation.MgmtClusterMemberNode
		for _, elem := range elems {
			if de := elem.FindElement(".//node-serial-num"); de != nil {
				clusterMember.NodeSerial = de.Text()
			}
			if de := elem.FindElement(".//node-switch-mac"); de != nil {
				clusterMember.NodeMac = de.Text()
			}
			if de := elem.FindElement(".//node-public-ip-address"); de != nil {
				clusterMember.NodeMgmtIP = de.Text()
			}
			if de := elem.FindElement(".//node-internal-ip-address"); de != nil {
				clusterMember.NodeInternalIP = de.Text()
			}
			if de := elem.FindElement(".//node-id"); de != nil {
				clusterMember.NodeID = de.Text()
			}
			if de := elem.FindElement(".//node-condition"); de != nil {
				clusterMember.NodeCondition = de.Text()
			}
			if de := elem.FindElement(".//node-status"); de != nil {
				clusterMember.NodeStatus = de.Text()
			}
			if de := elem.FindElement(".//node-is-principal"); de != nil {
				clusterMember.NodeIsPrincipal = de.Text()
			}
			if de := elem.FindElement(".//node-is-local"); de != nil {
				clusterMember.NodeIsLOcal = de.Text()
			}
			if de := elem.FindElement(".//node-switchtype"); de != nil {
				clusterMember.NodeSwitchType = de.Text()
			}
			if de := elem.FindElement(".//firmware-version"); de != nil {
				clusterMember.NodeFwVersion = de.Text()
			}
			if clusterMember.NodeIsPrincipal == "true" {
				principalNode = clusterMember.NodeMgmtIP
			}
			MgmtClusterStatus.MemberNodes = append(MgmtClusterStatus.MemberNodes, clusterMember)
		}
	}

	return RawResponse, MgmtClusterStatus, principalNode, nil
}

//ConfigureCluster is used to configure the cluster
func (base *SLXRoutingBase) ConfigureCluster(client *client.NetconfClient, clusterName string,
	clusterID string, peerType string, peerName string, peerIP string, peerLoopbackIP string,
	clusterControlVlan string, clusterControlVE string, sourceIP string) (string, error) {
	//Configure Peer Interface and Peer IP
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID,
		"peer_if_type": "Ve", "peer_if_name": clusterControlVE, "peer_ip": peerLoopbackIP}

	config, templateError := base.GetStringFromTemplate(mctClusterAddPeer, mctCluster)

	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	if err != nil {
		return resp, err
	}

	var mplsmap = map[string]interface{}{}
	config, templateError = base.GetStringFromTemplate(routerMPLSCreate, mplsmap)
	//fmt.Println(config)

	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)

	return resp, err
}

//DeleteCluster is used to delete "cluster <id> <name>" from the switching device
func (base *SLXRoutingBase) DeleteCluster(client *client.NetconfClient, clusterName string, clusterID string) (string, error) {
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID}

	config, templateError := base.GetStringFromTemplate(mctClusterDelete, mctCluster)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)

	var mplsmap = map[string]interface{}{}
	config, templateError = base.GetStringFromTemplate(routerMPLSDelete, mplsmap)
	//fmt.Println(config)

	if templateError != nil {
		return "", templateError
	}

	resp, err = client.EditConfig(config)

	return resp, err
}

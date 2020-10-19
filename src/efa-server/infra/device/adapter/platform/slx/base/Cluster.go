package base

import (
	"efa-server/domain/operation"
	"efa-server/infra/device/client"
	"errors"
	"fmt"
	"github.com/beevik/etree"
)

//GetManagementClusterStatus is used to get the operational state of the management cluster, from the switching device
func (base *SLXBase) GetManagementClusterStatus(client *client.NetconfClient) (string, operation.MgmtClusterStatus, string, error) {

	var MgmtClusterStatus operation.MgmtClusterStatus
	MgmtClusterStatus.TotalMemberNodeCount = "0"
	request := `<show-cluster-management xmlns="http://brocade.com/ns/brocade-cluster"></show-cluster-management>`
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

//ConfigureNodeID is used to execute "cluster management node-id <node-id>" on the switching device
func (base *SLXBase) ConfigureNodeID(sshClient *client.SSHClient, nodeID string) error {
	command := "cluster management node-id " + nodeID
	output := sshClient.ExecuteOperationalCommand(command)
	if output != "" {
		return errors.New(output)
	}
	return nil
}

//ConfigureNodePrincipalPriority is used to configure "management cluster principal-priority" on the switching device
func (base *SLXBase) ConfigureNodePrincipalPriority(client *client.NetconfClient, nodeID string,
	configureClusterPriority string, clusterPriority string) (string, error) {
	var nodeIDMap = map[string]interface{}{"node_id": nodeID, "configure_cluster_priority": configureClusterPriority,
		"cluster_priority": clusterPriority}

	config, templateError := base.GetStringFromTemplate(nodeIDClusterPriorityCreate, nodeIDMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//UnconfigureNodePrincipalPriority is used to unconfigure "management cluster principal-priority" from the switching device
func (base *SLXBase) UnconfigureNodePrincipalPriority(client *client.NetconfClient, nodeID string) (string, error) {
	var nodeIDMap = map[string]interface{}{"node_id": nodeID}

	config, templateError := base.GetStringFromTemplate(nodeIDClusterPriorityDelete, nodeIDMap)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//GetNodeIDAndPrincipalPriority is used to get "running-config" of node-id and principal-priority from the switching device
func (base *SLXBase) GetNodeIDAndPrincipalPriority(client *client.NetconfClient) (map[string]string, error) {
	ResultMap := make(map[string]string)
	resp, err := client.GetConfig("/node-id")
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElement("//principal-priority"); elem != nil {
		ResultMap["cluster-priority"] = elem.Text()
	}
	if elem := doc.FindElement("//node-id/node-id"); elem != nil {
		ResultMap["node-id"] = elem.Text()
	}
	return ResultMap, err
}

//CreateClusterControlVlan is used to create the "management cluster control VLAN" and "associate the control VLAN with control VE", on the switching device
func (base *SLXBase) CreateClusterControlVlan(client *client.NetconfClient, controlVlan string,
	controlVe string, description string) (string, error) {
	var mctControlVlan = map[string]interface{}{"control_vlan": controlVlan, "control_ve": controlVe,
		"description": description}

	config, templateError := base.GetStringFromTemplate(mctControlVlanCreate, mctControlVlan)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//GetClusterControlVlan is used to get the "management cluster control VLAN" from the switching device
func (base *SLXBase) GetClusterControlVlan(client *client.NetconfClient, controlVlan string) (map[string]string, error) {
	ResultMap := make(map[string]string)
	RequestMsg := fmt.Sprintf("/interface-vlan/vlan[name='%s']", controlVlan)

	resp, err := client.GetConfig(RequestMsg)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElement("//router-interface/ve-config"); elem != nil {
		ResultMap["control-ve"] = elem.Text()
	}
	if elem := doc.FindElement("//description"); elem != nil {
		ResultMap["description"] = elem.Text()
	}
	return ResultMap, err
}

//DeleteClusterControlVe is used to "dissociate management cluster control VLAN and control VE", from the switching device
func (base *SLXBase) DeleteClusterControlVe(client *client.NetconfClient, controlVlan string,
	controlVe string) (string, error) {
	var mctControlVlan = map[string]interface{}{"control_vlan": controlVlan, "control_ve": controlVe}

	config, templateError := base.GetStringFromTemplate(mctControlVeDelete, mctControlVlan)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//DeleteClusterControlVlan is used to delete "management cluster control VLAN" from the switching device
func (base *SLXBase) DeleteClusterControlVlan(client *client.NetconfClient, controlVlan string) (string, error) {
	var mctControlVlan = map[string]interface{}{"control_vlan": controlVlan}

	config, templateError := base.GetStringFromTemplate(mctControlVlanDelete, mctControlVlan)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//CreateCluster is used to create "cluster <id> <name>" on the switching device
func (base *SLXBase) CreateCluster(client *client.NetconfClient, clusterName string, clusterID string,
	peerType string, peerName string, peerIP string) (string, error) {

	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID}

	config, templateError := base.GetStringFromTemplate(mctClusterCreate, mctCluster)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//ConfigureClusterPeerIP is used to configure Peer IP on "cluster <id> <name>", on the switching device
func (base *SLXBase) ConfigureClusterPeerIP(client *client.NetconfClient, clusterName string,
	clusterID string, peerIP string) (string, error) {

	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID, "peer_ip": peerIP}

	config, templateError := base.GetStringFromTemplate(mctClusterAddPeerIP, mctCluster)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//GetCluster is used to get running-config of "cluster <id> <name>", from the switching device
func (base *SLXBase) GetCluster(client *client.NetconfClient) (string, string, string, string, string, string, error) {
	RequestMsg := fmt.Sprintf("//cluster")
	resp, err := client.GetConfig(RequestMsg)
	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return "", "", "", "", "", "", err
	}

	var clusterName, clusterID, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP string
	if elem := doc.FindElement("//cluster-name"); elem != nil {
		clusterName = elem.Text()
	}

	if elem := doc.FindElement("//cluster-id"); elem != nil {
		clusterID = elem.Text()
	}

	if elem := doc.FindElement("//cluster-control-vlan"); elem != nil {
		clusterVlan = elem.Text()
	}

	if elem := doc.FindElement("//peer-interface/peer-if-type"); elem != nil {
		clusterPeerIfType = elem.Text()
	}
	if elem := doc.FindElement("//peer-interface/peer-if-name"); elem != nil {
		clusterPeerIfName = elem.Text()
	}
	if elem := doc.FindElement("//peer/peer-ip"); elem != nil {
		clusterPeerIP = elem.Text()
	}

	return clusterName, clusterID, clusterVlan, clusterPeerIfType, clusterPeerIfName, clusterPeerIP, err
}

//GetClusterByName is used to get running-config of "cluster <id> <name>" for a given name, from the switching device
func (base *SLXBase) GetClusterByName(client *client.NetconfClient, name string) (map[string]string, error) {
	ResultMap := make(map[string]string)
	RequestMsg := fmt.Sprintf("/cluster[cluster-name='%s']", name)

	resp, err := client.GetConfig(RequestMsg)

	doc := etree.NewDocument()

	if err := doc.ReadFromBytes([]byte(resp)); err != nil {
		return ResultMap, err
	}

	if elem := doc.FindElement("//cluster-id"); elem != nil {
		ResultMap["cluster-id"] = elem.Text()
	}
	if elem := doc.FindElement("//cluster-control-vlan"); elem != nil {
		ResultMap["cluster-control-vlan"] = elem.Text()
	}
	if elem := doc.FindElement("//df-load-balance"); elem != nil {
		ResultMap["df-load-balance"] = "true"
	}
	if elem := doc.FindElement("//deploy"); elem != nil {
		ResultMap["deploy"] = "true"
	}
	if elem := doc.FindElement("//peer-interface/peer-if-type"); elem != nil {
		ResultMap["peer-type"] = elem.Text()
	}
	if elem := doc.FindElement("//peer-interface/peer-if-name"); elem != nil {
		ResultMap["peer-name"] = elem.Text()
	}
	if elem := doc.FindElement("//peer/peer-ip"); elem != nil {
		ResultMap["peer-ip"] = elem.Text()
	}
	if elem := doc.FindElement("//peer/source/source_ip"); elem != nil {
		ResultMap["source-ip"] = elem.Text()
	}
	if elem := doc.FindElement("//client-isolation/loose"); elem != nil {
		ResultMap["client-isolation-loose"] = "true"
	}
	return ResultMap, err
}

//UndeployCluster is used to "undeploy" "cluster <id> <name> on the switching device
func (base *SLXBase) UndeployCluster(client *client.NetconfClient, clusterName string, clusterID string) (string, error) {
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID}

	config, templateErr := base.GetStringFromTemplate(mctClusterUndeploy, mctCluster)
	if templateErr != nil {
		return "", templateErr
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//UnconfigureClusterPeerIP is used to unconfigure cluster Peer IP from "cluster <id> <name>" on the switching device
func (base *SLXBase) UnconfigureClusterPeerIP(client *client.NetconfClient, clusterName string, clusterID string, peerIP string) (string, error) {
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID, "peer_ip": peerIP}

	config, templateErr := base.GetStringFromTemplate(mctClusterRemovePeerIP, mctCluster)
	if templateErr != nil {
		return "", templateErr
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

//DeleteCluster is used to delete "cluster <id> <name>" from the switching device
func (base *SLXBase) DeleteCluster(client *client.NetconfClient, clusterName string, clusterID string) (string, error) {
	var mctCluster = map[string]interface{}{"cluster_name": clusterName, "cluster_id": clusterID}

	config, templateError := base.GetStringFromTemplate(mctClusterDelete, mctCluster)
	if templateError != nil {
		return "", templateError
	}

	resp, err := client.EditConfig(config)
	return resp, err
}

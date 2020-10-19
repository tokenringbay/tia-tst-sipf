package actions

import (
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	ad "efa-server/infra/device/adapter"

	"efa-server/infra/device/client"
	"fmt"
	nlog "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"

	"context"
	"errors"
)

//MgmtClusterStatePollingTimeOutInSec implies the timeout value for polling the operational status of management cluster to be up.
var MgmtClusterStatePollingTimeOutInSec = 90

//MgmtClusterStatePollingIntervalInSec implies the interval for polling the operational status of management cluster to be up.
var MgmtClusterStatePollingIntervalInSec = 15

//OperationError is used to represent any error during any of the fabric operation.
type OperationError struct {
	Operation string
	Error     error
	Host      string
}

func (b OperationError) String() string {
	return fmt.Sprintf("{Host=%s,Operation=%s,Error=%s}", b.Host, b.Operation, b.Error)
}

//ExecuteClearBgpEvpnNeighbourAll is used to execute the operational CLI "clear bgp evpn neighbor all" on the switch
func ExecuteClearBgpEvpnNeighbourAll(configSwitch *operation.ConfigSwitch) error {
	/*SSH client*/
	sshClient := &client.SSHClient{Host: configSwitch.Host, User: configSwitch.UserName, Password: configSwitch.Password}
	loginErr := sshClient.Login()
	if loginErr != nil {
		return loginErr
	}
	defer sshClient.Close()
	adapter := ad.GetAdapter(configSwitch.Model)
	err := adapter.ExecuteClearBgpEvpnNeighbourAll(sshClient)
	return err
}

//ExecuteClearBgpEvpnNeighbour is used to execute the operational CLI "clear bgp evpn neighbor <peer-ip>" on the switch
func ExecuteClearBgpEvpnNeighbour(configSwitch *operation.ConfigSwitch, neighbourIP string) error {
	/*SSH client*/
	sshClient := &client.SSHClient{Host: configSwitch.Host, User: configSwitch.UserName, Password: configSwitch.Password}
	loginErr := sshClient.Login()
	if loginErr != nil {
		return loginErr
	}
	defer sshClient.Close()
	adapter := ad.GetAdapter(configSwitch.Model)
	err := adapter.ExecuteClearBgpEvpnNeighbour(sshClient, neighbourIP)
	return err
}

//PollManagementClusterStatusOnANode is used to poll the management cluster status of a node
func PollManagementClusterStatusOnANode(ctx context.Context, clusterOperWaitGroup *sync.WaitGroup, intendedCluster *operation.ConfigCluster,
	mctNode *operation.ClusterMemberNode, clusterConfigErrors chan OperationError) {

	defer clusterOperWaitGroup.Done()

	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    intendedCluster.FabricName,
		"Operation": "Poll Management Cluster Status",
		"Switch":    mctNode.NodeMgmtIP,
	})

	/*Netconf client*/
	adapter := ad.GetAdapter(mctNode.NodeModel)
	client := &client.NetconfClient{Host: mctNode.NodeMgmtIP, User: mctNode.NodeMgmtUserName, Password: mctNode.NodeMgmtPassword}
	client.Login()
	defer client.Close()

	intendedClusterMembers := intendedCluster.ClusterMemberNodes

	Operation := "Poll for management cluster status"
	if len(intendedClusterMembers) > 2 {
		clusterConfigErrors <- OperationError{Operation: Operation, Error: errors.New("Management cluster is supported for a maximum of 2 nodes"), Host: mctNode.NodeMgmtIP}
		return
	}

	timeout := time.After(time.Duration(MgmtClusterStatePollingTimeOutInSec) * (time.Second))
	tick := time.Tick(time.Duration(MgmtClusterStatePollingIntervalInSec) * (time.Second))

	for {
		select {
		case <-timeout:
			clusterConfigErrors <- OperationError{Operation: Operation, Error: errors.New("Management Cluster is not operational. Polling timed out"), Host: mctNode.NodeMgmtIP}
			return
		case <-tick:
			fmt.Println("Management cluster status polled at", time.Now())

			output, operationalClusterMembers, principalNode, err := adapter.GetManagementClusterStatus(client)
			log.Infof("Principal Node IP obtained on <%s> is <%s>", mctNode.NodeMgmtIP, principalNode)
			operationalClusterMemberCount, err := strconv.Atoi(operationalClusterMembers.TotalMemberNodeCount)

			if err != nil {
				clusterConfigErrors <- OperationError{Operation: Operation, Error: err, Host: mctNode.NodeMgmtIP}
				return
			}

			if operationalClusterMemberCount == len(intendedClusterMembers) {
				var i = 0
				var j = 0
				var found = false
				for i = 0; i < operationalClusterMemberCount; i++ {
					found = false
					operationalClusterMember := operationalClusterMembers.MemberNodes[i]
					for j = 0; j < operationalClusterMemberCount; j++ {
						var ipClusterMember = intendedClusterMembers[j]
						if operationalClusterMember.NodeMgmtIP == ipClusterMember.NodeMgmtIP {
							found = true
							break
						}
					}
					if found == false {
						break
					}
				}
				if found == true {
					fmt.Println("Management Cluster is operational, hence exiting the poll on ", mctNode.NodeMgmtIP)
					return
				}
			}

			log.Info("Raw o/p of show-cluster-management", output)
			//fmt.Println("operationalClusterMembers", operationalClusterMembers)
		}
	}
}

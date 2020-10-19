package configurefabric

import (
	"context"
	"efa-server/domain"
	"efa-server/domain/operation"
	"efa-server/gateway/appcontext"
	"efa-server/infra/device/actions"
	"efa-server/infra/device/actions/deconfigurefabric"
	ad "efa-server/infra/device/adapter"
	netconf "efa-server/infra/device/client"
	nlog "github.com/sirupsen/logrus"
	"sync"
)

//ConfigureFabric Configures IP Fabric excluding MCT Cluster
func ConfigureFabric(ctx context.Context, config operation.ConfigFabricRequest, force bool, persist bool) []actions.OperationError {

	if config.FabricSettings.FabricType == domain.NonCLOSFabricType {
		return ConfigureNonClosFabric(ctx, config, force, persist)
	}

	clusterStatus := make(map[string]bool)
	log := appcontext.Logger(ctx)

	log.Info("Start")

	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	fabricErrors := make(chan actions.OperationError, 1)

	//For each Switch Invoke Configure Switch
	for iter := range config.Hosts {
		configSwitch := config.Hosts[iter]
		fabricGate.Add(1)
		go ConfigureSwitch(ctx, &fabricGate, configSwitch, force, persist, fabricErrors)

	}

	log.Info("Waiting for Switch Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		log.Info("Wait Completed")
		close(fabricErrors)

	}()

	//Check for errors in the sub-action
	for err := range fabricErrors {
		Errors = append(Errors, err)
	}

	if len(Errors) > 0 {
		log.Error("Configure Fabric Failed")
		return Errors
	}

	//Configure MCT Cluster
	if MCTErrors := configureMCTCluster(ctx, config, force, clusterStatus); len(MCTErrors) > 0 {
		log.Errorln("Error Configuring MCT cluster - ", MCTErrors)
		log.Error("Configure Fabric Failed")
		return MCTErrors
	}

	var overlayGate sync.WaitGroup
	overlayErrors := make(chan actions.OperationError, 1)
	//For each Switch Invoke Configure Overlay
	for iter := range config.Hosts {
		sw := config.Hosts[iter]
		markIfSwitchIsMCTSecondary(&sw, clusterStatus)
		if sw.Role == "Leaf" && sw.ConfigureOverlayGateway == "Yes" && sw.MctSecondaryNode == false {
			overlayGate.Add(1)
			go ConfigureOverlayGateway(ctx, &overlayGate, &sw, force, overlayErrors)
		}
	}
	log.Info("Waiting for Overlay Operations")

	//Utility go-routine waiting for actions to complete
	go func() {
		overlayGate.Wait()
		log.Info("overlay Wait Completed")
		close(overlayErrors)

	}()
	for err := range overlayErrors {
		Errors = append(Errors, err)
	}

	// save the configs on all the devices
	if persist {
		var saveConfig sync.WaitGroup
		saveConfigErrors := make(chan actions.OperationError, 1)
		//For each Switch Invoke Configure Overlay
		for iter := range config.Hosts {
			sw := config.Hosts[iter]
			saveConfig.Add(1)
			go persistConfig(ctx, &saveConfig, &sw, saveConfigErrors)
		}

		//Utility go-routine waiting for actions to complete
		go func() {
			saveConfig.Wait()
			log.Info("saving configuration wait Completed")
			close(saveConfigErrors)

		}()
		for err := range saveConfigErrors {
			Errors = append(Errors, err)
		}
	}

	if len(Errors) > 0 {
		log.Error("Configure Fabric Failed")
		return Errors
	}
	log.Debug("Configure Fabric Completed...")
	return Errors
}

func configureMCTCluster(ctx context.Context, config operation.ConfigFabricRequest, force bool, clusterStatus map[string]bool) []actions.OperationError {
	//Send the Config Object to Actions for configuring the switches
	var Error []actions.OperationError
	if len(config.MctCluster) > 0 {
		if len(config.MctCluster[domain.MctDelete]) > 0 {
			if Errors := ConfigureDeConfigureMctClusters(ctx, domain.MctDelete, config.MctCluster[domain.MctDelete], force); len(Errors) != 0 {
				return Errors
			}
		}
		if len(config.MctCluster[domain.MctUpdate]) > 0 {
			if Errors := ConfigureDeConfigureMctClusters(ctx, domain.MctUpdate, config.MctCluster[domain.MctUpdate], force); len(Errors) != 0 {
				return Errors
			}
		}
		if len(config.MctCluster[domain.MctCreate]) > 0 {
			if Errors := ConfigureDeConfigureMctClusters(ctx, domain.MctCreate, config.MctCluster[domain.MctCreate], force); len(Errors) != 0 {
				return Errors
			}
		}
	}
	if len(Error) == 0 {
		for iter := range config.MctCluster[domain.MctUpdate] {
			cluster := config.MctCluster[domain.MctUpdate][iter]
			getManagementClusterStatus(clusterStatus, cluster.ClusterMemberNodes[0].NodeMgmtIP, cluster.ClusterMemberNodes[0].NodeMgmtUserName, cluster.ClusterMemberNodes[0].NodeMgmtPassword,
				cluster.ClusterMemberNodes[0].NodeModel)
		}
		for iter := range config.MctCluster[domain.MctCreate] {
			cluster := config.MctCluster[domain.MctCreate][iter]
			getManagementClusterStatus(clusterStatus, cluster.ClusterMemberNodes[0].NodeMgmtIP, cluster.ClusterMemberNodes[0].NodeMgmtUserName, cluster.ClusterMemberNodes[0].NodeMgmtPassword,
				cluster.ClusterMemberNodes[0].NodeModel)

		}

	}
	return Error
}

//ConfigureDeConfigureMctClusters "configures" and "deconfigures" MCT cluster on a collection of devices
func ConfigureDeConfigureMctClusters(ctx context.Context, opcode uint, MctCluster []operation.ConfigCluster, force bool) []actions.OperationError {
	MctOperation := [4]string{1: "configure", 2: "deconfigure", 3: "update"}
	log := appcontext.Logger(ctx).WithFields(nlog.Fields{
		"App":       "dcfabric",
		"Fabric":    MctCluster[0].FabricName,
		"Operation": "Configure Fabric",
	})
	log.Infoln("MCT Cluster Start opcode - ", MctOperation[opcode])
	hasBit := func(n uint64, pos uint) bool {
		val := n & (1 << pos)
		return (val > 0)
	}
	//List to hold errors from sub-actions
	Errors := make([]actions.OperationError, 0)

	//Concurrency gate for sub-actions
	var fabricGate sync.WaitGroup
	//Each MCT cluster Has 2 Nodes hence first Multiple by 2 and each Request Needs
	//Length of 2.
	mctErrors := make(chan actions.OperationError, len(MctCluster)*2*2)
	//For each Cluster Invoke
	prevForce := force
	for iter := range MctCluster {
		configCluster := MctCluster[iter]
		log.Infoln("Printing MCT Cluster Details Before Sending - ", configCluster)
		fabricGate.Add(1)
		switch opcode {
		case domain.MctCreate:
			if hasBit(configCluster.OperationBitMap, domain.BitPositionForMctCreate) {
				configCluster.OperationBitMap = 0
				force = true
			}
			go ConfigureManagementCluster(ctx, &fabricGate, &configCluster, force, mctErrors)
			force = prevForce
		case domain.MctDelete:
			go deconfigurefabric.UnconfigureManagementCluster(ctx, &fabricGate, &configCluster, force, mctErrors)
		case domain.MctUpdate:
			go UpdateManagementCluster(ctx, &fabricGate, &configCluster, force, mctErrors)
			fabricGate.Wait()
		}
	}

	//Utility go-routine waiting for actions to complete
	go func() {
		fabricGate.Wait()
		close(mctErrors)

	}()

	log.Infof("Wait FOR MCT %s opcode to Complete", MctOperation[opcode])

	//Check for errors in the sub-action
	for err := range mctErrors {
		Errors = append(Errors, err)
	}
	return Errors
}

func getManagementClusterStatus(clusterStatus map[string]bool, MgmtIP, UserName, Password string, NodeModel string) error {
	/*Netconf client*/
	if _, ok := clusterStatus[MgmtIP]; ok {
		return nil
	}
	adapter := ad.GetAdapter(NodeModel)
	client := &netconf.NetconfClient{Host: MgmtIP, User: UserName, Password: Password}
	client.Login()
	defer client.Close()
	_, operationalClusterMembers, _, err := adapter.GetManagementClusterStatus(client)
	if err != nil {
		return err
	}
	for _, memberNode := range operationalClusterMembers.MemberNodes {
		if memberNode.NodeIsPrincipal == "true" {
			clusterStatus[memberNode.NodeMgmtIP] = true
		} else {
			clusterStatus[memberNode.NodeMgmtIP] = false
		}
	}
	return nil
}

func markIfSwitchIsMCTSecondary(sw *operation.ConfigSwitch, clusterStatus map[string]bool) {
	if val, ok := clusterStatus[sw.Host]; ok {
		if val == false {
			//MCT secondary
			sw.MctSecondaryNode = true
		}
	}
}

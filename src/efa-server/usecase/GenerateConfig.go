package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"efa-server/usecase/comparator/interfaceswitchconfig"
	"errors"
	"fmt"

	"efa-server/usecase/comparator/bgpconfig"
	"net"
	"strconv"
)

//This method builds the SwitchConfiguration for each Switch
func (sh *DeviceInteractor) buildSwitchConfig(ctx context.Context, device *domain.Device) (string, error) {

	LOG := appcontext.Logger(ctx)
	var switchConfig domain.SwitchConfig
	NeighborVTEPLoopBack := ""
	NeighborLoopBack := ""
	NeighborAsn := ""
	switchConfig.DeviceID = device.ID
	switchConfig.Role = device.DeviceRole
	switchConfig.FabricID = sh.FabricID
	switchConfig.DeviceIP = device.IPAddress
	switchConfig.ASConfigType = domain.ConfigNone

	//Fetch Fabric Properties
	FabricProperties, _ := sh.Db.GetFabricProperties(sh.FabricID)

	LOG.Infoln("Build switch config")
	//Fetch the existing Switch Config from Database
	oldSwitchConfig, err := sh.Db.GetSwitchConfigOnFabricIDAndDeviceID(device.FabricID, device.ID)
	if err != nil {
		//Switch Config is to be created
		switchConfig.ASConfigType = domain.ConfigCreate
	} else {
		//Setting the ID to the same one, so that DB gets updated
		switchConfig.ID = oldSwitchConfig.ID
	}
	if switchConfig.Role == LeafRole || switchConfig.Role == RackRole {
		NeighborVTEPLoopBack, NeighborLoopBack, NeighborAsn, err = sh.FetchMctNeighborVTEPLoopBackIPAndASN(ctx, device.ID)
		if err != nil {
			statusMsg := fmt.Sprintln("Error Fetching MCT neighbor VTEP LoopBack IP ", err)
			LOG.Error(statusMsg)
			return statusMsg, err
		}

	}
	//compute ASN and Configs to be sent to switch
	err = sh.computeASN(ctx, device, &FabricProperties, &oldSwitchConfig, &switchConfig, NeighborAsn)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to Compute ASN for %s", device.IPAddress)
		LOG.Infoln(statusMsg)
		return statusMsg, err

	}

	//Fetch the current Loopback IP from the Device
	lpOnDevice, _ := sh.createLoopbackIntrfaceIfNotExists(FabricProperties.LoopBackPortNumber, device.ID)
	OnSwitchLoopbackIP, err := sh.evaluateIP(lpOnDevice, FabricProperties.LoopBackIPRange)
	if err != nil {
		statusMsg := fmt.Sprintf("Loopback IP address Parse Failed for %s", device.IPAddress)
		return statusMsg, err
	}

	switchConfig.LoopbackIP, switchConfig.LoopbackIPConfigType, err = sh.computeLoopBackIP(ctx, device,
		FabricProperties.LoopBackPortNumber, FabricProperties.LoopBackIPRange,
		OnSwitchLoopbackIP, oldSwitchConfig.LoopbackIP, lpOnDevice.ID, "", NeighborLoopBack)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to GET/Reserve Loopback IP address for %s", device.IPAddress)
		LOG.Println(statusMsg)
		return statusMsg, err
	}
	statusMsg := fmt.Sprintf("Loopback IP address for %s is %s", device.IPAddress, switchConfig.LoopbackIP)
	LOG.Infoln(statusMsg)

	//Get VTEP LoopBack IP from Pool

	if switchConfig.Role == LeafRole || switchConfig.Role == RackRole {
		OldVTEPLoopIP, _ := sh.createLoopbackIntrfaceIfNotExists(FabricProperties.VTEPLoopBackPortNumber, device.ID)
		OnSwitchVTEPLoopbackIP, err := sh.evaluateIP(OldVTEPLoopIP, FabricProperties.LoopBackIPRange)
		if err != nil {
			statusMsg := fmt.Sprintf("VTEP Loopback IP address Parse Failed for %s", device.IPAddress)
			return statusMsg, err
		}

		switchConfig.VTEPLoopbackIP, switchConfig.VTEPLoopbackIPConfigType, err = sh.computeLoopBackIP(ctx, device,
			FabricProperties.VTEPLoopBackPortNumber, FabricProperties.LoopBackIPRange,
			OnSwitchVTEPLoopbackIP, oldSwitchConfig.VTEPLoopbackIP, OldVTEPLoopIP.ID, NeighborVTEPLoopBack, "")
		if err != nil {
			statusMsg := fmt.Sprintf("Failed to GET/Reserve VTEP Loopback IP address for %s", device.IPAddress)
			LOG.Println(statusMsg)
			return statusMsg, err
		}

	}

	//update the db entry
	if err := sh.Db.CreateSwitchConfig(&switchConfig); err != nil {
		statusMsg := fmt.Sprintf("Failed to save switch Config for %s", device.IPAddress)
		LOG.Println(statusMsg)
		return statusMsg, errors.New(statusMsg)
	}

	return "", nil
}

func (sh *DeviceInteractor) computeLoopBackIP(ctx context.Context, Device *domain.Device,
	LoopBackPortNumber string, LoopBackRange string,
	LoopBackOnDevice string, LoopBackInDB string, InterfaceID uint, NeighborVTEPLoopBack string, NeighborLoopback string) (string, string, error) {
	LOG := appcontext.Logger(ctx)
	LoopBackIPToConfigure := ""
	LoopBackIPConfigType := domain.ConfigNone
	LOG.Infoln("Compute Loopback", "LoopBack On Switch:", LoopBackOnDevice, ",ASN in DB:", LoopBackInDB, "Neighbor LoopbackIP:", NeighborVTEPLoopBack)
	if (Device.DeviceRole == LeafRole || Device.DeviceRole == RackRole) && LoopBackPortNumber == sh.FabricProperties.VTEPLoopBackPortNumber {
		if NeighborVTEPLoopBack != "" {
			LOG.Infoln("MCT Neighbor Device Found With NeighborVTEPLoopBack ", NeighborVTEPLoopBack)
			if NeighborVTEPLoopBack != LoopBackOnDevice {
				//TODO Find out expected Behaviour during this case
				//Other way out should be to through Error if device is Configured with different IP address
				//Than NeighborVTEPLoopBack
				LOG.Infof("Switch %s is configured with VTEP IP %s ", Device.IPAddress, LoopBackOnDevice)
				LOG.Infof("Assigning MCT Neighbor Devices VTEP Loop back IP to Device VTEP Loop Back IP ")
				if LoopBackOnDevice != "" {
					err := sh.ReleaseIP(ctx, sh.FabricID, Device.ID, domain.IntfTypeLoopback, LoopBackOnDevice, InterfaceID)
					if err != nil {
						LOG.Errorf("Error Releasing IP address %s for Device %s", LoopBackOnDevice, Device.IPAddress)
					}
				}
				LoopBackIPConfigType = domain.ConfigCreate
				return NeighborVTEPLoopBack, LoopBackIPConfigType, nil

				//LoopBackOnDevice = NeighborVTEPLoopBack

			}
			//They are same, nothing do
			LoopBackIPConfigType = domain.ConfigNone
			return NeighborVTEPLoopBack, LoopBackIPConfigType, nil

		}
	}

	//If On Switch Loopback is empty and DB Switch Loopback is empty, Get New Loopback
	if LoopBackOnDevice == "" && LoopBackInDB == "" {

		ip, err := sh.reserveOrObtainLoopbackIP(ctx, Device.ID, LoopBackOnDevice,
			LoopBackPortNumber, LoopBackRange, "Loopback", InterfaceID)
		if err != nil {
			return LoopBackIPToConfigure, LoopBackIPConfigType, err
		}
		LoopBackIPToConfigure = ip
		//Needs to push to Switch so set Config type as ConfigCreate
		LoopBackIPConfigType = domain.ConfigCreate
		LOG.Infoln("Compute Loopback: Loopback Allocated:", LoopBackIPToConfigure, "Config Type:", LoopBackIPConfigType)
		return LoopBackIPToConfigure, LoopBackIPConfigType, nil
	}
	//If configs are same at both the places, nothing needs to be done
	if LoopBackOnDevice == LoopBackInDB {
		//Nothing needs to be done
		LoopBackIPToConfigure = LoopBackOnDevice
		LoopBackIPConfigType = domain.ConfigNone
		LOG.Infoln("Compute Loopback: Loopback No change:", LoopBackIPToConfigure, "Config Type:", LoopBackIPConfigType)
		return LoopBackIPToConfigure, LoopBackIPConfigType, nil
	}
	//If On Switch Loopback is empty and DB Switch Loopback is valid, copy it to OnSwitch
	if LoopBackOnDevice == "" && LoopBackInDB != "" {
		//Needs to push to Switch so set Config type as ConfigCreate
		LoopBackIPToConfigure = LoopBackInDB
		LoopBackIPConfigType = domain.ConfigCreate
		LOG.Infoln("Compute Loopback: Loopback Re-use from Pool:", LoopBackIPToConfigure, "Config Type:", LoopBackIPConfigType)
		return LoopBackIPToConfigure, LoopBackIPConfigType, nil
	}

	//If On Switch Loopback has a value and DB Switch Loopback is empty, Try reserving the same value
	if LoopBackOnDevice != "" && LoopBackInDB == "" {
		//Needs to push to Switch so set Config type as ConfigCreate
		ip, err := sh.reserveOrObtainLoopbackIP(ctx, Device.ID, LoopBackOnDevice,
			LoopBackPortNumber, LoopBackRange, "Loopback", InterfaceID)
		if err != nil {
			return LoopBackIPToConfigure, LoopBackIPConfigType, err
		}
		LoopBackIPToConfigure = ip
		//No need to push to switch as we were able to reserve the same ASN
		LoopBackIPConfigType = domain.ConfigNone
		LOG.Infoln("Compute Loopback: Loopback Allocated:", LoopBackOnDevice, "Config Type:", LoopBackIPConfigType)
		return LoopBackIPToConfigure, LoopBackIPConfigType, nil
	}

	//If On Switch Loopback has a differnt Value to one on the DB, Reserve the one on Switch and Release on in DB
	if LoopBackOnDevice != LoopBackInDB {
		//Release the Loopback in the DB
		sh.ReleaseIP(ctx, sh.FabricID, Device.ID, domain.IntfTypeLoopback, LoopBackInDB, InterfaceID)

		//Needs to push to Switch so set Config type as ConfigCreate
		reserveLoopbackIP, err := sh.reserveOrObtainLoopbackIP(ctx, Device.ID, LoopBackOnDevice,
			LoopBackPortNumber, LoopBackRange, "Loopback", InterfaceID)
		if err != nil {
			return LoopBackIPToConfigure, LoopBackIPConfigType, err
		}
		LoopBackIPToConfigure = reserveLoopbackIP
		//We need to update the Loopback IP
		LoopBackIPConfigType = domain.ConfigUpdate
		LOG.Infoln("Compute Loopback: Old Loopback Released,ASN Allocated:", LoopBackOnDevice, "Config Type:", LoopBackIPConfigType)
		return LoopBackIPToConfigure, LoopBackIPConfigType, nil
	}

	return LoopBackIPToConfigure, LoopBackIPConfigType, nil
}

func (sh *DeviceInteractor) computeASN(ctx context.Context, Device *domain.Device, FabricProperties *domain.FabricProperties,
	DBSwitchConfig *domain.SwitchConfig, OnSwitchConfig *domain.SwitchConfig, NeighborAsn string) error {
	LOG := appcontext.Logger(ctx)
	CurrentRole := Device.DeviceRole

	//Set the ASNBlock based on Role
	asnBlock := FabricProperties.LeafASNBlock
	if CurrentRole == SpineRole {
		asnBlock = FabricProperties.SpineASNBlock
	}
	if CurrentRole == RackRole {
		asnBlock = FabricProperties.RackASNBlock
	}
	if Device.DeviceRole == LeafRole || Device.DeviceRole == RackRole {
		if NeighborAsn != "" {
			LOG.Infoln("MCT Neighbor Device Found With ASN ", NeighborAsn)
			if Device.LocalAs != NeighborAsn {
				//TODO Find out expected Behaviour during this case
				LOG.Infof("Switch %s is configured with ASN %s ", Device.IPAddress, Device.LocalAs)
				LOG.Info("Assigning MCT Neighbot Device Local ASN ")
				//For Now Letus Copy NeighborAsn As Device ASN
				if Device.LocalAs != "" {
					//Release ASN if Allocated
					usedAsn, _ := strconv.ParseUint(Device.LocalAs, 10, 64)
					err := sh.ReleaseASN(ctx, sh.FabricID, Device.ID, Device.DeviceRole, usedAsn)
					if err != nil {
						LOG.Errorf("Error Releasing ASN %s for Device %s", Device.LocalAs, Device.IPAddress)
					}
				}
				OnSwitchConfig.ASConfigType = domain.ConfigCreate
				OnSwitchConfig.LocalAS = NeighborAsn
				return nil
			}

		}
	}
	LOG.Infoln("Compute ASN", "ASN On Switch:", Device.LocalAs, ",ASN in DB:", DBSwitchConfig.LocalAS)
	//If On Switch AS is empty and DB Switch AS is empty, Get New ASN
	if Device.LocalAs == "" && DBSwitchConfig.LocalAS == "" {
		//Needs to push to Switch so set Config type as ConfigCreate
		asn, err := sh.reserveOrObtainASN(ctx, Device.ID, Device.LocalAs, CurrentRole, asnBlock)
		if err != nil {
			return err
		}
		OnSwitchConfig.LocalAS = asn
		//Needs to push to Switch so set Config type as ConfigCreate
		OnSwitchConfig.ASConfigType = domain.ConfigCreate
		LOG.Infoln("Compute ASN: ASN Allocated:", OnSwitchConfig.LocalAS, "Config Type:", OnSwitchConfig.ASConfigType)
		return nil
	}
	//If configs are same at both the places, nothing needs to be done
	if Device.LocalAs == DBSwitchConfig.LocalAS {
		//Nothing needs to be done
		OnSwitchConfig.LocalAS = Device.LocalAs
		OnSwitchConfig.ASConfigType = domain.ConfigNone
		LOG.Infoln("Compute ASN: ASN No change:", OnSwitchConfig.LocalAS, "Config Type:", OnSwitchConfig.ASConfigType)
		return nil
	}
	//If On Switch AS is empty and DB Switch AS is valid, copy it to OnSwitch
	if Device.LocalAs == "" && DBSwitchConfig.LocalAS != "" {
		//Needs to push to Switch so set Config type as ConfigCreate
		OnSwitchConfig.LocalAS = DBSwitchConfig.LocalAS
		OnSwitchConfig.ASConfigType = domain.ConfigCreate
		LOG.Infoln("Compute ASN: ASN Re-use from Pool:", OnSwitchConfig.LocalAS, "Config Type:", OnSwitchConfig.ASConfigType)
		return nil
	}

	//If On Switch AS has a vlaue and DB Switch AS is empty, Try reserving the same value
	if Device.LocalAs != "" && DBSwitchConfig.LocalAS == "" {
		//Needs to push to Switch so set Config type as ConfigCreate
		asn, err := sh.reserveOrObtainASN(ctx, Device.ID, Device.LocalAs, CurrentRole, asnBlock)
		if err != nil {
			return err
		}
		OnSwitchConfig.LocalAS = asn
		//No need to push to switch as we were able to reserve the same ASN
		OnSwitchConfig.ASConfigType = domain.ConfigNone
		LOG.Infoln("Compute ASN: ASN Allocated:", asn, "Config Type:", OnSwitchConfig.ASConfigType)
		return nil
	}

	//If On Switch AS has a differnt Value to one on the DB, Reserve the one on Switch and Release on in DB
	if Device.LocalAs != DBSwitchConfig.LocalAS {
		//Release the ASN in the DB
		asn, _ := strconv.ParseUint(DBSwitchConfig.LocalAS, 10, 64)
		sh.ReleaseASN(ctx, sh.FabricID, Device.ID, DBSwitchConfig.Role, asn)

		//Needs to push to Switch so set Config type as ConfigCreate
		reserverasn, err := sh.reserveOrObtainASN(ctx, Device.ID, Device.LocalAs, CurrentRole, asnBlock)
		if err != nil {
			return err
		}
		OnSwitchConfig.LocalAS = reserverasn
		//No need to push to switch as we were able to reserve the same ASN
		OnSwitchConfig.ASConfigType = domain.ConfigNone
		LOG.Infoln("Compute ASN: Old ASN Released,ASN Allocated:", asn, "Config Type:", OnSwitchConfig.ASConfigType)
		return nil
	}

	return nil
}

func (sh *DeviceInteractor) evaluateIP(oldlp domain.Interface, LoopbackRange string) (string, error) {
	switchIP := ""
	if oldlp.IPAddress != "" {
		_, fabricLoopBackNet, _ := net.ParseCIDR(LoopbackRange)
		switchLoopBackIP, _, err := net.ParseCIDR(oldlp.IPAddress)
		if err == nil {
			switchIP = switchLoopBackIP.String()
		}

		//Check if the IP address is in the Fabric Property range
		if fabricLoopBackNet.Contains(switchLoopBackIP) == false {
			statusMsg := fmt.Sprintf("%s %s on %s not in %s Range", switchLoopBackIP, domain.IntfTypeLoopback, oldlp.IntName,
				fabricLoopBackNet)
			//Not in Fabric Range Error
			return switchIP, errors.New(statusMsg)
		}
	}
	return switchIP, nil
}

func (sh *DeviceInteractor) createLoopbackIntrfaceIfNotExists(InterfaceName string, DeviceID uint) (domain.Interface, error) {
	intf, err := sh.Db.GetInterface(sh.FabricID, DeviceID, domain.IntfTypeLoopback, InterfaceName)
	//Object does not exist so create one
	if err != nil {
		intf = domain.Interface{FabricID: sh.FabricID, DeviceID: DeviceID, IntType: domain.IntfTypeLoopback, IntName: InterfaceName}
		sh.Db.CreateInterface(&intf)
	}

	if intf.ConfigType == domain.ConfigDelete {
		intf.IPAddress = ""
	}
	//If the Loopback interface is marked for deletion re-set it (Loopback interfaces even when they are delete
	//it needs to be re-created)
	sh.Db.ResetDeleteOperationOnInterface(sh.FabricID, intf.ID)

	return intf, nil
}

func (sh *DeviceInteractor) interfaceConfigsBasedOnDeletedLLDPNeighbors(ctx context.Context, device *domain.Device,
	FabricProperties *domain.FabricProperties) (string, error) {
	lldpNeighborsMarkedForDeletion, err := sh.Db.GetLLDPNeighborsOnDeviceMarkedForDeletion(sh.FabricID, device.ID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to Fetch LLDP Neighbors for %s", device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	//If there are lldpNeighbors marked for deletion then mark corresponding InterfaceConfigs and RemoteNeighbor configs
	//for deletion
	if len(lldpNeighborsMarkedForDeletion) > 0 {
		InterfaceIDs := make([]uint, 0)
		for _, lldpNeighbor := range lldpNeighborsMarkedForDeletion {
			InterfaceIDs = append(InterfaceIDs, lldpNeighbor.InterfaceOneID)
			InterfaceIDs = append(InterfaceIDs, lldpNeighbor.InterfaceTwoID)
			//Only for /31 (non loopback address release the IP address back to the pool)
			if FabricProperties.P2PIPType == domain.P2PIpTypeNumbered {
				InterfaceOneIP, _, _ := net.ParseCIDR(lldpNeighbor.InterfaceOneIP)
				InterfaceTwoIP, _, _ := net.ParseCIDR(lldpNeighbor.InterfaceTwoIP)
				//Release IPs from the Pool
				sh.ReleaseIPPair(ctx, sh.FabricID, lldpNeighbor.DeviceOneID, lldpNeighbor.DeviceTwoID, "P2P",
					InterfaceOneIP.String(), InterfaceTwoIP.String(), lldpNeighbor.InterfaceOneID, lldpNeighbor.InterfaceTwoID)
			}
		}
		//Mark the Interface SwitchConfigs for delete
		if err := sh.Db.UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs(sh.FabricID, InterfaceIDs, domain.ConfigDelete); err != nil {
			statusMsg := fmt.Sprintf("Unable to mark for Interface Switch Configs for deletion for %s", device.IPAddress)
			return statusMsg, errors.New(statusMsg)
		}
		//Mark the BGP Switch Configs for delete
		if err := sh.Db.UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID(sh.FabricID, InterfaceIDs, domain.ConfigDelete); err != nil {
			statusMsg := fmt.Sprintf("Unable to mark for BGP Switch Configs for deletion for %s", device.IPAddress)
			return statusMsg, errors.New(statusMsg)
		}
	}

	return "", nil
}

func (sh *DeviceInteractor) prepareMapOfSwitchConfigs(ctx context.Context) map[uint]domain.SwitchConfig {
	SwitchConfigs := sh.GetSwitchConfigs(ctx, sh.FabricName)
	SwitchConfigsMap := make(map[uint]domain.SwitchConfig, 0)
	for _, sw := range SwitchConfigs {
		SwitchConfigsMap[sw.DeviceID] = sw
	}
	return SwitchConfigsMap
}

func (sh *DeviceInteractor) buildInterfaceConfigs(ctx context.Context, device *domain.Device, rack bool) (string, error) {
	LOG := appcontext.Logger(ctx)
	LOG.Infoln("Build interface configs")
	FabricProperties, _ := sh.Db.GetFabricProperties(sh.FabricID)
	if msg, err := sh.interfaceConfigsBasedOnDeletedLLDPNeighbors(ctx, device, &FabricProperties); err != nil {
		return msg, err
	}

	lldpNeighbors, err := sh.Db.GetLLDPNeighborsOnDevice(sh.FabricID, device.ID)
	//fmt.Println("CRE", lldpNeighbors)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to Fetch LLDP Neighbors for %s", device.IPAddress)
		return statusMsg, errors.New(statusMsg)
	}
	InterfaceConfigs := make([]domain.InterfaceSwitchConfig, 0)
	RemoteInterfaceConfigs := make([]domain.RemoteNeighborSwitchConfig, 0)
	SwitchConfigMap := sh.prepareMapOfSwitchConfigs(ctx)
	//Interface IDs to be used for querying the InterfaceConfigs and BGP Remote Configs
	//for comparison.
	InterfaceIDs := make([]uint, 0)
	MCTInterfaceIDs := make([]uint, 0)
	MCTBGPConfigs := make([]domain.RemoteNeighborSwitchConfig, 0)

	//For each neighbor find the Interface Configs and BGP Neighbor configs
	for _, neighbor := range lldpNeighbors {

		if neighbor.DeviceOneRole == LeafRole && neighbor.DeviceTwoRole == LeafRole {
			//MCT Cluster create BGP neighbor of TYPE NSH

			LOG.Infoln("Handle BGP MCT case")
			MCTBGPOneConf, MCTBGPTwoConf, InterfaceOneID, InterrfaceTwoID, err := sh.buildMctConfigForNeighbor(ctx, &neighbor, &FabricProperties, SwitchConfigMap)
			if err != nil {
				return "Unable to build MCT BGP configurations", err

			}
			if InterfaceOneID != 0 && InterrfaceTwoID != 0 {
				MCTInterfaceIDs = append(MCTInterfaceIDs, InterfaceOneID)
				MCTInterfaceIDs = append(MCTInterfaceIDs, InterrfaceTwoID)
				if neighbor.ConfigType != domain.ConfigDelete {
					MCTBGPConfigs = append(MCTBGPConfigs, MCTBGPOneConf)
					MCTBGPConfigs = append(MCTBGPConfigs, MCTBGPTwoConf)
				}
			}
		} else {
			if neighbor.ConfigType == domain.ConfigDelete {
				continue
			}
			IntOneConf, IntTwoConf, RemoteIntOneConf, RemoteIntTwoConf, err := sh.buildInterfaceConfigForNeighbor(ctx,
				&neighbor, &FabricProperties, SwitchConfigMap)
			if err != nil {
				return "Unable to build Interface conf", err
			}
			//Use all the Interfaces for DB fetch
			InterfaceIDs = append(InterfaceIDs, neighbor.InterfaceOneID)
			InterfaceIDs = append(InterfaceIDs, neighbor.InterfaceTwoID)
			InterfaceConfigs = append(InterfaceConfigs, IntOneConf)
			InterfaceConfigs = append(InterfaceConfigs, IntTwoConf)
			RemoteInterfaceConfigs = append(RemoteInterfaceConfigs, RemoteIntOneConf)
			RemoteInterfaceConfigs = append(RemoteInterfaceConfigs, RemoteIntTwoConf)

		}
		if neighbor.ConfigType == domain.ConfigDelete {
			continue
		}
		//Update the neighbor IP Address to DB
		sh.Db.CreateLLDPNeighbor(&neighbor)
	}
	//Persist MCT Remote Interface Configs
	if _, err := sh.persistRemoteInterfaceConfigs(ctx, device, MCTBGPConfigs, MCTInterfaceIDs); err != nil {
		return "Failed to persist Remote Interface Configs", err
	}

	//Persist Interface Configs
	if _, err := sh.persistInterfaceConfigs(ctx, device, InterfaceConfigs, InterfaceIDs, &FabricProperties); err != nil {
		return "Failed to persist Interface Configs", err
	}
	//Persist Remote Interface Configs
	if _, err := sh.persistRemoteInterfaceConfigs(ctx, device, RemoteInterfaceConfigs, InterfaceIDs); err != nil {
		return "Failed to persist Remote Interface Configs", err
	}

	return "", nil
}

func (sh *DeviceInteractor) persistInterfaceConfigs(ctx context.Context, device *domain.Device, InterfaceConfigs []domain.InterfaceSwitchConfig,
	InterfaceIDs []uint, FabricProperties *domain.FabricProperties) (string, error) {
	LOG := appcontext.Logger(ctx)
	OldInterfaceConfigs, _ := sh.Db.GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion(sh.FabricID, InterfaceIDs)
	LOG.Infoln("Old Interface Configs", OldInterfaceConfigs)
	LOG.Infoln("New Interface Configs", InterfaceConfigs)
	//Methods to use as key for Set Operations
	GetInterfaceKey := func(data interface{}) string {
		s, _ := data.(domain.InterfaceSwitchConfig)
		return fmt.Sprintln(s.DeviceID, s.IntName, s.IntType)
	}
	//Method to compare two Interface objects to find whether they are updated
	InterfaceEqualMethod := func(first interface{}, second interface{}) (bool, domain.InterfaceSwitchConfig, domain.InterfaceSwitchConfig) {
		f, _ := first.(domain.InterfaceSwitchConfig)
		s, _ := second.(domain.InterfaceSwitchConfig)
		if f.DonorType == s.DonorType && f.DonorName == s.DonorName && f.IPAddress == s.IPAddress {
			return true, f, s
		}
		return false, f, s
	}
	CreatedInterfaceConfigs, DeletedInterfaceConfigs, UpdatedInterfaceConfigs := interfaceswitchconfig.Compare(GetInterfaceKey, InterfaceEqualMethod,
		OldInterfaceConfigs, InterfaceConfigs)
	LOG.Infoln("Created Interface Configs", CreatedInterfaceConfigs)
	LOG.Infoln("Updated Interface Configs", UpdatedInterfaceConfigs)
	LOG.Infoln("Deleted Interface Configs", DeletedInterfaceConfigs)

	for _, CIntf := range CreatedInterfaceConfigs {
		CIntf.ConfigType = domain.ConfigCreate
		if err := sh.Db.CreateInterfaceSwitchConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to create Interface Config %s %s", CIntf.IntType, CIntf.IntName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, UpdateData := range UpdatedInterfaceConfigs {
		//If IP address dont match release the IP from the Pool

		/* TODO need to check this -- IP would have got released
		since they are released in pairs
		if FabricProperties.P2PIPType == domain.p2PIpTypeNumbered {
			if UpdateData.New.IPAddress != UpdateData.Old.IPAddress {
				fmt.Println("Release", UpdateData.Old.IPAddress, UpdateData.Old.InterfaceID)
				err := sh.ReleaseIP(ctx, sh.FabricID, UpdateData.Old.DeviceID, "P2P", UpdateData.Old.IPAddress, UpdateData.Old.InterfaceID)
				fmt.Println(err)
			}
		}*/
		UpdateData.New.ConfigType = domain.ConfigUpdate
		if err := sh.Db.CreateInterfaceSwitchConfig(&UpdateData.New); err != nil {
			statusMsg := fmt.Sprintf("Failed to update Interface Config %s %s", UpdateData.New.IntType, UpdateData.New.IntName)
			return statusMsg, errors.New(statusMsg)
		}
	}

	//For Delete mark it as delete
	for _, CIntf := range DeletedInterfaceConfigs {
		CIntf.ConfigType = domain.ConfigDelete
		if err := sh.Db.CreateInterfaceSwitchConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to Update Interface Config %s %s", CIntf.IntType, CIntf.IntName)
			return statusMsg, errors.New(statusMsg)
		}
	}
	return "", nil
}

func (sh *DeviceInteractor) persistRemoteInterfaceConfigs(ctx context.Context, device *domain.Device,
	RemoteNeighborSwitchConfigs []domain.RemoteNeighborSwitchConfig, InterfaceIDs []uint) (string, error) {
	LOG := appcontext.Logger(ctx)
	OldRemoteNeighborSwitchConfigs, _ := sh.Db.GetBGPSwitchConfigsExcludingMarkedForDeletion(sh.FabricID, InterfaceIDs)
	LOG.Infoln("Old Remote Neighbor Configs", OldRemoteNeighborSwitchConfigs)
	LOG.Infoln("New Remote Neighbor Configs", RemoteNeighborSwitchConfigs)
	//Methods to use as key for Set Operations
	GetInterfaceKey := func(data interface{}) string {
		s, _ := data.(domain.RemoteNeighborSwitchConfig)
		return fmt.Sprintln(s.DeviceID, s.RemoteDeviceID, s.RemoteInterfaceID)
	}
	//Method to compare two Interface objects to find whether they are updated
	InterfaceEqualMethod := func(first interface{}, second interface{}) (bool, domain.RemoteNeighborSwitchConfig, domain.RemoteNeighborSwitchConfig) {
		f, _ := first.(domain.RemoteNeighborSwitchConfig)
		s, _ := second.(domain.RemoteNeighborSwitchConfig)
		if f.RemoteIPAddress == s.RemoteIPAddress && f.RemoteAS == s.RemoteAS {
			return true, f, s
		}
		return false, f, s
	}
	CreatedRemoteInterfaceConfigs, DeletedRemoteInterfaceConfigs, UpdatedRemoteInterfaceConfigs := bgpconfig.Compare(GetInterfaceKey, InterfaceEqualMethod,
		OldRemoteNeighborSwitchConfigs, RemoteNeighborSwitchConfigs)

	LOG.Infoln("Created Remote Neighbor Configs", CreatedRemoteInterfaceConfigs)
	LOG.Infoln("Updated Remote Neighbor Configs", UpdatedRemoteInterfaceConfigs)
	LOG.Infoln("Deleted Remote Neighbor Configs", DeletedRemoteInterfaceConfigs)

	for _, CIntf := range CreatedRemoteInterfaceConfigs {
		CIntf.ConfigType = domain.ConfigCreate
		if err := sh.Db.CreateBGPSwitchConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to create Interface Config %d %d", CIntf.RemoteDeviceID, CIntf.RemoteInterfaceID)
			return statusMsg, errors.New(statusMsg)
		}
	}
	for _, CIntf := range UpdatedRemoteInterfaceConfigs {
		CIntf.ConfigType = domain.ConfigUpdate
		if err := sh.Db.CreateBGPSwitchConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to update Interface Config %d %d", CIntf.RemoteDeviceID, CIntf.RemoteInterfaceID)
			return statusMsg, errors.New(statusMsg)
		}
	}
	//For Delete mark it as delete
	for _, CIntf := range DeletedRemoteInterfaceConfigs {
		CIntf.ConfigType = domain.ConfigDelete
		if err := sh.Db.CreateBGPSwitchConfig(&CIntf); err != nil {
			statusMsg := fmt.Sprintf("Failed to Update Interface Config %d %d", CIntf.RemoteDeviceID, CIntf.RemoteInterfaceID)
			return statusMsg, errors.New(statusMsg)
		}
	}
	return "", nil
}

func (sh *DeviceInteractor) buildMctConfigForNeighbor(ctx context.Context, neighbor *domain.LLDPNeighbor,
	FabricProperties *domain.FabricProperties, SwitchConfigMap map[uint]domain.SwitchConfig) (domain.RemoteNeighborSwitchConfig, domain.RemoteNeighborSwitchConfig, uint, uint, error) {
	LOG := appcontext.Logger(ctx)
	OldMcts, merr := sh.Db.GetMctClustersWithBothDevices(neighbor.FabricID, neighbor.DeviceOneID, neighbor.DeviceTwoID, []string{})
	if merr != nil {
		statusMsg := fmt.Sprintln("Unable to Retrieve MctCluster from DB for device and Neighbor ", SwitchConfigMap[neighbor.DeviceOneID], SwitchConfigMap[neighbor.DeviceTwoID])
		LOG.Errorln(statusMsg, merr)
		return domain.RemoteNeighborSwitchConfig{}, domain.RemoteNeighborSwitchConfig{}, 0, 0, merr
	}
	if len(OldMcts) == 0 {
		statusMsg := fmt.Sprintln("No MCT Cluster Found for Device and Neighbor", SwitchConfigMap[neighbor.DeviceOneID], SwitchConfigMap[neighbor.DeviceTwoID])
		LOG.Debug(statusMsg)
		return domain.RemoteNeighborSwitchConfig{}, domain.RemoteNeighborSwitchConfig{}, 0, 0, nil
	}
	if len(OldMcts) > 1 {
		statusMsg := fmt.Sprintln("More than One Cluster Found for Device  and Neighbor", SwitchConfigMap[neighbor.DeviceOneID], SwitchConfigMap[neighbor.DeviceTwoID])
		LOG.Info(statusMsg)
	}
	oldMct := OldMcts[0]
	if oldMct.ConfigType == domain.ConfigDelete {
		return domain.RemoteNeighborSwitchConfig{}, domain.RemoteNeighborSwitchConfig{}, oldMct.VEInterfaceOneID, oldMct.VEInterfaceTwoID, nil
	}
	remoteAS := sh.fetchASforDevice(ctx, oldMct.MCTNeighborDeviceID, oldMct.DeviceTwoMgmtIP)
	RemoteIntOneConf := sh.prepareRemoteInterfaceConfig(oldMct.DeviceID, oldMct.VEInterfaceTwoID,
		oldMct.MCTNeighborDeviceID, oldMct.PeerOneIP, remoteAS, domain.BGPEncapTypeForCluster)

	remoteAS = sh.fetchASforDevice(ctx, oldMct.DeviceID, oldMct.DeviceOneMgmtIP)
	RemoteIntTwoConf := sh.prepareRemoteInterfaceConfig(oldMct.MCTNeighborDeviceID, oldMct.VEInterfaceOneID,
		oldMct.DeviceID, oldMct.PeerTwoIP, remoteAS, domain.BGPEncapTypeForCluster)
	return RemoteIntOneConf, RemoteIntTwoConf, oldMct.VEInterfaceOneID, oldMct.VEInterfaceTwoID, nil
}

func (sh *DeviceInteractor) buildInterfaceConfigForNeighbor(ctx context.Context, neighbor *domain.LLDPNeighbor,
	FabricProperties *domain.FabricProperties, SwitchConfigMap map[uint]domain.SwitchConfig) (domain.InterfaceSwitchConfig, domain.InterfaceSwitchConfig,
	domain.RemoteNeighborSwitchConfig, domain.RemoteNeighborSwitchConfig, error) {

	var err error
	donorType := ""
	donorName := ""
	//For Numbered cases fetch the Interface IP from the Pool
	if FabricProperties.P2PIPType == domain.P2PIpTypeNumbered {
		if neighbor.InterfaceOneIP, neighbor.InterfaceTwoIP, err = sh.reserveOrObtainIPPair(ctx, neighbor.DeviceOneID, neighbor.DeviceTwoID,
			neighbor.InterfaceOneType, neighbor.InterfaceOneName, neighbor.InterfaceTwoType, neighbor.InterfaceTwoName,
			FabricProperties.P2PLinkRange, "P2P", neighbor.InterfaceOneID, neighbor.InterfaceTwoID); err != nil {
			return domain.InterfaceSwitchConfig{}, domain.InterfaceSwitchConfig{},
				domain.RemoteNeighborSwitchConfig{}, domain.RemoteNeighborSwitchConfig{}, err
		}
	} else {
		//For Un-numbered interfaces
		donorType = domain.IntfTypeLoopback
		donorName = FabricProperties.LoopBackPortNumber

		neighbor.InterfaceOneIP = SwitchConfigMap[neighbor.DeviceOneID].LoopbackIP
		neighbor.InterfaceTwoIP = SwitchConfigMap[neighbor.DeviceTwoID].LoopbackIP

	}
	//Interface One
	IntOneDescription := fmt.Sprintf("Link to %s %s", SwitchConfigMap[neighbor.DeviceTwoID].DeviceIP,
		SwitchConfigMap[neighbor.DeviceTwoID].Role)
	IntOneConf := sh.prepareInterfaceSwitchConfig(neighbor.DeviceOneID, neighbor.InterfaceOneID,
		neighbor.InterfaceOneIP, neighbor.InterfaceOneName, neighbor.InterfaceOneType,
		donorType, donorName, IntOneDescription)

	remoteAS := sh.fetchASforDevice(ctx, neighbor.DeviceTwoID, SwitchConfigMap[neighbor.DeviceTwoID].DeviceIP)
	RemoteIntOneConf := sh.prepareRemoteInterfaceConfig(neighbor.DeviceOneID, neighbor.InterfaceTwoID,
		neighbor.DeviceTwoID, neighbor.InterfaceTwoIP, remoteAS, "vxlan")

	//Interface Two
	IntTwoDescription := fmt.Sprintf("Link to %s %s", SwitchConfigMap[neighbor.DeviceOneID].DeviceIP,
		SwitchConfigMap[neighbor.DeviceOneID].Role)
	IntTwoConf := sh.prepareInterfaceSwitchConfig(neighbor.DeviceTwoID, neighbor.InterfaceTwoID,
		neighbor.InterfaceTwoIP, neighbor.InterfaceTwoName, neighbor.InterfaceTwoType,
		donorType, donorName, IntTwoDescription)

	remoteAS = sh.fetchASforDevice(ctx, neighbor.DeviceOneID, SwitchConfigMap[neighbor.DeviceOneID].DeviceIP)
	RemoteIntTwoConf := sh.prepareRemoteInterfaceConfig(neighbor.DeviceTwoID, neighbor.InterfaceOneID,
		neighbor.DeviceOneID, neighbor.InterfaceOneIP, remoteAS, "vxlan")

	return IntOneConf, IntTwoConf, RemoteIntOneConf, RemoteIntTwoConf, nil
}

func (sh *DeviceInteractor) fetchASforDevice(ctx context.Context, DeviceID uint, DeviceIP string) string {
	LOG := appcontext.Logger(ctx)
	var device domain.SwitchConfig

	device, err := sh.Db.GetSwitchConfigOnFabricIDAndDeviceID(sh.FabricID, DeviceID)
	if err == nil {
		LOG.Infof("ASN for device %s is %s", DeviceIP, device.LocalAS)
		return device.LocalAS
	}

	return ""

}

func (sh *DeviceInteractor) prepareRemoteInterfaceConfig(DeviceID uint, RemoteInterfaceID uint, RemoteDeviceID uint,
	RemoteIPAddress string, RemoteAS string, EncapsulationType string) domain.RemoteNeighborSwitchConfig {

	//oldRemoteInterfaceCondfig,err := sh.Db.GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID(sh.FabricID, RemoteInterfaceID)
	IntConf := domain.RemoteNeighborSwitchConfig{FabricID: sh.FabricID, DeviceID: DeviceID, RemoteInterfaceID: RemoteInterfaceID,
		RemoteDeviceID:    RemoteDeviceID,
		RemoteIPAddress:   RemoteIPAddress,
		RemoteAS:          RemoteAS,
		EncapsulationType: EncapsulationType}
	IntConf.ConfigType = domain.ConfigNone
	return IntConf
}

func (sh *DeviceInteractor) prepareInterfaceSwitchConfig(DeviceID uint, InterfaceID uint, ipaddress string,
	intfName string, intfType string, donorType string, donorName string, description string) domain.InterfaceSwitchConfig {

	//oldIterfaceConfig,err := sh.Db.GetInterfaceSwitchConfigOnFabricIDAndInterfaceID(sh.FabricID, InterfaceID)
	IntConf := domain.InterfaceSwitchConfig{FabricID: sh.FabricID, DeviceID: DeviceID, InterfaceID: InterfaceID,
		IPAddress: ipaddress, IntName: intfName, IntType: intfType,
		DonorType: donorType, DonorName: donorName, Description: description}
	IntConf.ConfigType = domain.ConfigNone
	return IntConf
}

func (sh *DeviceInteractor) reserveOrObtainASN(ctx context.Context, DeviceID uint,
	ASNOnDevice string, SwitchRole string, asnblock string) (string, error) {
	LOG := appcontext.Logger(ctx)
	AllocatedASN := ""
	var err error

	//Try reserving the same ASN as on the Device
	if ASNOnDevice != "" {
		asn, _ := strconv.ParseUint(ASNOnDevice, 10, 64)

		//Validate the range
		asnMin, asnMax := GetASNMinMax(asnblock)
		if asn > asnMax || asn < asnMin {
			return AllocatedASN, fmt.Errorf("ASN %s not in range %s", ASNOnDevice, asnblock)
		}

		LOG.Infof("Try reserving ASN %s for Device %d", ASNOnDevice, DeviceID)
		//If non-empty ASN try reserving
		if err = sh.ReserveASN(ctx, sh.FabricID, DeviceID, SwitchRole, asn); err == nil {
			AllocatedASN = ASNOnDevice
		}
	} else {

		//Obtain the ASN
		LOG.Infof("Get new ASN for Device")
		var asn uint64
		if asn, err = sh.GetASN(ctx, sh.FabricID, DeviceID, SwitchRole); err == nil {
			AllocatedASN = fmt.Sprint(asn)
		}

	}
	LOG.Infoln("ASN allocated for Device", AllocatedASN)
	return AllocatedASN, err
}

//This function tries to reserve an existing IP address from the Pool or obtain a new
//address from the pool
func (sh *DeviceInteractor) reserveOrObtainLoopbackIP(ctx context.Context, DeviceID uint, OnSwitchLoopbackIP string,
	InterfaceName string, LoopbackRange string, IPtype string, InterfaceID uint) (string, error) {
	LOG := appcontext.Logger(ctx)
	AllocatedIP := ""

	if OnSwitchLoopbackIP != "" {

		LOG.Infof("Try reserving interface IP %s for interface %s", OnSwitchLoopbackIP, InterfaceName)

		//Try reserving the IP address from the Pool
		reserveErr := sh.ReserveIP(ctx, sh.FabricID, DeviceID, IPtype, OnSwitchLoopbackIP, InterfaceID)
		if reserveErr != nil {
			//Cannot Reserve Error
			return AllocatedIP, reserveErr
		}
		//Reservation success so retain the Switch IP
		AllocatedIP = OnSwitchLoopbackIP
		LOG.Infof("Reserved IP %s for %s %s", OnSwitchLoopbackIP, domain.IntfTypeLoopback, InterfaceName)
		return AllocatedIP, nil

	}

	// Try getting a new one
	AllocatedIP, err := sh.GetIP(ctx, sh.FabricID, DeviceID, IPtype, InterfaceID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to get IP address from Pool")
		LOG.Println(statusMsg)

		return "", errors.New(statusMsg)
	}

	LOG.Infof("Allocated IP %s", AllocatedIP)
	return AllocatedIP, nil
}
func (sh *DeviceInteractor) findPeerIP(network string) (string, error) {
	ip, ipnet, _ := net.ParseCIDR(network)
	for peerip := ip.Mask(ipnet.Mask); ipnet.Contains(peerip); sh.inc(peerip) {
		if peerip.String() != ip.String() {
			return peerip.String() + "/31", nil
		}
	}

	return "", errors.New("Unable to obtain Peer IP")
}
func (sh *DeviceInteractor) reserveOrObtainIPPair(ctx context.Context, DeviceOneID uint, DeviceTwoID uint,
	InterfaceOneType string, InterfaceOneName string, InterfaceTwoType string, InterfaceTwoName string,
	IPRange string, IPtype string, InterfaceOneID uint, InterfaceTwoID uint) (string, string, error) {
	LOG := appcontext.Logger(ctx)
	AllocatedOneIP := ""
	AllocatedTwoIP := ""

	intfOne, intOneerr := sh.Db.GetInterface(sh.FabricID, DeviceOneID, InterfaceOneType, InterfaceOneName)
	intfTwo, intTwoerr := sh.Db.GetInterface(sh.FabricID, DeviceTwoID, InterfaceTwoType, InterfaceTwoName)
	//fmt.Println("Int One",InterfaceOneType,InterfaceOneName,InterfaceOneID,intfOne.IPAddress)
	//fmt.Println("Int Two",InterfaceTwoType,InterfaceTwoName,InterfaceTwoID,intfTwo.IPAddress)
	//Both end Interfaces should either have the IP address or not

	//If one of the interface has neighbors flag it
	if intOneerr == nil && intTwoerr == nil && intfOne.IPAddress != "" && intfTwo.IPAddress == "" {
		peerIP, perr := sh.findPeerIP(intfOne.IPAddress)
		if perr != nil {
			statusMsg := fmt.Sprintf("Both end of the neighbors are not pre-configured with IP Address %s %s %s", intfOne.IPAddress, InterfaceTwoType, InterfaceTwoName)
			return AllocatedOneIP, AllocatedTwoIP, errors.New(statusMsg)
		}
		LOG.Infof("Peer ip address is %s", peerIP)
		intfTwo.IPAddress = peerIP
		sh.Db.UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceTwoID, domain.ConfigUpdate)
		sh.Db.UpdateBGPSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceTwoID, domain.ConfigUpdate)
	}
	if intOneerr == nil && intTwoerr == nil && intfOne.IPAddress == "" && intfTwo.IPAddress != "" {
		peerIP, perr := sh.findPeerIP(intfTwo.IPAddress)
		if perr != nil {
			statusMsg := fmt.Sprintf("Both end of the neighbors are not pre-configured with IP Address %s %s %s", intfTwo.IPAddress, InterfaceOneType, InterfaceOneName)
			return AllocatedOneIP, AllocatedTwoIP, errors.New(statusMsg)
		}
		LOG.Infof("Peer ip address is %s", peerIP)
		intfOne.IPAddress = peerIP
		sh.Db.UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceOneID, domain.ConfigUpdate)
		sh.Db.UpdateBGPSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceOneID, domain.ConfigUpdate)
	}

	//Case for Both Interface have IP Address
	if intOneerr == nil && intfOne.IPAddress != "" && intTwoerr == nil && intfTwo.IPAddress != "" {
		//Interface has an IP address already configured on the Switch,
		//try reserving
		intfOneIP, _, intOneIPerr := net.ParseCIDR(intfOne.IPAddress)
		intfTwoIP, _, intTwoIPerr := net.ParseCIDR(intfTwo.IPAddress)
		_, IPRangeNet, _ := net.ParseCIDR(IPRange)
		if intOneIPerr == nil && intTwoIPerr == nil {
			//Check If IP Interfaces are  not in the IP Range
			if IPRangeNet.Contains(intfOneIP) == false {
				statusMsg := fmt.Sprintf("%s is not in %s range on %s %s", intfOneIP, IPRangeNet, intfOne.IntType, intfOne.IntName)
				//Not in Fabric Range Error
				return AllocatedOneIP, AllocatedTwoIP, errors.New(statusMsg)
			}
			if IPRangeNet.Contains(intfTwoIP) == false {
				statusMsg := fmt.Sprintf("%s is not in %s range on %s %s", intfTwoIP, IPRangeNet, intfTwo.IntType, intfTwo.IntName)
				//Not in Fabric Range Error
				return AllocatedOneIP, AllocatedTwoIP, errors.New(statusMsg)
			}

			_, IPSubnetRange, _ := net.ParseCIDR(intfOne.IPAddress)
			//If both the interfaces are not in the same subnet range then return the IP Address as requested
			//This will be caught in the validation fabric or get corrected when the peer device is Updated
			if IPSubnetRange.Contains(intfTwoIP) == false {
				AllocatedOneIP = intfOneIP.String()
				AllocatedTwoIP = intfTwoIP.String()
				return AllocatedOneIP, AllocatedTwoIP, nil
			}

			//Try reserving the IP address from the Pool
			reserveErr := sh.ReserveIPPair(ctx, sh.FabricID, DeviceOneID, DeviceTwoID, IPtype, intfOneIP.String(), intfTwoIP.String(), InterfaceOneID, InterfaceTwoID)
			if reserveErr != nil {
				//Cannot Reserve Error

				return AllocatedOneIP, AllocatedTwoIP, reserveErr
			}

			//Reservation success so retain the Switch IP
			AllocatedOneIP = intfOneIP.String()
			AllocatedTwoIP = intfTwoIP.String()
			LOG.Infof("Reserved IP %s for %s %s", intfOne.IPAddress, intfOne.IntType, intfOne.IntName)
			LOG.Infof("Reserved IP %s for %s %s", intfTwo.IPAddress, intfTwo.IntType, intfTwo.IntName)
			return AllocatedOneIP, AllocatedTwoIP, nil

		}
		//An error encountered with the IP on the interface
		statusMsg := fmt.Sprintf("Switch  IPPair (%s,%s) error Parsing", intfOne.IPAddress, intfTwo.IPAddress)
		return AllocatedOneIP, AllocatedTwoIP, errors.New(statusMsg)
	}
	var err error
	//Check if IP Address already allocated to this interfaces
	AllocatedOneIP, AllocatedTwoIP, err = sh.GetAlreadyAllocatedIPPair(ctx, sh.FabricID, DeviceOneID, DeviceTwoID, IPtype, InterfaceOneID, InterfaceTwoID)
	if err == nil {
		//IP Address was already there in the Pool and one retrieved from switch is empty
		//so mark the Config Objects as update
		sh.Db.UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceOneID, domain.ConfigUpdate)
		sh.Db.UpdateBGPSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceOneID, domain.ConfigUpdate)
		sh.Db.UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceTwoID, domain.ConfigUpdate)
		sh.Db.UpdateBGPSwitchConfigsOnInterfaceIDConfigType(sh.FabricID, InterfaceTwoID, domain.ConfigUpdate)

		statusMsg := fmt.Sprintf("IP Address already allocated to this interface")
		LOG.Println(statusMsg)
		return AllocatedOneIP, AllocatedTwoIP, nil
	}

	// Try getting a new one
	AllocatedOneIP, AllocatedTwoIP, err = sh.GetIPPair(ctx, sh.FabricID, DeviceOneID, DeviceTwoID, IPtype, InterfaceOneID, InterfaceTwoID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to get IP address from Pool")
		LOG.Println(statusMsg)

		return AllocatedOneIP, AllocatedTwoIP, errors.New(statusMsg)
	}

	LOG.Infof("Allocated IP %s %s", AllocatedOneIP, AllocatedTwoIP)
	return AllocatedOneIP, AllocatedTwoIP, nil
}

//GetSwitchConfigs TODO
func (sh *DeviceInteractor) GetSwitchConfigs(ctx context.Context, fabricName string) []domain.SwitchConfig {
	LOG := appcontext.Logger(ctx)
	//check for Switch
	var devices []domain.SwitchConfig
	devices, err := sh.Db.GetSwitchConfigs(fabricName)
	if err != nil {
		statusMsg := fmt.Sprint("No Switch Present in Fabric", fabricName)
		LOG.Println(statusMsg)
	}
	return devices
}

//GetSwitchConfig TODO
func (sh *DeviceInteractor) GetSwitchConfig(ctx context.Context, FabricName string, DeviceIP string) domain.SwitchConfig {
	LOG := appcontext.Logger(ctx)
	device, err := sh.Db.GetSwitchConfigOnDeviceIP(FabricName, DeviceIP)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to get Switch with IP %s %s", FabricName, DeviceIP)
		LOG.Println(statusMsg)
	}
	return device
}

//GetEVPNNeighborConfig TODO
func (sh *DeviceInteractor) GetEVPNNeighborConfig(ctx context.Context, FabricName string, DeviceID uint) []domain.RackEvpnNeighbors {
	LOG := appcontext.Logger(ctx)
	evpnNeighbors, err := sh.Db.GetRackEvpnConfigOnDeviceID(DeviceID)
	if err != nil {
		statusMsg := fmt.Sprintf("Failed to get Switch with IP %s %d", FabricName, DeviceID)
		LOG.Println(statusMsg)
	}
	return evpnNeighbors
}

//GetInterfaceSwitchConfigs TODO
func (sh *DeviceInteractor) GetInterfaceSwitchConfigs(ctx context.Context, FabricID uint, DeviceID uint) []domain.InterfaceSwitchConfig {
	LOG := appcontext.Logger(ctx)
	//check for Switch
	interfaces, err := sh.Db.GetInterfaceSwitchConfigsOnDeviceID(FabricID, DeviceID)
	if err != nil {
		statusMsg := fmt.Sprint("No Interface Present in Fabric", FabricID)
		LOG.Println(statusMsg)
	}
	return interfaces
}

//GetBGPSwitchConfigs TODO
func (sh *DeviceInteractor) GetBGPSwitchConfigs(ctx context.Context, FabricID uint, DeviceID uint) []domain.RemoteNeighborSwitchConfig {
	LOG := appcontext.Logger(ctx)
	//check for Switch
	bgpneighbors, err := sh.Db.GetBGPSwitchConfigsOnDeviceID(FabricID, DeviceID)
	if err != nil {
		statusMsg := fmt.Sprint("No Interface Present in Fabric", FabricID)
		LOG.Println(statusMsg)
	}
	return bgpneighbors
}

//GetMCTBGPSwitchConfigs gets MCT neighbor configs
func (sh *DeviceInteractor) GetMCTBGPSwitchConfigs(ctx context.Context, FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	LOG := appcontext.Logger(ctx)
	//check for Switch
	mctBgpNeighbors, err := sh.Db.GetMCTBGPSwitchConfigsOnDeviceID(FabricID, DeviceID)
	if err != nil {
		statusMsg := fmt.Sprintf("No BGP MCT Configs Found For Device %d", DeviceID)
		LOG.Println(statusMsg)
	}
	return mctBgpNeighbors, err
}

package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

//CloseTransaction closes the application database transaction
func (sh *DeviceInteractor) CloseTransaction(ctx context.Context, RollBack *bool) {
	LOG := appcontext.Logger(ctx)
	if *RollBack == true {
		DEC(LOG).Infoln("Roll Back Transaction")
		sh.Db.RollBackTransaction()
		return
	}
	LOG.Infoln("Commit Transaction")
	sh.Db.CommitTransaction()
}

//DatabaseUpgrade add routines for database upgrade
func (sh *DeviceInteractor) DatabaseUpgrade(ctx context.Context, FabricName string) error {
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Database Upgrade")
	ctx = context.WithValue(ctx, appcontext.FabricName, FabricName)
	LOG := appcontext.Logger(ctx)
	var Fabric domain.Fabric
	var err error
	RollBack := true
	//Check if Fabric Already exists
	if Fabric, err = sh.Db.GetFabric(FabricName); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s does not exist", FabricName)
		DEC(LOG).Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()
	err = sh.Db.OpenTransaction()
	if err != nil {
		return err
	}
	defer sh.CloseTransaction(ctx, &RollBack)

	//Check Non-CLOS ASN Pool
	FabricProp, err := sh.Db.GetFabricProperties(Fabric.ID)
	asnCount, err := sh.Db.GetASNCountOnRole(Fabric.ID, RackRole)
	LOG.Println("Number of ROle ASN", asnCount)
	if asnCount == 0 {
		LOG.Println("Creating RACk ASN Block")
		asnMin, asnMax := GetASNMinMax(FabricProp.RackASNBlock)
		if err := sh.PopulateASN(ctx, FabricName, asnMin, asnMax, Fabric.ID, RackRole); err != nil {
			LOG.Infoln(err)
			return err
		}
		LOG.Println("Creating MCT L3 Loopback IP Range")
		//Populate IPPair Pool with MCT L3 Loopback IP Range
		if err := sh.PopulateIPPairs(ctx, FabricName, Fabric.ID, FabricProp.MCTL3LBIPRange, domain.RackL3LoopBackPoolName, true); err != nil {
			LOG.Infoln(err)
			return err
		}
	}

	//Operation is Success, Set RollBack to False
	RollBack = false
	return nil
}

//AddFabric adds a given fabric to the application database and initializes the necessary IP Pools
func (sh *DeviceInteractor) AddFabric(ctx context.Context, FabricName string) error {
	ctx = context.WithValue(ctx, appcontext.UseCaseName, "Add Fabric")
	ctx = context.WithValue(ctx, appcontext.FabricName, FabricName)
	LOG := appcontext.Logger(ctx)

	var Fabric domain.Fabric
	var err error
	RollBack := true
	//Check if Fabric Already exists
	if Fabric, err = sh.Db.GetFabric(FabricName); err == nil {
		statusMsg := fmt.Sprintf("Fabric %s already exists", FabricName)
		DEC(LOG).Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	Fabric.Name = FabricName

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()
	err = sh.Db.OpenTransaction()
	if err != nil {
		return err
	}
	defer sh.CloseTransaction(ctx, &RollBack)

	//Create Fabric
	LOG.Infof("Create Fabric")
	if err := sh.Db.CreateFabric(&Fabric); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s create failed", FabricName)
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	//Save the Default Fabric Properties
	FabricProp := sh.getDefaultProperties()
	if err := sh.createAndUpdateFabricProperites(ctx, FabricName, Fabric.ID, FabricProp); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s update Fabric Property failed", FabricName)
		DEC(LOG).Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	//Populate ASN Block
	if err := sh.populateASNPool(ctx, FabricName, Fabric.ID, FabricProp); err != nil {
		statusMsg := fmt.Sprintf("Error populating ASN Pool for %s: %s",
			err.Error(), FabricName)
		DEC(LOG).Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	//Populate IP Pool
	if err := sh.populateIPPool(ctx, FabricName, Fabric.ID, FabricProp); err != nil {
		statusMsg := fmt.Sprintf("Error populating IP Pool for %s: %s",
			err.Error(), FabricName)
		DEC(LOG).Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	//Operation is Success, Set RollBack to False
	RollBack = false
	return nil
}

func updateFabricPropertiesDB(ctx context.Context, sh *DeviceInteractor, FabricID uint, FabricName string, NewFabricProp domain.FabricProperties, OldFabricProp domain.FabricProperties) error {
	var UpdateASNPool bool
	var UpdateIPPool bool
	var err error
	RollBack := true

	//Start Transaction
	sh.DBMutex.Lock()
	defer sh.DBMutex.Unlock()
	err = sh.Db.OpenTransaction()
	if err != nil {
		return err
	}
	defer sh.CloseTransaction(ctx, &RollBack)
	if NewFabricProp.P2PLinkRange != OldFabricProp.P2PLinkRange ||
		NewFabricProp.LoopBackIPRange != OldFabricProp.LoopBackIPRange ||
		NewFabricProp.MCTLinkIPRange != OldFabricProp.MCTLinkIPRange ||
		NewFabricProp.MCTL3LBIPRange != OldFabricProp.MCTL3LBIPRange {
		UpdateIPPool = true
		sh.Db.DeleteIPPool()
		sh.Db.DeleteIPPairPool()
		sh.Db.DeleteUsedIPPool()
		sh.Db.DeleteUsedIPPairPool()
	}
	if NewFabricProp.LeafASNBlock != OldFabricProp.LeafASNBlock ||
		NewFabricProp.SpineASNBlock != OldFabricProp.SpineASNBlock ||
		NewFabricProp.RackASNBlock != OldFabricProp.RackASNBlock {
		UpdateASNPool = true
		sh.Db.DeleteASNPool()
		sh.Db.DeleteUsedASNPool()
	}
	if err := sh.createAndUpdateFabricProperites(ctx, FabricName, FabricID, NewFabricProp); err != nil {
		statusMsg := fmt.Sprintf("Fabric %s update Fabric Property failed", FabricName)
		return errors.New(statusMsg)
	}

	//Populate ASN Block
	if UpdateASNPool {
		if err := sh.populateASNPool(ctx, FabricName, FabricID, NewFabricProp); err != nil {
			statusMsg := fmt.Sprintf("Error Populating ASN Pool: %s", err.Error())
			return errors.New(statusMsg)
		}
	}

	//Populate IP Pool
	if UpdateIPPool {
		if err := sh.populateIPPool(ctx, FabricName, FabricID, NewFabricProp); err != nil {
			statusMsg := fmt.Sprintf("Error populating IP Pool: %s for %s", err.Error(), FabricName)
			return errors.New(statusMsg)
		}
	}
	//Operation is Success, Set RollBack to False
	RollBack = false
	return nil

}

func (sh *DeviceInteractor) createAndUpdateFabricProperites(ctx context.Context,
	FabricName string, FabricID uint, FabricProp domain.FabricProperties) error {

	LOG := appcontext.Logger(ctx)

	FabricProp.FabricID = FabricID
	fProps, err := sh.Db.GetFabricProperties(FabricID)
	if err == nil {
		//Fabric Property Exists, So update
		LOG.Infof("Update fabric properties")
		FabricProp.ID = fProps.ID
		return sh.Db.UpdateFabricProperties(&FabricProp)
	}
	LOG.Infof("Create fabric properties")
	//Fabric Propety does not exist, so create
	return sh.Db.CreateFabricProperties(&FabricProp)
}

func (sh *DeviceInteractor) getDefaultProperties() domain.FabricProperties {
	var FabricProp domain.FabricProperties
	FabricProp.P2PLinkRange = "10.10.10.0/23"
	FabricProp.LoopBackIPRange = "172.31.254.0/24"
	FabricProp.MCTLinkIPRange = domain.DefaultMctLinkRange
	FabricProp.MCTL3LBIPRange = domain.DefaultMctL3LBRange
	//For now lets keep 200 later change to 4090
	FabricProp.ControlVlan = domain.DefaultControlVlan
	FabricProp.ControlVE = domain.DefaultControlVE
	FabricProp.MctPortChannel = domain.DefaultMctPortChannel
	FabricProp.RoutingMctPortChannel = domain.RoutingDefaultMctPortChannel
	FabricProp.P2PIPType = domain.P2PIpTypeNumbered
	FabricProp.LoopBackPortNumber = "1"
	FabricProp.SpineASNBlock = "64512"
	FabricProp.LeafASNBlock = "65000-65534"
	FabricProp.RackASNBlock = "4200000000-4200065534"
	FabricProp.VTEPLoopBackPortNumber = "2"
	FabricProp.AnyCastMac = "0201.0101.0101"
	FabricProp.IPV6AnyCastMac = "0201.0101.0102"
	FabricProp.ConfigureOverlayGateway = "Yes"
	FabricProp.VNIAutoMap = "Yes"
	FabricProp.BFDEnable = "No"
	FabricProp.BFDTx = "300"
	FabricProp.BFDRx = "300"
	FabricProp.BFDMultiplier = "3"
	FabricProp.BGPMultiHop = "2"
	FabricProp.MaxPaths = "8"
	FabricProp.AllowASIn = "0"
	FabricProp.MTU = "9216"   // <NUMBER:1548-9216>   MTU in bytes
	FabricProp.IPMTU = "9100" // <NUMBER:1300-9194>   IP MTU in bytes
	FabricProp.LeafPeerGroup = "spine-group"
	FabricProp.SpinePeerGroup = "leaf-group"

	//EVPN Fields
	FabricProp.ArpAgingTimeout = "300"
	FabricProp.MacAgingTimeout = "1800"
	FabricProp.MacAgingConversationalTimeout = "300"
	FabricProp.MacMoveLimit = "20"
	FabricProp.DuplicateMacTimer = "5"
	FabricProp.DuplicateMaxTimerMaxCount = "3"

	//Non Clos Fields
	FabricProp.RackPeerEBGPGroup = "underlay-ebgp-group"
	FabricProp.RackPeerOvgGroup = "overlay-ebgp-group"

	return FabricProp
}

func (sh *DeviceInteractor) populateIPPool(ctx context.Context, FabricName string,
	fabricID uint, prop domain.FabricProperties) error {
	if err := sh.PopulateIP(ctx, FabricName, fabricID, prop.LoopBackIPRange, "Loopback", true); err != nil {
		return err
	}
	//Populate IPPair Pool with MCT L3 Loopback IP Range
	if err := sh.PopulateIPPairs(ctx, FabricName, fabricID, prop.MCTL3LBIPRange, domain.RackL3LoopBackPoolName, true); err != nil {
		return err
	}
	//Populate P2P Pool only if the type is Numbered
	if prop.P2PIPType == domain.P2PIpTypeNumbered {
		//Populate P2P Pool as IPPair Pool
		if err := sh.PopulateIPPairs(ctx, FabricName, fabricID, prop.P2PLinkRange, "P2P", false); err != nil {
			return err
		}
	}
	//MCT Link IP range
	err := sh.PopulateIPPairs(ctx, FabricName, fabricID, prop.MCTLinkIPRange, domain.MCTPoolName, true)
	if err != nil {
		fmt.Printf("Error Populating MCT IP POOL %s ", err.Error())
		return err
	}
	return err
}

//GetASNMinMax returns the min and max ASN value based on the input ASNRange
func GetASNMinMax(asnRange string) (asnMin, asnMax uint64) {
	asn := strings.Split(asnRange, "-")
	asnMin, _ = strconv.ParseUint(asn[0], 10, 64)
	asnMax = asnMin
	if len(asn) == 2 {
		asnMax, _ = strconv.ParseUint(asn[1], 10, 64)
	}
	return
}

func (sh *DeviceInteractor) populateASNPool(ctx context.Context, FabricName string, fabricID uint, prop domain.FabricProperties) error {
	var asnMin, asnMax uint64
	LOG := appcontext.Logger(ctx)
	asnMin, asnMax = GetASNMinMax(prop.LeafASNBlock)
	if err := sh.PopulateASN(ctx, FabricName, asnMin, asnMax, fabricID, "Leaf"); err != nil {
		LOG.Infoln(err)
		return err
	}
	asnMin, asnMax = GetASNMinMax(prop.SpineASNBlock)
	if err := sh.PopulateASN(ctx, FabricName, asnMin, asnMax, fabricID, "Spine"); err != nil {
		LOG.Infoln(err)
		return err
	}

	asnMin, asnMax = GetASNMinMax(prop.RackASNBlock)
	if err := sh.PopulateASN(ctx, FabricName, asnMin, asnMax, fabricID, "Rack"); err != nil {
		LOG.Infoln(err)
		return err
	}
	return nil
}

func modifyUpdatedField(d *domain.FabricProperties, s *domain.FabricProperties) {
	if len(s.P2PLinkRange) != 0 && s.P2PLinkRange != d.P2PLinkRange {
		d.P2PLinkRange = s.P2PLinkRange
	}
	if len(s.P2PIPType) != 0 && s.P2PIPType != d.P2PIPType {
		d.P2PIPType = s.P2PIPType
	}
	if len(s.LoopBackIPRange) != 0 && s.LoopBackIPRange != d.LoopBackIPRange {
		d.LoopBackIPRange = s.LoopBackIPRange
	}
	if len(s.MCTLinkIPRange) != 0 && s.MCTLinkIPRange != d.MCTLinkIPRange {
		d.MCTLinkIPRange = s.MCTLinkIPRange
	}
	if len(s.MCTL3LBIPRange) != 0 && s.MCTL3LBIPRange != d.MCTL3LBIPRange {
		d.MCTL3LBIPRange = s.MCTL3LBIPRange
	}
	if len(s.LoopBackPortNumber) != 0 && s.LoopBackPortNumber != d.LoopBackPortNumber {
		d.LoopBackPortNumber = s.LoopBackPortNumber
	}
	if len(s.SpineASNBlock) != 0 && s.SpineASNBlock != d.SpineASNBlock {
		d.SpineASNBlock = s.SpineASNBlock
	}
	if len(s.LeafASNBlock) != 0 && s.LeafASNBlock != d.LeafASNBlock {
		d.LeafASNBlock = s.LeafASNBlock
	}
	if len(s.RackASNBlock) != 0 && s.RackASNBlock != d.RackASNBlock {
		d.RackASNBlock = s.RackASNBlock
	}
	if len(s.VTEPLoopBackPortNumber) != 0 && s.VTEPLoopBackPortNumber != d.VTEPLoopBackPortNumber {
		d.VTEPLoopBackPortNumber = s.VTEPLoopBackPortNumber
	}
	if len(s.AnyCastMac) != 0 && s.AnyCastMac != d.AnyCastMac {
		d.AnyCastMac = s.AnyCastMac
	}
	if len(s.IPV6AnyCastMac) != 0 && s.IPV6AnyCastMac != d.IPV6AnyCastMac {
		d.IPV6AnyCastMac = s.IPV6AnyCastMac
	}
	if len(s.ConfigureOverlayGateway) != 0 && s.ConfigureOverlayGateway != d.ConfigureOverlayGateway {
		d.ConfigureOverlayGateway = s.ConfigureOverlayGateway
	}
	if len(s.VNIAutoMap) != 0 && s.VNIAutoMap != d.VNIAutoMap {
		d.VNIAutoMap = s.VNIAutoMap
	}
	if len(s.BFDEnable) != 0 && s.BFDEnable != d.BFDEnable {
		d.BFDEnable = s.BFDEnable
	}
	if len(s.BFDTx) != 0 && s.BFDTx != d.BFDTx {
		d.BFDTx = s.BFDTx
	}
	if len(s.BFDRx) != 0 && s.BFDRx != d.BFDRx {
		d.BFDRx = s.BFDRx
	}
	if len(s.BFDMultiplier) != 0 && s.BFDMultiplier != d.BFDMultiplier {
		d.BFDMultiplier = s.BFDMultiplier
	}
	if len(s.BGPMultiHop) != 0 && s.BGPMultiHop != d.BGPMultiHop {
		d.BGPMultiHop = s.BGPMultiHop
	}
	if len(s.MaxPaths) != 0 && s.MaxPaths != d.MaxPaths {
		d.MaxPaths = s.MaxPaths
	}
	if len(s.AllowASIn) != 0 && s.AllowASIn != d.AllowASIn {
		d.AllowASIn = s.AllowASIn
	}
	if len(s.MTU) != 0 && s.MTU != d.MTU {
		d.MTU = s.MTU
	}
	if len(s.IPMTU) != 0 && s.IPMTU != d.IPMTU {
		d.IPMTU = s.IPMTU
	}
	if len(s.LeafPeerGroup) != 0 && s.LeafPeerGroup != d.LeafPeerGroup {
		d.LeafPeerGroup = s.LeafPeerGroup
	}
	if len(s.SpinePeerGroup) != 0 && s.SpinePeerGroup != d.SpinePeerGroup {
		d.SpinePeerGroup = s.SpinePeerGroup
	}
	//
	if len(s.ArpAgingTimeout) != 0 && s.ArpAgingTimeout != d.ArpAgingTimeout {
		d.ArpAgingTimeout = s.ArpAgingTimeout
	}
	if len(s.MacAgingTimeout) != 0 && s.MacAgingTimeout != d.MacAgingTimeout {
		d.MacAgingTimeout = s.MacAgingTimeout
	}
	if len(s.MacAgingConversationalTimeout) != 0 && s.MacAgingConversationalTimeout !=
		d.MacAgingConversationalTimeout {
		d.MacAgingConversationalTimeout = s.MacAgingConversationalTimeout
	}
	if len(s.MacMoveLimit) != 0 && s.MacMoveLimit != d.MacMoveLimit {
		d.MacMoveLimit = s.MacMoveLimit
	}
	if len(s.DuplicateMacTimer) != 0 && s.DuplicateMacTimer != d.DuplicateMacTimer {
		d.DuplicateMacTimer = s.DuplicateMacTimer
	}
	if len(s.DuplicateMaxTimerMaxCount) != 0 && s.DuplicateMaxTimerMaxCount != d.DuplicateMaxTimerMaxCount {
		d.DuplicateMaxTimerMaxCount = s.DuplicateMaxTimerMaxCount
	}
	if len(s.ControlVlan) != 0 && s.ControlVlan != d.ControlVlan {
		d.ControlVlan = s.ControlVlan
	}
	if len(s.ControlVE) != 0 && s.ControlVE != d.ControlVE {
		d.ControlVE = s.ControlVE
	}
	if len(s.MctPortChannel) != 0 && s.MctPortChannel != d.MctPortChannel {
		d.MctPortChannel = s.MctPortChannel
	}
	if len(s.RoutingMctPortChannel) != 0 && s.RoutingMctPortChannel != d.RoutingMctPortChannel {
		d.RoutingMctPortChannel = s.RoutingMctPortChannel
	}
	if len(s.FabricType) != 0 && s.FabricType != d.FabricType {
		d.FabricType = s.FabricType
	}
	if len(s.RackPeerEBGPGroup) != 0 && s.RackPeerEBGPGroup != d.RackPeerEBGPGroup {
		d.RackPeerEBGPGroup = s.RackPeerEBGPGroup
	}
	if len(s.RackPeerOvgGroup) != 0 && s.RackPeerOvgGroup != d.RackPeerOvgGroup {
		d.RackPeerOvgGroup = s.RackPeerOvgGroup
	}
}

func (sh *DeviceInteractor) validateOverlappingASNRange(ctx context.Context, prop *domain.FabricProperties) error {
	//Already ASN Blocks are sanitized
	LOG := appcontext.Logger(ctx)
	leafASNMin, leafASNMax := GetASNMinMax(prop.LeafASNBlock)
	spineASNMin, spineASNMax := GetASNMinMax(prop.SpineASNBlock)
	if leafASNMin <= spineASNMax && spineASNMin <= leafASNMax {
		ret := fmt.Sprintf("Leaf ASN - %s and Spine ASN - %s  Ranges OverLap", prop.LeafASNBlock, prop.SpineASNBlock)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	return nil
}

func (sh *DeviceInteractor) intersectIP(ip1, ip2 string) bool {
	//Already IP address are sanitized
	_, n1, _ := net.ParseCIDR(ip1)
	_, n2, _ := net.ParseCIDR(ip2)
	return n2.Contains(n1.IP) || n1.Contains(n2.IP)
}

func (sh *DeviceInteractor) validateOverlappingIPRange(ctx context.Context, prop *domain.FabricProperties) error {
	LOG := appcontext.Logger(ctx)
	if sh.intersectIP(prop.P2PLinkRange, prop.LoopBackIPRange) {
		ret := fmt.Sprintf("P2pLinkRange %s and LoopBackRange %s Overlap with each other", prop.P2PLinkRange, prop.LoopBackIPRange)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	if sh.intersectIP(prop.P2PLinkRange, prop.MCTLinkIPRange) {
		ret := fmt.Sprintf("P2pLinkRange %s and MctLinkRange %s Overlap with each other", prop.P2PLinkRange, prop.MCTLinkIPRange)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	if sh.intersectIP(prop.P2PLinkRange, prop.MCTL3LBIPRange) {
		ret := fmt.Sprintf("P2pLinkRange %s and L3BackupIpRange %s Overlap with each other", prop.P2PLinkRange, prop.MCTL3LBIPRange)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	if sh.intersectIP(prop.LoopBackIPRange, prop.MCTLinkIPRange) {
		ret := fmt.Sprintf("LoopBackRange %s and MctLinkRange %s Overlap with each other", prop.LoopBackIPRange, prop.MCTLinkIPRange)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	if sh.intersectIP(prop.LoopBackIPRange, prop.MCTL3LBIPRange) {
		ret := fmt.Sprintf("LoopBackRange %s and L3BackupIpRange %s Overlap with each other", prop.LoopBackIPRange, prop.MCTL3LBIPRange)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	if sh.intersectIP(prop.MCTLinkIPRange, prop.MCTL3LBIPRange) {
		ret := fmt.Sprintf("MctLinkRange %s and L3BackupIpRange %s Overlap with each other", prop.MCTLinkIPRange, prop.MCTL3LBIPRange)
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	return nil
}

func (sh *DeviceInteractor) validateFabricProperties(ctx context.Context, prop *domain.FabricProperties) error {
	// This Function Validates overlapping IP Ranges and ASN Blocks
	if err := sh.validateOverlappingASNRange(ctx, prop); err != nil {
		return err
	}

	if err := sh.validateOverlappingIPRange(ctx, prop); err != nil {
		return err
	}
	if prop.LoopBackPortNumber == prop.VTEPLoopBackPortNumber {
		return errors.New("VTEP Loopback Number and Loopback Number cannot be same")
	}
	if err := sh.validateAnyCastMac(ctx, prop); err != nil {
		return err
	}
	return nil
}

func (sh *DeviceInteractor) validateAnyCastMac(ctx context.Context, prop *domain.FabricProperties) error {
	//Already ASN Blocks are sanitized
	LOG := appcontext.Logger(ctx)
	if prop.AnyCastMac == prop.IPV6AnyCastMac {
		ret := fmt.Sprintf("Anycast Mac and IPV6 Anycast Mac should have distinct values.")
		LOG.Errorln(ret)
		return errors.New(ret)
	}
	return nil
}

//UpdateFabricProperties updates the fabric settings in application database
func (sh *DeviceInteractor) UpdateFabricProperties(ctx context.Context, FabricName string, FabricUpdateRequest *domain.FabricProperties) (string, uint, error) {
	var ret string
	var err error
	var Fabric domain.Fabric
	var FabricProperties domain.FabricProperties
	if Fabric, err = sh.Db.GetFabric(FabricName); err == nil {
		if deviceCount := sh.Db.GetDevicesCountInFabric(Fabric.ID); deviceCount != 0 {
			ret = fmt.Sprintf("%s: fabric is already active and cannot be updated\n", FabricName)
			return ret, sh.FabricID, domain.ErrFabricActive
		}
		if FabricProperties, err = sh.Db.GetFabricProperties(Fabric.ID); err == nil {
			oldFabricProperties := FabricProperties
			//Update only requested Fields
			modifyUpdatedField(&FabricProperties, FabricUpdateRequest)
			if FabricProperties == oldFabricProperties {
				ret = fmt.Sprintf("No Fabric Property Update is Requested For Fabric  %s\n", FabricName)
				return ret, sh.FabricID, domain.ErrFabricIncorrectValues
			}
			if err = sh.validateFabricProperties(ctx, &FabricProperties); err != nil {
				return err.Error(), sh.FabricID, domain.ErrFabricIncorrectValues
			}
			updateFabricPropertiesDB(ctx, sh, Fabric.ID, FabricName, FabricProperties, oldFabricProperties)
			ret = "Fabric Properties updated"
		} else {
			ret = fmt.Sprintf("Unable to retrieve Fabric Properties for %s\n", FabricName)
			return ret, sh.FabricID, domain.ErrFabricInternalError
		}
	} else {
		ret = fmt.Sprintf("Unable to retrieve Fabric %s\n", FabricName)
		return ret, sh.FabricID, domain.ErrFabricNotFound
	}

	return ret, Fabric.ID, nil
}

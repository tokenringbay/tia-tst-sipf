package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"errors"
	"fmt"
	"net"

	"strings"
)

func (sh *DeviceInteractor) skipInvalidIPPairs(ip net.IP) bool {
	var ret = false

	if strings.HasSuffix(ip.String(), ".0") {
		ret = true
	}
	if strings.HasSuffix(ip.String(), ".1") {
		ret = true
	}

	if strings.HasSuffix(ip.String(), ".254") {
		ret = true
	}
	if strings.HasSuffix(ip.String(), ".255") {
		ret = true
	}

	return ret
}

//PopulateIPPairs populates the IPPairPool withe IP address from the network string
func (sh *DeviceInteractor) PopulateIPPairs(ctx context.Context, FabricName string, fabricID uint, network string, IPType string, skip bool) error {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Populate IP Pool(%s) of type %s", network, IPType)

	ip, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		statusMsg := fmt.Sprintf("IP Pool Range %s provided is invalid for fabric %s type %s:%s", network,
			FabricName, IPType, err.Error())
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}
	//Prepare a list of IP Address
	IPAddressList := make([]string, 0)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); sh.inc(ip) {
		if skip {
			if sh.skipInvalidIPPairs(ip) {
				LOG.Debugf("Skiping IP %s for fabric %s type %s", ip.String(), FabricName, IPType)
				continue
			}
		}
		IPAddressList = append(IPAddressList, ip.String())
	}
	//Populate Pooltable with IPAddress list
	length := len(IPAddressList) / 2
	for i := 0; i < length; i++ {
		a := 2 * i
		b := a + 1
		IPEntry := domain.IPPairAllocationPool{FabricID: fabricID,
			IPAddressOne: IPAddressList[a], IPAddressTwo: IPAddressList[b], IPType: IPType}
		err := sh.Db.CreateIPPairEntry(&IPEntry)
		if err != nil {
			statusMsg := fmt.Sprintf("IP Pair Pool %s initialization failed for fabric %s type %s:%s", network,
				FabricName, IPType, err.Error())
			LOG.Errorln(statusMsg)
			return errors.New(statusMsg)
		}
	}

	return nil
}

//GetAlreadyAllocatedIPPair returns the Pair of IP Address allocated to the interfaces specified by the interfaceID'ss
func (sh *DeviceInteractor) GetAlreadyAllocatedIPPair(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint,
	IPType string, InterfaceOneID uint, InterfaceTwoID uint) (string, string, error) {
	//Get the Next Available IP from the Pool

	//check if its already be assigned to this Interface
	usedIPPair, err := sh.Db.GetUsedIPPairOnDeviceInterfaceIDAndType(FabricID, DeviceOneID, DeviceTwoID, IPType, InterfaceOneID, InterfaceTwoID)
	if err == nil {
		//Return the ip address in the order it was queried
		if InterfaceOneID == usedIPPair.InterfaceOneID {
			return usedIPPair.IPAddressOne, usedIPPair.IPAddressTwo, nil
		}
		return usedIPPair.IPAddressTwo, usedIPPair.IPAddressOne, nil
	}
	return "", "", err
}

//GetIPPair gets the next Pair of IP Address from the Pool and move to UsedIPPair Pool
func (sh *DeviceInteractor) GetIPPair(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPType string,
	InterfaceOneID uint, InterfaceTwoID uint) (string, string, error) {
	//Get the Next Available IP from the Pool
	LOG := appcontext.Logger(ctx)

	ipDB, err := sh.getNextIPPairFromPool(ctx, FabricID, IPType)
	if err != nil {
		LOG.Println(err)
		return "", "", err
	}

	sh.moveFromIPPairPoolToUsedIPPair(ctx, FabricID, DeviceOneID, DeviceTwoID, IPType, InterfaceOneID, InterfaceTwoID, ipDB)
	return ipDB.IPAddressOne, ipDB.IPAddressTwo, nil
}

//ReleaseIPPair releases a pair of IP Address back to the Pool
func (sh *DeviceInteractor) ReleaseIPPair(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint,
	IPType string, ipaddressOne string, ipaddressTwo string, InterfaceOneID uint, InterfaceTwoID uint) error {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Release IP %s %s for IPType %s", ipaddressOne, ipaddressTwo, IPType)
	//Remove the IP from Used IP
	if err := sh.removeUsedIPPairEntry(ctx, FabricID, DeviceOneID, DeviceTwoID, ipaddressOne, ipaddressTwo, IPType, InterfaceOneID, InterfaceTwoID); err != nil {
		return err
	}

	sh.returnIPPairToPool(ctx, FabricID, IPType, ipaddressOne, ipaddressTwo)

	return nil
}

//ReserveIPPair reserves a pair of IP Address for a set of interfaces
func (sh *DeviceInteractor) ReserveIPPair(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPType string,
	ipaddressOne string, ipaddressTwo string, InterfaceOneID uint, InterfaceTwoID uint) error {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Reserve IP (%s %s) for %d IPType %s", ipaddressOne, ipaddressTwo, FabricID, IPType)
	var ipDB domain.IPPairAllocationPool
	var ipCount int64

	//if Requested IP in the Pool,then move to Used IP Pair
	if ipCount, ipDB = sh.GetIPPairCountInPool(ctx, FabricID, ipaddressOne, ipaddressTwo, IPType); ipCount == 1 {
		//If the Requested IP is in the available and there are interfaces are UsedIP then release the existing IP Address
		eIP1, eIP2, eerr := sh.GetAlreadyAllocatedIPPair(ctx, FabricID, DeviceOneID, DeviceTwoID, IPType, InterfaceOneID, InterfaceTwoID)
		if eerr == nil {
			sh.ReleaseIPPair(ctx, FabricID, DeviceOneID, DeviceTwoID, IPType, eIP1, eIP2, InterfaceOneID, InterfaceTwoID)
		}
		LOG.Infof("IP available in  IP Pair Pool %s %s for IPType %s", ipaddressOne, ipaddressTwo, IPType)
		sh.moveReservedFromIPPairPoolToUsedIPPair(ctx, FabricID, DeviceOneID, DeviceTwoID, IPType, ipaddressOne, ipaddressTwo, InterfaceOneID, InterfaceTwoID, ipDB)
		return nil
	}

	//check if its already be assigned to this Interface pair
	_, err := sh.Db.GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType(FabricID, DeviceOneID, DeviceTwoID, ipaddressOne, ipaddressTwo, IPType, InterfaceOneID, InterfaceTwoID)
	if err != nil {
		statusMsg := fmt.Sprintf("IPPair(%s,%s) not present in the Used IP Table for Fabric %d Device (%d,%d)", ipaddressOne, ipaddressTwo, FabricID, DeviceOneID, DeviceTwoID)
		return errors.New(statusMsg)
	}
	return nil
}

//GetIPPairCountInPool returns the count of entries that are in the pool
//Count of Zero indicates that the entry is not in the Pool
func (sh *DeviceInteractor) GetIPPairCountInPool(ctx context.Context, FabricID uint, ipaddressOne string, ipaddressTwo string, IPType string) (int64, domain.IPPairAllocationPool) {
	LOG := appcontext.Logger(ctx)

	ipCount, IPEntry, err := sh.Db.GetIPPairEntryAndCountOnIPAddressAndType(FabricID, ipaddressOne, ipaddressTwo, IPType)

	if err != nil {
		LOG.Infoln("No Entry")
	}
	LOG.Infof("IP Pool: Count of IP (%s %s) for IPType %s %d", ipaddressOne, ipaddressTwo, IPType, ipCount)
	return ipCount, IPEntry
}

func (sh *DeviceInteractor) removeUsedIPPairEntry(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint, ipaddressOne string,
	ipaddressTwo string, IPType string, InterfaceOneID uint, InterfaceTwoID uint) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Remove IP from Used IP Table (%s %s) for IPType %s", ipaddressOne, ipaddressTwo, IPType)
	usedIP, err := sh.Db.GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType(FabricID, DeviceOneID, DeviceTwoID, ipaddressOne, ipaddressTwo, IPType, InterfaceOneID, InterfaceTwoID)
	if err != nil {
		statusMsg := fmt.Sprintf("IP %s %s not present in the Used IP PAIR Table for Fabric %d Device %d", ipaddressOne, ipaddressTwo, FabricID, DeviceOneID)
		LOG.Infoln(statusMsg)
		return errors.New(statusMsg)
	}
	return sh.Db.DeleteUsedIPPairEntry(&usedIP)
}

func (sh *DeviceInteractor) returnIPPairToPool(ctx context.Context, FabricID uint, IPType string, ipaddressOne string, ipaddressTwo string) {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Add IP back in Pool  for IPType %s IP (%s %s)", IPType, ipaddressOne, ipaddressTwo)
	IPEntry := domain.IPPairAllocationPool{FabricID: FabricID,
		IPAddressOne: ipaddressOne, IPAddressTwo: ipaddressTwo, IPType: IPType}

	err := sh.Db.CreateIPPairEntry(&IPEntry)
	if err != nil {
		LOG.Infoln(err)
	}
}

func (sh *DeviceInteractor) getNextIPPairFromPool(ctx context.Context, FabricID uint, IPType string) (domain.IPPairAllocationPool, error) {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Fetch next IP from Pool for IPType %s", IPType)
	var IPEntry domain.IPPairAllocationPool
	IPEntry, err := sh.Db.GetNextIPPairEntryOnType(FabricID, IPType)
	if err != nil {
		statusMsg := fmt.Sprintf("IP Exhausted for %d", FabricID)
		LOG.Println(statusMsg)
		return IPEntry, errors.New(statusMsg)
	}
	LOG.Infof("Value of next IP from Pool for IPType %s IP (%s %s)", IPType, IPEntry.IPAddressOne, IPEntry.IPAddressTwo)
	return IPEntry, nil
}

func (sh *DeviceInteractor) moveFromIPPairPoolToUsedIPPair(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint, IPType string, InterfaceOneID uint, InterfaceTwoID uint,
	IPEntry domain.IPPairAllocationPool) {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Delete  IP from Allocation Table for IPType %s", IPEntry.IPType)
	if err := sh.Db.DeleteIPPairEntry(&IPEntry); err != nil {
		LOG.Println(err)
	}

	//Make Allocation Entry for the IP in Used IP Table
	LOG.Infof("Make IP entry in Used IP Table for IPType %s ip (%s %s)",
		IPType, IPEntry.IPAddressOne, IPEntry.IPAddressTwo)
	usedEntry := domain.UsedIPPair{FabricID: FabricID, DeviceOneID: DeviceOneID, DeviceTwoID: DeviceTwoID, IPAddressOne: IPEntry.IPAddressOne, IPAddressTwo: IPEntry.IPAddressTwo, IPType: IPType,
		InterfaceOneID: InterfaceOneID, InterfaceTwoID: InterfaceTwoID}
	sh.Db.CreateUsedIPPairEntry(&usedEntry)
}

func (sh *DeviceInteractor) moveReservedFromIPPairPoolToUsedIPPair(ctx context.Context, FabricID uint, DeviceOneID uint, DeviceTwoID uint,
	IPType string, IPAddressOne string, IPAddressTwo string, InterfaceOneID uint, InterfaceTwoID uint, IPEntry domain.IPPairAllocationPool) {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Delete IP from allocation table for IPType %s", IPEntry.IPType)
	if err := sh.Db.DeleteIPPairEntry(&IPEntry); err != nil {
		LOG.Println(err)
	}

	//Make Allocation Entry for the IP in Used IP Table
	LOG.Infof("Make IP entry in Used IP Table for IPType %s ip (%s %s)",
		IPType, IPEntry.IPAddressOne, IPEntry.IPAddressTwo)
	usedEntry := domain.UsedIPPair{FabricID: FabricID, DeviceOneID: DeviceOneID, DeviceTwoID: DeviceTwoID, IPAddressOne: IPAddressOne, IPAddressTwo: IPAddressTwo, IPType: IPType,
		InterfaceOneID: InterfaceOneID, InterfaceTwoID: InterfaceTwoID}
	sh.Db.CreateUsedIPPairEntry(&usedEntry)
}

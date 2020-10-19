package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"net"
	"strings"
)

//PopulateIP populates IPAllocationPool instances
func (sh *DeviceInteractor) PopulateIP(ctx context.Context, FabricName string,
	fabricID uint, network string, IPType string, skip bool) error {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Populate IP Pool(%s) of type %s", network, IPType)
	ip, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		statusMsg := fmt.Sprintf("IP Pool Range %s provided is invalid for fabric %s type %s:%s", network,
			FabricName, IPType, err.Error())
		LOG.Errorln(statusMsg)
		return errors.New(statusMsg)
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); sh.inc(ip) {
		if skip && strings.HasSuffix(ip.String(), ".0") {
			continue
		}
		if skip && strings.HasSuffix(ip.String(), ".254") {
			continue
		}
		if skip && strings.HasSuffix(ip.String(), ".255") {
			continue
		}
		if skip && IPType == "MCTLink" && strings.HasSuffix(ip.String(), ".1") {
			continue
		}
		IPEntry := domain.IPAllocationPool{FabricID: fabricID,
			IPAddress: ip.String(), IPType: IPType}
		err := sh.Db.CreateIPEntry(&IPEntry)
		if err != nil {
			LOG.Println(err)
			statusMsg := fmt.Sprintf("IP Pool Initialization Failed for type %s", IPType)
			LOG.Infoln(statusMsg)
			return errors.New(statusMsg)
		}
	}

	return nil
}

func (sh *DeviceInteractor) inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

//GetAlreadyAllocatedIP gets already allocated IP for a given interface, IP type
func (sh *DeviceInteractor) GetAlreadyAllocatedIP(ctx context.Context, FabricID uint,
	DeviceID uint, IPType string, InterfaceID uint) (string, error) {

	//check if its already be assigned to this Interface
	usedIP, err := sh.Db.GetUsedIPOnDeviceInterfaceIDAndType(FabricID, DeviceID, IPType, InterfaceID)
	if err == nil {
		return usedIP.IPAddress, nil
	}
	return "IP address not already allocated to this interface", err
}

//GetIP returns an IP address from the IPAllocationPool and moves the same to IPUsedPool
func (sh *DeviceInteractor) GetIP(ctx context.Context, FabricID uint, DeviceID uint, IPType string, InterfaceID uint) (string, error) {
	//Get the Next Available IP from the Pool
	LOG := appcontext.Logger(ctx)

	ipDB, err := sh.getNextIPFromPool(ctx, FabricID, IPType)
	if err != nil {
		DEC(LOG).Println(err)
		return "", err
	}
	sh.moveFromIPPoolToUsedIP(ctx, FabricID, DeviceID, IPType, InterfaceID, ipDB)
	return ipDB.IPAddress, nil
}

//ReleaseIP releases the IP address from IPUsedPool to IPAllocationPool
func (sh *DeviceInteractor) ReleaseIP(ctx context.Context, FabricID uint, DeviceID uint, IPType string, ipaddress string, InterfaceID uint) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Release IP %s for IPType %s", ipaddress, IPType)
	//Remove the IP from Used IP
	if err := sh.removeUsedIPEntry(ctx, FabricID, DeviceID, ipaddress, IPType, InterfaceID); err != nil {
		if err == gorm.ErrRecordNotFound {
			LOG.Infof("IP %s %s not present in Used Table for Device ID %d", IPType, ipaddress, DeviceID)
			return nil
		}
		return err
	}

	sh.returnIPToPool(ctx, FabricID, IPType, ipaddress)

	return nil
}

//ReserveIP reserves the IP address for an interface
func (sh *DeviceInteractor) ReserveIP(ctx context.Context, FabricID uint, DeviceID uint, IPType string,
	ipaddress string, InterfaceID uint) error {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Reserve IP %s for IPType  %s", ipaddress, IPType)
	var ipDB domain.IPAllocationPool
	var ipCount int64

	//if Requested IP in the Pool,then move to Used IP
	if ipCount, ipDB = sh.GetIPCountInPool(ctx, FabricID, ipaddress, IPType); ipCount == 1 {
		DEC(LOG).Infof("IP available in  IP Pool %s for IPType  %s", ipaddress, IPType)
		sh.moveFromIPPoolToUsedIP(ctx, FabricID, DeviceID, IPType, InterfaceID, ipDB)
		return nil
	}
	//check if its already be assigned to this Interface
	_, err := sh.Db.GetUsedIPOnDeviceInterfaceIDIPAddresssAndType(FabricID, DeviceID, ipaddress, IPType, InterfaceID)
	if err != nil {
		statusMsg := fmt.Sprintf("IP %s not present in the Used IP Table for IPType  %s ", ipaddress, IPType)
		return errors.New(statusMsg)
	}
	return nil
}

//GetIPCountInPool gets the count of IP from the pool, for a given IP address and IPType
func (sh *DeviceInteractor) GetIPCountInPool(ctx context.Context, FabricID uint, ipaddress string, IPType string) (int64, domain.IPAllocationPool) {
	LOG := appcontext.Logger(ctx)
	ipCount, IPEntry, err := sh.Db.GetIPEntryAndCountOnIPAddressAndType(FabricID, ipaddress, IPType)

	if err != nil {
		LOG.Infoln("No Entry")
	}
	LOG.Infof("IP Pool: Count of  IP %s for IPType  %s %d", ipaddress, IPType, ipCount)
	return ipCount, IPEntry
}

func (sh *DeviceInteractor) removeUsedIPEntry(ctx context.Context, FabricID uint, DeviceID uint, ipaddress string, IPType string, InterfaceID uint) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Remove IP %s from Used IP Table for IPType %s", ipaddress, IPType)
	usedIP, err := sh.Db.GetUsedIPOnDeviceInterfaceIDIPAddresssAndType(FabricID, DeviceID, ipaddress, IPType, InterfaceID)
	if err != nil {
		statusMsg := fmt.Sprintf("IP %s not present in the Used IP Table for Fabric %d Device %d", ipaddress, FabricID, DeviceID)
		LOG.Infoln(statusMsg)
		//TODO Need to Handle the ERROR gracefully
		//This Done since Same LOOP Back IP can not be assigned to two different Device
		//In MCT that is the requirement
		return nil
		//return errors.New(statusMsg)
	}

	return sh.Db.DeleteUsedIPEntry(&usedIP)
}

func (sh *DeviceInteractor) returnIPToPool(ctx context.Context, FabricID uint, IPType string, ipAddress string) {
	LOG := appcontext.Logger(ctx)

	LOG.Infof("Add IP %s back in Pool for IPType %s", ipAddress, IPType)
	IPEntry := domain.IPAllocationPool{FabricID: FabricID,
		IPAddress: ipAddress, IPType: IPType}

	err := sh.Db.CreateIPEntry(&IPEntry)
	if err != nil {
		LOG.Infoln(err)
	}

}

func (sh *DeviceInteractor) getNextIPFromPool(ctx context.Context, FabricID uint, IPType string) (domain.IPAllocationPool, error) {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Fetch next IP from Pool for  IPType %s", IPType)
	var IPEntry domain.IPAllocationPool
	IPEntry, err := sh.Db.GetNextIPEntryOnType(FabricID, IPType)
	if err != nil {
		statusMsg := fmt.Sprintf("IP Exhausted for %d", FabricID)
		LOG.Println(statusMsg)
		return IPEntry, errors.New(statusMsg)
	}
	LOG.Infof("Value of next IP from Pool for IPType %s IP %s", IPType, IPEntry.IPAddress)
	return IPEntry, nil
}

func (sh *DeviceInteractor) moveFromIPPoolToUsedIP(ctx context.Context, FabricID uint, DeviceID uint, IPType string, InterfaceID uint,
	IPEntry domain.IPAllocationPool) {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Delete IP from allocation table IPType %s", IPEntry.IPType)
	if err := sh.Db.DeleteIPEntry(&IPEntry); err != nil {
		LOG.Println(err)
	}

	//Make allocation Entry for the IP in Used IP Table
	LOG.Infof("Make IP %s entry in Used IP Table IPType %s", IPEntry.IPAddress, IPType)
	usedEntry := domain.UsedIP{FabricID: FabricID, DeviceID: DeviceID, IPAddress: IPEntry.IPAddress, IPType: IPType,
		InterfaceID: InterfaceID}
	sh.Db.CreateUsedIPEntry(&usedEntry)
}

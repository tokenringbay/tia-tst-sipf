package usecase

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	"errors"
	"fmt"
	nLOG "github.com/sirupsen/logrus"
)

func init() {

}

//PopulateASN creates the ASNAllocationPool for a given fabric, device role
func (sh *DeviceInteractor) PopulateASN(ctx context.Context, FabricName string, ASNMin uint64, ASNMax uint64, fabricID uint, role string) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Populate ASN Pool(%d-%d) for role %s", ASNMin, ASNMax, role)
	for i := ASNMin; i <= ASNMax; i++ {
		asn := domain.ASNAllocationPool{FabricID: fabricID,
			ASN: i, DeviceRole: role}
		if err := sh.Db.CreateASN(&asn); err != nil {
			DEC(LOG).Println(err)
			statusMsg := fmt.Sprintf("ASN Pool Initialization Failed for Role %s", role)
			DEC(LOG).Infoln(statusMsg)
			return errors.New(statusMsg)
		}
	}
	return nil
}

//GetASN returns an ASN from the ASNAllocationPool and moves the returned ASN from ASNAllocationPool to ASNUsedPool
func (sh *DeviceInteractor) GetASN(ctx context.Context, FabricID uint, DeviceID uint, role string) (uint64, error) {
	LOG := appcontext.Logger(ctx)

	var asn domain.ASNAllocationPool
	var err error
	//Get the Next Available ASN from the Pool
	if asn, err = sh.getNextASNFromPool(LOG, FabricID, role); err != nil {
		DEC(LOG).Errorln(err)
		return 0, err
	}
	sh.moveFromASNPoolToUsedASN(LOG, FabricID, DeviceID, role, asn)
	return asn.ASN, nil
}

//ReleaseASN releases the input ASN from ASNUsedPool to the ASNAllocationPool
func (sh *DeviceInteractor) ReleaseASN(ctx context.Context, FabricID uint, DeviceID uint, role string, asn uint64) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Release ASN %d for role  %s", asn, role)
	//Remove the ASN from Used ASN
	if err := sh.removeAllocationEntry(LOG, FabricID, DeviceID, asn, role); err != nil {
		if err != nil {
			LOG.Infof("Used ASN %d not found on used Table for Device %d", asn, DeviceID)
			//TODO Need to Handle the ERROR gracefully
			//This Done since Same ASN  can not be assigned to two different Device
			//In MCT that is the requirement
			return nil
		}
		return err
	}
	//If there are no more entires for this ASN across Device in Used ASN
	if sh.getASNCountInUsedASN(LOG, FabricID, asn, role) == 0 {
		sh.returnASNToPool(LOG, FabricID, role, asn)
	}
	return nil
}

//ReserveASN reserves an ASN to the device,
// either by moving the ASN from ASNAllocationPool to ASNUsedPool
// or by marking the ASN against a device in ASNUsedPool
func (sh *DeviceInteractor) ReserveASN(ctx context.Context, FabricID uint, DeviceID uint, role string, asn uint64) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infof("Reserve ASN %d for role  %s", asn, role)
	var asndb domain.ASNAllocationPool
	var asnCount int64

	//if Requested ASN in the Pool,then move to Used ASN
	if asnCount, asndb = sh.getASNCountInPool(LOG, FabricID, asn, role); asnCount == 1 {
		LOG.Infof("ASN available in  ASN Pool %d for role  %s", asn, role)
		sh.moveFromASNPoolToUsedASN(LOG, FabricID, DeviceID, role, asndb)
		return nil
	}

	//If Requested ASN in the Used ASN table,make entry for device
	if sh.getASNCountInUsedASN(LOG, FabricID, asn, role) >= 1 {
		DEC(LOG).Infof("ASN available in  Used ASN Pool %d for role  %s", asn, role)
		sh.makeUsedASNEntryForDevice(LOG, FabricID, DeviceID, asn, role)
		return nil
	}
	return nil
}

func (sh *DeviceInteractor) moveFromASNPoolToUsedASN(LOG *nLOG.Entry, FabricID uint, DeviceID uint, role string, asn domain.ASNAllocationPool) {
	//Get the Count of the ASN in the Pool for that role
	asnCount, _ := sh.getASNCountForRole(LOG, FabricID, role)

	//Remove from the Pool only if there are more than ASN's of the role
	if asnCount > 1 {
		sh.deleteASN(LOG, &asn)
	}
	//Make Allocation Entry for the ASN in Used ASN Table
	sh.makeUsedASNEntryForDevice(LOG, FabricID, DeviceID, asn.ASN, role)
}

func (sh *DeviceInteractor) getASNCountInPool(LOG *nLOG.Entry, FabricID uint, asn uint64, role string) (int64, domain.ASNAllocationPool) {

	asnCount, asndb := sh.Db.GetASNAndCountOnASNAndRole(FabricID, asn, role)

	LOG.Infof("ASN Pool: Count of  ASN %d role  %s %d", asn, role, asnCount)
	return asnCount, asndb
}

func (sh *DeviceInteractor) getASNCountInUsedASN(LOG *nLOG.Entry, FabricID uint, asn uint64, role string) int64 {
	var asnCount int64
	asnCount, err := sh.Db.GetUsedASNCountOnASNAndRole(FabricID, asn, role)
	if err != nil {
		DEC(LOG).Infoln("No Entry")
	}
	LOG.Infof("Used ASN Table : Count of  ASN %d for role %s %d", asn, role, asnCount)
	return asnCount
}

func (sh *DeviceInteractor) removeAllocationEntry(LOG *nLOG.Entry, FabricID uint, DeviceID uint, asn uint64, role string) error {
	LOG.Infof("Remove ASN from Used ASN Table  %d", asn)

	usedASN, err := sh.Db.GetUsedASNOnASNAndDeviceAndRole(FabricID, DeviceID, asn, role)
	if err != nil {
		statusMsg := fmt.Sprintf("ASN %d not present in the Used ASN Table", asn)
		DEC(LOG).Infoln(statusMsg)
		return errors.New(statusMsg)
	}
	LOG.Infoln(usedASN)
	return sh.Db.DeleteUsedASN(&usedASN)
}

func (sh *DeviceInteractor) makeUsedASNEntryForDevice(LOG *nLOG.Entry, FabricID uint, DeviceID uint, asn uint64, role string) error {

	var asnCount int64
	//Same ASN can be shared by devices
	asnCount, err := sh.Db.GetUsedASNCountOnASNAndDevice(FabricID, asn, DeviceID)

	if err != nil {
		DEC(LOG).Infoln("No Entry")
	}
	if asnCount == 0 {
		LOG.Infof("Make ASN entry in Used ASN Table for role %s  ASN %d",
			role, asn)
		usedEntry := domain.UsedASN{FabricID: FabricID, DeviceID: DeviceID, ASN: asn, DeviceRole: role}
		return sh.Db.CreateUsedASN(&usedEntry)
	}
	return nil
}

func (sh *DeviceInteractor) getASNCountForRole(LOG *nLOG.Entry, FabricID uint, role string) (int64, error) {

	LOG.Infof("Get ASN Count in Allocation Table for role %s", role)
	asnCount, err := sh.Db.GetASNCountOnRole(FabricID, role)
	if err != nil {
		return 0, nil
	}
	LOG.Infof("Value of ASN Count in Allocation Table for role %s Count %d", role, asnCount)
	return asnCount, nil

}

func (sh *DeviceInteractor) returnASNToPool(LOG *nLOG.Entry, FabricID uint, role string, asn uint64) {
	asnCount, err := sh.Db.GetASNCountOnASN(FabricID, asn)
	if err != nil {
		LOG.Infof("No Entry for asn %d", asn)
	}
	if asnCount == 0 {
		LOG.Infof("Add ASN back in Allocation Table for role %s ASN %d", role, asn)
		asndb := domain.ASNAllocationPool{FabricID: FabricID,
			ASN: asn, DeviceRole: role}

		if err := sh.Db.CreateASN(&asndb); err != nil {
			DEC(LOG).Infoln(err)
		}

	}

}

func (sh *DeviceInteractor) getNextASNFromPool(LOG *nLOG.Entry, FabricID uint, role string) (domain.ASNAllocationPool, error) {
	LOG.Infof("Fetch next ASN from Allocation Table for role %s", role)
	asn, err := sh.Db.GetNextASNForRole(FabricID, role)
	if err != nil {
		statusMsg := fmt.Sprintf("ASN Exhausted for %d", FabricID)
		DEC(LOG).Println(statusMsg)
		return asn, errors.New(statusMsg)
	}
	LOG.Infof("Value of next ASN from Allocation Table for Role %s ASN %d", role, asn.ASN)
	return asn, nil
}

func (sh *DeviceInteractor) deleteASN(LOG *nLOG.Entry, asn *domain.ASNAllocationPool) error {
	LOG.Infof("Delete  ASN from Allocation Table for Fabric %d Role %s", asn.FabricID, asn.DeviceRole)
	return sh.Db.DeleteASN(asn)
}

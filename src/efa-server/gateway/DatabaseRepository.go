package gateway

import (
	//"fmt"
	"efa-server/domain"
	"efa-server/infra/constants"
	"efa-server/infra/database"
	"efa-server/infra/util"
	"github.com/jinzhu/gorm"
)

//DatabaseRepository represents the Application Database Repository
type DatabaseRepository struct {
	Database    *database.Database
	Transaction *gorm.DB
}

//GetDBHandle returns a handle to DB, using which the database operations can be performed
func (dbRepo *DatabaseRepository) GetDBHandle() *gorm.DB {
	if dbRepo.Transaction != nil {
		return dbRepo.Transaction
	}
	return dbRepo.Database.Instance
}

//Backup represents Database
func (dbRepo *DatabaseRepository) Backup() error {
	return dbRepo.Database.BackupDB()
}

//OpenTransaction begins the database transaction
func (dbRepo *DatabaseRepository) OpenTransaction() error {
	dbRepo.Transaction = dbRepo.GetDBHandle().Begin()
	return dbRepo.Transaction.Error
}

//CommitTransaction commits the database transaction
func (dbRepo *DatabaseRepository) CommitTransaction() error {

	err := dbRepo.GetDBHandle().Commit().Error
	dbRepo.Transaction = nil
	return err
}

//RollBackTransaction rolls back the database transaction
func (dbRepo *DatabaseRepository) RollBackTransaction() error {

	err := dbRepo.GetDBHandle().Rollback().Error
	dbRepo.Transaction = nil
	return err
}

//GetFabric returns the DC Fabric instance from the database
func (dbRepo *DatabaseRepository) GetFabric(FabricName string) (domain.Fabric, error) {
	var fabric database.Fabric
	err := dbRepo.GetDBHandle().First(&fabric, "Name = ?", FabricName).Error

	var DomainFabric domain.Fabric
	Copy(&DomainFabric, fabric)

	return DomainFabric, err
}

//CreateFabric creates the DC Fabric instance in the database
func (dbRepo *DatabaseRepository) CreateFabric(Fabric *domain.Fabric) error {
	var DBFabric database.Fabric
	Copy(&DBFabric, Fabric)
	err := dbRepo.GetDBHandle().Create(&DBFabric).Error
	if err == nil {
		Fabric.ID = DBFabric.ID
	}
	return err
}

//DeleteFabric deletes the DC Fabric instance from the databse
func (dbRepo *DatabaseRepository) DeleteFabric(FabricName string) error {
	var fabric database.Fabric
	return dbRepo.GetDBHandle().Where("Name = ?", FabricName).Delete(fabric).Error
}

//GetFabricProperties returns the properties of the DC Fabric from the database
func (dbRepo *DatabaseRepository) GetFabricProperties(FabricID uint) (domain.FabricProperties, error) {
	var DBFabricProps database.FabricProperties
	err := dbRepo.GetDBHandle().First(&DBFabricProps, "fabric_id = ?", FabricID).Error

	var DomainFabricProps domain.FabricProperties
	Copy(&DomainFabricProps, DBFabricProps)

	return DomainFabricProps, err
}

//CreateFabricProperties creates the properties of the DC Fabric in the database
func (dbRepo *DatabaseRepository) CreateFabricProperties(FabricProperties *domain.FabricProperties) error {
	var DBFabricProps database.FabricProperties
	Copy(&DBFabricProps, &FabricProperties)
	err := dbRepo.GetDBHandle().Create(&DBFabricProps).Error
	if err == nil {
		FabricProperties.ID = DBFabricProps.ID
	}
	return err

}

//UpdateFabricProperties updates the properties of the DC Fabric in the database
func (dbRepo *DatabaseRepository) UpdateFabricProperties(FabricProperties *domain.FabricProperties) error {

	var DBFabricProps database.FabricProperties
	Copy(&DBFabricProps, &FabricProperties)
	err := dbRepo.GetDBHandle().Save(&DBFabricProps).Error
	if err == nil {
		FabricProperties.ID = DBFabricProps.ID
	}
	return err
}

//CreateASN creates ASNAllocationPool instance in the database
func (dbRepo *DatabaseRepository) CreateASN(ASNAllocationPool *domain.ASNAllocationPool) error {
	var DBASNAllocationPool database.ASNAllocationPool
	Copy(&DBASNAllocationPool, ASNAllocationPool)
	err := dbRepo.GetDBHandle().Save(&DBASNAllocationPool).Error
	if err == nil {
		ASNAllocationPool.ID = DBASNAllocationPool.ID
	}
	return err
}

//DeleteASNPool deletes all ASNAllocationPool instances from the database
func (dbRepo *DatabaseRepository) DeleteASNPool() error {
	return dbRepo.GetDBHandle().Model(&database.ASNAllocationPool{}).Delete(&database.ASNAllocationPool{}).Error
}

//DeleteASN deletes an input ASNAllocationPool instance from the database
func (dbRepo *DatabaseRepository) DeleteASN(ASNAllocationPool *domain.ASNAllocationPool) error {
	var DBASNAllocationPool database.ASNAllocationPool
	Copy(&DBASNAllocationPool, &ASNAllocationPool)
	err := dbRepo.GetDBHandle().Delete(&DBASNAllocationPool).Error
	return err
}

//GetASNAndCountOnASNAndRole returns the number of instances of ASNAllocationPool and ASNAllocationPool from the database, for a given ASN and device role
func (dbRepo *DatabaseRepository) GetASNAndCountOnASNAndRole(FabricID uint, asn uint64, role string) (int64, domain.ASNAllocationPool) {
	var asnCount int64
	var asndb database.ASNAllocationPool

	dbRepo.GetDBHandle().Model(database.ASNAllocationPool{}).
		Where("fabric_id = ? AND asn = ? AND device_role = ?", FabricID, asn, role).Find(&asndb).Count(&asnCount)
	var DomASN domain.ASNAllocationPool
	Copy(&DomASN, asndb)
	return asnCount, DomASN
}

//GetNextASNForRole returns the available ASNAllocationPool instance from the database
func (dbRepo *DatabaseRepository) GetNextASNForRole(FabricID uint, role string) (domain.ASNAllocationPool, error) {
	var dbasn database.ASNAllocationPool
	err := dbRepo.GetDBHandle().Order("asn desc").Find(&dbasn, "device_role = ?", role).Error
	var DomainASN domain.ASNAllocationPool
	Copy(&DomainASN, &dbasn)
	return DomainASN, err
}

//GetASNCountOnASN returns the count of available ASNAllocationPool instances from the database, for a given ASN
func (dbRepo *DatabaseRepository) GetASNCountOnASN(FabricID uint, asn uint64) (int64, error) {
	var asnCount int64
	err := dbRepo.GetDBHandle().Model(domain.ASNAllocationPool{}).
		Where("fabric_id = ? AND asn = ?", FabricID, asn).Count(&asnCount).Error
	return asnCount, err
}

//GetASNCountOnRole returns the count of available ASNAllocationPool instances from the database, for a given device role
func (dbRepo *DatabaseRepository) GetASNCountOnRole(FabricID uint, role string) (int64, error) {
	var asnCount int64
	err := dbRepo.GetDBHandle().Model(domain.ASNAllocationPool{}).
		Where("fabric_id = ? AND device_role = ?", FabricID, role).Count(&asnCount).Error
	return asnCount, err
}

//CreateUsedASN creates an instance of UsedASN in the database
func (dbRepo *DatabaseRepository) CreateUsedASN(UsedASN *domain.UsedASN) error {
	var DBUsedASN database.UsedASN
	Copy(&DBUsedASN, &UsedASN)
	err := dbRepo.GetDBHandle().Create(&DBUsedASN).Error
	if err == nil {
		DBUsedASN.ID = UsedASN.ID
	}
	return err
}

//DeleteUsedASN deletes an instance of UsedASN in the database
func (dbRepo *DatabaseRepository) DeleteUsedASN(UsedASN *domain.UsedASN) error {
	var DBUsedASN database.UsedASN
	Copy(&DBUsedASN, &UsedASN)
	return dbRepo.GetDBHandle().Delete(&DBUsedASN).Error

}

//DeleteUsedASNPool deletes all the instances of UsedASN from the database
func (dbRepo *DatabaseRepository) DeleteUsedASNPool() error {
	return dbRepo.GetDBHandle().Model(&database.UsedASN{}).Delete(&database.UsedASN{}).Error
}

//GetUsedASNOnASNAndDeviceAndRole returns an instance of UsedASN for a given "fabric,device,device-role and ASN" input
func (dbRepo *DatabaseRepository) GetUsedASNOnASNAndDeviceAndRole(FabricID uint, DeviceID uint, asn uint64, role string) (domain.UsedASN, error) {
	var usedASN database.UsedASN
	err := dbRepo.GetDBHandle().Model(database.UsedASN{}).
		Where("fabric_id = ? AND asn = ? AND device_id = ? AND device_role = ?",
			FabricID, asn, DeviceID, role).Find(&usedASN).Error
	var DomUsedASN domain.UsedASN
	Copy(&DomUsedASN, usedASN)
	return DomUsedASN, err
}

//GetUsedASNCountOnASNAndDevice returns an instance of UsedASN for a given "fabric,device and ASN" input
func (dbRepo *DatabaseRepository) GetUsedASNCountOnASNAndDevice(FabricID uint, asn uint64, DeviceID uint) (int64, error) {
	var asnCount int64
	err := dbRepo.GetDBHandle().Model(domain.UsedASN{}).
		Where("fabric_id = ? AND asn = ? AND device_id = ?", FabricID, asn, DeviceID).Count(&asnCount).Error
	return asnCount, err
}

//GetUsedASNCountOnASNAndRole returns count of UsedASN instances for a given "fabric, asn and device-role" input
func (dbRepo *DatabaseRepository) GetUsedASNCountOnASNAndRole(FabricID uint, asn uint64, role string) (int64, error) {
	var asnCount int64
	err := dbRepo.GetDBHandle().Model(domain.UsedASN{}).
		Where("fabric_id = ? AND asn = ? AND device_role = ?", FabricID, asn, role).Count(&asnCount).Error
	return asnCount, err
}

//CreateIPEntry creates an instance of IPAllocationPool
func (dbRepo *DatabaseRepository) CreateIPEntry(IPEntry *domain.IPAllocationPool) error {
	var DBIPEntry database.IPAllocationPool
	Copy(&DBIPEntry, IPEntry)
	err := dbRepo.GetDBHandle().Save(&DBIPEntry).Error
	if err == nil {
		IPEntry.ID = DBIPEntry.ID
	}
	return err
}

//DeleteIPEntry deletes an instance of IPAllocationPool
func (dbRepo *DatabaseRepository) DeleteIPEntry(IPEntry *domain.IPAllocationPool) error {
	var DBIPEntry database.IPAllocationPool
	Copy(&DBIPEntry, IPEntry)
	err := dbRepo.GetDBHandle().Delete(&DBIPEntry).Error
	return err
}

//DeleteIPPool deletes all the instances of IPAllocationPool
func (dbRepo *DatabaseRepository) DeleteIPPool() error {
	return dbRepo.GetDBHandle().Model(&database.IPAllocationPool{}).Delete(&database.IPAllocationPool{}).Error
}

//DeleteUsedIPPool deletes all the instances of UsedIP
func (dbRepo *DatabaseRepository) DeleteUsedIPPool() error {
	return dbRepo.GetDBHandle().Model(&database.UsedIP{}).Delete(&database.UsedIP{}).Error
}

//GetIPEntryAndCountOnIPAddressAndType returns an instance of IPAllocationPool for a given "FabricID, IPAddress, IPType" input
func (dbRepo *DatabaseRepository) GetIPEntryAndCountOnIPAddressAndType(FabricID uint, ipaddress string, IPType string) (int64, domain.IPAllocationPool, error) {
	var ipCount int64
	var DBIPEntry database.IPAllocationPool

	err := dbRepo.GetDBHandle().Model(database.IPAllocationPool{}).
		Where("fabric_id = ? AND ip_address = ? AND ip_type = ?", FabricID, ipaddress, IPType).Find(&DBIPEntry).Count(&ipCount).Error

	var IPEntry domain.IPAllocationPool
	Copy(&IPEntry, &DBIPEntry)
	return ipCount, IPEntry, err
}

//GetNextIPEntryOnType returns an instance of IPAllocationPool, for a given "FabricID, IPType" input
func (dbRepo *DatabaseRepository) GetNextIPEntryOnType(FabricID uint, IPType string) (domain.IPAllocationPool, error) {
	var IPEntry domain.IPAllocationPool
	var DBIPEntry database.IPAllocationPool

	err := dbRepo.GetDBHandle().Order("id desc").Find(&DBIPEntry, "ip_type = ?", IPType).Error
	Copy(&IPEntry, &DBIPEntry)
	return IPEntry, err
}

//GetUsedIPOnDeviceInterfaceIDIPAddresssAndType returns an instance of UsedIP for a given "FabricID, DeviceID, IPAddress, IPType, InterfaceID" input
func (dbRepo *DatabaseRepository) GetUsedIPOnDeviceInterfaceIDIPAddresssAndType(FabricID uint, DeviceID uint, ipaddress string,
	IPType string, InterfaceID uint) (domain.UsedIP, error) {
	var usedIP domain.UsedIP
	var DBusedIP database.UsedIP

	err := dbRepo.GetDBHandle().Model(database.UsedIP{}).
		Where("fabric_id = ? AND ip_address = ? AND device_id = ? AND ip_type = ? AND interface_id = ?",
			FabricID, ipaddress, DeviceID, IPType, InterfaceID).Find(&DBusedIP).Error
	Copy(&usedIP, &DBusedIP)
	return usedIP, err
}

//GetUsedIPOnDeviceInterfaceIDAndType returns an instance of UsedIP, for a given "FabricID, DeviceID, IPType, InterfaceID" input
func (dbRepo *DatabaseRepository) GetUsedIPOnDeviceInterfaceIDAndType(FabricID uint, DeviceID uint,
	IPType string, InterfaceID uint) (domain.UsedIP, error) {
	var usedIP domain.UsedIP
	var DBusedIP database.UsedIP

	err := dbRepo.GetDBHandle().Model(database.UsedIP{}).
		Where("fabric_id = ? AND device_id = ? AND ip_type = ? AND interface_id = ?",
			FabricID, DeviceID, IPType, InterfaceID).Find(&DBusedIP).Error
	Copy(&usedIP, &DBusedIP)
	return usedIP, err
}

//GetUsedIPSOnDeviceAndType returns an array of UsedIP for a given "FabricID, DeviceID"
func (dbRepo *DatabaseRepository) GetUsedIPSOnDeviceAndType(FabricID uint, DeviceID uint) ([]domain.UsedIP, error) {
	var DBusedIPs []database.UsedIP
	err := dbRepo.GetDBHandle().Model(database.UsedIP{}).
		Where("fabric_id = ?  AND device_id = ?",
			FabricID, DeviceID).Find(&DBusedIPs).Error

	UsedIPs := make([]domain.UsedIP, 0, len(DBusedIPs))
	for _, DBusedIP := range DBusedIPs {
		var UsedIP domain.UsedIP
		Copy(&UsedIP, DBusedIP)
		UsedIPs = append(UsedIPs, UsedIP)
	}

	return UsedIPs, err
}

//CreateUsedIPEntry creates an instance of UsedIP in the database
func (dbRepo *DatabaseRepository) CreateUsedIPEntry(UsedIPEntry *domain.UsedIP) error {
	var DBusedIPEntry database.UsedIP
	Copy(&DBusedIPEntry, &UsedIPEntry)
	err := dbRepo.GetDBHandle().Save(&DBusedIPEntry).Error
	if err == nil {
		UsedIPEntry.ID = DBusedIPEntry.ID
	}
	return err
}

//DeleteUsedIPEntry deletes an instance of UsedIP from the database
func (dbRepo *DatabaseRepository) DeleteUsedIPEntry(UsedIPEntry *domain.UsedIP) error {
	var DBusedIPEntry database.UsedIP
	Copy(&DBusedIPEntry, &UsedIPEntry)
	err := dbRepo.GetDBHandle().Delete(&DBusedIPEntry).Error
	return err
}

//CreateIPPairEntry creates an instance of IPPairAllocationPool in the database
func (dbRepo *DatabaseRepository) CreateIPPairEntry(IPPairEntry *domain.IPPairAllocationPool) error {
	var DBIPPairEntry database.IPPairAllocationPool
	Copy(&DBIPPairEntry, IPPairEntry)
	err := dbRepo.GetDBHandle().Save(&DBIPPairEntry).Error
	if err == nil {
		IPPairEntry.ID = DBIPPairEntry.ID
	}
	return err
}

//DeleteIPPairEntry deletes an instance of IPPairAllocationPool from the database
func (dbRepo *DatabaseRepository) DeleteIPPairEntry(IPPairEntry *domain.IPPairAllocationPool) error {
	var DBIPEntry database.IPPairAllocationPool
	Copy(&DBIPEntry, IPPairEntry)
	err := dbRepo.GetDBHandle().Delete(&DBIPEntry).Error
	return err
}

//DeleteIPPairPool deletes all the instances of IPPairAllocationPool from the database
func (dbRepo *DatabaseRepository) DeleteIPPairPool() error {
	return dbRepo.GetDBHandle().Model(&database.IPPairAllocationPool{}).Delete(&database.IPPairAllocationPool{}).Error
}

//GetIPPairEntryAndCountOnIPAddressAndType returns an instance of IPPairAllocationPool, for a given "FabricID, IPAddressOne, IPAddressTwo, IPType" input
func (dbRepo *DatabaseRepository) GetIPPairEntryAndCountOnIPAddressAndType(FabricID uint, ipaddressOne string, ipaddressTwo string, IPType string) (int64, domain.IPPairAllocationPool, error) {
	var ipCount int64
	var DBIPPairEntry database.IPPairAllocationPool
	IPList := []string{ipaddressOne, ipaddressTwo}
	err := dbRepo.GetDBHandle().Model(database.IPPairAllocationPool{}).
		Where("fabric_id = ? AND ip_address_one IN (?)  AND ip_address_two IN (?)  AND ip_type = ?", FabricID, IPList, IPList, IPType).Find(&DBIPPairEntry).Count(&ipCount).Error

	var IPEntry domain.IPPairAllocationPool
	Copy(&IPEntry, &DBIPPairEntry)
	return ipCount, IPEntry, err
}

//GetIPPairEntryAndCountOnEitherIPAddressAndType returns an instance of IPPairAllocationPool for a given "FabricID, IPAddress, IPType" input
func (dbRepo *DatabaseRepository) GetIPPairEntryAndCountOnEitherIPAddressAndType(FabricID uint, ipaddress string, IPType string) (int64, domain.IPPairAllocationPool, error) {
	var ipCount int64
	var DBIPPairEntry database.IPPairAllocationPool

	err := dbRepo.GetDBHandle().Model(database.IPPairAllocationPool{}).
		Where("fabric_id = ? AND (ip_address_one =  ?  OR ip_address_two = ? )  AND ip_type = ?", FabricID, ipaddress, ipaddress, IPType).Find(&DBIPPairEntry).Count(&ipCount).Error

	var IPEntry domain.IPPairAllocationPool
	Copy(&IPEntry, &DBIPPairEntry)
	return ipCount, IPEntry, err
}

//GetNextIPPairEntryOnType returns an instance of next available IPPairAllocationPool for a given "FabricID, IPType" input
func (dbRepo *DatabaseRepository) GetNextIPPairEntryOnType(FabricID uint, IPType string) (domain.IPPairAllocationPool, error) {
	var IPPairEntry domain.IPPairAllocationPool
	var DBIPPairEntry database.IPPairAllocationPool

	err := dbRepo.GetDBHandle().Order("id desc").Find(&DBIPPairEntry, "ip_type = ?", IPType).Error
	Copy(&IPPairEntry, &DBIPPairEntry)
	return IPPairEntry, err
}

//GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType returns an instance of UsedIPPair,
// for a given "FabricID, DeviceOneID, DeviceTwoID, IPAddressOne, IPAddressTwo, IPType, InterfaceOneId, InterfaceTwoId" input
func (dbRepo *DatabaseRepository) GetUsedIPPairOnDeviceInterfaceIDIPAddresssAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint, ipaddressOne string, ipaddressTwo string, IPType string,
	InterfaceOneID uint, InterfaceTwoID uint) (domain.UsedIPPair, error) {
	var usedIP domain.UsedIPPair
	var DBusedIP database.UsedIPPair
	IPList := []string{ipaddressOne, ipaddressTwo}
	DeviceIDList := []uint{DeviceOneID, DeviceTwoID}
	InterfaceIDList := []uint{InterfaceOneID, InterfaceTwoID}
	err := dbRepo.GetDBHandle().Model(database.UsedIPPair{}).
		Where("fabric_id = ? AND ip_address_one IN (?) AND ip_address_two IN (?) AND device_one_id IN (?) AND device_two_id IN (?) AND ip_type = ? AND interface_one_id IN (?) AND interface_two_id IN (?) ",
			FabricID, IPList, IPList, DeviceIDList, DeviceIDList, IPType, InterfaceIDList, InterfaceIDList).Find(&DBusedIP).Error

	Copy(&usedIP, &DBusedIP)
	return usedIP, err
}

//GetUsedIPPairOnDeviceInterfaceIDAndType returns an instance of UsedIPPair,
//for a given "FabricID, DeviceOneID, DeviceTwoID, IPType, InterfaceOneId, InterfaceTwoId" input
func (dbRepo *DatabaseRepository) GetUsedIPPairOnDeviceInterfaceIDAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint,
	IPType string, InterfaceOneID uint, InterfaceTwoID uint) (domain.UsedIPPair, error) {
	var usedIP domain.UsedIPPair
	var DBusedIP database.UsedIPPair
	DeviceIDList := []uint{DeviceOneID, DeviceTwoID}
	InterfaceIDList := []uint{InterfaceOneID, InterfaceTwoID}
	err := dbRepo.GetDBHandle().Model(database.UsedIPPair{}).
		Where("fabric_id = ? AND device_one_id IN (?) AND device_two_id IN (?) AND ip_type = ? AND interface_one_id IN (?) AND interface_two_id IN (?)",
			FabricID, DeviceIDList, DeviceIDList, IPType, InterfaceIDList, InterfaceIDList).Find(&DBusedIP).Error
	Copy(&usedIP, &DBusedIP)
	return usedIP, err
}

//GetUsedIPPairsSOnDeviceAndType returns an array of UsedIPPair, for a given "FabricID, DeviceOneID, DeviceTwoID"
func (dbRepo *DatabaseRepository) GetUsedIPPairsSOnDeviceAndType(FabricID uint, DeviceOneID uint, DeviceTwoID uint) ([]domain.UsedIPPair, error) {
	var DBusedIPs []database.UsedIPPair
	DeviceIDList := []uint{DeviceOneID, DeviceTwoID}
	err := dbRepo.GetDBHandle().Model(database.UsedIPPair{}).
		Where("fabric_id = ?  AND device_one_id IN (?) AND device_two_id IN (?)",
			FabricID, DeviceIDList, DeviceIDList).Find(&DBusedIPs).Error

	UsedIPs := make([]domain.UsedIPPair, 0, len(DBusedIPs))
	for _, DBusedIP := range DBusedIPs {
		var UsedIP domain.UsedIPPair
		Copy(&UsedIP, DBusedIP)
		UsedIPs = append(UsedIPs, UsedIP)
	}

	return UsedIPs, err
}

//CreateUsedIPPairEntry creates an instance of UsedIPPair in the database
func (dbRepo *DatabaseRepository) CreateUsedIPPairEntry(UsedIPEntry *domain.UsedIPPair) error {
	var DBusedIPEntry database.UsedIPPair
	Copy(&DBusedIPEntry, &UsedIPEntry)
	err := dbRepo.GetDBHandle().Save(&DBusedIPEntry).Error
	if err == nil {
		UsedIPEntry.ID = DBusedIPEntry.ID
	}
	return err
}

//DeleteUsedIPPairEntry deletes an instance of UsedIPPair from the database
func (dbRepo *DatabaseRepository) DeleteUsedIPPairEntry(UsedIPEntry *domain.UsedIPPair) error {
	var DBusedIPEntry database.UsedIPPair
	Copy(&DBusedIPEntry, &UsedIPEntry)
	err := dbRepo.GetDBHandle().Delete(&DBusedIPEntry).Error
	return err
}

//DeleteUsedIPPairPool deletes all the instances of UsedIPPair from the database
func (dbRepo *DatabaseRepository) DeleteUsedIPPairPool() error {
	return dbRepo.GetDBHandle().Model(&database.UsedIPPair{}).Delete(&database.UsedIPPair{}).Error
}

//GetDeviceUsingDeviceID returns an instance of Device for a given "FabricID, DeviceID" input
func (dbRepo *DatabaseRepository) GetDeviceUsingDeviceID(FabricID uint, DeviceID uint) (domain.Device, error) {

	var DBDevice database.Device
	err := dbRepo.GetDBHandle().First(&DBDevice, "id = ?", DeviceID).Error
	var Device domain.Device
	Copy(&Device, DBDevice)
	if err == nil {
		Device.Password, err = util.AesDecrypt(constants.AESEncryptionKey, Device.Password)
		Device.IsPasswordEncrypted = false
	}
	return Device, err
}

//GetDevice returns an instance of Device for a given "FabricName, Device Mgmt IPAddress" i/p
func (dbRepo *DatabaseRepository) GetDevice(FabricName string, IPAddress string) (domain.Device, error) {

	var DBDevice database.Device
	err := dbRepo.GetDBHandle().First(&DBDevice, "ip_address = ?", IPAddress).Error
	var Device domain.Device
	Copy(&Device, DBDevice)
	if err == nil {
		Device.Password, err = util.AesDecrypt(constants.AESEncryptionKey, Device.Password)
		Device.IsPasswordEncrypted = false
	}
	return Device, err
}

//GetDeviceInAnyFabric returns an instance of Device which exists in any fabric given device IP or hostname
func (dbRepo *DatabaseRepository) GetDeviceInAnyFabric(IPAddress string) (domain.Device, error) {

	var DBDevice database.Device
	err := dbRepo.GetDBHandle().First(&DBDevice, "ip_address = ?", IPAddress).Error
	var Device domain.Device
	Copy(&Device, DBDevice)
	if err == nil {
		Device.Password, err = util.AesDecrypt(constants.AESEncryptionKey, Device.Password)
		Device.IsPasswordEncrypted = false
	}
	return Device, err
}

//GetDevicesInFabric returns an array of devices in a given fabric
func (dbRepo *DatabaseRepository) GetDevicesInFabric(FabricID uint) ([]domain.Device, error) {
	var DBDevices []database.Device
	err := dbRepo.GetDBHandle().Model(database.Device{}).Where("fabric_id = ?", FabricID).Find(&DBDevices).Error

	Devices := make([]domain.Device, 0, len(DBDevices))
	if err == nil {
		for _, DBDevice := range DBDevices {
			var Device domain.Device
			Copy(&Device, DBDevice)
			if Device.Password, err = util.AesDecrypt(constants.AESEncryptionKey, Device.Password); err != nil {
				return Devices, err
			}
			Device.IsPasswordEncrypted = false
			Devices = append(Devices, Device)
		}
	}

	return Devices, err
}

//GetDevicesInFabricMatching returns an array of devices in a given fabric
func (dbRepo *DatabaseRepository) GetDevicesInFabricMatching(FabricID uint, device []string) ([]domain.Device, error) {

	var DBDevices []database.Device
	err := dbRepo.GetDBHandle().Model(database.Device{}).Where("fabric_id = ? and ip_address IN(?)", FabricID, device).Find(&DBDevices).Error

	Devices := make([]domain.Device, 0, len(DBDevices))
	if err == nil {
		for _, DBDevice := range DBDevices {
			var Device domain.Device
			Copy(&Device, DBDevice)
			if Device.Password, err = util.AesDecrypt(constants.AESEncryptionKey, Device.Password); err != nil {
				return Devices, err
			}
			Device.IsPasswordEncrypted = false
			Devices = append(Devices, Device)
		}
	}

	return Devices, err
}

//GetDevicesInFabricNotMatching returns an array of devices in a given fabric
func (dbRepo *DatabaseRepository) GetDevicesInFabricNotMatching(FabricID uint, device []string) ([]domain.Device, error) {

	var DBDevices []database.Device
	err := dbRepo.GetDBHandle().Model(database.Device{}).Where("fabric_id = ? and ip_address NOT IN(?)", FabricID, device).Find(&DBDevices).Error

	Devices := make([]domain.Device, 0, len(DBDevices))
	if err == nil {
		for _, DBDevice := range DBDevices {
			var Device domain.Device
			Copy(&Device, DBDevice)
			if Device.Password, err = util.AesDecrypt(constants.AESEncryptionKey, Device.Password); err != nil {
				return Devices, err
			}
			Device.IsPasswordEncrypted = false
			Devices = append(Devices, Device)
		}
	}

	return Devices, err
}

//GetDevicesCountInFabric returns the count of devices in a given fabric
func (dbRepo *DatabaseRepository) GetDevicesCountInFabric(FabricID uint) (deviceCount uint16) {
	dbRepo.GetDBHandle().Model(database.Device{}).Where("fabric_id = ?", FabricID).Count(&deviceCount)
	return
}

//GetRack returns an instance of Rack for a given "FabricName, Rack Mgmt IPAddress" i/p
func (dbRepo *DatabaseRepository) GetRack(FabricName string, IP1 string, IP2 string) (domain.Rack, error) {

	var DBRack database.Rack
	err := dbRepo.GetDBHandle().First(&DBRack,
		"(device_one_ip = ? AND device_two_ip = ?) OR (device_one_ip = ? AND device_two_ip = ?) ",
		IP1, IP2, IP2, IP1).Error
	var Rack domain.Rack
	Copy(&Rack, DBRack)

	return Rack, err
}

//GetRackbyIP returns an instance of Rack for a given "FabricName, Rack Mgmt IPAddress" i/p
func (dbRepo *DatabaseRepository) GetRackbyIP(FabricName string, IP string) (domain.Rack, error) {

	var DBRack database.Rack
	err := dbRepo.GetDBHandle().First(&DBRack,
		"(device_one_ip = ?) OR (device_two_ip = ?) ",
		IP, IP).Error
	var Rack domain.Rack
	Copy(&Rack, DBRack)

	return Rack, err
}

//GetRackAll returns all instance of Rack for a given "FabricName" i/p
func (dbRepo *DatabaseRepository) GetRackAll(FabricName string) ([]domain.Rack, error) {

	var DBRacks []database.Rack
	err := dbRepo.GetDBHandle().Find(&DBRacks).Error
	var Racks []domain.Rack
	Racks = make([]domain.Rack, 0)
	Copy(&Racks, DBRacks)

	return Racks, err
}

//GetRacksInFabric returns an array of devices in a given fabric
func (dbRepo *DatabaseRepository) GetRacksInFabric(FabricID uint) ([]domain.Rack, error) {
	var DBRacks []database.Rack
	err := dbRepo.GetDBHandle().Model(database.Rack{}).Where("fabric_id = ?", FabricID).Find(&DBRacks).Error

	Racks := make([]domain.Rack, 0, len(DBRacks))
	if err == nil {
		for _, DBRack := range DBRacks {
			var Rack domain.Rack
			Copy(&Rack, DBRack)
			Racks = append(Racks, Rack)
		}
	}

	return Racks, err
}

//CreateRack creates an instance of Rack in the database
func (dbRepo *DatabaseRepository) CreateRack(Rack *domain.Rack) error {
	//Has as RackID so call Save
	if Rack.ID != 0 {
		return dbRepo.SaveRack(Rack)
	}
	// Encrypt the password before creating Racks in DB.
	var err error

	var DBRack database.Rack
	Copy(&DBRack, Rack)
	err = dbRepo.GetDBHandle().Create(&DBRack).Error
	if err == nil {
		Rack.ID = DBRack.ID
	}
	return err
}

//DeleteRack deletes instances of Rack for a given array of RackID
func (dbRepo *DatabaseRepository) DeleteRack(RackID []uint) error {
	err := dbRepo.GetDBHandle().Where("id IN (?)", RackID).Delete(&database.Rack{}).Error
	return err
}

//SaveRack creates an instance of Rack in the database
func (dbRepo *DatabaseRepository) SaveRack(Rack *domain.Rack) error {
	var err error

	var DBRack database.Rack
	Copy(&DBRack, Rack)

	err = dbRepo.GetDBHandle().Save(&DBRack).Error
	return err
}

//CreateDevice creates an instance of Device in the database
func (dbRepo *DatabaseRepository) CreateDevice(Device *domain.Device) error {
	//Has as deviceID so call Save
	if Device.ID != 0 {
		return dbRepo.SaveDevice(Device)
	}
	// Encrypt the password before creating Devices in DB.
	var err error
	if !Device.IsPasswordEncrypted {
		if Device.Password, err = util.AesEncrypt(constants.AESEncryptionKey, Device.Password); err != nil {
			return err
		}
		Device.IsPasswordEncrypted = true
	}

	var DBDevice database.Device
	Copy(&DBDevice, Device)
	err = dbRepo.GetDBHandle().Create(&DBDevice).Error
	if err == nil {
		Device.ID = DBDevice.ID
	}
	return err
}

//DeleteDevice deletes instances of Device for a given array of DeviceID
func (dbRepo *DatabaseRepository) DeleteDevice(DeviceID []uint) error {
	err := dbRepo.GetDBHandle().Where("id IN (?)", DeviceID).Delete(&database.Device{}).Error
	return err
}

//SaveDevice creates an instance of Device in the database
//TODO : Relook at CreateDevice and SaveDevice
func (dbRepo *DatabaseRepository) SaveDevice(Device *domain.Device) error {

	var err error
	if !Device.IsPasswordEncrypted {
		// Encrypt the password before saving Devices in DB.
		if Device.Password, err = util.AesEncrypt(constants.AESEncryptionKey, Device.Password); err != nil {
			return err
		}
		Device.IsPasswordEncrypted = true
	}

	var DBDevice database.Device
	Copy(&DBDevice, Device)

	err = dbRepo.GetDBHandle().Save(&DBDevice).Error
	return err
}

//GetInterface returns an instance of Interface for a given "FabricId, DeviceID, InterfaceType, InterfaceName" input
func (dbRepo *DatabaseRepository) GetInterface(FabricID uint, DeviceID uint,
	InterfaceType string, InterfaceName string) (domain.Interface, error) {

	var DBInterface database.PhysInterface
	err := dbRepo.GetDBHandle().First(&DBInterface, "fabric_id = ? AND device_id = ? AND int_type = ? AND int_name = ?",
		FabricID, DeviceID, InterfaceType, InterfaceName).Error
	var Interface domain.Interface
	Copy(&Interface, DBInterface)
	return Interface, err
}

//GetInterfacesonDevice returns an array of instance of Interface, for a given "FabricID, DeviceID"
func (dbRepo *DatabaseRepository) GetInterfacesonDevice(FabricID uint, DeviceID uint) ([]domain.Interface, error) {
	var DBInterfaces []database.PhysInterface
	err := dbRepo.GetDBHandle().Model(database.PhysInterface{}).
		Where("fabric_id = ? AND device_id = ?", FabricID, DeviceID).Find(&DBInterfaces).Error

	Interfaces := make([]domain.Interface, 0, len(DBInterfaces))
	for _, DBInterface := range DBInterfaces {
		var Interface domain.Interface
		Copy(&Interface, DBInterface)
		Interfaces = append(Interfaces, Interface)
	}

	return Interfaces, err
}

//GetInterfaceIDsMarkedForDeletion returns a map of "InterfaceID, DeviceID" which have been marked for deletion,
//for a given "FabricID" input
func (dbRepo *DatabaseRepository) GetInterfaceIDsMarkedForDeletion(FabricID uint) (map[uint]uint, error) {
	var DBInterfaces []database.PhysInterface
	err := dbRepo.GetDBHandle().Model(database.PhysInterface{}).
		Where("fabric_id = ? and config_type = ?", FabricID, domain.ConfigDelete).Find(&DBInterfaces).Error

	Interfaces := make(map[uint]uint)
	for _, DBInterface := range DBInterfaces {

		Interfaces[DBInterface.ID] = DBInterface.DeviceID
	}
	return Interfaces, err

}

//DeleteInterface deletes an instance of "PhysInterface" from the database
func (dbRepo *DatabaseRepository) DeleteInterface(Interface *domain.Interface) error {
	var DBInterface database.PhysInterface
	Copy(&DBInterface, &Interface)
	err := dbRepo.GetDBHandle().Delete(&DBInterface).Error
	return err
}

//DeleteInterfaceUsingID deletes an instance of "PhysInterface" for a given "FabricID, InterfaceID" input
func (dbRepo *DatabaseRepository) DeleteInterfaceUsingID(FabricID uint, InterfaceID uint) error {
	return dbRepo.GetDBHandle().Table("phys_interfaces").Where(
		"fabric_id = ? AND id = (?)",
		FabricID, InterfaceID).Delete(database.PhysInterface{}).Error
}

//CreateInterface creates an instance of "PhysInterface" in the database
func (dbRepo *DatabaseRepository) CreateInterface(Interface *domain.Interface) error {

	var DBInterface database.PhysInterface
	Copy(&DBInterface, Interface)
	//Save
	if Interface.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBInterface).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBInterface).Error
	if err == nil {
		Interface.ID = DBInterface.ID
	}
	return err
}

//ResetDeleteOperationOnInterface resets the "config_type" of "phys_interfaces" to "ConfigCreate",
//for a given "FabricID, InterfaceID" input
func (dbRepo *DatabaseRepository) ResetDeleteOperationOnInterface(FabricID uint, InterfaceID uint) error {
	return dbRepo.GetDBHandle().Table("phys_interfaces").Where(
		"fabric_id = ? AND id = (?) and config_type = (?)", FabricID, InterfaceID, domain.ConfigDelete).
		UpdateColumn("config_type", domain.ConfigCreate).Error
}

//GetInterfaceOnMac returns an instance of "Interface" for a given "MAC, FabricID" input
func (dbRepo *DatabaseRepository) GetInterfaceOnMac(Mac string, FabricID uint) (domain.Interface, error) {
	var DBInterface database.PhysInterface
	err := dbRepo.GetDBHandle().First(&DBInterface, "mac = ? AND fabric_id = ?", Mac, FabricID).Error
	var Interface domain.Interface
	Copy(&Interface, DBInterface)
	return Interface, err
}

//GetLLDP returns an instance of "LLDP" for a given "FabricID, DeviceID, LocalInterfaceType, LocalInterfaceName" input
func (dbRepo *DatabaseRepository) GetLLDP(FabricID uint, DeviceID uint,
	LocalInterfaceType string, LocalInterfaceName string) (domain.LLDP, error) {

	var DBLLDP database.LLDPData
	err := dbRepo.GetDBHandle().First(&DBLLDP, "fabric_id = ? AND device_id = ? AND local_int_type = ? AND local_int_name = ?",
		FabricID, DeviceID, LocalInterfaceType, LocalInterfaceName).Error
	var LLDP domain.LLDP
	Copy(&LLDP, DBLLDP)
	return LLDP, err
}

//GetLLDPsonDevice returns an array of instances of "LLDP", for a given "FabricID, DeviceID" input
func (dbRepo *DatabaseRepository) GetLLDPsonDevice(FabricID uint, DeviceID uint) ([]domain.LLDP, error) {
	var DBLLDPS []database.LLDPData
	err := dbRepo.GetDBHandle().Model(database.PhysInterface{}).
		Where("fabric_id = ? AND device_id = ?", FabricID, DeviceID).Find(&DBLLDPS).Error

	LLDPS := make([]domain.LLDP, 0, len(DBLLDPS))
	for _, DBLLDP := range DBLLDPS {
		var lldp domain.LLDP
		Copy(&lldp, DBLLDP)
		LLDPS = append(LLDPS, lldp)
	}

	return LLDPS, err
}

//CreateLLDP creates an instance of "DBLLDP" in the database
func (dbRepo *DatabaseRepository) CreateLLDP(LLDP *domain.LLDP) error {

	var DBLLDP database.LLDPData
	Copy(&DBLLDP, LLDP)
	//Save
	if LLDP.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBLLDP).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBLLDP).Error
	if err == nil {
		LLDP.ID = DBLLDP.ID
	}
	return err
}

//DeleteLLDP deletes an instance of "LLDPData" from the database
func (dbRepo *DatabaseRepository) DeleteLLDP(LLDP *domain.LLDP) error {
	var DBLLDP database.LLDPData
	Copy(&DBLLDP, &LLDP)
	err := dbRepo.GetDBHandle().Delete(&DBLLDP).Error
	return err
}

//GetLLDPNeighbor returns an instance of "domain.LLDPNeighbor",
//for a given "FabricID, DeviceOneID, DeviceTwoID, InterfaceOneID, InterfaceTwoID" input
func (dbRepo *DatabaseRepository) GetLLDPNeighbor(FabricID uint, DeviceOneID uint,
	DeviceTwoID uint, InterfaceOneID uint, InterfaceTwoID uint) (domain.LLDPNeighbor, error) {

	var DBLLDPNeighbor database.LLDPNeighbor
	err := dbRepo.GetDBHandle().First(&DBLLDPNeighbor,
		"fabric_id = ? AND device_one_id = ? AND device_two_id = ? AND interface_one_id = ? AND interface_two_id = ?",
		FabricID, DeviceOneID, DeviceTwoID, InterfaceOneID, InterfaceTwoID).Error
	var LLDPNeighbor domain.LLDPNeighbor
	Copy(&LLDPNeighbor, DBLLDPNeighbor)
	return LLDPNeighbor, err
}

//CreateLLDPNeighbor creates an instance of "LLDPNeighbor" in the database
func (dbRepo *DatabaseRepository) CreateLLDPNeighbor(LLDPNeighbor *domain.LLDPNeighbor) error {
	var DBLLDPNeighbor database.LLDPNeighbor
	Copy(&DBLLDPNeighbor, LLDPNeighbor)
	if LLDPNeighbor.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBLLDPNeighbor).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBLLDPNeighbor).Error
	if err == nil {
		LLDPNeighbor.ID = DBLLDPNeighbor.ID
	}
	return err
}

//DeleteLLDPNeighbor deletes an instance of "LLDPNeighbor" from the database
func (dbRepo *DatabaseRepository) DeleteLLDPNeighbor(BGPNeighbor *domain.LLDPNeighbor) error {
	var DBLLDPNeighbor database.LLDPNeighbor
	Copy(&DBLLDPNeighbor, &BGPNeighbor)
	err := dbRepo.GetDBHandle().Delete(&DBLLDPNeighbor).Error
	return err
}

//GetLLDPNeighborsOnEitherDevice returns an array of instances of "domain.LLDPNeighbor", for a given "FabricID, DeviceID" input
func (dbRepo *DatabaseRepository) GetLLDPNeighborsOnEitherDevice(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	var DBLLDPNeighbors []database.LLDPNeighbor
	//fmt.Println(RemoteMac)
	err := dbRepo.GetDBHandle().Model(database.LLDPNeighbor{}).Where("(fabric_id = ?) AND (device_one_id = ? OR device_two_id = ?)", FabricID, DeviceID, DeviceID).Find(&DBLLDPNeighbors).Error

	LLDPNeighbors := make([]domain.LLDPNeighbor, 0, len(DBLLDPNeighbors))
	for _, DBLLDPNeighbor := range DBLLDPNeighbors {
		var LLDPNeighbor domain.LLDPNeighbor
		Copy(&LLDPNeighbor, DBLLDPNeighbor)
		LLDPNeighbors = append(LLDPNeighbors, LLDPNeighbor)
	}

	return LLDPNeighbors, err
}

//GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion returns an array of instances of "domain.LLDPNeighbor" excluding the one's which are marked for deletion,
//for a given "FabcirID, DeviceID" input
func (dbRepo *DatabaseRepository) GetLLDPNeighborsOnDeviceExcludingMarkedForDeletion(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	var DBLLDPNeighbors []database.LLDPNeighbor
	//fmt.Println(RemoteMac)
	err := dbRepo.GetDBHandle().Model(database.LLDPNeighbor{}).Where("device_one_id = ? AND fabric_id = ? AND config_type <> ?",
		DeviceID, FabricID, domain.ConfigDelete).Find(&DBLLDPNeighbors).Error

	LLDPNeighbors := make([]domain.LLDPNeighbor, 0, len(DBLLDPNeighbors))
	for _, DBLLDPNeighbor := range DBLLDPNeighbors {
		var LLDPNeighbor domain.LLDPNeighbor
		Copy(&LLDPNeighbor, DBLLDPNeighbor)
		LLDPNeighbors = append(LLDPNeighbors, LLDPNeighbor)
	}

	return LLDPNeighbors, err
}

//GetLLDPNeighborsOnDeviceMarkedForDeletion returns an array of instances of domain.LLDPNeighbor" marked for deletion,
//for a given "FabcirID, DeviceID" input
func (dbRepo *DatabaseRepository) GetLLDPNeighborsOnDeviceMarkedForDeletion(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	var DBLLDPNeighbors []database.LLDPNeighbor
	err := dbRepo.GetDBHandle().Model(database.LLDPNeighbor{}).Where("device_one_id = ? AND fabric_id = ? AND config_type = ?",
		DeviceID, FabricID, domain.ConfigDelete).Find(&DBLLDPNeighbors).Error

	LLDPNeighbors := make([]domain.LLDPNeighbor, 0, len(DBLLDPNeighbors))
	for _, DBLLDPNeighbor := range DBLLDPNeighbors {
		var LLDPNeighbor domain.LLDPNeighbor
		Copy(&LLDPNeighbor, DBLLDPNeighbor)
		LLDPNeighbors = append(LLDPNeighbors, LLDPNeighbor)
	}

	return LLDPNeighbors, err
}

//GetLLDPNeighborsOnDevice returns an array of instances of "domain.LLDPNeighbor" for a given "FabridID, DeviceID" input
func (dbRepo *DatabaseRepository) GetLLDPNeighborsOnDevice(FabricID uint, DeviceID uint) ([]domain.LLDPNeighbor, error) {
	var DBLLDPNeighbors []database.LLDPNeighbor
	//fmt.Println(RemoteMac)
	err := dbRepo.GetDBHandle().Model(database.LLDPNeighbor{}).Where("device_one_id = ? AND fabric_id = ?",
		DeviceID, FabricID).Find(&DBLLDPNeighbors).Error

	LLDPNeighbors := make([]domain.LLDPNeighbor, 0, len(DBLLDPNeighbors))
	for _, DBLLDPNeighbor := range DBLLDPNeighbors {
		var LLDPNeighbor domain.LLDPNeighbor
		Copy(&LLDPNeighbor, DBLLDPNeighbor)
		LLDPNeighbors = append(LLDPNeighbors, LLDPNeighbor)
	}

	return LLDPNeighbors, err
}

//GetLLDPNeighborsOnRemoteDeviceID returns an array of instances of "domain.LLDPNeighbor" for a given "FabridID, DeviceID" input
func (dbRepo *DatabaseRepository) GetLLDPNeighborsOnRemoteDeviceID(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.LLDPNeighbor, error) {
	var DBLLDPNeighbors []database.LLDPNeighbor
	//fmt.Println(RemoteMac)
	err := dbRepo.GetDBHandle().Model(database.LLDPNeighbor{}).Where("device_one_id = ? AND fabric_id = ?  AND device_two_id IN(?)",
		DeviceID, FabricID, RemoteDeviceIDs).Find(&DBLLDPNeighbors).Error

	LLDPNeighbors := make([]domain.LLDPNeighbor, 0, len(DBLLDPNeighbors))
	for _, DBLLDPNeighbor := range DBLLDPNeighbors {
		var LLDPNeighbor domain.LLDPNeighbor
		Copy(&LLDPNeighbor, DBLLDPNeighbor)
		LLDPNeighbors = append(LLDPNeighbors, LLDPNeighbor)
	}

	return LLDPNeighbors, err
}

//GetLLDPOnRemoteMacExcludingMarkedForDeletion returns an instance of "domain.LLDP" excluding the instances which have beeen marked for deletion,
//for a given "RemoteMAC, FabricID" input
func (dbRepo *DatabaseRepository) GetLLDPOnRemoteMacExcludingMarkedForDeletion(RemoteMac string, FabricID uint) (domain.LLDP, error) {
	var DBLLDP database.LLDPData
	err := dbRepo.GetDBHandle().First(&DBLLDP, "remote_int_mac = ? AND fabric_id = ? and config_type <> ?",
		RemoteMac, FabricID, domain.ConfigDelete).Error
	var LLDP domain.LLDP
	Copy(&LLDP, DBLLDP)
	return LLDP, err
}

//GetLLDPNeighborsBetweenTwoDevices returns array of instances of "domain.LLDPNeighbor" between two devices in fabric
func (dbRepo *DatabaseRepository) GetLLDPNeighborsBetweenTwoDevices(FabricID uint, DeviceOneID uint, DeviceTwoID uint) ([]domain.LLDPNeighbor, error) {
	var DBLLDPNeighbors []database.LLDPNeighbor
	//fmt.Println(RemoteMac)
	err := dbRepo.GetDBHandle().Model(database.LLDPNeighbor{}).Where("(fabric_id = ?) AND (device_one_id = ? AND device_two_id = ?)", FabricID, DeviceOneID, DeviceTwoID).Find(&DBLLDPNeighbors).Error

	LLDPNeighbors := make([]domain.LLDPNeighbor, 0, len(DBLLDPNeighbors))
	for _, DBLLDPNeighbor := range DBLLDPNeighbors {
		var LLDPNeighbor domain.LLDPNeighbor
		Copy(&LLDPNeighbor, DBLLDPNeighbor)
		LLDPNeighbors = append(LLDPNeighbors, LLDPNeighbor)
	}

	return LLDPNeighbors, err
}

//GetMctClusters returns an array of instances of "domain.MctClusterConfig", for a given "FabricID, DeviceID, ConfigType[]" input
func (dbRepo *DatabaseRepository) GetMctClusters(FabricID uint, DeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error) {
	var DBMctClusterConfigs []database.MctClusterConfig
	var err error
	db := dbRepo.GetDBHandle().Model(database.MctClusterConfig{}).Order("local_node_id")
	db = db.Where("(fabric_id = ?) AND (device_id = ? )", FabricID, DeviceID)
	if len(ConfigType) > 0 {
		db = db.Where("config_type in (?)", ConfigType)
	}
	err = db.Find(&DBMctClusterConfigs).Error
	MctClusterConfigs := make([]domain.MctClusterConfig, 0, len(DBMctClusterConfigs))
	for _, DBMctClusterConfig := range DBMctClusterConfigs {
		var MctClusterConfig domain.MctClusterConfig
		Copy(&MctClusterConfig, DBMctClusterConfig)
		MctClusterConfigs = append(MctClusterConfigs, MctClusterConfig)
	}
	return MctClusterConfigs, err
}

//GetMctClustersWithBothDevices based on device ID and Neighbor Device ID retirieves MCT cluster from DB
func (dbRepo *DatabaseRepository) GetMctClustersWithBothDevices(FabricID uint, DeviceID uint, NeighborDeviceID uint, ConfigType []string) ([]domain.MctClusterConfig, error) {
	var DBMctClusterConfigs []database.MctClusterConfig
	var err error
	if len(ConfigType) > 0 {
		err = dbRepo.GetDBHandle().Model(database.MctClusterConfig{}).Where(
			"(fabric_id = ?) AND (device_id = ? ) AND (mct_neighbor_device_id = ?) "+
				"AND (config_type in (?))", FabricID, DeviceID, NeighborDeviceID,
			ConfigType).Find(&DBMctClusterConfigs).Error
	} else {
		err = dbRepo.GetDBHandle().Model(database.MctClusterConfig{}).Where(
			"(fabric_id = ?) AND (device_id = ? ) AND (mct_neighbor_device_id = ?)",
			FabricID, DeviceID, NeighborDeviceID).Find(&DBMctClusterConfigs).Error

	}
	MctClusterConfigs := make([]domain.MctClusterConfig, 0, len(DBMctClusterConfigs))
	for _, DBMctClusterConfig := range DBMctClusterConfigs {
		var MctClusterConfig domain.MctClusterConfig
		Copy(&MctClusterConfig, DBMctClusterConfig)
		MctClusterConfigs = append(MctClusterConfigs, MctClusterConfig)
	}
	return MctClusterConfigs, err
}

//GetMctClustersCount returns "count" of instances of "MctClusterConfig" in the database, for a given "FabricID, DeviceID, ConfigType[]" input
func (dbRepo *DatabaseRepository) GetMctClustersCount(FabricID uint, DeviceID uint, ConfigType []string) (uint64, error) {
	var err error
	var count uint64
	if len(ConfigType) > 0 {
		err = dbRepo.GetDBHandle().Model(database.MctClusterConfig{}).Where(
			"(fabric_id = ?) AND (device_id = ? OR mct_neighbor_device_id = ?)  AND (config_type in (?))",
			FabricID, DeviceID, DeviceID, ConfigType).Count(&count).Error
	} else {
		err = dbRepo.GetDBHandle().Model(database.MctClusterConfig{}).Where(
			"(fabric_id = ?) AND (device_id = ? OR mct_neighbor_device_id = ?)",
			FabricID, DeviceID, DeviceID).Count(&count).Error

	}
	return count, err
}

//GetMctClusterConfigWithDeviceIP returns an instance of "domain.MctClusterConfig" for a given "Device Mgmt IPAddress" input
func (dbRepo *DatabaseRepository) GetMctClusterConfigWithDeviceIP(IPAddress string) ([]domain.MctClusterConfig, error) {
	var DBSMCTConfigs []database.MctClusterConfig
	var MCTConfigs []domain.MctClusterConfig
	err := dbRepo.GetDBHandle().Model(database.MctClusterConfig{}).Where("(device_one_mgmt_ip = ? OR device_two_mgmt_ip = ? )",
		IPAddress, IPAddress).Find(&DBSMCTConfigs).Error
	if err == gorm.ErrRecordNotFound {
		return MCTConfigs, nil
	}
	if err != nil {
		log.Errorf("Error Querying MCT cluster Details with IP Address %s", IPAddress)
		return MCTConfigs, err
	}
	for _, DBSMCTConfig := range DBSMCTConfigs {
		var MCTConfig domain.MctClusterConfig
		Copy(&MCTConfig, DBSMCTConfig)
		MCTConfigs = append(MCTConfigs, MCTConfig)
	}
	return MCTConfigs, nil
}

//CreateMctClusters creates an intsnace of "MctClusterConfig" in the database
func (dbRepo *DatabaseRepository) CreateMctClusters(MCTConfig *domain.MctClusterConfig) error {
	var DBSMCTConfig database.MctClusterConfig
	Copy(&DBSMCTConfig, MCTConfig)
	if MCTConfig.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBSMCTConfig).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBSMCTConfig).Error
	if err == nil {
		MCTConfig.ID = DBSMCTConfig.ID
	}
	return err
}

//CreateMctClustersMembers creates instances of "ClusterMember" in the database
func (dbRepo *DatabaseRepository) CreateMctClustersMembers(MCTConfigs []domain.MCTMemberPorts, ClusterID uint16, ConfigType string) error {
	var DBSMCTConfig database.ClusterMember
	var err error
	for _, MCTConfig := range MCTConfigs {
		MCTConfig.ConfigType = ConfigType
		if MCTConfig.ClusterID == 0 {
			MCTConfig.ClusterID = ClusterID
		}
		Copy(&DBSMCTConfig, MCTConfig)
		if MCTConfig.ID != 0 {
			return dbRepo.GetDBHandle().Save(&DBSMCTConfig).Error
		}
		err = dbRepo.GetDBHandle().Create(&DBSMCTConfig).Error
		if err == nil {
			MCTConfig.ID = DBSMCTConfig.ID
		} else {
			return err
		}
	}
	return err
}

//GetMctMemberPortsConfig returns an array of instances of "domain.MCTMemberPorts" for a given "FabricID, DeviceID, ConfigType[]" input
func (dbRepo *DatabaseRepository) GetMctMemberPortsConfig(FabricID uint, DeviceID uint,
	RemoteDeviceID uint, ConfigType []string) ([]domain.MCTMemberPorts, error) {
	var DBMCTlusterMembers []database.ClusterMember
	var err error
	db := dbRepo.GetDBHandle().Model(database.ClusterMember{}).Order("interface_speed")
	db = db.Where("(device_id = ? )", DeviceID)
	if RemoteDeviceID != 0 {
		db = db.Where("(remote_device_id = ? )", RemoteDeviceID)
	}
	if len(ConfigType) > 0 {
		db = db.Where("config_type in (?)", ConfigType)
	}
	MctClusterMembers := make([]domain.MCTMemberPorts, 0, len(DBMCTlusterMembers))
	err = db.Find(&DBMCTlusterMembers).Error
	if err != nil {
		return MctClusterMembers, err
	}
	for _, DBMctClusterConfig := range DBMCTlusterMembers {
		var MctClusterConfig domain.MCTMemberPorts
		Copy(&MctClusterConfig, DBMctClusterConfig)
		MctClusterMembers = append(MctClusterMembers, MctClusterConfig)
	}
	return MctClusterMembers, err
}

//DeleteMCTPortConfig deletes instances of "ClusterMember" from the database
//TODO : Bulk delete??
func (dbRepo *DatabaseRepository) DeleteMCTPortConfig(MctPorts []domain.MCTMemberPorts) error {
	var DBMCTConfig database.ClusterMember
	var err error
	for _, MctPortConfig := range MctPorts {
		Copy(&DBMCTConfig, MctPortConfig)
		err = dbRepo.GetDBHandle().Delete(&DBMCTConfig).Error
		if err != nil {
			return err
		}
	}
	return err
}

//CreateSwitchConfig creates an instance of "SwitchConfig" in the database
func (dbRepo *DatabaseRepository) CreateSwitchConfig(SwitchConfig *domain.SwitchConfig) error {
	var DBSwitchConfig database.SwitchConfig
	Copy(&DBSwitchConfig, SwitchConfig)
	if SwitchConfig.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBSwitchConfig).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBSwitchConfig).Error
	if err == nil {
		SwitchConfig.ID = DBSwitchConfig.ID
	}
	return err
}

//UpdateSwitchConfigsASConfigType updates the "as_config_type" attribute of "switch_configs", based on the input criteria
func (dbRepo *DatabaseRepository) UpdateSwitchConfigsASConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	return dbRepo.GetDBHandle().Table("switch_configs").Where(
		"fabric_id = ? AND as_config_type IN (?)", FabricID, QueryconfigTypes).
		UpdateColumn("as_config_type", configType).Error
}

//UpdateSwitchConfigsLoopbackConfigType updates "loopback_ip_config_type" attribute of "switch_configs", based on the input criteria
func (dbRepo *DatabaseRepository) UpdateSwitchConfigsLoopbackConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	return dbRepo.GetDBHandle().Table("switch_configs").Where(
		"fabric_id = ? AND loopback_ip_config_type IN (?)", FabricID, QueryconfigTypes).
		Update("loopback_ip_config_type", configType).Error
}

//UpdateSwitchConfigsVTEPLoopbackConfigType updates the "vtep_loopback_ip_config_type" field of "switch_configs", based on the input criteria
func (dbRepo *DatabaseRepository) UpdateSwitchConfigsVTEPLoopbackConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	return dbRepo.GetDBHandle().Table("switch_configs").Where(
		"fabric_id = ? AND vtep_loopback_ip_config_type IN (?)", FabricID, QueryconfigTypes).
		Update("vtep_loopback_ip_config_type", configType).Error
}

//GetSwitchConfigOnFabricIDAndDeviceID returns an instance of "domain.SwitchConfig" for a given "FabricID, DeviceID" input
func (dbRepo *DatabaseRepository) GetSwitchConfigOnFabricIDAndDeviceID(FabricID uint, DeviceID uint) (domain.SwitchConfig, error) {
	var DBSwitchConfig database.SwitchConfig
	err := dbRepo.GetDBHandle().Model(domain.SwitchConfig{}).Where("device_id = ? AND fabric_id= ?", DeviceID, FabricID).Find(&DBSwitchConfig).Error
	var SwitchConfig domain.SwitchConfig
	Copy(&SwitchConfig, DBSwitchConfig)
	return SwitchConfig, err
}

//GetSwitchConfigs returns an array of instance of "domain.SwitchConfig" for a given "FabricName"
func (dbRepo *DatabaseRepository) GetSwitchConfigs(FabricName string) ([]domain.SwitchConfig, error) {
	var DBSwitchConfigs []database.SwitchConfig
	err := dbRepo.GetDBHandle().Find(&DBSwitchConfigs).Error

	SwitchConfigs := make([]domain.SwitchConfig, 0, len(DBSwitchConfigs))
	for _, DBSwitchConfig := range DBSwitchConfigs {
		var SwitchConfig domain.SwitchConfig
		Copy(&SwitchConfig, DBSwitchConfig)
		SwitchConfigs = append(SwitchConfigs, SwitchConfig)
	}

	return SwitchConfigs, err

}

//GetSwitchConfigOnDeviceIP returns an instance of "domain.SwitchConfig" for a given "FabricName, Device ManagementIP" input
func (dbRepo *DatabaseRepository) GetSwitchConfigOnDeviceIP(FabricName string, DeviceIP string) (domain.SwitchConfig, error) {
	var DBSwitchConfig database.SwitchConfig
	var SwitchConfig domain.SwitchConfig
	err := dbRepo.GetDBHandle().Model(database.SwitchConfig{}).Where("device_ip = ?", DeviceIP).First(&DBSwitchConfig).Error
	if err == gorm.ErrRecordNotFound {
		return SwitchConfig, nil
	}
	Copy(&SwitchConfig, DBSwitchConfig)
	return SwitchConfig, err
}

//CreateInterfaceSwitchConfig creates an instance of "InterfaceSwitchConfig" in the database
func (dbRepo *DatabaseRepository) CreateInterfaceSwitchConfig(InterfaceSwitchConfig *domain.InterfaceSwitchConfig) error {
	var DBInterfaceSwitchConfig database.InterfaceSwitchConfig
	Copy(&DBInterfaceSwitchConfig, InterfaceSwitchConfig)
	if InterfaceSwitchConfig.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBInterfaceSwitchConfig).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBInterfaceSwitchConfig).Error
	if err == nil {
		InterfaceSwitchConfig.ID = DBInterfaceSwitchConfig.ID
	}
	return err
}

//UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType updates "config_type" atttibute of "interface_switch_configs" instance, based on the input criteria
func (dbRepo *DatabaseRepository) UpdateInterfaceSwitchConfigsOnInterfaceIDConfigType(FabricID uint, interfaceID uint, configType string) error {
	return dbRepo.GetDBHandle().Table("interface_switch_configs").Where(
		"fabric_id = ? AND interface_id IN (?)", FabricID, interfaceID).
		Update("config_type", configType).Error
}

//UpdateInterfaceSwitchConfigsConfigType updates "config_type" attribute of "interface_switch_configs" instance, based on the input criteria
func (dbRepo *DatabaseRepository) UpdateInterfaceSwitchConfigsConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	return dbRepo.GetDBHandle().Table("interface_switch_configs").Where(
		"fabric_id = ? AND config_type IN (?)", FabricID, QueryconfigTypes).
		Update("config_type", configType).Error
}

//GetInterfaceSwitchConfigOnFabricIDAndInterfaceID returns an instance of "domain.InterfaceSwitchConfig" for a given "FabricID, InterfaceID" input
func (dbRepo *DatabaseRepository) GetInterfaceSwitchConfigOnFabricIDAndInterfaceID(FabricID uint, InterfaceID uint) (domain.InterfaceSwitchConfig, error) {
	var DBInterfaceSwitchConfig database.InterfaceSwitchConfig
	err := dbRepo.GetDBHandle().Model(domain.InterfaceSwitchConfig{}).Where("interface_id = ? AND fabric_id = ?", InterfaceID, FabricID).
		Find(&DBInterfaceSwitchConfig).Error
	var InterfaceSwitchConfig domain.InterfaceSwitchConfig
	Copy(&InterfaceSwitchConfig, DBInterfaceSwitchConfig)
	return InterfaceSwitchConfig, err
}

//GetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID returns "count" of instances of "interface_switch_configs", for a given "FabcirID, InterfaceID" input
func (dbRepo *DatabaseRepository) GetInterfaceSwitchConfigCountOnFabricIDAndInterfaceID(FabricID uint, InterfaceID uint) int64 {
	var interfaceCount int64
	dbRepo.GetDBHandle().Model(domain.InterfaceSwitchConfig{}).Where("interface_id = ? AND fabric_id = ?", InterfaceID, FabricID).
		Count(&interfaceCount)
	return interfaceCount
}

//GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion returns "count" of instances of "interface_switch_configs" excluding the ones marked for deletion,
// for a given "FabcirID, InterfaceID" input
func (dbRepo *DatabaseRepository) GetInterfaceSwitchConfigsOnInterfaceIDsExcludingMarkedForDeletion(FabricID uint, InterfaceIDs []uint) ([]domain.InterfaceSwitchConfig, error) {
	var DBInterfaceSwitchConfigs []database.InterfaceSwitchConfig

	//err := dbRepo.GetDBHandle().Model(db.InterfaceSwitchConfig{}).Where("fabric_id = ? AND device_id IN (?)",FabricID,DeviceIDStr).Find(&DBInterfaceSwitchConfigs).Error
	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("fabric_id = ? AND interface_id IN (?) AND config_type <> ?",
		FabricID, InterfaceIDs, domain.ConfigDelete).Find(&DBInterfaceSwitchConfigs).Error
	InterfaceSwitchConfigs := make([]domain.InterfaceSwitchConfig, 0, len(DBInterfaceSwitchConfigs))
	for _, DBInterfaceSwitchConfig := range DBInterfaceSwitchConfigs {
		var InterfaceSwitchConfig domain.InterfaceSwitchConfig
		Copy(&InterfaceSwitchConfig, DBInterfaceSwitchConfig)
		InterfaceSwitchConfigs = append(InterfaceSwitchConfigs, InterfaceSwitchConfig)
	}

	return InterfaceSwitchConfigs, err
}

//UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs updates "config_type" attribute of "interface_switch_configs",
// for a given "FabricID, InterfaceIDs[]", with the input "config_type"
func (dbRepo *DatabaseRepository) UpdateConfigTypeForInterfaceSwitchConfigsOnIntefaceIDs(FabricID uint, InterfaceIDs []uint, configType string) error {
	return dbRepo.GetDBHandle().Table("interface_switch_configs").Where(
		"fabric_id = ? AND interface_id IN (?)", FabricID, InterfaceIDs).
		UpdateColumn("config_type", configType).Error
}

//GetInterfaceSwitchConfigsOnDeviceID returns an array of instances of "domain.InterfaceSwitchConfig" for a given "FabricID, DeviceID" input
func (dbRepo *DatabaseRepository) GetInterfaceSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.InterfaceSwitchConfig, error) {
	var DBInterfaceSwitchConfigs []database.InterfaceSwitchConfig

	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("device_id = ? AND fabric_id = ?", DeviceID, FabricID).Find(&DBInterfaceSwitchConfigs).Error

	InterfaceSwitchConfigs := make([]domain.InterfaceSwitchConfig, 0, len(DBInterfaceSwitchConfigs))
	for _, DBInterfaceSwitchConfig := range DBInterfaceSwitchConfigs {
		var InterfaceSwitchConfig domain.InterfaceSwitchConfig
		Copy(&InterfaceSwitchConfig, DBInterfaceSwitchConfig)
		InterfaceSwitchConfigs = append(InterfaceSwitchConfigs, InterfaceSwitchConfig)
	}

	return InterfaceSwitchConfigs, err
}

//GetBGPSwitchConfigsOnDeviceID returns an array of instances of "domain.RemoteNeighborSwitchConfig" for a given "FabricID, DeviceID" input
func (dbRepo *DatabaseRepository) GetBGPSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	var DBRemoteNeighborSwitchConfigs []database.RemoteNeighborSwitchConfig
	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("device_id = ? AND fabric_id = ? AND encapsulation_type != ?", DeviceID, FabricID, domain.BGPEncapTypeForCluster).Find(&DBRemoteNeighborSwitchConfigs).Error

	RemoteNeighborSwitchConfigs := make([]domain.RemoteNeighborSwitchConfig, 0, len(DBRemoteNeighborSwitchConfigs))
	for _, DBRemoteNeighborSwitchConfig := range DBRemoteNeighborSwitchConfigs {
		var RemoteNeighborSwitchConfig domain.RemoteNeighborSwitchConfig
		Copy(&RemoteNeighborSwitchConfig, DBRemoteNeighborSwitchConfig)
		RemoteNeighborSwitchConfigs = append(RemoteNeighborSwitchConfigs, RemoteNeighborSwitchConfig)
	}

	return RemoteNeighborSwitchConfigs, err
}

//GetBGPSwitchConfigsOnRemoteDeviceID returns  MCT BGP  Neighbors
func (dbRepo *DatabaseRepository) GetBGPSwitchConfigsOnRemoteDeviceID(FabricID uint, DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	var DBRemoteNeighborSwitchConfigs []database.RemoteNeighborSwitchConfig
	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("device_id = ? AND fabric_id = ? AND remote_device_id IN (?)", DeviceID, FabricID, RemoteDeviceIDs).Find(&DBRemoteNeighborSwitchConfigs).Error

	RemoteNeighborSwitchConfigs := make([]domain.RemoteNeighborSwitchConfig, 0, len(DBRemoteNeighborSwitchConfigs))
	for _, DBRemoteNeighborSwitchConfig := range DBRemoteNeighborSwitchConfigs {
		var RemoteNeighborSwitchConfig domain.RemoteNeighborSwitchConfig
		Copy(&RemoteNeighborSwitchConfig, DBRemoteNeighborSwitchConfig)
		RemoteNeighborSwitchConfigs = append(RemoteNeighborSwitchConfigs, RemoteNeighborSwitchConfig)
	}

	return RemoteNeighborSwitchConfigs, err
}

//GetMCTBGPSwitchConfigsOnDeviceID MCT BGP Leaf Neighbors
func (dbRepo *DatabaseRepository) GetMCTBGPSwitchConfigsOnDeviceID(FabricID uint, DeviceID uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	var DBRemoteNeighborSwitchConfigs []database.RemoteNeighborSwitchConfig
	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("device_id = ? AND fabric_id = ? AND encapsulation_type = ?", DeviceID, FabricID, domain.BGPEncapTypeForCluster).Find(&DBRemoteNeighborSwitchConfigs).Error

	RemoteNeighborSwitchConfigs := make([]domain.RemoteNeighborSwitchConfig, 0, len(DBRemoteNeighborSwitchConfigs))
	for _, DBRemoteNeighborSwitchConfig := range DBRemoteNeighborSwitchConfigs {
		var RemoteNeighborSwitchConfig domain.RemoteNeighborSwitchConfig
		Copy(&RemoteNeighborSwitchConfig, DBRemoteNeighborSwitchConfig)
		RemoteNeighborSwitchConfigs = append(RemoteNeighborSwitchConfigs, RemoteNeighborSwitchConfig)
	}
	return RemoteNeighborSwitchConfigs, err
}

//GetBGPSwitchConfigs returns an array of instances of "domain.RemoteNeighborSwitchConfig" for a given "FabricID, InterfaceIDs"
func (dbRepo *DatabaseRepository) GetBGPSwitchConfigs(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	var DBRemoteNeighborSwitchConfigs []database.RemoteNeighborSwitchConfig
	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("fabric_id = ? and remote_interface_id IN (?)", FabricID, InterfaceIDs).Find(&DBRemoteNeighborSwitchConfigs).Error

	RemoteNeighborSwitchConfigs := make([]domain.RemoteNeighborSwitchConfig, 0, len(DBRemoteNeighborSwitchConfigs))
	for _, DBRemoteNeighborSwitchConfig := range DBRemoteNeighborSwitchConfigs {
		var RemoteNeighborSwitchConfig domain.RemoteNeighborSwitchConfig
		Copy(&RemoteNeighborSwitchConfig, DBRemoteNeighborSwitchConfig)
		RemoteNeighborSwitchConfigs = append(RemoteNeighborSwitchConfigs, RemoteNeighborSwitchConfig)
	}

	return RemoteNeighborSwitchConfigs, err
}

//GetBGPSwitchConfigsExcludingMarkedForDeletion gets instances of "domain.RemoteNeighborSwitchConfig" excluding the ones marked for deletion,
//for a given "FabricID, InterfaceIDs"
func (dbRepo *DatabaseRepository) GetBGPSwitchConfigsExcludingMarkedForDeletion(FabricID uint, InterfaceIDs []uint) ([]domain.RemoteNeighborSwitchConfig, error) {
	var DBRemoteNeighborSwitchConfigs []database.RemoteNeighborSwitchConfig
	err := dbRepo.GetDBHandle().Model(database.InterfaceSwitchConfig{}).Where("fabric_id = ? and remote_interface_id IN (?) and config_type <> ?",
		FabricID, InterfaceIDs, domain.ConfigDelete).Find(&DBRemoteNeighborSwitchConfigs).Error

	RemoteNeighborSwitchConfigs := make([]domain.RemoteNeighborSwitchConfig, 0, len(DBRemoteNeighborSwitchConfigs))
	for _, DBRemoteNeighborSwitchConfig := range DBRemoteNeighborSwitchConfigs {
		var RemoteNeighborSwitchConfig domain.RemoteNeighborSwitchConfig
		Copy(&RemoteNeighborSwitchConfig, DBRemoteNeighborSwitchConfig)
		RemoteNeighborSwitchConfigs = append(RemoteNeighborSwitchConfigs, RemoteNeighborSwitchConfig)
	}

	return RemoteNeighborSwitchConfigs, err
}

//UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID updates the "config_type" attribute of "remote_neighbor_switch_configs", for an input criteria
func (dbRepo *DatabaseRepository) UpdateConfigTypeForBGPSwitchConfigsOnIntefaceID(FabricID uint, InterfaceIDs []uint, configType string) error {
	return dbRepo.GetDBHandle().Table("remote_neighbor_switch_configs").Where(
		"fabric_id = ? AND remote_interface_id IN (?)", FabricID, InterfaceIDs).
		UpdateColumn("config_type", configType).Error
}

//GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID returns an instance of "domain.RemoteNeighborSwitchConfig" for a given "FabricID, RemoteInterfaceID"
func (dbRepo *DatabaseRepository) GetBGPSwitchConfigOnFabricIDAndRemoteInterfaceID(FabricID uint, RemoteInterfaceID uint) (domain.RemoteNeighborSwitchConfig, error) {
	var DBRemoteNeighborSwitchConfig database.RemoteNeighborSwitchConfig
	err := dbRepo.GetDBHandle().Model(database.RemoteNeighborSwitchConfig{}).Where("remote_interface_id = ? AND fabric_id = ?", RemoteInterfaceID, FabricID).
		Find(&DBRemoteNeighborSwitchConfig).Error
	var RemoteNeighborSwitchConfig domain.RemoteNeighborSwitchConfig
	Copy(&RemoteNeighborSwitchConfig, DBRemoteNeighborSwitchConfig)

	return RemoteNeighborSwitchConfig, err

}

//GetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID returns "count" of instances "RemoteNeighborSwitchConfig" in the database, for a given "FabricID, RemoteInterfaceID"
func (dbRepo *DatabaseRepository) GetBGPSwitchConfigCountOnFabricIDAndRemoteInterfaceID(FabricID uint, RemoteInterfaceID uint) int64 {
	var interfaceCount int64
	dbRepo.GetDBHandle().Model(database.RemoteNeighborSwitchConfig{}).Where("remote_interface_id = ? AND fabric_id = ?", RemoteInterfaceID, FabricID).
		Count(&interfaceCount)

	return interfaceCount

}

//CreateBGPSwitchConfig creates an instance of "RemoteNeighborSwitchConfig" in the database
func (dbRepo *DatabaseRepository) CreateBGPSwitchConfig(BGPSwitchConfig *domain.RemoteNeighborSwitchConfig) error {
	var DBRemoteNeighborSwitchConfig database.RemoteNeighborSwitchConfig
	Copy(&DBRemoteNeighborSwitchConfig, BGPSwitchConfig)
	if BGPSwitchConfig.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBRemoteNeighborSwitchConfig).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBRemoteNeighborSwitchConfig).Error
	if err == nil {
		BGPSwitchConfig.ID = DBRemoteNeighborSwitchConfig.ID
	}
	return err
}

//UpdateBGPSwitchConfigsOnInterfaceIDConfigType updates "config_type" attribute of "remote_neighbor_switch_configs" for a given "FabricID, InterfaceID, ConfigType" input
func (dbRepo *DatabaseRepository) UpdateBGPSwitchConfigsOnInterfaceIDConfigType(FabricID uint, InterfaceID uint, configType string) error {
	return dbRepo.GetDBHandle().Table("remote_neighbor_switch_configs").Where(
		"fabric_id = ? AND remote_interface_id IN (?)", FabricID, InterfaceID).
		Update("config_type", configType).Error
}

//UpdateBGPSwitchConfigsConfigType updates "config_type" attribute of "remote_neighbor_switch_configs" based on the input criteria
func (dbRepo *DatabaseRepository) UpdateBGPSwitchConfigsConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	err := dbRepo.GetDBHandle().Table("remote_neighbor_switch_configs").Where(
		"fabric_id = ? AND config_type IN (?)", FabricID, QueryconfigTypes).
		Update("config_type", configType).Error
	if err != nil {
		return err
	}
	err = dbRepo.GetDBHandle().Table("remote_neighbor_switch_configs").Where(
		"fabric_id = ?", FabricID).Delete(database.RemoteNeighborSwitchConfig{}, "config_type IN (?)", domain.ConfigDelete).Error
	if err != nil {
		return err
	}
	return nil
}

//CreateRackEvpnConfig creates an instance of "RackEvpnNeighbors" in the database
func (dbRepo *DatabaseRepository) CreateRackEvpnConfig(RackEvpnConfig *domain.RackEvpnNeighbors) error {
	var DBRackEvpnConfig database.RackEvpnNeighbors
	Copy(&DBRackEvpnConfig, RackEvpnConfig)
	if RackEvpnConfig.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBRackEvpnConfig).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBRackEvpnConfig).Error
	if err == nil {
		RackEvpnConfig.ID = DBRackEvpnConfig.ID
	}
	return err
}

//GetRackEvpnConfig returns RackEvpnNeighbors
func (dbRepo *DatabaseRepository) GetRackEvpnConfig(RackID uint) ([]domain.RackEvpnNeighbors, error) {
	var DBRackEvpnNeighbors []database.RackEvpnNeighbors
	var RackEvpnNeighbors []domain.RackEvpnNeighbors

	err := dbRepo.GetDBHandle().Model(database.RackEvpnNeighbors{}).Where("local_rack_id = ?", RackID).Find(&DBRackEvpnNeighbors).Error
	if err == gorm.ErrRecordNotFound {
		return RackEvpnNeighbors, nil
	}
	if err != nil {
		log.Errorf("Error Querying Rack Evpn Neighbors")
		return RackEvpnNeighbors, err
	}
	for _, DBRackEvpnNeighbor := range DBRackEvpnNeighbors {
		var RackEvpnNeighbor domain.RackEvpnNeighbors
		Copy(&RackEvpnNeighbor, DBRackEvpnNeighbor)
		RackEvpnNeighbors = append(RackEvpnNeighbors, RackEvpnNeighbor)
	}
	return RackEvpnNeighbors, nil
}

//GetRackEvpnConfigOnDeviceID returns RackEvpnNeighbors
func (dbRepo *DatabaseRepository) GetRackEvpnConfigOnDeviceID(DeviceID uint) ([]domain.RackEvpnNeighbors, error) {
	var DBRackEvpnNeighbors []database.RackEvpnNeighbors
	var RackEvpnNeighbors []domain.RackEvpnNeighbors

	err := dbRepo.GetDBHandle().Model(database.RackEvpnNeighbors{}).Where("local_device_id = ?", DeviceID).Find(&DBRackEvpnNeighbors).Error
	if err == gorm.ErrRecordNotFound {
		return RackEvpnNeighbors, nil
	}
	if err != nil {
		log.Errorf("Error Querying Rack Evpn Neighbors")
		return RackEvpnNeighbors, err
	}
	for _, DBRackEvpnNeighbor := range DBRackEvpnNeighbors {
		var RackEvpnNeighbor domain.RackEvpnNeighbors
		Copy(&RackEvpnNeighbor, DBRackEvpnNeighbor)
		RackEvpnNeighbors = append(RackEvpnNeighbors, RackEvpnNeighbor)
	}
	return RackEvpnNeighbors, nil
}

//GetRackEvpnConfigOnRemoteDeviceID returns RackEvpnNeighbors
func (dbRepo *DatabaseRepository) GetRackEvpnConfigOnRemoteDeviceID(DeviceID uint, RemoteDeviceIDs []uint) ([]domain.RackEvpnNeighbors, error) {
	var DBRackEvpnNeighbors []database.RackEvpnNeighbors
	var RackEvpnNeighbors []domain.RackEvpnNeighbors

	err := dbRepo.GetDBHandle().Model(database.RackEvpnNeighbors{}).Where("local_device_id = ? AND remote_device_id IN(?)", DeviceID, RemoteDeviceIDs).Find(&DBRackEvpnNeighbors).Error
	if err == gorm.ErrRecordNotFound {
		return RackEvpnNeighbors, nil
	}
	if err != nil {
		log.Errorf("Error Querying Rack Evpn Neighbors")
		return RackEvpnNeighbors, err
	}
	for _, DBRackEvpnNeighbor := range DBRackEvpnNeighbors {
		var RackEvpnNeighbor domain.RackEvpnNeighbors
		Copy(&RackEvpnNeighbor, DBRackEvpnNeighbor)
		RackEvpnNeighbors = append(RackEvpnNeighbors, RackEvpnNeighbor)
	}
	return RackEvpnNeighbors, nil
}

//CreateExecutionLog creates an instance of "ExecutionLog" in the database
func (dbRepo *DatabaseRepository) CreateExecutionLog(ExecutionLog *domain.ExecutionLog) error {
	var DBExecutionLog database.ExecutionLog
	Copy(&DBExecutionLog, ExecutionLog)

	err := dbRepo.GetDBHandle().Create(&DBExecutionLog).Error
	if err == nil {
		ExecutionLog.ID = DBExecutionLog.ID
	}
	return err
}

//GetExecutionLogList returns an array of "domain.ExecutionLog" and the number of elements in the array will be limited by the input
func (dbRepo *DatabaseRepository) GetExecutionLogList(limit int, status string) ([]domain.ExecutionLog, error) {
	var DBExecutionLogs []database.ExecutionLog

	db := dbRepo.GetDBHandle().Model(database.ExecutionLog{}).Order("start_time desc")
	db = db.Where("status != 'Recieved'")

	if status != "all" {
		if status == "succeeded" {
			db = db.Where("status like 'Completed%'")
		}
		if status == "failed" {
			db = db.Where("status like 'Failed%'")
		}
	}

	if limit != 0 {
		db = db.Limit(limit)
	}

	err := db.Find(&DBExecutionLogs).Error

	ExecutionLogs := make([]domain.ExecutionLog, 0, len(DBExecutionLogs))
	for _, DBExecutionLog := range DBExecutionLogs {
		var ExecutionLog domain.ExecutionLog
		Copy(&ExecutionLog, DBExecutionLog)
		ExecutionLogs = append(ExecutionLogs, ExecutionLog)
	}

	return ExecutionLogs, err
}

//GetExecutionLogByUUID returns an instance of "domain.ExecutionLog" for a given execution ID
func (dbRepo *DatabaseRepository) GetExecutionLogByUUID(uuid string) (domain.ExecutionLog, error) {
	var DBExecutionLog database.ExecutionLog

	err := dbRepo.GetDBHandle().Model(database.ExecutionLog{}).Where("UUID = ? AND status != 'Recieved'", uuid).Find(&DBExecutionLog).Error
	//fmt.Println(DBExecutionLog)
	var ExecutionLog domain.ExecutionLog
	Copy(&ExecutionLog, DBExecutionLog)
	return ExecutionLog, err
}

//UpdateExecutionLog updates an instance of "ExecutionLog" in the database
func (dbRepo *DatabaseRepository) UpdateExecutionLog(ExecutionLog *domain.ExecutionLog) error {

	var DBExecutionLog database.ExecutionLog
	Copy(&DBExecutionLog, &ExecutionLog)
	err := dbRepo.GetDBHandle().Save(&DBExecutionLog).Error
	if err == nil {
		ExecutionLog.ID = DBExecutionLog.ID
	}
	return err
}

//CreateMctClusterConfig creates an instance of "MCTClusterDetail" in the database
func (dbRepo *DatabaseRepository) CreateMctClusterConfig(MCTConfig *domain.MCTClusterDetails) error {
	var DBSMCTConfig database.MCTClusterDetail
	Copy(&DBSMCTConfig, MCTConfig)
	if MCTConfig.ID != 0 {
		return dbRepo.GetDBHandle().Save(&DBSMCTConfig).Error
	}
	err := dbRepo.GetDBHandle().Create(&DBSMCTConfig).Error
	if err == nil {
		MCTConfig.ID = DBSMCTConfig.ID
	}
	return err
}

//DeleteMCTCluster deletes an instance of "MCTClusterDetail" from the database, for a given DeviceID
func (dbRepo *DatabaseRepository) DeleteMCTCluster(DeviceID uint) error {
	err := dbRepo.GetDBHandle().Delete(database.MCTClusterDetail{}, "device_id = ?", DeviceID).Error
	return err
}

//GetMCTCluster returns an instance of "domain.MCTClusterDetails" for a given DeviceID
func (dbRepo *DatabaseRepository) GetMCTCluster(DeviceID uint) (domain.MCTClusterDetails, error) {
	var DBMCTConfig database.MCTClusterDetail
	err := dbRepo.GetDBHandle().First(&DBMCTConfig, "device_id = ?", DeviceID).Error
	var MCTConfig domain.MCTClusterDetails
	Copy(&MCTConfig, DBMCTConfig)
	return MCTConfig, err
}

//UpdateMctPortConfigType updates "config_type" attribute of "cluster_members" based on the input criteria
func (dbRepo *DatabaseRepository) UpdateMctPortConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	return dbRepo.GetDBHandle().Table("cluster_members").Where(
		"fabric_id = ? AND config_type IN (?)", FabricID, QueryconfigTypes).
		UpdateColumn("config_type", configType).Error
}

//DeleteMctPortsMarkedForDelete deletes instances of "cluster_members" which have been marked for deletion, for a given "FabricID"
func (dbRepo *DatabaseRepository) DeleteMctPortsMarkedForDelete(FabricID uint) error {
	return dbRepo.GetDBHandle().Table("cluster_members").Where(
		"fabric_id = ?", FabricID).Delete(database.ClusterMember{}, "config_type IN (?)", domain.ConfigDelete).Error
}

//UpdateLLDPConfigType updates the "config_type" attribute of "lldp_data" and "lldp_neighbors" for a given input criteria
func (dbRepo *DatabaseRepository) UpdateLLDPConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {

	err := dbRepo.GetDBHandle().Table("lldp_data").Where(
		"fabric_id = ? AND config_type IN (?)", FabricID, QueryconfigTypes).
		UpdateColumn("config_type", configType).Error
	if err != nil {
		return err
	}
	err = dbRepo.GetDBHandle().Table("lldp_neighbors").Where(
		"fabric_id = ? AND config_type IN (?)", FabricID, QueryconfigTypes).
		UpdateColumn("config_type", configType).Error
	if err != nil {
		return err
	}
	return err
}

//DeleteLLDPMarkedForDelete deletes instances of "lldp_data" and "lldp_neighbors" which have been marked for deletion, for a given "FabricID"
func (dbRepo *DatabaseRepository) DeleteLLDPMarkedForDelete(FabricID uint) error {
	err := dbRepo.GetDBHandle().Table("lldp_data").Where(
		"fabric_id = ?", FabricID).Delete(database.LLDPData{}, "config_type IN (?)", domain.ConfigDelete).Error
	if err != nil {
		return err
	}
	err = dbRepo.GetDBHandle().Table("lldp_neighbors").Where(
		"fabric_id = ?", FabricID).Delete(database.LLDPNeighbor{}, "config_type IN (?)", domain.ConfigDelete).Error
	if err != nil {
		return err
	}
	return err
}

//UpdateMctClusterConfigType updates "config_type" and "updated_attributes" attributes of "mct_cluster_configs" for a given input criteria
func (dbRepo *DatabaseRepository) UpdateMctClusterConfigType(FabricID uint, QueryconfigTypes []string, configType string) error {
	updatedAttr := 0
	err := dbRepo.GetDBHandle().Table("mct_cluster_configs").Where(
		"fabric_id = ? AND config_type IN (?)", FabricID, QueryconfigTypes).
		UpdateColumn("config_type", configType).Error
	if err != nil {
		return err
	}
	err = dbRepo.GetDBHandle().Table("mct_cluster_configs").Where(
		"fabric_id = ?", FabricID, QueryconfigTypes).
		UpdateColumn("updated_attributes", updatedAttr).Error
	if err != nil {
		return err
	}
	return err
}

//DeleteMctClustersMarkedForDelete deletes instances of "mct_cluster_configs" for a given "FabricID"
func (dbRepo *DatabaseRepository) DeleteMctClustersMarkedForDelete(FabricID uint) error {
	return dbRepo.GetDBHandle().Table("mct_cluster_configs").Where(
		"fabric_id = ?", FabricID).Delete(database.MctClusterConfig{}, "config_type IN (?)", domain.ConfigDelete).Error
}

//MarkMctClusterForDelete marks instances of "mct_cluster_configs" for delete, for a given "FabricID, DeviceID"
func (dbRepo *DatabaseRepository) MarkMctClusterForDelete(FabricID uint, DeviceID uint) error {
	err := dbRepo.GetDBHandle().Table("mct_cluster_configs").Where("fabric_id = ? AND mct_neighbor_device_id = ?",
		FabricID, DeviceID).UpdateColumn("config_type", domain.ConfigDelete).Error
	if err != nil {
		return err
	}
	err = dbRepo.GetDBHandle().Table("mct_cluster_configs").Where("fabric_id = ? AND device_id = ?",
		FabricID, DeviceID).UpdateColumn("config_type", domain.ConfigDelete).Error

	return err
}

//MarkMctClusterForDeleteWithBothDevices marks instances of "mct_cluster_configs" for delete, for a given "FabricID, DeviceID, MctNeighborDeviceID" input
func (dbRepo *DatabaseRepository) MarkMctClusterForDeleteWithBothDevices(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error {
	return dbRepo.GetDBHandle().Table("mct_cluster_configs").Where(
		"fabric_id = ? AND device_id = ? AND mct_neighbor_device_id = ?", FabricID, DeviceID,
		MctNeighborDeviceID).UpdateColumn("config_type", domain.ConfigDelete).Error

}

//MarkMctClusterMemberPortsForDelete marks instances of "cluster_members" for delete, for a given "FabricID, DeviceID"
func (dbRepo *DatabaseRepository) MarkMctClusterMemberPortsForDelete(FabricID uint, DeviceID uint) error {
	err := dbRepo.GetDBHandle().Table("cluster_members").Where(
		"fabric_id = ? AND device_id = ?", FabricID, DeviceID).
		UpdateColumn("config_type", domain.ConfigDelete).Error
	if err != nil {
		return err
	}
	err = dbRepo.GetDBHandle().Table("cluster_members").Where(
		"fabric_id = ? AND remote_device_id = ?", FabricID, DeviceID).
		UpdateColumn("config_type", domain.ConfigDelete).Error

	return err

}

//MarkMctClusterMemberPortsForCreate marks instances of "cluster_members" for create, for a given "FabricID, DeviceID, RemoteDeviceID" input
func (dbRepo *DatabaseRepository) MarkMctClusterMemberPortsForCreate(FabricID uint, DeviceID uint, RemoteDeviceID uint) error {
	return dbRepo.GetDBHandle().Table("cluster_members").Where(
		"fabric_id = ? AND device_id = ? AND remote_device_id = ?", FabricID, DeviceID, RemoteDeviceID).
		UpdateColumn("config_type", domain.ConfigCreate).Error

}

//MarkMctClusterMemberPortsForDeleteWithBothDevices marks "cluster_members" for delete, for a given "FabricID, DeviceID, MctNeighborDeviceID" input
func (dbRepo *DatabaseRepository) MarkMctClusterMemberPortsForDeleteWithBothDevices(FabricID uint, DeviceID uint, MctNeighborDeviceID uint) error {
	return dbRepo.GetDBHandle().Table("cluster_members").Where(
		"fabric_id = ? AND device_id = ? AND remote_device_id = ?", FabricID, DeviceID,
		MctNeighborDeviceID).UpdateColumn("config_type", domain.ConfigDelete).Error

}

//DeleteMctClustersUsingClusterObject deletes instances of "mct_cluster_configs", for a given "FabricID, DeviceID"
func (dbRepo *DatabaseRepository) DeleteMctClustersUsingClusterObject(oldMct domain.MctClusterConfig) error {
	var mct database.MctClusterConfig
	Copy(&mct, &oldMct)
	return dbRepo.GetDBHandle().Delete(&mct).Error
}

//DeleteMctClustersWithMgmtIP deletes the instance of mct_cluster_configs for a given device management IP
func (dbRepo *DatabaseRepository) DeleteMctClustersWithMgmtIP(IPAddress string) error {
	dbRepo.GetDBHandle().Table("mct_cluster_configs").Where(
		"device_one_mgmt_ip = ? OR device_two_mgmt_ip = ?",
		IPAddress, IPAddress).Delete(database.MctClusterConfig{})
	return nil
}

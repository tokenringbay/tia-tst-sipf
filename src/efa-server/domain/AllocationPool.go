package domain

//ASNAllocationPool holds the ASN entries
//that are available in the pool to be allocated
type ASNAllocationPool struct {
	ID         uint
	FabricID   uint
	ASN        uint64
	DeviceRole string
}

//UsedASN holds the ASN entries that have
//been allocated to a Device
type UsedASN struct {
	ID         uint
	FabricID   uint
	DeviceID   uint
	ASN        uint64
	DeviceRole string
}

//IPAllocationPool holds the IP Address
//that are available for allocation
type IPAllocationPool struct {
	ID        uint
	FabricID  uint
	IPAddress string
	IPType    string
}

//UsedIP holds the IP Address
//that have been allocated to an Interface
type UsedIP struct {
	ID          uint
	FabricID    uint
	DeviceID    uint
	IPAddress   string
	IPType      string
	InterfaceID uint
}

//IPPairAllocationPool holds the pair of IPAddress
//that are available for allocation
type IPPairAllocationPool struct {
	ID           uint
	FabricID     uint
	IPAddressOne string
	IPAddressTwo string
	IPType       string
}

//UsedIPPair holds the pair of IPAddress
//that have been allocated to pair of interfaces
type UsedIPPair struct {
	ID             uint
	FabricID       uint
	DeviceOneID    uint
	DeviceTwoID    uint
	IPAddressOne   string
	IPAddressTwo   string
	IPType         string
	InterfaceOneID uint
	InterfaceTwoID uint
}

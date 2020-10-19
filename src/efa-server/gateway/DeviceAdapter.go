package gateway

import (
	"context"
	"efa-server/domain"
	"efa-server/gateway/appcontext"
	ada "efa-server/infra/device/adapter"
	"efa-server/infra/device/client"
	Interactor "efa-server/usecase/interactorinterface"
	nlog "github.com/sirupsen/logrus"
)

var log *nlog.Entry

func init() {
	log = nlog.WithFields(nlog.Fields{
		"Module": "FabricAdd",
	})
}

//DeviceAdapter is the reciever object for DeviceAdapter interfaces
type DeviceAdapter struct {
	client *client.NetconfClient
	detail domain.DeviceDetail
}

//DeviceAdapterFactory is a factory method to instantiate and return the DeviceAdapter
func DeviceAdapterFactory(ctx context.Context, IPAddress string, UserName string, Password string) (Interactor.DeviceAdapter, error) {
	client := &client.NetconfClient{Host: IPAddress, User: UserName, Password: Password}
	err := client.Login()
	if err != nil {
		return &DeviceAdapter{client: client}, err
	}
	detail, _ := ada.GetDeviceDetail(client)
	return &DeviceAdapter{client: client, detail: detail}, err
}

//CloseConnection closess the Connection to the Device
func (ad *DeviceAdapter) CloseConnection(ctx context.Context) error {
	LOG := appcontext.Logger(ctx)
	LOG.Infoln("Close Connection Called")
	return ad.client.Close()

}

//GetInterfaces fetches the Interfaces from the Device
func (ad *DeviceAdapter) GetInterfaces(FabricID uint, DeviceID uint, ControlVlan string) ([]domain.Interface, error) {

	adapter := ada.GetAdapter(ad.detail.Model)
	var Interfaces []domain.Interface
	SwitchInterfaces, _ := adapter.GetInterfaces(ad.client, ControlVlan)

	for _, Interface := range SwitchInterfaces {
		phy := domain.Interface{FabricID: FabricID, DeviceID: DeviceID,
			IntType: Interface.InterfaceType, IntName: Interface.InterfaceName, Mac: Interface.InterfaceMac,
			IPAddress: Interface.IPAddress, ConfigState: Interface.ConfigState,
			InterfaceSpeed: Interface.InterfaceSpeed}

		Interfaces = append(Interfaces, phy)
	}

	return Interfaces, nil
}

//EnableInterfaces enables the interfaces on the Device
func (ad *DeviceAdapter) EnableInterfaces(InterfaceNames []string) (string, error) {
	adapter := ada.GetAdapter(ad.detail.Model)
	return adapter.EnableInterfaces(ad.client, InterfaceNames)
}

//GetLLDPs fetches the LLDP data from the Device
func (ad *DeviceAdapter) GetLLDPs(FabricID uint, DeviceID uint) ([]domain.LLDP, error) {
	adapter := ada.GetAdapter(ad.detail.Model)
	var LLDPS []domain.LLDP
	Neighbors, _ := adapter.GetLLDPNeighbors(ad.client)
	for _, Neighbor := range Neighbors {
		lldp := domain.LLDP{FabricID: FabricID, DeviceID: DeviceID,
			LocalIntName: Neighbor.LocalInterfaceName, LocalIntMac: Neighbor.LocalInterfaceMac,
			RemoteIntName: Neighbor.RemoteInterfaceName, RemoteIntMac: Neighbor.RemoteInterfaceMac,
			LocalIntType: Neighbor.LocalInterfaceType, RemoteIntType: Neighbor.RemoteInterfaceType,
		}

		LLDPS = append(LLDPS, lldp)
	}
	return LLDPS, nil
}

//GetASN fetches the ASN for the Device
func (ad *DeviceAdapter) GetASN(FabricID uint, Device uint) (string, error) {
	adapter := ada.GetAdapter(ad.detail.Model)
	return adapter.GetLocalAsn(ad.client)
}

//GetDeviceDetail gets device details about the device
func (ad *DeviceAdapter) GetDeviceDetail(FabricID uint, DeviceID uint, DeviceIP string) (
	domain.DeviceDetail, error) {

	var err error
	ad.detail, err = ada.GetDeviceDetail(ad.client)
	ad.detail.DeviceID = DeviceID
	return ad.detail, err
}

//GetInterfaceSpeed fetches speed for interface
func (ad *DeviceAdapter) GetInterfaceSpeed(InterfaceType string, InterfaceName string) (int, error) {

	adapter := ada.GetAdapter(ad.detail.Model)
	InterfaceSpeed, err := adapter.GetInterfaceSpeed(ad.client, InterfaceType, InterfaceName)
	if err != nil {
		log.Error("Fetching Interface Speed Failed")
	}
	return InterfaceSpeed, err
}

//GetInterfaceVe fetches VE interface from Switch
func (ad *DeviceAdapter) GetInterfaceVe(name string) (map[string]string, error) {
	adapter := ada.GetAdapter(ad.detail.Model)
	InterfaceVE, err := adapter.GetInterfaceVe(ad.client, name)
	if err != nil {
		log.Errorf("Fetching Interface VE %s Failed", name)
	}
	return InterfaceVE, err
}

//GetClusterByName fetches "cluster" config, from the switch
func (ad *DeviceAdapter) GetClusterByName(name string) (map[string]string, error) {
	adapter := ada.GetAdapter(ad.detail.Model)
	Cluster, err := adapter.GetClusterByName(ad.client, name)
	if err != nil {
		log.Errorf("Fetching MCT Cluster %s Failed", name)
	}
	return Cluster, err

}

//GetInterfacePoMember fetches the member ports of a port-channel, from the switch
func (ad *DeviceAdapter) GetInterfacePoMember(name string) (map[string]string, error) {
	adapter := ada.GetAdapter(ad.detail.Model)
	Port, err := adapter.GetInterfacePoMember(ad.client, name)
	if err != nil {
		log.Errorf("Fetching PO Member %s Failed", name)
	}
	return Port, err

}

//GetSwitchHostName fetches host-name of the switches
func (ad *DeviceAdapter) GetSwitchHostName(DeviceIP string) (string, error) {

	adapter := ada.GetAdapter(ad.detail.Model)
	HostName, err := adapter.GetSwitchHostName(ad.client)
	if err != nil {
		log.Errorf("Fetching Switch HostName for %s Failed", DeviceIP)
	}
	return HostName, err
}

//CheckSupportedFirmware checks Firmware and model
func (ad *DeviceAdapter) CheckSupportedFirmware(DeviceIP string) error {

	return ada.CheckSupportedVersion(ad.detail.Model)

}

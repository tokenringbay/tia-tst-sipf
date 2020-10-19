package interfaces

//Switch provides collection of methods supported for Switch
type Switch interface {
	OverlayGateway
	LLDP
	BGP
	Interface
	Cluster
	System
}

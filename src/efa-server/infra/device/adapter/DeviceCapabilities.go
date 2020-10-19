package adapter

import (
	"context"
	"efa-server/gateway/appcontext"
	"fmt"
)

//CapabilityType for the devices
type CapabilityType int

const (
	//MCTCapability type for MCT feature
	MCTCapability CapabilityType = iota
)

//DeviceCapabilityMap holds the map of capabilities of different device types
var DeviceCapabilityMap map[string]map[CapabilityType]bool

func init() {
	DeviceCapabilityMap = make(map[string]map[CapabilityType]bool, 0)

	//For Device Avalance
	DeviceCapabilityMap[AvalancheType] = make(map[CapabilityType]bool, 0)
	DeviceCapabilityMap[AvalancheType][MCTCapability] = true

	//For Freedom
	DeviceCapabilityMap[FreedomType] = make(map[CapabilityType]bool, 0)
	DeviceCapabilityMap[FreedomType][MCTCapability] = true

	//For Cedar
	DeviceCapabilityMap[CedarType] = make(map[CapabilityType]bool, 0)
	DeviceCapabilityMap[CedarType][MCTCapability] = true

	//For Orca/OrcaT
	DeviceCapabilityMap[OrcaType] = make(map[CapabilityType]bool, 0)
	DeviceCapabilityMap[OrcaType][MCTCapability] = true
	DeviceCapabilityMap[OrcaTType] = make(map[CapabilityType]bool, 0)
	DeviceCapabilityMap[OrcaTType][MCTCapability] = true
}

//IsCapabilitySupported returns the capability of the device
func IsCapabilitySupported(ctx context.Context, capability CapabilityType, deviceType string) bool {
	LOG := appcontext.Logger(ctx)
	capabilityMap, ok := DeviceCapabilityMap[deviceType]
	if !ok {
		statusMsg := fmt.Sprintf("Capability map not defined for %s", deviceType)
		LOG.Errorln(statusMsg)
		return false
	}
	capabilityEnabled, ok := capabilityMap[capability]
	if !ok {
		statusMsg := fmt.Sprintf("Capability %d not defined for %s", capability, deviceType)
		LOG.Errorln(statusMsg)
		return false
	}

	return capabilityEnabled
}

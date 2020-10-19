package utils

import (
	"net"
	"regexp"
)

const (
	//CLOSFabricType represent CLOS Fabric
	CLOSFabricType = "clos"

	//NonCLOSFabricType represent CLOS Fabric
	NonCLOSFabricType = "non-clos"
)

//IsValidIP check if given IP is valid IPv4 or IPv6
func IsValidIP(ipaddress string) bool {
	ip := net.ParseIP(ipaddress)
	if ip == nil {
		return false
	}
	if isZeros(ip) {
		return false
	}
	return true
}

// Is p all zeros?
func isZeros(p net.IP) bool {
	if p.Equal(net.IPv4zero) || p.Equal(net.IPv6zero) {
		return true
	}
	return false
}

//IsValidIPs check if given IP is valid IPv4 or IPv6
func IsValidIPs(ipaddresses []string) bool {
	for _, ip := range ipaddresses {
		if !IsValidIP(ip) {
			return false
		}
	}
	return true
}

//IsValidHostName check if hostname is valid
func IsValidHostName(hostName string) bool {
	// Regular expression used to validate RFC1035 hostnames
	var hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)
	if !hostnameRegex.MatchString(hostName) {
		return false
	}
	return true
}

//IsIPinSubnet check if given ipaddress falls under the given network
func IsIPinSubnet(network string, ipaddress string) bool {
	_, subnet, _ := net.ParseCIDR(network)
	ip := net.ParseIP(ipaddress)
	if subnet.Contains(ip) {
		return true
	}
	return false
}

//IsValidMAC check if given MAC is valid
func IsValidMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	if err != nil {
		return false
	}
	return true
}

//IsValidCIDR check if given CIDR is valid
func IsValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return true
}

//IsValidDeviceRole check if given device role is valid
func IsValidDeviceRole(role string) bool {
	if role != "super-spine" && role != "spine" && role != "leaf" {
		return false
	}
	return true
}

//IsValidLeafType check if given leafType is valid
func IsValidLeafType(leafType string) bool {
	if leafType != "single-homed" && leafType != "multi-homed" {
		return false
	}
	return true
}

package handler

import (
	"net/http"

	"efa-server/domain"
	"efa-server/infra"
	"efa-server/infra/constants"
	"efa-server/infra/logging"
	Restmodel "efa-server/infra/rest/generated/server/go"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"
)

//UpdateFabricSettings is a REST handler which handles the Fabric Settings Update REST request
func UpdateFabricSettings(w http.ResponseWriter, r *http.Request) {
	constants.RestLock.Lock()
	defer constants.RestLock.Unlock()
	success := true
	statusMsg := ""

	var FabricSettings Restmodel.FabricSettings
	var FabricUpdate domain.FabricProperties

	alog := logging.AuditLog{Request: &logging.Request{Command: "Update Fabric Settings"}}
	ctx := alog.LogMessageInit()
	defer alog.LogMessageEnd(&success, &statusMsg)

	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &FabricSettings)
	if err != nil {
		success = false
		return
	}
	errMap := make(map[string]string, 0)
	FabricName := FabricSettings.Name
	for _, FabricParameter := range FabricSettings.Keyval {
		switch FabricParameter.Key {
		case "ConfigureOverlayGateway":
			FabricUpdate.ConfigureOverlayGateway = FabricParameter.Value
		case "MCTLinkIPRange":
			FabricUpdate.MCTLinkIPRange = FabricParameter.Value
		case "MCTL3LBIPRange":
			FabricUpdate.MCTL3LBIPRange = FabricParameter.Value
		case "ControlVlan":
			FabricUpdate.ControlVlan = FabricParameter.Value
		case "P2PIPType":
			FabricUpdate.P2PIPType = FabricParameter.Value
		case "MTU":
			FabricUpdate.MTU = FabricParameter.Value
		case "IPMTU":
			FabricUpdate.IPMTU = FabricParameter.Value
		case "SpineASNBlock":
			FabricUpdate.SpineASNBlock = FabricParameter.Value
		case "LeafASNBlock":
			FabricUpdate.LeafASNBlock = FabricParameter.Value
		case "RackASNBlock":
			FabricUpdate.RackASNBlock = FabricParameter.Value
		case "BGPMultiHop":
			FabricUpdate.BGPMultiHop = FabricParameter.Value
		case "MaxPaths":
			FabricUpdate.MaxPaths = FabricParameter.Value
		case "AllowASIn":
			FabricUpdate.AllowASIn = FabricParameter.Value
		case "LeafPeerGroup":
			FabricUpdate.LeafPeerGroup = FabricParameter.Value
		case "SpinePeerGroup":
			FabricUpdate.SpinePeerGroup = FabricParameter.Value
		case "P2PLinkRange":
			FabricUpdate.P2PLinkRange = FabricParameter.Value
		case "LoopBackIPRange":
			FabricUpdate.LoopBackIPRange = FabricParameter.Value
		case "LoopBackPortNumber":
			FabricUpdate.LoopBackPortNumber = FabricParameter.Value
		case "BFDEnable":
			FabricUpdate.BFDEnable = FabricParameter.Value
		case "BFDTx":
			FabricUpdate.BFDTx = FabricParameter.Value
		case "BFDRx":
			FabricUpdate.BFDRx = FabricParameter.Value
		case "BFDMultiplier":
			FabricUpdate.BFDMultiplier = FabricParameter.Value
		case "VTEPLoopBackPortNumber":
			FabricUpdate.VTEPLoopBackPortNumber = FabricParameter.Value
		case "VNIAutoMap":
			FabricUpdate.VNIAutoMap = FabricParameter.Value
		case "AnyCastMac":
			FabricUpdate.AnyCastMac = FabricParameter.Value
		case "IPV6AnyCastMac":
			FabricUpdate.IPV6AnyCastMac = FabricParameter.Value
		case "ArpAgingTimeout":
			FabricUpdate.ArpAgingTimeout = FabricParameter.Value
		case "MacAgingTimeout":
			FabricUpdate.MacAgingTimeout = FabricParameter.Value
		case "MacAgingConversationalTimeOut":
			FabricUpdate.MacAgingConversationalTimeout = FabricParameter.Value
		case "MacMoveLimit":
			FabricUpdate.MacMoveLimit = FabricParameter.Value
		case "DuplicateMacTimer":
			FabricUpdate.DuplicateMacTimer = FabricParameter.Value
		case "DuplicateMaxTimerMaxCount":
			FabricUpdate.DuplicateMaxTimerMaxCount = FabricParameter.Value
		case "MacAgingConversationalTimeout":
			FabricUpdate.MacAgingConversationalTimeout = FabricParameter.Value
		case "ControlVE":
			FabricUpdate.ControlVE = FabricParameter.Value
		case "MctPortChannel":
			FabricUpdate.MctPortChannel = FabricParameter.Value
		case "RoutingMctPortChannel":
			FabricUpdate.RoutingMctPortChannel = FabricParameter.Value
		case "FabricType":
			FabricUpdate.FabricType = FabricParameter.Value
			if FabricUpdate.FabricType == domain.NonCLOSFabricType {
				FabricUpdate.BGPMultiHop = "4"
				FabricUpdate.P2PIPType = "numbered"
				FabricUpdate.ConfigureOverlayGateway = "Yes"
			} else if FabricUpdate.FabricType == domain.CLOSFabricType {
				FabricUpdate.BGPMultiHop = "2"
			}
		case "RackPeerEBGPGroup":
			FabricUpdate.RackPeerEBGPGroup = FabricParameter.Value
		case "RackPeerOvgGroup":
			FabricUpdate.RackPeerOvgGroup = FabricParameter.Value
		default:
			errMap[FabricParameter.Key] = fmt.Sprintf("Invalid Parameter: %s", FabricParameter.Key)
		}
	}
	alog.LogMessageReceived()

	//update Request object after all parameters are received
	alog.Request.Params = map[string]interface{}{
		"FabricName":                    FabricName,
		"P2PLinkRange":                  FabricUpdate.P2PLinkRange,
		"P2PIPType":                     FabricUpdate.P2PIPType,
		"LoopBackIPRange":               FabricUpdate.LoopBackIPRange,
		"MCTLinkIPRange":                FabricUpdate.MCTLinkIPRange,
		"MCTL3LBIPRange":                FabricUpdate.MCTL3LBIPRange,
		"LoopBackPortNumber":            FabricUpdate.LoopBackPortNumber,
		"LeafAsnBlock":                  FabricUpdate.LeafASNBlock,
		"SpineAsnBlock":                 FabricUpdate.SpineASNBlock,
		"VTEPLoopBackPortNumber":        FabricUpdate.VTEPLoopBackPortNumber,
		"AnyCastMac":                    FabricUpdate.AnyCastMac,
		"IPV6AnyCastMac":                FabricUpdate.IPV6AnyCastMac,
		"ConfigureOverlayGateway":       FabricUpdate.ConfigureOverlayGateway,
		"VNIAutoMap":                    FabricUpdate.VNIAutoMap,
		"BFDEnable":                     FabricUpdate.BFDEnable,
		"BFDTx":                         FabricUpdate.BFDTx,
		"BFDRx":                         FabricUpdate.BFDRx,
		"BFDMultiplier":                 FabricUpdate.BFDMultiplier,
		"BGPMultiHop":                   FabricUpdate.BGPMultiHop,
		"MaxPaths":                      FabricUpdate.MaxPaths,
		"AllowASIn":                     FabricUpdate.AllowASIn,
		"MTU":                           FabricUpdate.MTU,
		"IPMTU":                         FabricUpdate.IPMTU,
		"LeafPeerGroup":                 FabricUpdate.LeafPeerGroup,
		"SpinePeerGroup":                FabricUpdate.SpinePeerGroup,
		"ArpAgingTimeout":               FabricUpdate.ArpAgingTimeout,
		"MacAgingTimeout":               FabricUpdate.MacAgingTimeout,
		"MacAgingConversationaltimeout": FabricUpdate.MacAgingConversationalTimeout,
		"MacMoveLimit":                  FabricUpdate.MacMoveLimit,
		"DuplicateMacTimer":             FabricUpdate.DuplicateMacTimer,
		"DuplicateMaxTimerMaxCount":     FabricUpdate.DuplicateMaxTimerMaxCount,
		"ControlVlan":                   FabricUpdate.ControlVlan,
		"ControlVE":                     FabricUpdate.ControlVE,
		"MctPortChannel":                FabricUpdate.MctPortChannel,
		"RoutingMctPortChannel":         FabricUpdate.RoutingMctPortChannel,
		"FabricType":                    FabricUpdate.FabricType,
		"RackPeerEBGPGroup":             FabricUpdate.RackPeerEBGPGroup,
		"RackPeerOvgGroup":              FabricUpdate.RackPeerOvgGroup,
	}

	alog.LogMessageReceived()
	if len(errMap) > 0 {
		fmt.Println("Fabric Update  Invalild Parameter")
		http.Error(w, "", http.StatusBadRequest)
		OpenAPIError := Restmodel.FabricdataErrorResponse{FabricName: FabricName, FabricSettings: errMap}
		bytess, _ := json.Marshal(&OpenAPIError)
		w.Write(bytess)
		return
	}

	errMap = ValidateFabricProperties(FabricName, &FabricUpdate)
	if len(errMap) != 0 {
		fmt.Println("Fabric Update Parameter Validation Failed")
		//ret, _ := json.MarshalIndent(errMap, "", "\n")
		http.Error(w, "", http.StatusBadRequest)
		OpenAPIError := Restmodel.FabricdataErrorResponse{FabricName: FabricName, FabricSettings: errMap}
		bytess, _ := json.Marshal(&OpenAPIError)
		w.Write(bytess)
		return
	}

	UseCaseInteractor := infra.GetUseCaseInteractor()
	ret, FabricID, err := UseCaseInteractor.UpdateFabricProperties(ctx, FabricName, &FabricUpdate)

	//Indicating there is an overall Failure
	if err != nil {
		switch err {
		case domain.ErrFabricActive:
			http.Error(w, "", http.StatusConflict)
		case domain.ErrFabricNotFound:
			http.Error(w, "", http.StatusNotFound)
		case domain.ErrFabricIncorrectValues:
			http.Error(w, "", http.StatusBadRequest)
		default:
			http.Error(w, "", http.StatusInternalServerError)
		}
		errMap[err.Error()] = ret
		OpenAPIError := Restmodel.FabricdataErrorResponse{FabricName: FabricName, FabricSettings: errMap}
		bytess, _ := json.Marshal(&OpenAPIError)
		w.Write(bytess)
	} else {
		//Send  Fabric config Show Response
		success = true
		statusMsg = fmt.Sprint("Fabric Update Succeeded.")
		OpenAPIResp := Restmodel.FabricdataResponse{
			FabricName: FabricName,
			FabricId:   int32(FabricID),
		}
		//Write Success Structure to the Body
		bytess, _ := json.Marshal(&OpenAPIResp)
		w.Write(bytess)
	}

}

func transformYesAndNo(data string) string {
	if strings.ToUpper(data) == "YES" {
		return "Yes"
	}
	if strings.ToUpper(data) == "NO" {
		return "No"
	}
	return data
}

//ValidateFabricProperties is used to validate the fabric-setting
func ValidateFabricProperties(FabricName string, FabricUpdateRequest *domain.FabricProperties) map[string]string {
	var e error
	err := make(map[string]string)

	_, e = validateFabricName(FabricName)
	if e != nil {
		err["fabric-name"] = e.Error()
	}

	FabricUpdateRequest.ConfigureOverlayGateway = transformYesAndNo(FabricUpdateRequest.ConfigureOverlayGateway)
	_, e = validateConfigureOverlayGateway(FabricUpdateRequest.ConfigureOverlayGateway)
	if e != nil {
		err["configure-overlay-gateway"] = e.Error()
	}

	_, e = validateMTU(FabricUpdateRequest.MTU)
	if e != nil {
		err["mtu"] = e.Error()
	}
	_, e = validateVlan(FabricUpdateRequest.ControlVlan)
	if e != nil {
		ret := fmt.Sprintf("%s is not Valid control VLAN , Valid control VLAN is 2-4090", FabricUpdateRequest.ControlVlan)
		err["control-vlan"] = ret
	}

	_, e = validateVlan(FabricUpdateRequest.ControlVE)
	if e != nil {
		ret := fmt.Sprintf("%s is not Valid control VE , Valid control VE is 2-4090", FabricUpdateRequest.ControlVE)
		err["control-ve"] = ret
	}

	_, e = validatePortChannel(FabricUpdateRequest.MctPortChannel)
	if e != nil {
		err["mct-port-channel"] = e.Error()
	}

	_, e = validateRoutingPortChannel(FabricUpdateRequest.RoutingMctPortChannel)
	if e != nil {
		err["routing-mct-port-channel"] = e.Error()
	}
	_, e = validateIPMTU(FabricUpdateRequest.IPMTU)
	if e != nil {
		err["ip-mtu"] = e.Error()
	}

	_, e = validateASN(FabricUpdateRequest.LeafASNBlock)
	if e != nil {
		err["leaf-asn-block"] = e.Error()
	}
	if e == nil {
		if len(FabricUpdateRequest.LeafASNBlock) > 0 {
			min, max := getASNMinMax(FabricUpdateRequest.LeafASNBlock)
			if min == max {
				err["leaf-asn-block"] = fmt.Sprintf("%s is not valid ASN Range , Valid ASN Range is 1-4294967295", FabricUpdateRequest.LeafASNBlock)
			}
		}
	}
	_, e = validateASN(FabricUpdateRequest.SpineASNBlock)
	if e != nil {
		err["spine-asn-block"] = e.Error()
	}

	_, e = validateASN(FabricUpdateRequest.RackASNBlock)
	if e != nil {
		err["rack-asn-block"] = e.Error()
	}
	if e == nil {
		if len(FabricUpdateRequest.RackASNBlock) > 0 {
			min, max := getASNMinMax(FabricUpdateRequest.RackASNBlock)
			if min == max {
				err["rack-asn-block"] = fmt.Sprintf("%s is not valid ASN Range , Valid ASN Range is 1-4294967295", FabricUpdateRequest.LeafASNBlock)
			}
		}
	}

	_, e = validateBgpMultiHop(FabricUpdateRequest.BGPMultiHop)
	if e != nil {
		err["bgp-multihop"] = e.Error()
	}

	_, e = validateBgpMaxPaths(FabricUpdateRequest.MaxPaths)
	if e != nil {
		err["bgp-maxpaths"] = e.Error()
	}

	_, e = validateBgpMaxPaths(FabricUpdateRequest.MaxPaths)
	if e != nil {
		err["bgp-maxpaths"] = e.Error()
	}

	_, e = validateAllowAsIn(FabricUpdateRequest.AllowASIn)
	if e != nil {
		err["allow-as-in"] = e.Error()
	}

	_, e = validatePeerGroupName(FabricUpdateRequest.LeafPeerGroup)
	if e != nil {
		err["leaf-peer-group-name"] = e.Error()
	}

	_, e = validatePeerGroupName(FabricUpdateRequest.SpinePeerGroup)
	if e != nil {
		err["spine-peer-group-name"] = e.Error()
	}

	_, e = validateIPRange(FabricUpdateRequest.P2PLinkRange)
	if e != nil {
		err["p2p-link-range"] = e.Error()
	}

	_, e = validateIPRange(FabricUpdateRequest.LoopBackIPRange)
	if e != nil {
		err["loopback-ip-range"] = e.Error()
	}
	_, e = validateIPRange(FabricUpdateRequest.MCTLinkIPRange)
	if e != nil {
		err["mctlink-ip-range"] = e.Error()
	}
	_, e = validateIPRange(FabricUpdateRequest.MCTL3LBIPRange)
	if e != nil {
		err["mctlink-l3-lb-range"] = e.Error()
	}

	_, e = validateP2PIPType(FabricUpdateRequest.P2PIPType)
	if e != nil {
		err["p2p-ip-type"] = e.Error()
	}

	_, e = validateLoopBackPortNumber("loopback", FabricUpdateRequest.LoopBackPortNumber)
	if e != nil {
		err["loopback-port-number"] = e.Error()
	}

	_, e = validateLoopBackPortNumber("VTEP loopback", FabricUpdateRequest.VTEPLoopBackPortNumber)
	if e != nil {
		err["vtep-loopback-port-number"] = e.Error()
	}

	if (FabricUpdateRequest.VTEPLoopBackPortNumber != "") && (FabricUpdateRequest.LoopBackPortNumber != "") &&
		(FabricUpdateRequest.VTEPLoopBackPortNumber == FabricUpdateRequest.LoopBackPortNumber) {
		err["loopback-port-number"] = "VTEP Loopback Number and Loopback Number cannot be same"
	}

	FabricUpdateRequest.BFDEnable = transformYesAndNo(FabricUpdateRequest.BFDEnable)
	_, e = validateBFDEnable(FabricUpdateRequest.BFDEnable)
	if e != nil {
		err["bfd-enable"] = e.Error()
	}
	_, e = validateBFDTx(FabricUpdateRequest.BFDTx)
	if e != nil {
		err["bfd-tx"] = e.Error()
	}

	_, e = validateBFDRx(FabricUpdateRequest.BFDRx)
	if e != nil {
		err["bfd-rx"] = e.Error()
	}
	_, e = validateBFDMultiplier(FabricUpdateRequest.BFDMultiplier)
	if e != nil {
		err["bfd-multiplier"] = e.Error()
	}
	FabricUpdateRequest.VNIAutoMap = transformYesAndNo(FabricUpdateRequest.VNIAutoMap)
	_, e = validateVNIAutoMap(FabricUpdateRequest.VNIAutoMap)
	if e != nil {
		err["vni-auto-map"] = e.Error()
	}
	_, e = validateAnyCastMacAddres(FabricUpdateRequest.AnyCastMac)
	if e != nil {
		err["anycast-mac"] = e.Error()
	}
	_, e = validateAnyCastMacAddres(FabricUpdateRequest.IPV6AnyCastMac)
	if e != nil {
		err["ipv6-anycast-mac"] = e.Error()
	}
	_, e = validateArpAgingTimeOut(FabricUpdateRequest.ArpAgingTimeout)
	if e != nil {
		err["arp-aging-timeout"] = e.Error()
	}
	_, e = validateMacAgingTimeOut(FabricUpdateRequest.MacAgingTimeout)
	if e != nil {
		err["mac-aging-timeout"] = e.Error()
	}
	_, e = validateMacAgingConversationalTimeOut(FabricUpdateRequest.MacAgingConversationalTimeout)
	if e != nil {
		err["mac-aging-conversation-timeout"] = e.Error()
	}
	_, e = validateMacMoveLimit(FabricUpdateRequest.MacMoveLimit)
	if e != nil {
		err["mac-move-limit"] = e.Error()
	}
	_, e = validateDuplicateMacTimer(FabricUpdateRequest.DuplicateMacTimer)
	if e != nil {
		err["duplicate-mac-timer-timeout"] = e.Error()
	}
	_, e = validateDuplicateMacTimerMaxCount(FabricUpdateRequest.DuplicateMaxTimerMaxCount)
	if e != nil {
		err["duplicate-mac-timer-max-count-timeout"] = e.Error()
	}
	FabricUpdateRequest.FabricType = cleanupString(FabricUpdateRequest.FabricType)
	_, e = validateFabricType(FabricUpdateRequest.FabricType)
	if e != nil {
		err["fabric-type"] = e.Error()
	}

	_, e = validatePeerGroupName(FabricUpdateRequest.RackPeerEBGPGroup)
	if e != nil {
		err["leaf-peer-ebgp-group"] = e.Error()
	}

	_, e = validatePeerGroupName(FabricUpdateRequest.RackPeerOvgGroup)
	if e != nil {
		err["leaf-peer-overlay-evpn-group"] = e.Error()
	}

	//TODO CALL VALIDATION FOR DuplicateMacTimer and DuplicateMaxTimerMaxCount
	return err
}

func isValidPeerGroupName(str string) bool {
	if len(str) > 63 {
		return false
	}
	//regular expression picked from switch yang
	ret := regexp.MustCompile(`^[a-zA-Z]{1}([-a-zA-Z0-9\\\\@#\+\*\(\)=\{~\}%<>=$_\[\]\|]{0,62})$`).MatchString(str)
	return ret
}

func isAlphaNumericWithUnderScore(str string) bool {
	ret := regexp.MustCompile("^[a-zA-Z0-9_]*$").MatchString(str)
	return ret
}

func isAlphaNumeric(str string) bool {
	ret := regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString(str)
	return ret
}

func cleanupString(str string) string {
	str = strings.TrimRight(str, " \n")
	str = strings.ToLower(str)
	return str
}

func isValidRange(data string, MinRange int64, MaxRange int64) bool {
	var val int64
	val, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return false
	}
	if val < MinRange || val > MaxRange {
		return false
	}
	return true
}

func validateFabricName(FabricName string) (bool, error) {
	FabricName = cleanupString(FabricName)
	if isAlphaNumericWithUnderScore(FabricName) {
		return true, nil
	}
	return false, errors.New("Fabric Name Should only contain Alpha Numeric Characters")
}

func validateConfigureOverlayGateway(ovg string) (bool, error) {
	if len(ovg) == 0 {
		return true, nil
	}
	ovg = cleanupString(ovg)
	if ovg == "yes" || ovg == "no" {
		return true, nil
	}
	ret := fmt.Sprintf("%s is not valid . Valid Value for configure-overlay-gateway is yes/no", ovg)
	return false, errors.New(ret)
}

func validateMTU(MTU string) (bool, error) {
	if len(MTU) == 0 {
		return true, nil
	}
	if !isValidRange(MTU, 1548, 9216) {
		ret := fmt.Sprintf("%s is not Valid MTU , Valid MTU is 1548-9216", MTU)
		return false, errors.New(ret)
	}
	return true, nil
}

func validatePortChannel(port string) (bool, error) {
	if len(port) == 0 {
		return true, nil
	}
	if !isValidRange(port, 1, 1024) {
		ret := fmt.Sprintf("%s is not Valid PORT-CHANNEL , Valid PORT_CHANNEL is 1-1024", port)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateRoutingPortChannel(port string) (bool, error) {
	if len(port) == 0 {
		return true, nil
	}
	if !isValidRange(port, 1, 64) {
		ret := fmt.Sprintf("%s is not Valid PORT-CHANNEL , Valid PORT_CHANNEL is 1-64", port)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateVlan(vlan string) (bool, error) {
	if len(vlan) == 0 {
		return true, nil
	}
	if !isValidRange(vlan, 2, 4090) {
		ret := fmt.Sprintf("%s is not Valid VLAN , Valid VLAN is 2-4090", vlan)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateIPMTU(IPMTU string) (bool, error) {
	if len(IPMTU) == 0 {
		return true, nil
	}
	if !isValidRange(IPMTU, 1300, 9194) {
		ret := fmt.Sprintf("%s is not Valid MTU , Valid MTU is 1300-9194", IPMTU)
		return false, errors.New(ret)
	}
	return true, nil
}

func getASNMinMax(asnRange string) (asnMin, asnMax uint64) {
	asn := strings.Split(asnRange, "-")
	asnMin, _ = strconv.ParseUint(asn[0], 10, 64)
	asnMax = asnMin
	if len(asn) == 2 {
		asnMax, _ = strconv.ParseUint(asn[1], 10, 64)
	}
	return
}

func validateASN(asn string) (bool, error) {
	if len(asn) == 0 {
		return true, nil
	}
	ValidAsn := regexp.MustCompile("[0-9]|[0-9]+(-[0-9])").MatchString
	if ValidAsn(asn) {
		min, max := getASNMinMax(asn)
		if min > max || min <= 0 || max == 0 || max > 4294967295 {
			ret := fmt.Sprintf("%s is not valid ASN Range , Valid ASN Range is 1-4294967295", asn)
			return false, errors.New(ret)
		}
	} else {
		ret := fmt.Sprintf("%s is not valid ASN Range , Valid ASN Range is 1-4294967295", asn)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateBgpMultiHop(MultiHop string) (bool, error) {
	if len(MultiHop) == 0 {
		return true, nil
	}
	if !isValidRange(MultiHop, 1, 255) {
		ret := fmt.Sprintf("%s is not Valid BgpMultiHop , Valid BGP MultiHop is 0-255", MultiHop)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateBgpMaxPaths(MaxPaths string) (bool, error) {
	if len(MaxPaths) == 0 {
		return true, nil
	}
	if !isValidRange(MaxPaths, 1, 64) {
		ret := fmt.Sprintf("%s is not Valid BgpMaxPaths , Valid BGP MaxPaths is 1-64", MaxPaths)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateAllowAsIn(AllowAsIn string) (bool, error) {
	if len(AllowAsIn) == 0 {
		return true, nil
	}
	if !isValidRange(AllowAsIn, 0, 10) {
		ret := fmt.Sprintf("%s is not Valid AllowAsIn , Valid BGP AllowAsIn is 0-10", AllowAsIn)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateIPRange(ip string) (bool, error) {
	if len(ip) == 0 {
		return true, nil
	}
	_, _, err := net.ParseCIDR(ip)
	if err != nil {
		ret := fmt.Sprintf("%s is not Valid IP Address , Valid IP in the format w.x.y.z/m", ip)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateP2PIPType(IPType string) (bool, error) {
	if len(IPType) == 0 {
		return true, nil
	}
	IPType = cleanupString(IPType)
	if IPType == "numbered" || IPType == "unnumbered" {
		return true, nil
	}
	ret := fmt.Sprintf("%s is not valid . Valid Value for p2p-type is numbered/unnumbered", IPType)
	return false, errors.New(ret)
}

func validateLoopBackPortNumber(loopbackType, port string) (bool, error) {
	if len(port) == 0 {
		return true, nil
	}
	if !isValidRange(port, 1, 255) {
		ret := fmt.Sprintf("%s is not Valid %s portnumber,Valid range is 1-255", port, loopbackType)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateBFDTx(BFDTx string) (bool, error) {
	if len(BFDTx) == 0 {
		return true, nil
	}
	if !isValidRange(BFDTx, 50, 30000) {
		ret := fmt.Sprintf("%s is not Valid BFDTx , Valid BGP BFDTx is 50-30000", BFDTx)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateBFDRx(BFDRx string) (bool, error) {
	if len(BFDRx) == 0 {
		return true, nil
	}
	if !isValidRange(BFDRx, 50, 30000) {
		ret := fmt.Sprintf("%s is not Valid BFDRx , Valid BGP BFDRx is 50-30000", BFDRx)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateBFDMultiplier(BFDMultiplier string) (bool, error) {
	if len(BFDMultiplier) == 0 {
		return true, nil
	}
	if !isValidRange(BFDMultiplier, 3, 50) {
		ret := fmt.Sprintf("%s is not Valid BFDMultiplier , Valid BGP BFDMultiplier is 3-50", BFDMultiplier)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateMacAddres(Mac string) (bool, error) {
	if len(Mac) == 0 {
		return true, nil
	}
	_, err := net.ParseMAC(Mac)
	if err != nil {
		ret := fmt.Sprintf("%s is Not Valid Mac Address, Valid MAC Address in the form HHHH.HHHH.HHHH", Mac)
		return false, errors.New(ret)
	}
	return true, nil
}

//This is internel funtion of net/parse.go hence copied here cant directly call their API
const big = 0xFFFFFF

func xtoi(s string) (n int, i int, ok bool) {
	n = 0
	for i = 0; i < len(s); i++ {
		if '0' <= s[i] && s[i] <= '9' {
			n *= 16
			n += int(s[i] - '0')
		} else if 'a' <= s[i] && s[i] <= 'f' {
			n *= 16
			n += int(s[i]-'a') + 10
		} else if 'A' <= s[i] && s[i] <= 'F' {
			n *= 16
			n += int(s[i]-'A') + 10
		} else {
			break
		}
		if n >= big {
			return 0, i, false
		}
	}
	if i == 0 {
		return 0, i, false
	}
	return n, i, true
}

func validateAnyCastMacAddres(Mac string) (bool, error) {
	if len(Mac) == 0 {
		return true, nil
	}
	_, e := validateMacAddres(Mac)
	if e != nil {
		ret := fmt.Sprintf("%s is Not Valid AnyCast Mac Address, Valid AnyCast MAC Address in the form HHHH.HHHH.HHHH.[0200.dea1.0001]", Mac)
		return false, errors.New(ret)
	}
	czero := func(x uint8) int {
		r, _, _ := xtoi(string(x))
		return r
	}
	flag := false
	for it := 0; it <= 6; it++ {
		if it == 4 {
			//skip dot
			continue
		}
		if czero(Mac[it]) != 0 {
			flag = true
			break
		}

	}
	if flag == false {
		ret := fmt.Sprintf("%s is Not Valid AnyCast Mac Address, Valid AnyCast MAC Address in the form HHHH.HHHH.HHHH.[0200.dea1.0001]", Mac)
		return false, errors.New(ret)
	}
	n, _, _ := xtoi(string(Mac[1]))
	if (n & 0x01) != 0 {
		//Multicast Address
		ret := fmt.Sprintf("%s is Not Valid AnyCast Mac Address, Valid AnyCast MAC Address in the form HHHH.HHHH.HHHH.[0200.dea1.0001]", Mac)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateVNIAutoMap(VNIAutoMap string) (bool, error) {
	if len(VNIAutoMap) == 0 {
		return true, nil
	}
	VNIAutoMap = cleanupString(VNIAutoMap)
	if VNIAutoMap == "yes" || VNIAutoMap == "no" {
		return true, nil
	}
	ret := fmt.Sprintf("%s is not valid . Valid Value for vni-auto-map is yes/no", VNIAutoMap)
	return false, errors.New(ret)
}

func validateArpAgingTimeOut(time string) (bool, error) {
	if len(time) == 0 {
		return true, nil
	}
	if !isValidRange(time, 60, 100000) {
		ret := fmt.Sprintf("%s is not Valid ARPAgingTimeOut, Valid Arp AgingTimeOut is 60-100000", time)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateMacAgingTimeOut(time string) (bool, error) {
	if len(time) == 0 {
		return true, nil
	}
	if !isValidRange(time, 60, 86400) {
		val, err := strconv.ParseInt(time, 10, 64)
		if err != nil {
			ret := fmt.Sprintf("%s is not Valid MACAgingTimeOut, Valid MACAgingTimeOut is 0|60-86400", time)
			return false, errors.New(ret)
		}
		if val == 0 {
			return true, nil
		}
		ret := fmt.Sprintf("%s is not Valid MACAgingTimeOut, Valid MACAgingTimeOut is 0|60-86400", time)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateMacAgingConversationalTimeOut(time string) (bool, error) {
	if len(time) == 0 {
		return true, nil
	}
	if !isValidRange(time, 60, 100000) {
		val, err := strconv.ParseInt(time, 10, 64)
		if err != nil {
			ret := fmt.Sprintf("%s is not Valid MACAgingConversationalTimeOut, Valid MACAgingConversationalTimeOut is 0|60-100000", time)
			return false, errors.New(ret)
		}
		if val == 0 {
			return true, nil
		}
		ret := fmt.Sprintf("%s is not Valid MACAgingConversationalTimeOut, Valid MACAgingConversationalTimeOut is 0|60-100000", time)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateDuplicateMacTimer(time string) (bool, error) {
	if len(time) == 0 {
		return true, nil
	}
	if !isValidRange(time, 5, 300) {
		ret := fmt.Sprintf("%s is not Valid DuplicateMacTimer, Valid ValidateDuplicateMacTimer is 5-300", time)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateDuplicateMacTimerMaxCount(count string) (bool, error) {
	if len(count) == 0 {
		return true, nil
	}
	if !isValidRange(count, 3, 10) {
		ret := fmt.Sprintf("%s is not Valid DuplicateMacTimerMaxCount, Valid ValidateDuplicateMacTimerMaxCount is 3-10", count)
		return false, errors.New(ret)
	}
	return true, nil
}

func validateMacMoveLimit(Count string) (bool, error) {
	if len(Count) == 0 {
		return true, nil
	}
	if !isValidRange(Count, 5, 500) {
		ret := fmt.Sprintf("%s is not Valid MacMoveLimit, Valid MacMoveLimit is 5-500", Count)
		return false, errors.New(ret)
	}
	return true, nil
}

func validatePeerGroupName(Peer string) (bool, error) {
	if len(Peer) == 0 {
		return true, nil
	}
	if isValidPeerGroupName(Peer) {
		return true, nil
	}
	ret := fmt.Sprintf("%s is not Valid Peer Group Name, Valid Peer Group Name is <WORD: 1-63>", Peer)
	return false, errors.New(ret)
}

func validateFabricType(FabricType string) (bool, error) {
	if len(FabricType) == 0 {
		return true, nil
	}
	FabricType = cleanupString(FabricType)
	if FabricType == domain.CLOSFabricType || FabricType == domain.NonCLOSFabricType {
		return true, nil
	}
	ret := fmt.Sprintf("%s is not Valid fabric type, Valid  fabric-type is <clos/non-clos>", FabricType)
	return false, errors.New(ret)
}

func validateBFDEnable(BFDEnable string) (bool, error) {
	if len(BFDEnable) == 0 {
		return true, nil
	}
	BFDEnable = cleanupString(BFDEnable)
	if BFDEnable == "yes" || BFDEnable == "no" {
		return true, nil
	}
	ret := fmt.Sprintf("%s is not valid . Valid Value for bfd-enable is yes/no", BFDEnable)
	return false, errors.New(ret)
}

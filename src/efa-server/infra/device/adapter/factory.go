package adapter

import (
	"efa-server/domain"
	"efa-server/infra/device/adapter/interface"
	"efa-server/infra/device/client"
)
import switchingbase "efa-server/infra/device/adapter/platform/slx/switching/base"
import cedarbase "efa-server/infra/device/adapter/platform/slx/switching/cedar/base"
import orcabase "efa-server/infra/device/adapter/platform/slx/routing/orca/base"
import slxbase "efa-server/infra/device/adapter/platform/slx/base"
import (
	avalanchebase "efa-server/infra/device/adapter/platform/slx/routing/avalanche/base"
	avalanche18r200 "efa-server/infra/device/adapter/platform/slx/routing/avalanche/fv18r200"
	"fmt"
	"regexp"
	"strings"
)

var adapterMap map[string]interfaces.Switch

//SwitchVersion holds the firmware version of the Switch
type SwitchVersion struct {
	//18r.1.01a
	Year  string //18r
	Major string //1
	Minor string //01
	Patch string //a
}

func (s *SwitchVersion) getPatch() string {
	return s.Year + "." + s.Major + "." + s.Minor + s.Patch
}
func (s *SwitchVersion) getMinor() string {
	return s.Year + "." + s.Major + "." + s.Minor
}
func (s *SwitchVersion) getMajor() string {
	return s.Year + "." + s.Major
}

//Holding Release variable
const (
	//SWR18R1 - Avalanche/Fusion 18r.1
	SWR18R1 = "18r.1" //Venus
	//SWR181 Version returned as 18.1 instead of 18r.1
	SWR181 = "18.1" //Venus

	//SWR18R2 - Avalanche/Fusion 18r.1
	SWR18R2 = "18r.2" //Puppis
	SWR182  = "18.2"  //Puppis

	//SWS17S1 - Cedar/Freedom 17s.1
	SWS17S1 = "17s.1" //Davinci

	//SWS171 - Cedar/Freedom 17.1 //Version returned as 17.1 instead of 17s.1
	SWS171 = "17.1" //Davinci

	//SWS18S1 - Cedar/Freedom 18s.1
	SWS18S1 = "18s.1" //Picasso
	SWS181  = "18.1"  //Certain firmware version's are returning 18.1 instead of 18s.1

	//SWX18X1 - Orca 18x.1
	SWX18X1 = "18x.1" //Pluto
	SWX181  = "18.1"  //Pluto //Certain firmware version's can return 18.1 instead of 18x.1

)

const (
	//AvalancheType type
	AvalancheType = "4000" //"BR-SLX9540"

	//FusionType type
	FusionType = "2000" //"BR-SLX9850"

	//CedarType type
	CedarType = "3000" //"BR-SLX9240"

	//FreedomType type
	FreedomType = "3001" //"BR-SLX9140"

	//OrcaType type
	OrcaType = "3006" //"EN-SLX-9030-48S"

	//OrcaTType type
	OrcaTType = "3007" //"EN-SLX-9030-48T"
)

var modelMap map[string]string

//TranslateModelString to Model
func TranslateModelString(modelVersion string) string {
	if strings.Contains(modelVersion, "_") {
		data := strings.Split(modelVersion, "_")
		Model := data[0]

		if val, ok := modelMap[Model]; ok {
			return val
		}
	}
	return modelVersion
}

func init() {
	adapterMap = make(map[string]interfaces.Switch, 0)
	modelMap = make(map[string]string, 0)
	modelMap["4000"] = "BR-SLX9540"
	modelMap["2000"] = "BR-SLX9850"
	modelMap["3000"] = "BR-SLX9240"
	modelMap["3001"] = "BR-SLX9140"
	modelMap["3006"] = "EN-SLX-9030-48S"
	modelMap["3007"] = "EN-SLX-9030-48T"

	//BASE Versions

	//Avalanche - Routing SLX
	adapterMap[version(AvalancheType, "base")] = &avalanchebase.SLXAvalancheBase{}

	//Orca - Routing SLX
	adapterMap[version(OrcaType, "base")] = &orcabase.SLXOrcaBase{}

	//OrcaT - Routing SLX
	adapterMap[version(OrcaTType, "base")] = &orcabase.SLXOrcaBase{}

	//Fusion - Routing SLX
	adapterMap[version(FusionType, "base")] = &switchingbase.SLXSwitchingBase{}

	//Cedar - Switching SLX
	adapterMap[version(CedarType, "base")] = &cedarbase.SLXCedarBase{}

	//Freedom - Switching SLX
	adapterMap[version(FreedomType, "base")] = &switchingbase.SLXSwitchingBase{}

	//Base Versions

	//SWR18R1 --- 18r.1 starts here  =============Avalanche==================
	//Avalanche - Routing SLX
	adapterMap[version(AvalancheType, SWR18R1)] = &avalanchebase.SLXAvalancheBase{}
	adapterMap[version(AvalancheType, SWR181)] = &avalanchebase.SLXAvalancheBase{}

	//Fusion - Routing SLX
	adapterMap[version(FusionType, SWR18R1)] = &switchingbase.SLXSwitchingBase{}
	adapterMap[version(FusionType, SWR181)] = &switchingbase.SLXSwitchingBase{}
	//SWR18R1 --- 18r.1 ends here============================================

	//SWR18R2 --- 18r.2 starts here  =============Avalanche==================
	//Avalanche - Routing SLX
	adapterMap[version(AvalancheType, SWR18R2)] = &avalanche18r200.SLXAvalancheFV18R200{}
	adapterMap[version(AvalancheType, SWR182)] = &avalanche18r200.SLXAvalancheFV18R200{}

	//Fusion - Routing SLX
	adapterMap[version(FusionType, SWR18R2)] = &switchingbase.SLXSwitchingBase{}
	adapterMap[version(FusionType, SWR182)] = &switchingbase.SLXSwitchingBase{}
	//SWR18R2 --- 18r.2 ends here============================================

	//SWX18X1 --- 18x.1 starts here  =============Orca==================
	//Orca - Routing SLX
	adapterMap[version(OrcaType, SWX18X1)] = &orcabase.SLXOrcaBase{}
	adapterMap[version(OrcaType, SWX181)] = &orcabase.SLXOrcaBase{}

	//OrcaT - Routing SLX
	adapterMap[version(OrcaTType, SWX18X1)] = &orcabase.SLXOrcaBase{}
	adapterMap[version(OrcaTType, SWX181)] = &orcabase.SLXOrcaBase{}

	//SWX18X1 --- 18x.1 starts here  ===========================================

	//SWS17S1 --- 17s.1 starts here ==============Switching===================
	//Cedar - Switching SLX
	adapterMap[version(CedarType, SWS17S1)] = &cedarbase.SLXCedarBase{}
	adapterMap[version(CedarType, SWS171)] = &cedarbase.SLXCedarBase{}

	//Freedom - Switching SLX
	adapterMap[version(FreedomType, SWS17S1)] = &switchingbase.SLXSwitchingBase{}
	adapterMap[version(FreedomType, SWS171)] = &switchingbase.SLXSwitchingBase{}
	//SWS17S1 --- 17s.1 ends here =============================================

	//SWS18S1 --- 18s.1 starts here ==============Switching===================
	//Cedar - Switching SLX
	adapterMap[version(CedarType, SWS18S1)] = &cedarbase.SLXCedarBase{}
	adapterMap[version(CedarType, SWS181)] = &cedarbase.SLXCedarBase{}

	//Freedom - Switching SLX
	adapterMap[version(FreedomType, SWS18S1)] = &switchingbase.SLXSwitchingBase{}
	adapterMap[version(FreedomType, SWS181)] = &switchingbase.SLXSwitchingBase{}
	//SWS18S1 --- 18s.1 ends here =============================================

}

func version(model string, ver string) string {
	return fmt.Sprint(model, "_", ver)
}

//GetAdapter provides the adapter based on the model
func GetAdapter(model string) interfaces.Switch {
	if strings.Contains(model, "_") {
		data := strings.Split(model, "_")
		ver := data[1]

		swVer := getParsedVersion(ver)

		//First Match on Patch
		key := version(data[0], swVer.getPatch())
		if val, ok := adapterMap[key]; ok {
			return val
		}
		//First Match on Minor
		key = version(data[0], swVer.getMinor())
		if val, ok := adapterMap[key]; ok {
			return val
		}
		//First Match on Major
		key = version(data[0], swVer.getMajor())
		if val, ok := adapterMap[key]; ok {
			return val
		}
	}
	//Default when the version is not present, upgrade from 18s1.0.1
	key := version(model, "base")
	if val, ok := adapterMap[key]; ok {
		return val
	}
	//For 17s* upgrade Release special handling for Cedar/Freedom
	if strings.Contains(model, "SLX9240") {
		return &cedarbase.SLXCedarBase{}
	}
	return &switchingbase.SLXSwitchingBase{}
}

//GetDeviceDetail provides the device detail for the device
func GetDeviceDetail(client *client.NetconfClient) (domain.DeviceDetail, error) {
	var DeviceDetail domain.DeviceDetail
	adapter := &slxbase.SLXBase{}
	resp, err := adapter.GetDeviceDetail(client)
	if err != nil {
		return DeviceDetail, err
	}

	DeviceDetail.FirmwareVersion = resp["firmware-full-version"]
	DeviceDetail.Model = resp["switch-type"] + "_" + resp["os-version"]

	return DeviceDetail, err
}

func getParsedVersion(version string) SwitchVersion {
	re := regexp.MustCompile(`(\d+[rsx]?).([0-9]+).([0-9]+)([a-z]*)`)
	match := re.FindStringSubmatch(version)
	ver := SwitchVersion{Year: match[1], Major: match[2], Minor: match[3], Patch: match[4]}
	return ver
}

//CheckSupportedVersion based on Model and Version
func CheckSupportedVersion(model string) error {
	if strings.Contains(model, "_") {
		data := strings.Split(model, "_")
		ver := data[1]
		re := regexp.MustCompile(`(\d+[rsx]?).([0-9]+).([0-9]+)([a-z]*)`)
		match := re.FindStringSubmatch(ver)
		if len(match) != 5 {
			return fmt.Errorf("unsupported model and firmware version %s", model)
		}

		swVer := SwitchVersion{Year: match[1], Major: match[2], Minor: match[3], Patch: match[4]}
		key := version(data[0], swVer.getPatch())
		if _, ok := adapterMap[key]; ok {
			return nil
		}
		//First Match on Minor
		key = version(data[0], swVer.getMinor())
		if _, ok := adapterMap[key]; ok {
			return nil
		}
		//First Match on Major
		key = version(data[0], swVer.getMajor())
		if _, ok := adapterMap[key]; ok {
			return nil
		}
	}
	return fmt.Errorf("unsupported model and firmware version %s", model)
}

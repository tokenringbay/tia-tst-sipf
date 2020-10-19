package interfaceswitchconfig

import (
	"efa-server/domain"
	"efa-server/infra/util"
	//"fmt"
)

//UpdateData sends both Old and New objects
type UpdateData struct {
	Old domain.InterfaceSwitchConfig
	New domain.InterfaceSwitchConfig
}

//Writing seperate Comparators for Each Object type for bettwr performance
type equalComparator func(first interface{}, second interface{}) (bool, domain.InterfaceSwitchConfig,
	domain.InterfaceSwitchConfig)

func list(s util.Interface) []domain.InterfaceSwitchConfig {
	m := s.GetMap()
	list := make([]domain.InterfaceSwitchConfig, 0, len(m))

	for _, item := range m {
		intf, _ := item.(domain.InterfaceSwitchConfig)
		list = append(list, intf)
	}

	return list
}

func populateSet(GetKey util.GetKey, domains []domain.InterfaceSwitchConfig) util.Interface {

	Set := util.NewSet(GetKey)
	for _, dom := range domains {
		Set.Add(dom)
	}
	return Set
}

//Compare is used to compare old and new InterfaceSwitchConfig and
// return "created", "updated" and "deleted" set of  InterfaceSwitchConfig
func Compare(GetKey util.GetKey, EqualMethod equalComparator, Old []domain.InterfaceSwitchConfig,
	New []domain.InterfaceSwitchConfig) ([]domain.InterfaceSwitchConfig, []domain.InterfaceSwitchConfig,
	[]UpdateData) {
	OldSet := populateSet(GetKey, Old)
	NewSet := populateSet(GetKey, New)

	//CreatedSet
	CreatedSet := util.Difference(NewSet, OldSet)

	//DeletedSet
	DeletedSet := util.Difference(OldSet, NewSet)

	//Updated sets
	CommonSet := util.Intersection(NewSet, OldSet)

	UpdateDomains := compareUpdatedSet(EqualMethod, OldSet, NewSet, CommonSet)

	return list(CreatedSet), list(DeletedSet), UpdateDomains
}

func compareUpdatedSet(EqualMethod equalComparator, OldSet util.Interface, NewSet util.Interface, CommonSet util.Interface) []UpdateData {
	commonMap := CommonSet.GetMap()
	list := make([]UpdateData, 0)

	oldMap := OldSet.GetMap()
	newMap := NewSet.GetMap()

	for key := range commonMap {
		oldValue := oldMap[key]
		newValue := newMap[key]
		//change in value
		if ok, oldDomain, newDomain := EqualMethod(oldValue, newValue); ok == false {
			newDomain.ID = oldDomain.ID
			list = append(list, UpdateData{oldDomain, newDomain})
		}
	}
	return list
}

package lldp

import (
	"efa-server/domain"
	"efa-server/infra/util"
	//"fmt"
)

//Writing seperate Comparators for Each Object type for bettwr performance
type equalComparator func(first interface{}, second interface{}) (bool, domain.LLDP, domain.LLDP)

func list(s util.Interface) []domain.LLDP {
	m := s.GetMap()
	list := make([]domain.LLDP, 0, len(m))

	for _, item := range m {
		intf, _ := item.(domain.LLDP)
		list = append(list, intf)
	}

	return list
}

func populateSet(GetKey util.GetKey, domains []domain.LLDP) util.Interface {

	Set := util.NewSet(GetKey)
	for _, dom := range domains {
		Set.Add(dom)
	}
	return Set
}

//Compare is used to compare old and new LLDP and
// return "created", "updated" and "deleted" set of  LLDP
func Compare(GetKey util.GetKey, EqualMethod equalComparator,
	Old []domain.LLDP, New []domain.LLDP) ([]domain.LLDP, []domain.LLDP, []domain.LLDP) {
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

func compareUpdatedSet(EqualMethod equalComparator, OldSet util.Interface, NewSet util.Interface, CommonSet util.Interface) []domain.LLDP {
	commonMap := CommonSet.GetMap()
	list := make([]domain.LLDP, 0)

	oldMap := OldSet.GetMap()
	newMap := NewSet.GetMap()

	for key := range commonMap {
		oldValue := oldMap[key]
		newValue := newMap[key]
		//change in value
		if ok, oldDomain, newDomain := EqualMethod(oldValue, newValue); ok == false {
			newDomain.ID = oldDomain.ID
			newDomain.DeviceID = oldDomain.DeviceID
			list = append(list, newDomain)
		}
	}
	return list
}

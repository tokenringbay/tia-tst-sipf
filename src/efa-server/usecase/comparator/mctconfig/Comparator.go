package mctconfig

import (
	"efa-server/domain"
	"efa-server/infra/util"
	//"fmt"
)

//Writing seperate Comparators for Each Object type for bettwr performance
type equalComparator func(first interface{}, second interface{}) (bool, domain.MCTMemberPorts,
	domain.MCTMemberPorts)

func list(s util.Interface) []domain.MCTMemberPorts {
	m := s.GetMap()
	list := make([]domain.MCTMemberPorts, 0, len(m))

	for _, item := range m {
		mct, _ := item.(domain.MCTMemberPorts)
		list = append(list, mct)
	}

	return list
}

func populateSet(GetKey util.GetKey, domains []domain.MCTMemberPorts) util.Interface {

	Set := util.NewSet(GetKey)
	for _, dom := range domains {
		Set.Add(dom)
	}
	return Set
}

//Compare is used to compare old and new MCTMemberPorts and
// return "created", "updated" and "deleted" set of  MCTMemberPorts
func Compare(GetKey util.GetKey, EqualMethod equalComparator,
	Old []domain.MCTMemberPorts, New []domain.MCTMemberPorts) ([]domain.MCTMemberPorts,
	[]domain.MCTMemberPorts, []domain.MCTMemberPorts) {
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

func compareUpdatedSet(EqualMethod equalComparator, OldSet util.Interface, NewSet util.Interface,
	CommonSet util.Interface) []domain.MCTMemberPorts {
	commonMap := CommonSet.GetMap()
	list := make([]domain.MCTMemberPorts, 0)

	oldMap := OldSet.GetMap()
	newMap := NewSet.GetMap()

	for key := range commonMap {
		oldValue := oldMap[key]
		newValue := newMap[key]
		//change in value
		EqualMethod(oldValue, newValue)
	}
	return list
}

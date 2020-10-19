package util

import (
	"fmt"
	"strings"
)

//GetKey describes a function to return a key
type GetKey func(interface{}) string

// Interface is describing a Set. Sets are an unordered, unique list of values.
type Interface interface {
	New(GetKey GetKey, items ...interface{}) Interface
	GetMap() map[string]interface{}
	Add(items ...interface{})
	Remove(items ...interface{})
	Pop() interface{}
	Has(items ...interface{}) bool
	Size() int
	Clear()
	IsEmpty() bool
	IsEqual(s Interface) bool
	IsSubset(s Interface) bool
	IsSuperset(s Interface) bool
	Each(func(interface{}) bool)
	String() string
	List() []interface{}
	Copy() Interface
	Merge(s Interface)
	Separate(s Interface)
}

// Union is the merger of multiple sets. It returns a new set with all the elements present in all the sets that are passed.
func Union(set1, set2 Interface, sets ...Interface) Interface {

	u := set1.Copy()

	set2.Each(func(item interface{}) bool {

		u.Add(item)
		return true
	})
	for _, set := range sets {
		set.Each(func(item interface{}) bool {
			u.Add(item)
			return true
		})
	}

	return u
}

// Difference returns a new set which contains items which are in in the first set but not in the others.
func Difference(set1, set2 Interface, sets ...Interface) Interface {
	s := set1.Copy()

	s.Separate(set2)

	for _, set := range sets {
		s.Separate(set) // seperate is thread safe
	}

	return s
}

//Intersection returns a new set which contains items that only exist in all given sets.
func Intersection(set1, set2 Interface, sets ...Interface) Interface {
	all := Union(set1, set2, sets...)
	result := Union(set1, set2, sets...)

	all.Each(func(item interface{}) bool {
		if !set1.Has(item) || !set2.Has(item) {
			result.Remove(item)
		}

		for _, set := range sets {
			if !set.Has(item) {
				result.Remove(item)
			}
		}
		return true
	})
	return result
}

//SymmetricDifference returns a new set which s is the difference of items which are in one of either, but not in both.
// i.e. union of 2 sets minus intersection of 2 sets
func SymmetricDifference(s Interface, t Interface) Interface {
	u := Difference(s, t)
	v := Difference(t, s)
	return Union(u, v)
}

//Set represents set data structure
type Set struct {
	m      map[string]interface{}
	getKey GetKey
}

// NewSet creates and initialize a new Set. It's accept a variable number of
// arguments to populate the initial set. If nothing passed a Set with zero
// size is created.
func NewSet(KeyMethod GetKey, items ...interface{}) *Set {
	s := &Set{}
	s.getKey = KeyMethod
	s.m = make(map[string]interface{})

	s.Add(items...)
	return s
}

// New creates and initalizes a new Set interface. It accepts a variable
// number of arguments to populate the initial set. If nothing is passed a
// zero size Set based on the struct is created.
func (s *Set) New(GetKey GetKey, items ...interface{}) Interface {
	return NewSet(GetKey, items...)
}

//GetMap gets the map for a given set.
func (s *Set) GetMap() map[string]interface{} {
	return s.m
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *Set) Add(items ...interface{}) {
	if len(items) == 0 {
		return
	}
	for _, item := range items {
		s.m[s.getKey(item)] = item
	}
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *Set) Remove(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		delete(s.m, s.getKey(item))
	}
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *Set) Pop() interface{} {
	for item := range s.m {
		delete(s.m, item)
		return item
	}
	return nil
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *Set) Has(items ...interface{}) bool {
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	has := true
	for _, item := range items {
		if _, has = s.m[s.getKey(item)]; !has {
			break
		}
	}
	return has
}

// Size returns the number of items in a set.
func (s *Set) Size() int {
	return len(s.m)
}

// Clear removes all items from the set.
func (s *Set) Clear() {
	s.m = make(map[string]interface{})
}

// IsEmpty reports whether the Set is empty.
func (s *Set) IsEmpty() bool {
	return s.Size() == 0
}

// IsEqual test whether s and t are the same in size and have the same items.
func (s *Set) IsEqual(t Interface) bool {

	// return false if they are no the same size
	if sameSize := len(s.m) == t.Size(); !sameSize {
		return false
	}

	equal := true
	t.Each(func(item interface{}) bool {
		_, equal = s.m[s.getKey(item)]
		return equal // if false, Each() will end
	})

	return equal
}

// IsSubset tests whether t is a subset of s.
func (s *Set) IsSubset(t Interface) (subset bool) {
	subset = true

	t.Each(func(item interface{}) bool {
		_, subset = s.m[s.getKey(item)]
		return subset
	})

	return
}

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s *Set) Each(f func(item interface{}) bool) {
	for _, item := range s.m {
		if !f(item) {
			break
		}
	}
}

// List returns a slice of all items.
func (s *Set) List() []interface{} {
	list := make([]interface{}, 0, len(s.m))

	for _, item := range s.m {
		list = append(list, item)
	}

	return list
}

// String returns a string representation of s
func (s *Set) String() string {
	t := make([]string, 0, len(s.List()))
	for _, item := range s.List() {
		t = append(t, fmt.Sprintf("%v", item))
	}

	return fmt.Sprintf("[%s]", strings.Join(t, ", "))
}

// Copy returns a new Set with a copy of s.
func (s *Set) Copy() Interface {
	return NewSet(s.getKey, s.List()...)
}

// IsSuperset tests whether t is a superset of s.
func (s *Set) IsSuperset(t Interface) bool {
	return t.IsSubset(s)
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
func (s *Set) Merge(t Interface) {
	t.Each(func(item interface{}) bool {
		s.m[s.getKey(item)] = item
		return true
	})
}

// Separate removes the set items containing in t from set s.
// Please be aware that it's not the opposite of Merge.
func (s *Set) Separate(t Interface) {
	s.Remove(t.List()...)
}

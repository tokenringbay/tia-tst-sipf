package usecase

import (
	"efa-server/infra/util"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestStruct struct {
	x string
	y uint
	z string
}

func TestStringSet_Create(t *testing.T) {
	GetKey := func(data interface{}) string {
		x, _ := data.(string)
		return x
	}
	s := util.NewSet(GetKey, "1", "2", "3")
	r := util.NewSet(GetKey, "3", "4", "5")

	assert.Equal(t, s.Size(), 3, "")
	assert.Equal(t, s.Has("1"), true)
	assert.Equal(t, s.Has("2"), true)
	assert.Equal(t, s.Has("3"), true)

	assert.Equal(t, r.Size(), 3, "")
	assert.Equal(t, r.Has("3"), true)
	assert.Equal(t, r.Has("4"), true)
	assert.Equal(t, r.Has("5"), true)
}

func TestStringSet_Union(t *testing.T) {
	GetKey := func(data interface{}) string {
		x, _ := data.(string)
		return x
	}
	s := util.NewSet(GetKey, "1", "2", "3")
	r := util.NewSet(GetKey, "3", "4", "5")
	u := util.Union(s, r)

	assert.Equal(t, u.Size(), 5, "")
	assert.Equal(t, u.Has("1"), true)
	assert.Equal(t, u.Has("2"), true)

	assert.Equal(t, u.Has("3"), true)
	assert.Equal(t, u.Has("4"), true)
	assert.Equal(t, u.Has("5"), true)

}

func TestStringSet_Intersect(t *testing.T) {
	GetKey := func(data interface{}) string {
		x, _ := data.(string)
		return x
	}
	s := util.NewSet(GetKey, "1", "2", "3")
	r := util.NewSet(GetKey, "3", "4", "5")
	u := util.Intersection(s, r)

	assert.Equal(t, u.Size(), 1, "")

	assert.Equal(t, u.Has("3"), true)

}

func TestStringSet_Difference(t *testing.T) {
	GetKey := func(data interface{}) string {
		x, _ := data.(string)
		return x
	}
	s := util.NewSet(GetKey, "1", "2", "3")
	r := util.NewSet(GetKey, "3", "4", "5")
	u := util.Difference(s, r)
	m := util.Difference(r, s)

	assert.Equal(t, u.Size(), 2, "")
	assert.Equal(t, u.Has("1"), true)
	assert.Equal(t, u.Has("2"), true)

	assert.Equal(t, m.Size(), 2, "")
	assert.Equal(t, m.Has("4"), true)
	assert.Equal(t, m.Has("5"), true)

}

func TestStructSet_Create(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}
	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"3", 3, "300"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "300"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"})

	assert.Equal(t, s.Size(), 3, "")
	assert.Equal(t, s.Has(TestStruct{"1", 1, "100"}), true)
	assert.Equal(t, s.Has(TestStruct{"2", 2, "200"}), true)
	assert.Equal(t, s.Has(TestStruct{"3", 3, "300"}), true)

	assert.Equal(t, r.Size(), 3, "")
	assert.Equal(t, r.Has(TestStruct{"3", 3, "300"}), true)
	assert.Equal(t, r.Has(TestStruct{"4", 4, "400"}), true)
	assert.Equal(t, r.Has(TestStruct{"5", 5, "500"}), true)

}

func TestStructSet_Union(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}
	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"3", 3, "300"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "300"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"})
	u := util.Union(s, r)

	assert.Equal(t, u.Size(), 5, "")
	assert.Equal(t, u.Has(TestStruct{"1", 1, "100"}), true)
	assert.Equal(t, u.Has(TestStruct{"2", 2, "200"}), true)
	assert.Equal(t, u.Has(TestStruct{"3", 3, "300"}), true)

	assert.Equal(t, u.Has(TestStruct{"3", 3, "300"}), true)
	assert.Equal(t, u.Has(TestStruct{"4", 4, "400"}), true)
	assert.Equal(t, u.Has(TestStruct{"5", 5, "500"}), true)

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"})
	r = util.NewSet(GetKey, TestStruct{"3", 3, "300"})
	q := util.NewSet(GetKey, TestStruct{"2", 2, "200"})
	u = util.Union(q, r, s)
	assert.Equal(t, u.Has(TestStruct{"3", 3, "300"}), true)
	assert.Equal(t, u.Has(TestStruct{"1", 1, "100"}), true)
	assert.Equal(t, u.Has(TestStruct{"2", 2, "200"}), true)

}

func TestStructSet_Difference(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}
	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"3", 3, "300"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "300"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"})
	q := util.NewSet(GetKey, TestStruct{"2", 2, "200"})
	u := util.Difference(s, r, q)

	assert.Equal(t, u.Size(), 1, "")
	assert.Equal(t, u.Has(TestStruct{"1", 1, "100"}), true)
	assert.Equal(t, !u.Has(TestStruct{"2", 2, "200"}), true)
	assert.Equal(t, !u.Has(TestStruct{"3", 3, "300"}), true)
}

func TestStructSet_Intersect(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}
	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"3", 3, "300"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "300"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"})
	q := util.NewSet(GetKey, TestStruct{"3", 3, "300"}, TestStruct{"6", 6, "600"}, TestStruct{"5", 5, "500"})
	u := util.Intersection(s, r, q)

	assert.Equal(t, u.Size(), 1, "")
	assert.Equal(t, u.Has(TestStruct{"3", 3, "300"}), true)

}

func TestStructSet_IntersectSameKey(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}
	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"3", 3, "300"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "3000"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"})
	u := util.Intersection(s, r)
	x := util.Intersection(r, s)

	assert.Equal(t, u.Size(), 1, "")
	assert.Equal(t, u.Has(TestStruct{"3", 3, "300"}), true)

	assert.Equal(t, x.Size(), 1, "")
	assert.Equal(t, x.Has(TestStruct{"3", 3, "3000"}), true)

}

func TestStructSet_SymmetricDifference(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}
	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"3", 3, "300"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "3000"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"})
	u := util.SymmetricDifference(s, r)

	fmt.Println(u)
	assert.Equal(t, 4, u.Size(), "")

	if !u.Has(TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"}, TestStruct{"4", 4, "400"}, TestStruct{"5", 5, "500"}) {
		assert.Fail(t, "The set returned by SymmetricDifference missing an expected item")
	}

	if u.Has(TestStruct{"3", 3, "3000"}) {
		assert.Fail(t, "The set returned by SymmetricDifference contains an unexpected item")
	}
}

func TestStructSet_IsEqual(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r := util.NewSet(GetKey, TestStruct{"3", 3, "3000"}, TestStruct{"4", 4, "400"})
	assert.False(t, s.IsEqual(r))

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"1", 1, "100"})
	assert.False(t, s.IsEqual(r))

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"1", 1, "3000"}, TestStruct{"2", 2, "400"})
	assert.True(t, s.IsEqual(r))
}

func TestStructSet_IsEmpty(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r := util.NewSet(GetKey)
	assert.False(t, s.IsEmpty())
	assert.True(t, r.IsEmpty())

}

func TestStructSet_Clear(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	assert.Equal(t, 2, s.Size())
	s.Clear()
	assert.Equal(t, 0, s.Size())

}

func TestStructSet_Pop(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprint(s.x, s.y)
	}

	//Create a set of size 2.
	s := util.NewSet(GetKey, TestStruct{"2", 2, "100"}, TestStruct{"1", 1, "200"})
	t1 := s.Pop()
	assert.Equal(t, 1, s.Size()) // After popping 1 item, set size should be 1.

	// Since map doesnt guarantee any order, evaluating the first pop item and then trying to add "expected" value in second pop
	if "22" == t1 {
		assert.Equal(t, "11", s.Pop())
	} else {
		assert.Equal(t, "22", s.Pop())
	}

	t2 := s.Pop()
	assert.Equal(t, nil, t2)
}

func TestStructSet_IsSubset(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r := util.NewSet(GetKey, TestStruct{"1", 1, "200"})
	assert.True(t, s.IsSubset(r))

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"3", 4, "100"})
	assert.False(t, s.IsSubset(r))

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"1", 1, "3000"}, TestStruct{"2", 2, "400"})
	assert.True(t, s.IsSubset(r))
}

func TestStructSet_IsSuperset(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r := util.NewSet(GetKey, TestStruct{"1", 1, "200"})
	assert.True(t, r.IsSuperset(s))

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"3", 4, "100"})
	assert.False(t, r.IsSuperset(s))

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"1", 1, "3000"}, TestStruct{"2", 2, "400"})
	assert.True(t, r.IsSuperset(s))
}

func TestStructSet_Merge(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r := util.NewSet(GetKey, TestStruct{"1", 1, "200"})
	s.Merge(r)
	assert.Equal(t, 2, s.Size())

	s = util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r = util.NewSet(GetKey, TestStruct{"3", 4, "100"})
	s.Merge(r)
	assert.Equal(t, 3, s.Size())

}

func TestStructSet_New(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	r := s.New(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	assert.Equal(t, 2, r.Size())
}

func TestStructSet_Has(t *testing.T) {
	GetKey := func(data interface{}) string {
		s, _ := data.(TestStruct)
		return fmt.Sprintln(s.x, s.y)
	}

	s := util.NewSet(GetKey, TestStruct{"1", 1, "100"}, TestStruct{"2", 2, "200"})
	assert.True(t, s.Has(TestStruct{"1", 1, "100"}))
	assert.False(t, s.Has())
}

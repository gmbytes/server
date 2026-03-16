package container

import (
	"testing"
)

func TestSet_Basic(t *testing.T) {
	s := NewSet[string]()

	// Test Add
	s.Add("a")
	s.Add("b")
	s.Add("c")

	if s.Len() != 3 {
		t.Errorf("Expected len=3, got %d", s.Len())
	}

	// Test Contains
	if !s.Contains("a") {
		t.Error("Expected Contains(a)=true")
	}

	if s.Contains("d") {
		t.Error("Expected Contains(d)=false")
	}

	// Test duplicate add
	s.Add("a")
	if s.Len() != 3 {
		t.Errorf("Expected len=3 after duplicate add, got %d", s.Len())
	}
}

func TestSet_Remove(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 2, 3, 4, 5)

	if s.Len() != 5 {
		t.Errorf("Expected len=5, got %d", s.Len())
	}

	// Remove existing element
	if !s.Remove(3) {
		t.Error("Expected Remove(3)=true")
	}

	if s.Len() != 4 {
		t.Errorf("Expected len=4 after remove, got %d", s.Len())
	}

	if s.Contains(3) {
		t.Error("Expected 3 to be removed")
	}

	// Remove non-existing element
	if s.Remove(10) {
		t.Error("Expected Remove(10)=false")
	}
}

func TestSet_Union(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2, 3)

	s2 := NewSet[int]()
	s2.AddAll(3, 4, 5)

	union := s1.Union(s2)

	if union.Len() != 5 {
		t.Errorf("Expected union len=5, got %d", union.Len())
	}

	for i := 1; i <= 5; i++ {
		if !union.Contains(i) {
			t.Errorf("Expected union to contain %d", i)
		}
	}
}

func TestSet_Intersection(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2, 3, 4)

	s2 := NewSet[int]()
	s2.AddAll(3, 4, 5, 6)

	intersection := s1.Intersection(s2)

	if intersection.Len() != 2 {
		t.Errorf("Expected intersection len=2, got %d", intersection.Len())
	}

	if !intersection.Contains(3) || !intersection.Contains(4) {
		t.Error("Expected intersection to contain 3 and 4")
	}
}

func TestSet_Difference(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2, 3, 4)

	s2 := NewSet[int]()
	s2.AddAll(3, 4, 5, 6)

	diff := s1.Difference(s2)

	if diff.Len() != 2 {
		t.Errorf("Expected difference len=2, got %d", diff.Len())
	}

	if !diff.Contains(1) || !diff.Contains(2) {
		t.Error("Expected difference to contain 1 and 2")
	}

	if diff.Contains(3) || diff.Contains(4) {
		t.Error("Expected difference not to contain 3 and 4")
	}
}

func TestSet_SymmetricDifference(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2, 3)

	s2 := NewSet[int]()
	s2.AddAll(3, 4, 5)

	symDiff := s1.SymmetricDifference(s2)

	if symDiff.Len() != 4 {
		t.Errorf("Expected symmetric difference len=4, got %d", symDiff.Len())
	}

	if !symDiff.Contains(1) || !symDiff.Contains(2) || !symDiff.Contains(4) || !symDiff.Contains(5) {
		t.Error("Expected symmetric difference to contain 1, 2, 4, 5")
	}

	if symDiff.Contains(3) {
		t.Error("Expected symmetric difference not to contain 3")
	}
}

func TestSet_IsSubset(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2)

	s2 := NewSet[int]()
	s2.AddAll(1, 2, 3, 4)

	if !s1.IsSubset(s2) {
		t.Error("Expected s1 to be subset of s2")
	}

	if s2.IsSubset(s1) {
		t.Error("Expected s2 not to be subset of s1")
	}
}

func TestSet_IsSuperset(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2, 3, 4)

	s2 := NewSet[int]()
	s2.AddAll(1, 2)

	if !s1.IsSuperset(s2) {
		t.Error("Expected s1 to be superset of s2")
	}

	if s2.IsSuperset(s1) {
		t.Error("Expected s2 not to be superset of s1")
	}
}

func TestSet_Equals(t *testing.T) {
	s1 := NewSet[int]()
	s1.AddAll(1, 2, 3)

	s2 := NewSet[int]()
	s2.AddAll(3, 2, 1)

	if !s1.Equals(s2) {
		t.Error("Expected s1 to equal s2")
	}

	s2.Add(4)
	if s1.Equals(s2) {
		t.Error("Expected s1 not to equal s2 after adding element")
	}
}

func TestSet_Clone(t *testing.T) {
	s := NewSet[string]()
	s.AddAll("a", "b", "c")

	cloned := s.Clone()

	if !cloned.Equals(s) {
		t.Error("Expected cloned set to equal original")
	}

	// Modify original
	s.Add("d")

	if cloned.Contains("d") {
		t.Error("Expected cloned set not to be affected by original modification")
	}
}

func TestSet_Filter(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 2, 3, 4, 5, 6)

	// Filter even numbers
	evens := s.Filter(func(item int) bool {
		return item%2 == 0
	})

	if evens.Len() != 3 {
		t.Errorf("Expected filtered len=3, got %d", evens.Len())
	}

	if !evens.Contains(2) || !evens.Contains(4) || !evens.Contains(6) {
		t.Error("Expected filtered set to contain 2, 4, 6")
	}
}

func TestSet_Any(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 3, 5, 7)

	// Check if any even number exists
	hasEven := s.Any(func(item int) bool {
		return item%2 == 0
	})

	if hasEven {
		t.Error("Expected no even numbers")
	}

	s.Add(4)
	hasEven = s.Any(func(item int) bool {
		return item%2 == 0
	})

	if !hasEven {
		t.Error("Expected to find even number")
	}
}

func TestSet_All(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(2, 4, 6, 8)

	// Check if all are even
	allEven := s.All(func(item int) bool {
		return item%2 == 0
	})

	if !allEven {
		t.Error("Expected all numbers to be even")
	}

	s.Add(3)
	allEven = s.All(func(item int) bool {
		return item%2 == 0
	})

	if allEven {
		t.Error("Expected not all numbers to be even")
	}
}

func TestSet_ToSlice(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 2, 3)

	slice := s.ToSlice()

	if len(slice) != 3 {
		t.Errorf("Expected slice len=3, got %d", len(slice))
	}

	// Check all elements are present
	found := make(map[int]bool)
	for _, item := range slice {
		found[item] = true
	}

	for i := 1; i <= 3; i++ {
		if !found[i] {
			t.Errorf("Expected slice to contain %d", i)
		}
	}
}

func TestSet_FromSlice(t *testing.T) {
	slice := []int{1, 2, 3, 2, 1}
	s := FromSlice(slice)

	if s.Len() != 3 {
		t.Errorf("Expected set len=3, got %d", s.Len())
	}

	if !s.Contains(1) || !s.Contains(2) || !s.Contains(3) {
		t.Error("Expected set to contain 1, 2, 3")
	}
}

func TestSet_Clear(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 2, 3)

	s.Clear()

	if !s.IsEmpty() {
		t.Error("Expected set to be empty after clear")
	}

	if s.Len() != 0 {
		t.Errorf("Expected len=0 after clear, got %d", s.Len())
	}
}

func TestSet_ForEach(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 2, 3)

	sum := 0
	s.ForEach(func(item int) {
		sum += item
	})

	if sum != 6 {
		t.Errorf("Expected sum=6, got %d", sum)
	}
}

func TestSet_ForEachBreakable(t *testing.T) {
	s := NewSet[int]()
	s.AddAll(1, 2, 3, 4, 5)

	count := 0
	s.ForEachBreakable(func(item int) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Errorf("Expected count=3, got %d", count)
	}
}

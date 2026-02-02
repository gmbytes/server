package container

// Set 集合（基于 map 实现，无锁设计）
// 注意：非线程安全，需要外部同步
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet 创建集合
func NewSet[T comparable](capacity ...int) *Set[T] {
	if len(capacity) > 0 && capacity[0] > 0 {
		return &Set[T]{
			items: make(map[T]struct{}, capacity[0]),
		}
	}
	return &Set[T]{
		items: make(map[T]struct{}),
	}
}

// Add 添加元素
func (s *Set[T]) Add(item T) {
	s.items[item] = struct{}{}
}

// AddAll 批量添加元素
func (s *Set[T]) AddAll(items ...T) {
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

// Remove 删除元素
func (s *Set[T]) Remove(item T) bool {
	if _, exists := s.items[item]; exists {
		delete(s.items, item)
		return true
	}
	return false
}

// Contains 检查元素是否存在
func (s *Set[T]) Contains(item T) bool {
	_, exists := s.items[item]
	return exists
}

// Len 返回元素数量
func (s *Set[T]) Len() int {
	return len(s.items)
}

// IsEmpty 检查集合是否为空
func (s *Set[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Clear 清空集合
func (s *Set[T]) Clear() {
	s.items = make(map[T]struct{})
}

// ToSlice 转换为切片
func (s *Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// ForEach 遍历所有元素
func (s *Set[T]) ForEach(fn func(item T)) {
	for item := range s.items {
		fn(item)
	}
}

// ForEachBreakable 遍历所有元素，支持中断
// 返回 false 可以中断遍历
func (s *Set[T]) ForEachBreakable(fn func(item T) bool) {
	for item := range s.items {
		if !fn(item) {
			break
		}
	}
}

// Clone 克隆集合
func (s *Set[T]) Clone() *Set[T] {
	newSet := NewSet[T](len(s.items))
	for item := range s.items {
		newSet.items[item] = struct{}{}
	}
	return newSet
}

// Union 并集（返回新集合）
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := s.Clone()
	for item := range other.items {
		result.items[item] = struct{}{}
	}
	return result
}

// Intersection 交集（返回新集合）
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	// 遍历较小的集合以提高性能
	smaller, larger := s, other
	if len(other.items) < len(s.items) {
		smaller, larger = other, s
	}
	for item := range smaller.items {
		if larger.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// Difference 差集（返回新集合，包含在 s 中但不在 other 中的元素）
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// SymmetricDifference 对称差集（返回新集合，包含只在其中一个集合中的元素）
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	for item := range other.items {
		if !s.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// IsSubset 检查是否为子集
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	if len(s.items) > len(other.items) {
		return false
	}
	for item := range s.items {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// IsSuperset 检查是否为超集
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	return other.IsSubset(s)
}

// Equals 检查两个集合是否相等
func (s *Set[T]) Equals(other *Set[T]) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for item := range s.items {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// Filter 过滤元素，返回新集合
func (s *Set[T]) Filter(fn func(item T) bool) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if fn(item) {
			result.Add(item)
		}
	}
	return result
}

// Any 检查是否存在满足条件的元素
func (s *Set[T]) Any(fn func(item T) bool) bool {
	for item := range s.items {
		if fn(item) {
			return true
		}
	}
	return false
}

// All 检查是否所有元素都满足条件
func (s *Set[T]) All(fn func(item T) bool) bool {
	for item := range s.items {
		if !fn(item) {
			return false
		}
	}
	return true
}

// FromSlice 从切片创建集合
func FromSlice[T comparable](items []T) *Set[T] {
	s := NewSet[T](len(items))
	for _, item := range items {
		s.Add(item)
	}
	return s
}

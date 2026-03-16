package container

// LMap 有序Map（基于 slice + map 实现，内存占用小，无锁设计）
// map 存储 key -> slice index 的映射
// slice 按插入顺序存储实际的值
// 注意：非线程安全，需要外部同步
type LMap[K comparable, V any] struct {
	items   []V       // 按插入顺序存储的值
	indices map[K]int // key -> items 中的索引
}

// NewLMap 创建有序Map
func NewLMap[K comparable, V any](capacity ...int) *LMap[K, V] {
	if len(capacity) > 0 && capacity[0] > 0 {
		return &LMap[K, V]{
			items:   make([]V, 0, capacity[0]),
			indices: make(map[K]int, capacity[0]),
		}
	}
	return &LMap[K, V]{
		items:   make([]V, 0),
		indices: make(map[K]int),
	}
}

// Set 设置键值对
func (m *LMap[K, V]) Set(key K, value V) {
	if idx, exists := m.indices[key]; exists {
		// 键已存在，更新值
		m.items[idx] = value
		return
	}

	// 添加新键值对
	idx := len(m.items)
	m.items = append(m.items, value)
	m.indices[key] = idx
}

// Get 获取值
func (m *LMap[K, V]) Get(key K) (V, bool) {
	if idx, exists := m.indices[key]; exists {
		return m.items[idx], true
	}

	var zero V
	return zero, false
}

// Delete 删除键值对
// 注意：删除操作会改变最后一个元素的位置，但不影响整体的插入顺序
func (m *LMap[K, V]) Delete(key K) bool {
	idx, exists := m.indices[key]
	if !exists {
		return false
	}

	// 删除策略：将最后一个元素移到被删除的位置
	lastIdx := len(m.items) - 1

	if idx != lastIdx {
		// 将最后一个元素移到 idx 位置
		m.items[idx] = m.items[lastIdx]

		// 需要找到最后一个元素对应的 key 并更新其索引
		// 遍历 indices 找到值为 lastIdx 的 key
		for k, i := range m.indices {
			if i == lastIdx {
				m.indices[k] = idx
				break
			}
		}
	}

	// 删除最后一个元素
	m.items = m.items[:lastIdx]
	delete(m.indices, key)

	return true
}

// Has 检查键是否存在
func (m *LMap[K, V]) Has(key K) bool {
	_, exists := m.indices[key]
	return exists
}

// Len 返回元素数量
func (m *LMap[K, V]) Len() int {
	return len(m.items)
}

// Clear 清空Map
func (m *LMap[K, V]) Clear() {
	m.items = m.items[:0]
	m.indices = make(map[K]int)
}

// ForEach 按插入顺序遍历
func (m *LMap[K, V]) ForEach(fn func(value V)) {
	for i := range m.items {
		fn(m.items[i])
	}
}

// ForEachBreakable 按插入顺序遍历，支持中断
// 返回 false 可以中断遍历
func (m *LMap[K, V]) ForEachBreakable(fn func(value V) bool) {
	for i := range m.items {
		if !fn(m.items[i]) {
			break
		}
	}
}

// ForEachReverse 按插入顺序反向遍历
func (m *LMap[K, V]) ForEachReverse(fn func(value V)) {
	for i := len(m.items) - 1; i >= 0; i-- {
		fn(m.items[i])
	}
}

// Keys 返回所有键（按插入顺序）
func (m *LMap[K, V]) Keys() []K {
	// 创建 index -> key 的映射
	indexToKey := make(map[int]K, len(m.indices))
	for k, idx := range m.indices {
		indexToKey[idx] = k
	}

	// 按索引顺序收集键
	keys := make([]K, 0, len(m.items))
	for i := range m.items {
		if k, ok := indexToKey[i]; ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// Values 返回所有值（按插入顺序）
func (m *LMap[K, V]) Values() []V {
	values := make([]V, len(m.items))
	copy(values, m.items)
	return values
}

// Entry 键值对
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// Entries 返回所有键值对（按插入顺序）
func (m *LMap[K, V]) Entries() []Entry[K, V] {
	// 创建 index -> key 的映射
	indexToKey := make(map[int]K, len(m.indices))
	for k, idx := range m.indices {
		indexToKey[idx] = k
	}

	entries := make([]Entry[K, V], 0, len(m.items))
	for i, v := range m.items {
		if k, ok := indexToKey[i]; ok {
			entries = append(entries, Entry[K, V]{
				Key:   k,
				Value: v,
			})
		}
	}
	return entries
}

// Clone 克隆Map
func (m *LMap[K, V]) Clone() *LMap[K, V] {
	newMap := &LMap[K, V]{
		items:   make([]V, len(m.items)),
		indices: make(map[K]int, len(m.indices)),
	}

	copy(newMap.items, m.items)
	for k, v := range m.indices {
		newMap.indices[k] = v
	}

	return newMap
}

// Filter 过滤元素，返回新Map
func (m *LMap[K, V]) Filter(fn func(value V) bool) *LMap[K, V] {
	newMap := NewLMap[K, V]()

	// 创建 index -> key 的映射
	indexToKey := make(map[int]K, len(m.indices))
	for k, idx := range m.indices {
		indexToKey[idx] = k
	}

	for i, v := range m.items {
		if fn(v) {
			if k, ok := indexToKey[i]; ok {
				newMap.Set(k, v)
			}
		}
	}
	return newMap
}

// Compact 压缩内部存储，释放未使用的容量
func (m *LMap[K, V]) Compact() {
	if cap(m.items) > len(m.items)*2 {
		// 如果容量超过实际使用的2倍，重新分配
		newItems := make([]V, len(m.items))
		copy(newItems, m.items)
		m.items = newItems
	}
}

// GetAt 获取指定索引位置的值（按插入顺序）
func (m *LMap[K, V]) GetAt(index int) (V, bool) {
	if index < 0 || index >= len(m.items) {
		var zero V
		return zero, false
	}

	return m.items[index], true
}

// Map 转换元素，返回新Map
func Map[K comparable, V any, R any](m *LMap[K, V], fn func(value V) R) *LMap[K, R] {
	newMap := NewLMap[K, R]()

	// 创建 index -> key 的映射
	indexToKey := make(map[int]K, len(m.indices))
	for k, idx := range m.indices {
		indexToKey[idx] = k
	}

	for i, v := range m.items {
		if k, ok := indexToKey[i]; ok {
			newMap.Set(k, fn(v))
		}
	}
	return newMap
}

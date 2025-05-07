package set

import (
	"math/rand"
	"strconv"
	"time"
)

type HashSet struct {
	dict     map[string]struct{}
	intset   *IntSet //存储整数集合
	isIntset bool    //标识当前集合是否为整数集合（用于切换底层实现）
}

// NewHashSet creates a new HashSet
func NewHashSet() *HashSet {
	return &HashSet{
		dict:     make(map[string]struct{}),
		intset:   NewIntSet(),
		isIntset: true, // 默认使用整数集合
	}
}
func (set *HashSet) Add(member string) int {
	// 如果当前集合是整数集合
	if set.isIntset {
		// 尝试将成员转换为64位整数
		if val, err := strconv.ParseInt(member, 10, 64); err == nil {
			if ok := set.intset.Add(val); ok {
				if set.intset.Len() > SET_MAX_INTSET_ENTRIES {
					set.convertToHashTable()
				}
				return 1 // 添加成功返回1
			}
			return 0 // 如果添加失败，说明已经存在
		} else {
			// 如果转换失败，说明不是整数，转换为哈希表
			set.convertToHashTable()
		}
	}
	// 如果当前集合是哈希表，直接添加到哈希表中
	if _, ok := set.dict[member]; ok {
		return 0 // 如果成员已经存在，返回0
	}
	set.dict[member] = struct{}{} // 添加成员到哈希表中
	return 1                      // 添加成功返回1
}

// convertToHashTable 将整数集合转换为哈希表
func (set *HashSet) convertToHashTable() {
	// 检查当前集合是否已经是哈希表
	if !set.isIntset {
		return
	}
	// 复制元素到哈希表中
	set.intset.ForEach(func(value int64) bool {
		set.dict[strconv.FormatInt(value, 10)] = struct{}{}
		return true
	})
	set.isIntset = false
	set.intset = nil // 释放整数集合的内存
}

func (set *HashSet) ForEach(consumer func(member string) bool) {
	if set.isIntset {
		set.intset.ForEach(func(value int64) bool {
			return consumer(strconv.FormatInt(value, 10))
		})
	} else {
		for member := range set.dict {
			if !consumer(member) {
				break
			}
		}
	}
}

// Len 返回集合的长度
func (set *HashSet) Len() int {
	if set.isIntset {
		return set.intset.Len()
	}
	return len(set.dict)
}

// Contains 判断集合中是否包含某个成员
func (set *HashSet) Contains(member string) bool {
	if set.isIntset {
		if val, err := strconv.ParseInt(member, 10, 64); err == nil {
			return set.intset.Contains(val)
		}
		return false
	}
	_, ok := set.dict[member]
	return ok
}

// Members 返回集合中的所有成员
func (set *HashSet) Members() []string {
	if set.isIntset { // 如果当前集合是整数集合
		members := make([]string, 0, set.intset.Len())
		set.intset.ForEach(func(value int64) bool {
			members = append(members, strconv.FormatInt(value, 10))
			return true
		})
		return members
	}
	// 如果当前集合是哈希表
	members := make([]string, 0, len(set.dict))
	for member := range set.dict {
		members = append(members, member)
	}
	return members
}

// Remove 移除集合中的一个成员，返回移除的成员数量
func (set *HashSet) Remove(member string) int {
	if set.isIntset {
		// 如果当前集合是整数集合，尝试将成员转换为64位整数
		if val, err := strconv.ParseInt(member, 10, 64); err == nil {
			if ok := set.intset.Remove(val); ok {
				return 1
			}
			return 0
		}
		return 0 // 如果转换失败，说明不是整数
	}

	if _, exists := set.dict[member]; !exists {
		return 0 // 如果成员不存在，返回0
	}
	// 如果当前集合是哈希表，直接从哈希表中删除成员
	delete(set.dict, member)
	return 1
}

// RandomDistinctMembers 随机返回集合中的不重复成员
func (set *HashSet) RandomDistinctMembers(count int) []string {
	size := set.Len()
	if count <= 0 || size == 0 {
		return []string{}
	}
	if count > size {
		return set.Members() // 返回所有成员
	}

	members := set.Members()
	// 使用 rand.Shuffle 方法对集合中的元素进行随机打乱，然后返回前 count 个元素
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	return members[:count]
}

// RandomMembers 随机返回集合中的成员
func (set *HashSet) RandomMembers(count int) []string {
	size := set.Len()
	if count <= 0 || size == 0 {
		return []string{}
	}
 
	res := make([]string, count)
	members := set.Members()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
 
	for i := 0; i < count; i++ {
		res[i] = members[r.Intn(size)]
	}
	return res
}

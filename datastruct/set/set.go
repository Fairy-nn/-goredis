package set

type Set interface {
	Add(member string) int                     // 添加成员到集合中，返回添加的成员数量
	Len() int                                  // 返回集合的长度
	ForEach(consumer func(member string) bool) // 遍历集合中的每个元素
	Contains(member string) bool               // 判断集合中是否包含某个成员
	Members() []string                         // 返回集合中的所有成员
	Remove(member string) int                  // 移除集合中的一个成员，返回移除的成员数量
	RandomDistinctMembers(count int) []string  // 随机返回集合中的不重复成员
	RandomMembers(count int) []string          // 随机返回集合中的成员

}

const (
	// intset 的最大元素数量,超过这个的时候就会转换为哈希表
	SET_MAX_INTSET_ENTRIES = 512
)

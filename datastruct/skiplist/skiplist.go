package skiplist

import (
	"math/rand"
	"time"
)

// 跳跃表允许的最大层数
const maxLevel = 16

// 跳跃表的节点
type Node struct {
	Member  string  //存储该节点关联的成员信息
	Score   float64 //表示该节点的分数
	Forward []*Node // 指向下一层的节点
}

// 跳跃表的结构体
type SkipList struct {
	header *Node      // 头节点
	tail   *Node      // 尾节点
	level  int        // 当前跳跃表的层数
	length int        // 跳跃表的长度
	rand   *rand.Rand // 随机数生成器,跳跃表中每个节点的层数是随机确定的
}

// NewSkipList 创建一个新的跳跃表
func NewSkipList() *SkipList {
	header := &Node{
		Forward: make([]*Node, maxLevel), // 初始化头节点的前向指针数组
	}
	return &SkipList{
		header: header,
		level:  1,                                               // 初始层数为1
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())), // 使用当前时间的纳秒数作为随机数种子
	}
}

// 随机生成跳跃表节点的层数
func (sl *SkipList) randomLevel() int {
	level := 1 // 初始层数为1
	//  for 循环以 25% 的概率增加层数
	// 使得大部分节点只在底层出现，少数节点出现在高层
	// 形成金字塔结构
	for sl.rand.Float64() < 0.25 && level < maxLevel {
		level++
	}
	return level
}

// Insert 向跳跃表中插入一个节点
func (sl *SkipList) Insert(member string, score float64) {
	// 用于记录每一层中，新节点插入位置的前驱节点
	update := make([]*Node, maxLevel)
	x := sl.header // 从头节点开始查找

	// 查找插入位置：从最高层向下查找
	for i := sl.level - 1; i >= 0; i-- {
		// 在当前层向右查找，直到找到第一个 Score 更大或 Member 更大的节点
		for x.Forward[i] != nil &&
			(x.Forward[i].Score < score ||
				(x.Forward[i].Score == score && x.Forward[i].Member < member)) {
			x = x.Forward[i]
		}
		// 记录下这一层需要修改 Forward 指针的节点 (即新节点的前驱)
		// 这里的 update[i] 是指在第 i 层中，x 的下一个节点是新节点
		// 也就是新节点的前面一个节点
		update[i] = x
	}

	// 生成新节点的层数
	level := sl.randomLevel()

	// 如果需要的话，更新最大层数
	if level > sl.level {
		// 扩展 update 数组
		for i := sl.level; i < level; i++ {
			update[i] = sl.header // 新增层级的前驱节点是 header（因为这个新的层级为空）
		}
		sl.level = level // 更新 SkipList 的当前最大层级
	}

	newNode := &Node{
		Member:  member,
		Score:   score,
		Forward: make([]*Node, level), // Forward 切片大小为新节点的层级
	}

	// 更新指针，将新节点链入 SkipList
	for i := 0; i < level; i++ {
		newNode.Forward[i] = update[i].Forward[i] // 新节点的 Forward 指向原前驱节点的下一个节点
		update[i].Forward[i] = newNode            // 前驱节点的 Forward 指向新节点
	}

	// 更新尾节点指针 (如果新节点是最后一个节点)
	if newNode.Forward[0] == nil {
		sl.tail = newNode
	}

	// 更新跳跃表的长度
	sl.length++
}

// Delete 从跳跃表中删除一个节点
func (sl *SkipList) Delete(member string, score float64) bool {
	update := make([]*Node, maxLevel)
	x := sl.header

	// 查找目标结点的前驱节点
	for i := sl.level - 1; i >= 0; i-- {
		// 从当前节点 x 开始，沿着 Forward 指针向右查找
		for x.Forward[i] != nil &&
			(x.Forward[i].Score < score ||
				(x.Forward[i].Score == score && x.Forward[i].Member < member)) {
			x = x.Forward[i]
		}
		update[i] = x // 将该层的前驱节点记录到 update 数组的对应位置
	}

	// 定位目标节点，x 是最底层中目标节点的前驱节点
	targetNode := x.Forward[0]

	// 检查节点是否存在且是否匹配
	if targetNode != nil && targetNode.Score == score && targetNode.Member == member {
		// 更新指针，在每一层中跳过目标节点
		for i := 0; i < sl.level; i++ {
			// 如果 update[i] 的下一个节点不是目标节点，说明目标节点不在这一层或更高层
			if update[i].Forward[i] != targetNode {
				break // 可以提前结束
			}
			// 将前驱节点的 Forward 指向目标节点的下一个节点，完成移除
			update[i].Forward[i] = targetNode.Forward[i]
		}

		// 如果目标节点是尾节点，更新尾节点指针
		if targetNode == sl.tail {
			// 新的尾节点是 update[0] (最底层的前驱)
			// 如果 update[0] 是 header，说明列表空了，tail 应为 nil
			if update[0] == sl.header {
				sl.tail = nil
			} else {
				sl.tail = update[0]
			}
		}

		// 如果目标节点是最高层的节点，更新跳跃表的层数
		for sl.level > 1 && sl.header.Forward[sl.level-1] == nil {
			sl.level-- // 降低跳跃表的层数
		}

		// 更新跳跃表的长度
		sl.length--
		return true // 删除成功
	}
	return false // 删除失败，节点不存在
}

// CountInRange 计算在指定范围内的节点数量
func (sl *SkipList) CountInRange(min, max float64) int {
	// 初始化计数器和起始节点
	count := 0
	x := sl.header

	// 找到第一个分数大于等于 min 的节点
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil && x.Forward[i].Score < min {
			x = x.Forward[i]
		}
	}

	// 遍历分数小于等于 max 的节点并计数
	x = x.Forward[0]
	for x != nil && x.Score <= max {
		count++
		x = x.Forward[0]
	}

	return count
}

// RangeByScore 返回在指定分数范围内的节点
func (sl *SkipList) RangeByScore(min, max float64, offset, count int) []string {
	result := []string{}
	x := sl.header

	// 找到第一个分数大于等于 min 的节点
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil && x.Forward[i].Score < min {
			x = x.Forward[i]
		}
	}

	// 遍历分数小于等于 max 的节点并筛选成员
	x = x.Forward[0]
	skipped := 0

	for x != nil && x.Score <= max {
		if offset < 0 || skipped >= offset {
			result = append(result, x.Member)
			// Stop if we've collected enough elements
			if count > 0 && len(result) >= count {
				break
			}
		} else {
			skipped++
		}
		x = x.Forward[0]
	}
	return result
}

// 
// RangeByRank 返回在指定排名范围内的节点
// 排名从 0 开始，0 表示第一个节点
func (sl *SkipList) RangeByRank(start, stop int) []string {
	result := []string{}
 
	// 处理负索引
	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}
	if start > stop || start >= sl.length {
		return result
	}
 
	// 找到第一个节点
	x := sl.header.Forward[0]
	for i := 0; i < start && x != nil; i++ {
		x = x.Forward[0]
	}
 
	// 遍历到指定范围的节点
	for i := start; i <= stop && x != nil; i++ {
		result = append(result, x.Member)
		x = x.Forward[0]
	}
 
	return result
}

// GetRank 返回指定成员的排名
func (sl *SkipList) GetRank(member string, score float64) int {
	rank := 0
	x := sl.header
 
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil &&
			(x.Forward[i].Score < score ||
				(x.Forward[i].Score == score && x.Forward[i].Member < member)) {
			rank += 1 // // 统计跳过的节点数
			x = x.Forward[i]
		}
	}
 
	x = x.Forward[0]
	if x != nil && x.Member == member {
		return rank
	}
 
	return -1 //没有找到该成员
}
package zset

import (
	"fmt"
	"goredis/datastruct/skiplist"
	"sort"
	"strconv"
)

// 有序集合的接口定义
type ZSet interface {
	Add(member string, score float64) bool // 添加成员和分数到有序集合中，返回是否添加成功
	Score(member string) (float64, bool)   // 获取成员的分数，返回分数和是否存在
	Len() int                              // 获取有序集合的长度
	RangeByRank(start, stop int) []string  // 获取指定排名范围内的成员
	Remove(member string) bool
	Count(min, max float64) int      // 获取指定分数范围内的成员数量
	Encoding() int                   // 获取当前编码类型
	GetSkiplist() *skiplist.SkipList // 获取跳跃表实例

}

const (
	encodingListpack = iota
	encodingSkiplist
)

// 用于限制 Listpack 的最大长度，超过长度后，使用 Skiplist 来存储
const listpackMaxSize = 128

type zset struct {
	encoding int
	listpack [][2]string
	dict     map[string]float64
	skiplist *skiplist.SkipList
}

// 创建一个新的 ZSet
func NewZSet() ZSet {
	return &zset{
		encoding: encodingListpack,
		listpack: make([][2]string, 0),
	}
}

func (z *zset) Add(member string, score float64) bool {
	if z.encoding == encodingListpack {
		// 检查成员是否已经存在于 listpack 中
		for i, pair := range z.listpack {
			if pair[0] == member {
				// 如果成员已经存在，更新分数
				z.listpack[i][1] = formatScore(score)
				return false
			}
		}
		// 添加成员和分数到 listpack 中
		z.listpack = append(z.listpack, [2]string{member, formatScore(score)})
		// 检查 listpack 的大小是否超过限制
		if len(z.listpack) > listpackMaxSize {
			z.convertToSkiplist()
		}
		return true
	} else {
		// 检查成员是否已经存在于 dict 中
		if existingScore, exists := z.dict[member]; exists {
			if existingScore != score {
				z.skiplist.Delete(member, existingScore) // 删除旧的节点
				z.skiplist.Insert(member, score)         // 添加新的节点
				z.dict[member] = score                   // 更新字典中的分数
			}
			return false
		}
		z.dict[member] = score
		z.skiplist.Insert(member, score)
		return true
	}
}

// 将分数转换为字符串格式
func formatScore(score float64) string {
	return fmt.Sprintf("%f", score)
}

// 将 listpack 转换为 skiplist
func (z *zset) convertToSkiplist() {
	// 如果当前有序集合已经是 skiplist 编码，直接返回，无需转换
	if z.encoding == encodingSkiplist {
		return
	}

	// 创建一个新的跳跃表实例，用于存储有序集合的成员和分数
	z.skiplist = skiplist.NewSkipList()
	// 创建一个新的字典，用于存储成员和分数的映射关系
	z.dict = make(map[string]float64, len(z.listpack))

	// 将 listpack 中的所有元素转移到跳跃表和字典中
	for _, pair := range z.listpack {
		member := pair[0]
		score, _ := parseScore(pair[1])
		z.dict[member] = score
		z.skiplist.Insert(member, score)
	}

	// 更新编码类型为 skiplist
	z.encoding = encodingSkiplist
	// 清空 listpack
	z.listpack = nil
}

// 将分数的字符串表示解析为 float64 类型
func parseScore(scoreStr string) (float64, error) {
	// 调用 strconv 包的 ParseFloat 函数进行解析，指定精度为 64 位
	return strconv.ParseFloat(scoreStr, 64)
}

func (z *zset) Score(member string) (float64, bool) {
	if z.encoding == encodingListpack {
		// 遍历 listpack 查找成员
		for _, pair := range z.listpack {
			if pair[0] == member {
				score, err := parseScore(pair[1])
				if err != nil {
					return 0, false
				}
				return score, true
			}
		}
		return 0, false // 成员不存在
	} else {
		score, exists := z.dict[member]
		return score, exists
	}
}

// 获取有序集合的长度
func (z *zset) Len() int {
	if z.encoding == encodingListpack { // 如果当前编码是 Listpack
		return len(z.listpack)
	} else { // 如果当前编码是dict
		return len(z.dict)
	}
}

// 获取指定排名范围内的成员
func (z *zset) RangeByRank(start, stop int) []string {
	if z.encoding == encodingListpack {
		// 创建一个新的切片用于存储成员和分数
		pairs := make([][2]string, len(z.listpack))
		copy(pairs, z.listpack) // 复制 listpack 的内容到新的切片中

		// 对切片进行排序，按照分数升序排列
		sort.Slice(pairs, func(i, j int) bool {
			score1, _ := parseScore(pairs[i][1])
			score2, _ := parseScore(pairs[j][1])
			return score1 < score2
		})

		size := len(pairs)
		if start < 0 {
			start = size + start
		}
		if stop < 0 {
			stop = size + stop
		}
		if start < 0 {
			start = 0
		}
		if stop >= size {
			stop = size - 1
		}
		if start > stop || start >= size {
			return []string{}
		}

		// 创建一个切片用于存储结果
		result := make([]string, 0, stop-start+1)
		for i := start; i <= stop; i++ {
			result = append(result, pairs[i][0]) // 添加成员到结果切片中
		}
		return result // 返回结果切片
	}
	return z.skiplist.RangeByRank(start, stop) // 如果当前编码是 Skiplist，直接调用跳跃表的 RangeByRank 方法
}

// 移除指定成员
func (z *zset) Remove(member string) bool {
	if z.encoding == encodingListpack {
		for i, pair := range z.listpack {
			if pair[0] == member {
				// 删除成员
				// 使用 append 方法将 listpack 切片的前后部分连接起来，跳过要删除的元素
				z.listpack = append(z.listpack[:i], z.listpack[i+1:]...)
				return true
			}
		}
		return false
	} else {
		score, exists := z.dict[member]
		if exists {
			z.skiplist.Delete(member, score)
			delete(z.dict, member)
			return true
		}
		return false
	}
}

// 获取指定分数范围内的成员数量
func (z *zset) Count(min, max float64) int {
	if z.encoding == encodingListpack {
		count := 0
		for _, pair := range z.listpack {
			score, _ := parseScore(pair[1])
			if score >= min && score <= max {
				count++
			}
		}
		return count
	}

	// Using skiplist encoding
	return z.skiplist.CountInRange(min, max)
}

// 获取当前编码类型
func (z *zset) Encoding() int {
	return z.encoding
}

// GetSkiplist 返回跳跃表实例
func (z *zset) GetSkiplist() *skiplist.SkipList {
	if z.encoding == encodingSkiplist {
		return z.skiplist
	}
	return nil
}

package set

import (
	"encoding/binary"
	"fmt"
	"math"
)

// 使用动态编码，通过根据集合中整数的范围选择最小的编码类型
const (
	INTSET_ENC_INT16 = 2
	INTSET_ENC_INT32 = 4
	INTSET_ENC_INT64 = 8
)

type IntSet struct {
	encoding uint32 //存储当前集合的编码类型
	length   uint32 //存储当前集合的长度
	contents []byte //存储当前集合的内容
}

// NewIntSet creates a new IntSet with the given encoding
func NewIntSet() *IntSet {
	return &IntSet{
		encoding: INTSET_ENC_INT16, // Default encoding is 16-bit integer
		length:   0,
		contents: make([]byte, 0),
	}
}

// 返回集合的长度
func (is *IntSet) Len() int {
	return int(is.length)
}

// 向整数集合里添加一个整数
func (is *IntSet) Add(value int64) bool {
	var requiredEncoding uint32
	if value < math.MinInt16 || value > math.MaxInt16 {
		if value < math.MinInt32 || value > math.MaxInt32 {
			requiredEncoding = INTSET_ENC_INT64
		} else {
			requiredEncoding = INTSET_ENC_INT32
		}
	} else {
		requiredEncoding = INTSET_ENC_INT16
	}
	// 必要时升级编码
	if requiredEncoding > is.encoding {
		is.upgradeEncoding(requiredEncoding)
	}
	// 检查值是否已经存在
	pos := is.findPosition(value)
	if pos >= 0 {
		return false // 值已经存在，返回false
	}
	// 添加值到集合中
	pos = -(pos + 1) // 计算插入位置
	is.insertAt(pos, value)
	return true // 返回true表示添加成功
}

// 升级编码
func (is *IntSet) upgradeEncoding(newEncoding uint32) {
	if newEncoding <= is.encoding {
		return // 不需要升级
	}
	// 保存旧值
	oldValues := is.ToSlice()
	// 创建新的 IntSet
	is.encoding = newEncoding
	is.length = 0
	is.contents = make([]byte, 0, len(oldValues)*int(newEncoding))
	// 重新添加旧值
	for _, value := range oldValues {
		is.Add(value)
	}
}

// 二分查找元素位置
func (is *IntSet) findPosition(value int64) int {
	low, high := 0, int(is.length)-1
	for low <= high {
		mid := (low + high) / 2
		midVal := is.getValueAt(uint32(mid))

		if midVal < value {
			low = mid + 1
		} else if midVal > value {
			high = mid - 1
		} else {
			return mid // Found
		}
	}
	return -(low + 1)
}

// 将当前集合中的所有元素保存到一个切片中
func (is *IntSet) ToSlice() []int64 {
	result := make([]int64, is.length)
	for i := uint32(0); i < is.length; i++ {
		result[i] = is.getValueAt(i)
	}
	return result
}

// 获取元素值
func (is *IntSet) getValueAt(index uint32) int64 {
	if index >= is.length {
		fmt.Println("getvalue:Index out of range")
		return 0 // 超出范围，返回0
	}
	// 计算偏移量
	offset := index * is.encoding
	switch is.encoding {
	case INTSET_ENC_INT16:
		return int64(int16(binary.LittleEndian.Uint16(is.contents[offset:])))
	case INTSET_ENC_INT32:
		return int64(int32(binary.LittleEndian.Uint32(is.contents[offset:])))
	case INTSET_ENC_INT64:
		return int64(binary.LittleEndian.Uint64(is.contents[offset:]))
	}
	panic("Invalid encoding")
}

// 在指定位置插入值
func (is *IntSet) insertAt(pos int, value int64) {
	// 扩展存储内容
	oldLen := len(is.contents)
	newLen := oldLen + int(is.encoding)
	if newLen > cap(is.contents) {
		// 容量不足，扩展容量
		newContents := make([]byte, newLen*2)
		copy(newContents, is.contents)
		is.contents = newContents
	} else {
		is.contents = is.contents[:newLen]
	}
	// 移动元素来腾出插入位置
	offset := pos * int(is.encoding)
	if pos < int(is.length) {
		copy(is.contents[offset+int(is.encoding):], is.contents[offset:])
	}
	// 插入新值
	switch is.encoding {
	case INTSET_ENC_INT16:
		binary.LittleEndian.PutUint16(is.contents[offset:], uint16(value))
	case INTSET_ENC_INT32:
		binary.LittleEndian.PutUint32(is.contents[offset:], uint32(value))
	case INTSET_ENC_INT64:
		binary.LittleEndian.PutUint64(is.contents[offset:], uint64(value))
	}
	is.length++ // 更新长度
}

// 遍历集合中的每个元素
func (is *IntSet) ForEach(consumer func(value int64) bool) {
	for i := uint32(0); i < is.length; i++ {
		if !consumer(is.getValueAt(i)) {
			break
		}
	}
}

// contains 判断集合中是否包含某个整数
func (is *IntSet) Contains(value int64) bool {
	pos := is.findPosition(value)
	return pos >= 0 // 如果pos大于等于0，说明存在
}

// Remove 删除集合中的一个整数
func (is *IntSet) Remove(value int64) bool {
	pos := is.findPosition(value)
	if pos < 0 {
		return false
	}

	is.removeAt(pos)
	return true
}

// removeAt 删除指定位置的元素
func (is *IntSet) removeAt(pos int) {
	if pos < 0 || pos >= int(is.length) {
		return // 超出范围，返回
	}
	offset := pos * int(is.encoding)
	endOffset := int(is.length) * int(is.encoding)

	copy(is.contents[offset:], is.contents[offset+int(is.encoding):endOffset])
	is.contents = is.contents[:endOffset-int(is.encoding)]
	is.length-- // 更新长度
}

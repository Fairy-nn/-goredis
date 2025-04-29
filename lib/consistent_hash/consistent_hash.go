package consistenthash

import (
	"hash/crc32"
	"sort"
)

type NodeMap struct {
	nodehashMap map[int]string // Hash value to node mapping
	nodeHashs   []int          // Sorted hash values
	// 数据都是以字节形式存储和传输的，将 []byte 作为参数
	// 能让哈希函数直接处理底层数据，减少数据转换的开销
	hashFunc func(data []byte) uint32
}

func NewNodeMap(hashfunc func(data []byte) uint32) *NodeMap {
	if hashfunc == nil { // If no hash function is provided, use crc32 as default
		hashfunc = crc32.ChecksumIEEE
	}
	return &NodeMap{
		nodehashMap: make(map[int]string),
		nodeHashs:   make([]int, 0),
		hashFunc:    hashfunc,
	}
}

// Add a node to the hash ring
func (n *NodeMap) AddNode(node ...string) {
	for _, v := range node {
		hash := int(n.hashFunc([]byte(v)))      // Calculate the hash value of the node
		n.nodeHashs = append(n.nodeHashs, hash) // Add the hash value to the slice
		n.nodehashMap[hash] = v                 // Map the hash value to the node
	}
	// Sort the hash values in ascending order
	sort.Ints(n.nodeHashs)
}

// Get the node for a given key,return the node corresponding to the key
func (n *NodeMap) GetNode(key string) string {
	if len(n.nodehashMap) == 0 {
		return ""
	}
	hash := int(n.hashFunc([]byte(key)))
	// 查找离给定键值最近的节点
	index := sort.Search(len(n.nodeHashs), func(i int) bool {
		return n.nodeHashs[i] >= hash // Find the first hash value greater than or equal to the key's hash value
	})
	if index == len(n.nodeHashs) { // If the index is equal to the length of the slice, wrap around to the first node
		index = 0
	}
	return n.nodehashMap[n.nodeHashs[index]] // Return the corresponding node

}

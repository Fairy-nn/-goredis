package hash

// 当哈希中数据的长度超过此值时，将转换为哈希表
const (
	hashMaxListpackEntries = 512
	hashMaxListpackValue   = 64
)

// 定义哈希的编码类型
const (
	encodingListpack = iota
	encodingHashTable
)

// 定义一个哈希数据结构
type Hash struct {
	encoding int               // 编码类型
	listpack [][2]string       //使用切片存储键值对，模拟listpack
	dict     map[string]string // 使用map存储键值对，模拟哈希表
}

func NewHash() *Hash {
	return &Hash{
		encoding: encodingListpack,
		listpack: make([][2]string, 0),
		dict:     nil,
	}
}

// Get函数从哈希中检索与给定键关联的值
func (h *Hash) Get(key string) (string, bool) {
	// 如果使用listpack编码，遍历listpack查找键
	if h.encoding == encodingListpack {
		for _, kv := range h.listpack {
			if kv[0] == key {
				return kv[1], true
			}
		}
	}
	// 如果使用哈希表编码，直接从哈希表中查找键
	if h.encoding == encodingHashTable {
		if val, ok := h.dict[key]; ok {
			return val, true
		}
	}
	return "", false
}

// Set函数在哈希中设置给定键的值
// 如果字段已存在，它将更新值并返回0；如果是新条目，则返回1
func (h *Hash) Set(key, value string) int {
	// 如果使用listpack编码，检查长度和键值对的大小
	if h.encoding == encodingListpack {
		if len(h.listpack)>= hashMaxListpackEntries ||
		len(key) > hashMaxListpackValue ||
		len(value) > hashMaxListpackValue {
			// 转换为哈希表编码
			h.covertToHashTable()
		}
	}
	// 使用listpack编码
	if h.encoding == encodingListpack {
		for i, kv := range h.listpack {
			if kv[0] == key {
				h.listpack[i][1] = value // 更新值
				return 0
			}
		}
		h.listpack = append(h.listpack, [2]string{key, value}) // 添加新键值对
		return 1
	}
	// 使用哈希表编码
	if h.encoding == encodingHashTable {
		if _, ok := h.dict[key]; ok {
			h.dict[key] = value // 更新值
			return 0
		}
		h.dict[key] = value // 添加新键值对
		return 1
	}
	return 0
}

// Delete函数从哈希中删除给定的字段
func (h *Hash) Delete(key string) int {
	count := 0
	// 如果使用listpack编码，遍历listpack查找键
	if h.encoding == encodingListpack {
		for i, kv := range h.listpack {
			if kv[0] == key {
				// 删除条目，并将最后一个条目移到当前位置以减小大小
                // 因为哈希不需要保持顺序，我们可以简单地将最后一个条目与当前条目交换
                lastIndex := len(h.listpack) - 1
                h.listpack[i] = h.listpack[lastIndex]
                h.listpack = h.listpack[:lastIndex]
				count++
				break
			}
		}
	}
	// 如果使用哈希表编码，直接从哈希表中删除键
	if h.encoding == encodingHashTable {
		if _, ok := h.dict[key]; ok {
			delete(h.dict, key)
			count++
		}
	}
	return count
}

// Len函数返回哈希中键值对的数量
func (h *Hash) Len() int {
	// listpack
	if h.encoding == encodingListpack {
		return len(h.listpack)
	}
	// hash table
	if h.encoding == encodingHashTable {
		return len(h.dict)
	}
	return 0
}

// GetAll函数返回哈希中的所有字段和值
func (h *Hash) GetAll() map[string]string {
	result := make(map[string]string)
	// listpack
	if h.encoding == encodingListpack {
		for _, kv := range h.listpack {
			result[kv[0]] = kv[1]
		}
	}
	// hash table
	if h.encoding == encodingHashTable {
		for k, v := range h.dict {
			result[k] = v
		}
	}
	return result
}

// Fields函数返回哈希中的所有字段
func (h *Hash) Fields() []string {
	result := make([]string, 0)
	// listpack
	if h.encoding == encodingListpack {
		for _, kv := range h.listpack {
			result = append(result, kv[0])
		}
	}
	// hash table
	if h.encoding == encodingHashTable {
		for k := range h.dict {
			result = append(result, k)
		}
	}
	return result
}

// Values函数返回哈希中的所有值
func (h *Hash) Values() []string {
	result := make([]string, 0)
	// listpack
	if h.encoding == encodingListpack {
		for _, kv := range h.listpack {
			result = append(result, kv[1])
		}
	}
	// hash table
	if h.encoding == encodingHashTable {
		for _, v := range h.dict {
			result = append(result, v)
		}
	}
	return result
}

// Exists函数检查字段是否存在于哈希中
func (h *Hash) Exists(key string) bool {
	_, exists := h.Get(key)
    return exists
}

// convertToHashTable函数将哈希从listpack编码转换为哈希表编码
func (h *Hash) covertToHashTable() {
	if h.encoding == encodingHashTable {
		return
	}

	h.dict = make(map[string]string, len(h.listpack))
	for _, kv := range h.listpack {
		h.dict[kv[0]] = kv[1]
	}
	h.listpack = nil // 清空listpack以释放内存
	h.encoding = encodingHashTable // 更新编码类型
}

// Encoding函数返回哈希的编码类型
func (h *Hash) Encoding() int {
	return h.encoding
}

// Clear函数清空哈希中的所有键值对
func (h *Hash) Clear() {
	h.listpack = make([][2]string, 0)
	h.dict = nil
	h.encoding = encodingListpack
}
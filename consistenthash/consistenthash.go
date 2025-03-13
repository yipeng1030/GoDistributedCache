package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HashF maps bytes to uint32
type HashF func(data []byte) uint32

// HashNodes constains all hashed keys
type HashNodes struct {
	hash           HashF          // HashF function
	replicas       int            // Number of virtual nodes
	keys           []int          // Sorted
	virtualNodeMap map[int]string // virtual node and actual node
}

// NewHashNodes creates a HashNodes instance
func NewHashNodes(replicas int, fn HashF) *HashNodes {
	m := &HashNodes{
		replicas:       replicas,
		hash:           fn,
		virtualNodeMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some nodes to the hash
func (m *HashNodes) Add(nodes ...string) {
	// 为了解决倾斜的问题，引入虚拟节点，虚拟节点的个数是replicas，用这样多个节点再哈希，然后再映射到真实的节点
	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
			m.keys = append(m.keys, hash)
			m.virtualNodeMap[hash] = node
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *HashNodes) Get(key string) string {
	// 先做hash值，然后在环上找到最近的一个节点，再考虑环的问题，然后用Map映射到真实的节点
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 所以只要加了一个%操作，就是一个环了
	return m.virtualNodeMap[m.keys[idx%len(m.keys)]]
}

package GoDistributedCache

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented to get the value
// 用来从对应 group 查找缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

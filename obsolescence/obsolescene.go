package obsolescence

type Cache interface {
	Add(key string, value Value)
	Get(key string) (value Value, ok bool)
	Del(key string)
	Len() int
	RemoveOldest()
}

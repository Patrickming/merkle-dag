package merkledag

// 保存KV的存储器接口
type KVStore interface {
	Has(key []byte) (bool, error)
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
}

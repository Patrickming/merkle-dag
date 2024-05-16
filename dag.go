package merkledag

import (
	"encoding/json"
	"hash"
)

// // 定义一些常量用于限制列表和块的大小。
// const (
// 	LIST_LIMIT  = 2048
// 	BLOCK_LIMIT = 256 * 1024 // 每个块256KB
// )

// // 定义Merkle DAG中的节点类型。
// const (
// 	BLOB = "blob"
// 	LIST = "list"
// 	TREE = "tree"
// )

// Link 结构体表示 DAG 中的链接，包含了文件名、哈希值和文件大小等信息。
type Link struct {
	Name string // 文件名
	Hash []byte // 文件哈希值
	Size int    // 文件大小
}

// Object 结构体表示 DAG 中的对象，包含了链接列表和数据。
type Object struct {
	Links []Link // 链接列表
	Data  []byte // 数据
}

// Add 函数是主函数，它接受一个存储器 KVStore，一个节点 Node 和一个哈希算法 hash.Hash 作为参数，并返回 Merkle 根哈希值。
// 根据节点的类型（文件或文件夹），调用相应的处理函数。
func Add(store KVStore, node Node, h hash.Hash) []byte {
	//将分片写入到KVStore中，并返回Merkle Root
	var rootHash []byte
	if node.Type() == FILE {
		rootHash, _ = StoreFile(store, node, h)
	} else if node.Type() == DIR {
		rootHash = StoreDir(store, node, h)
	}

	return rootHash
}

// StoreFile 函数用于处理文件节点，将文件分块存储到 KVStore 中，并返回文件的哈希值和类型。
func StoreFile(store KVStore, node Node, h hash.Hash) ([]byte, []byte) {
	content := node.(File).Bytes() // 获取文件内容
	fileType := []byte("blob")     // 文件类型，默认为 blob
	// 如果文件大小超过 256KB，则将文件类型设为 list，并进行分块处理
	if node.Size() > 256*1024 {
		fileType = []byte("list")
		obj := Object{}
		n := node.Size() % 256 * 1024
		m := node.Size() / 256 * 1024
		if m > 0 {
			n++
		}
		// 对文件内容进行分块处理
		for i := 0; i < int(n); i++ {
			start := i * 256 * 1024
			end := (i + 1) * 256 * 1024
			if end > len(content) {
				end = len(content)
			}
			chunk := content[start:end]        // 获取分块内容
			jsonData := Object{Data: chunk}    // 创建分块对象
			value, _ := json.Marshal(jsonData) // 将对象序列化为 JSON
			key := CalHash(h, value)           // 计算哈希值
			store.Put(key, value)              // 将分块存储到 KVStore 中
			// 更新对象数据和链接列表
			obj.Data = append(obj.Data, []byte("blob")...)
			obj.Links = append(obj.Links, Link{Hash: key, Size: end - start})
		}
		// 将该文件存入 KVStore 中
		jsonData := Object{Data: obj.Data, Links: obj.Links}
		value, _ := json.Marshal(jsonData)
		key := CalHash(h, value)
		store.Put(key, value)
		return key, fileType
	} else {
		// 如果文件大小不超过 256KB，直接存储文件内容
		jsonData := Object{Data: content}
		value, _ := json.Marshal(jsonData)
		key := CalHash(h, value)
		store.Put(key, value)
		return key, fileType
	}
}

// StoreDir 函数用于处理文件夹节点，递归处理文件夹下的所有文件和子文件夹，并构建出对应的 DAG 结构。
func StoreDir(store KVStore, node Node, h hash.Hash) []byte {
	tree := Object{
		Links: make([]Link, 0),
		Data:  make([]byte, 0),
	}
	dirNode := node.(Dir)
	it := dirNode.It() // 获取迭代器
	for it.Next() {
		childNode := it.Node()
		if childNode.Type() == FILE {
			// 处理文件节点
			key, fileType := StoreFile(store, childNode, h)
			// 更新 DAG 结构中的链接列表和数据
			tree.Data = append(tree.Data, fileType...)
			tree.Links = append(tree.Links, Link{Size: int(childNode.Size()), Hash: key})
			// 将更新后的 DAG 结构存入 KVStore 中
			value, _ := json.Marshal(tree)
			key = CalHash(h, value)
			store.Put(key, value)
		} else if childNode.Type() == DIR {
			// 处理文件夹节点
			key := StoreDir(store, childNode, h)
			// 更新 DAG 结构中的链接列表和数据
			tree.Links = append(tree.Links, Link{Name: childNode.Name(), Hash: key})
			tree.Data = append(tree.Data, []byte("tree")...)
		}
		// 将更新后的 DAG 结构存入 KVStore 中
		value, _ := json.Marshal(tree)
		key := CalHash(h, value)
		store.Put(key, value)
		return key
	}
	// 将 DAG 结构存入 KVStore 中
	value, _ := json.Marshal(tree)
	key := CalHash(h, value)
	store.Put(key, value)
	return key
}

// CalHash 函数用于计算数据的哈希值。
func CalHash(h hash.Hash, value []byte) []byte {
	h.Reset()
	key := h.Sum(value)
	return key
}

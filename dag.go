package merkledag

import (
	"encoding/json"
	"hash"
)

// 定义一些常量用于限制列表和块的大小。
const (
	LIST_LIMIT  = 2048
	BLOCK_LIMIT = 256 * 1024 // 每个块256KB
)

// 定义Merkle DAG中的节点类型。
const (
	BLOB = "blob"
	LIST = "list"
	TREE = "tree"
)

// Link 结构体定义了Merkle DAG中的链接节点。
type Link struct {
	Name string // 链接的名称
	Hash []byte // 链接指向的数据的哈希值
	Size int    // 数据的大小
}

// Object 结构体代表Merkle DAG中的一个节点，可以是文件或目录。
type Object struct {
	Links []Link // 存储指向其他Object的链接
	Data  []byte // 节点包含的数据
}

// Add 函数将节点添加到存储中，并返回Merkle根的哈希值。
func Add(store KVStore, node Node, h hash.Hash) []byte {
	// 判断节点类型并处理
	if node.Type() == FILE {
		file := node.(File)
		fileSlice := storeFile(file, store, h)
		jsonData, _ := json.Marshal(fileSlice)
		h.Write(jsonData)
		return h.Sum(nil)
	} else {
		dir := node.(Dir)
		dirSlice := storeDirectory(dir, store, h)
		jsonData, _ := json.Marshal(dirSlice)
		h.Write(jsonData)
		return h.Sum(nil)
	}
}

// compute 函数根据节点高度和分片的位置递归计算和存储节点。
func compute(height int, node File, store KVStore, seedId int, h hash.Hash) (*Object, int) {
	// 如果是叶子节点
	if height == 1 {
		if (len(node.Bytes()) - seedId) <= BLOCK_LIMIT {
			data := node.Bytes()[seedId:]
			blob := Object{
				Links: nil,
				Data:  data,
			}
			jsonData, _ := json.Marshal(blob)
			h.Reset()
			h.Write(jsonData)
			exists, _ := store.Has(h.Sum(nil))
			if !exists {
				store.Put(h.Sum(nil), data)
			}
			return &blob, len(data)
		}
		// 构造分支节点
		links := &Object{}
		totalLen := 0
		for i := 1; i <= 4096; i++ {
			end := seedId + BLOCK_LIMIT
			if len(node.Bytes()) < end {
				end = len(node.Bytes())
			}
			data := node.Bytes()[seedId:end]
			blob := Object{
				Links: nil,
				Data:  data,
			}
			totalLen += len(data)
			jsonData, _ := json.Marshal(blob)
			h.Reset()
			h.Write(jsonData)
			exists, _ := store.Has(h.Sum(nil))
			if !exists {
				store.Put(h.Sum(nil), data)
			}
			links.Links = append(links.Links, Link{
				Hash: h.Sum(nil),
				Size: len(data),
			})
			links.Data = append(links.Data, []byte("data")...)
			seedId += BLOCK_LIMIT
			if seedId >= len(node.Bytes()) {
				break
			}
		}
		jsonData, _ := json.Marshal(links)
		h.Reset()
		h.Write(jsonData)
		exists, _ := store.Has(h.Sum(nil))
		if !exists {
			store.Put(h.Sum(nil), jsonData)
		}
		return links, totalLen
	} else {
		// 递归处理高层节点
		links := &Object{}
		totalLen := 0
		for i := 1; i <= 4096; i++ {
			if seedId >= len(node.Bytes()) {
				break
			}
			child, childLen := compute(height-1, node, store, seedId, h)
			totalLen += childLen
			jsonData, _ := json.Marshal(child)
			h.Reset()
			h.Write(jsonData)
			links.Links = append(links.Links, Link{
				Hash: h.Sum(nil),
				Size: childLen,
			})
			typeName := "link"
			if child.Links == nil {
				typeName = "data"
			}
			links.Data = append(links.Data, []byte(typeName)...)
		}
		jsonData, _ := json.Marshal(links)
		h.Reset()
		h.Write(jsonData)
		exists, _ := store.Has(h.Sum(nil))
		if !exists {
			store.Put(h.Sum(nil), jsonData)
		}
		return links, totalLen
	}
}

// storeFile 函数用于存储文件类型的节点。
func storeFile(node File, store KVStore, h hash.Hash) *Object {
	if len(node.Bytes()) <= BLOCK_LIMIT {
		data := node.Bytes()
		blob := Object{
			Links: nil,
			Data:  data,
		}
		jsonData, _ := json.Marshal(blob)
		h.Reset()
		h.Write(jsonData)
		exists, _ := store.Has(h.Sum(nil))
		if !exists {
			store.Put(h.Sum(nil), data)
		}
		return &blob
	}
	// 计算需要的层级数
	linkLen := (len(node.Bytes()) + (BLOCK_LIMIT - 1)) / BLOCK_LIMIT
	height := 0
	tmp := linkLen
	for {
		height++
		tmp /= 4096
		if tmp == 0 {
			break
		}
	}
	res, _ := compute(height, node, store, 0, h)
	return res
}

// storeDirectory 函数用于存储目录类型的节点。
func storeDirectory(node Dir, store KVStore, h hash.Hash) *Object {
	iter := node.It()
	tree := &Object{}
	for iter.Next() {
		elem := iter.Node()
		if elem.Type() == FILE {
			file := elem.(File)
			fileSlice := storeFile(file, store, h)
			jsonData, _ := json.Marshal(fileSlice)
			h.Reset()
			h.Write(jsonData)
			tree.Links = append(tree.Links, Link{
				Hash: h.Sum(nil),
				Size: int(file.Size()),
				Name: file.Name(),
			})
			elemType := "link"
			if fileSlice.Links == nil {
				elemType = "data"
			}
			tree.Data = append(tree.Data, []byte(elemType)...)
		} else {
			dir := elem.(Dir)
			dirSlice := storeDirectory(dir, store, h)
			jsonData, _ := json.Marshal(dirSlice)
			h.Reset()
			h.Write(jsonData)
			tree.Links = append(tree.Links, Link{
				Hash: h.Sum(nil),
				Size: int(dir.Size()),
				Name: dir.Name(),
			})
			elemType := "tree"
			tree.Data = append(tree.Data, []byte(elemType)...)
		}
	}
	jsonData, _ := json.Marshal(tree)
	h.Reset()
	h.Write(jsonData)
	exists, _ := store.Has(h.Sum(nil))
	if !exists {
		store.Put(h.Sum(nil), jsonData)
	}
	return tree
}

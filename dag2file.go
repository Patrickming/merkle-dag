package merkledag

import (
	"encoding/json"
	"strings"
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
const STEP = 4 // 定义步进值，用于从Object的Data字段中提取信息

// Hash2File 根据给定的哈希和路径从KV存储中检索文件。
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// 检查哈希是否存在于存储中，hash对应的类型应该是tree
	flag, _ := store.Has(hash)
	if flag {
		objBinary, _ := store.Get(hash)               // 从存储中获取对应的二进制数据
		obj := binaryToObj(objBinary)                 // 将二进制数据转换为Object对象
		pathArr := strings.Split(path, "\\")          // 将路径分割成数组
		cur := 1                                      // 从路径的第一段开始处理
		return getFileByDir(obj, pathArr, cur, store) // 递归地检索文件
	}
	return nil // 如果哈希不存在，则返回nil
}

// getFileByDir 递归地在目录对象中查找指定路径的文件。
func getFileByDir(obj *Object, pathArr []string, cur int, store KVStore) []byte {
	if cur >= len(pathArr) { // 如果已经处理完所有路径部分
		return nil
	}
	index := 0                 // Data字段中的索引位置
	for i := range obj.Links { // 遍历所有链接
		objType := string(obj.Data[index : index+STEP]) // 获取当前链接的对象类型 从Data字段中读取4个字节来确定链接的类型
		index += STEP                                   // 移动到下一个类型标记的位置
		objInfo := obj.Links[i]
		if objInfo.Name != pathArr[cur] { // 如果名称不匹配，跳过
			continue
		}
		switch objType {
		case TREE: // 如果是目录，递归查找
			objDirBinary, _ := store.Get(objInfo.Hash)
			objDir := binaryToObj(objDirBinary)
			ans := getFileByDir(objDir, pathArr, cur+1, store)
			if ans != nil {
				return ans
			}
		case BLOB: // 如果是文件，直接返回文件内容
			ans, _ := store.Get(objInfo.Hash)
			return ans
		case LIST: // 如果是列表，递归处理列表
			objLinkBinary, _ := store.Get(objInfo.Hash)
			objList := binaryToObj(objLinkBinary)
			ans := getFileByList(objList, store)
			return ans
		}
	}
	return nil // 如果找不到，返回nil
}

// getFileByList 递归地处理列表类型的节点，合并所有BLOB类型的数据。
func getFileByList(obj *Object, store KVStore) []byte {
	ans := make([]byte, 0)
	index := 0                 // Data字段中的索引位置
	for i := range obj.Links { // 遍历所有链接
		curObjType := string(obj.Data[index : index+STEP]) // 读取当前对象的类型
		index += STEP
		curObjLink := obj.Links[i]
		curObjBinary, _ := store.Get(curObjLink.Hash)
		curObj := binaryToObj(curObjBinary)
		if curObjType == BLOB { // 如果是BLOB，直接添加到结果中
			ans = append(ans, curObjBinary...)
		} else { // 如果是LIST，递归处理
			tmp := getFileByList(curObj, store)
			ans = append(ans, tmp...)
		}
	}
	return ans
}

// binaryToObj 将二进制数据转换为Object对象。
func binaryToObj(objBinary []byte) *Object {
	var res Object
	json.Unmarshal(objBinary, &res) // 解析JSON数据填充Object结构
	return &res
}

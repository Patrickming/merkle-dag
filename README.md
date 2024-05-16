# 分布式存储实验


## 背景介绍

### 1. `file.go`文件中定义了以下几种接口： 

- `Node`为文件或者文件夹，根据Type可以判断
- `File`为文件，可以通过[]byte获取文件内容(不需要通过io从文件系统或者网络中读取文件)
- `Dir`为文件夹，可以通过It()函数获取到遍历文件的迭代器
- `DirIterator`为文件夹迭代器，可以获取当前文件夹下的文件/文件夹

### 2. `kvstore` 为保存KV的存储器接口，具体实现不需要关心，由实验测试系统来实现

## 实验一  File to DAG

- 根据上面的两个接口，实现 `dag.go` 中的 `Add` 函数，将 `Node` 中的数据保存在 `KVStore` 中，然后计算出Merkle Root
- 在`dag.go`实现

## 实验二  DAG to File

- 根据上面的接口，实现 `dag2file.go` 中的 `Hash2File` 函数，将 `hash` 对应的数据从 `KVStore` 中读出来，然后根据`path`返回对应的文件内容。
- 在`dag2file.go`实现


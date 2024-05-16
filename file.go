package merkledag

const (
	FILE = iota
	DIR
)

// 文件或者文件夹，根据Type可以判断
type Node interface {
	Size() uint64
	Name() string
	Type() int
}

// 文件，可以通过[]byte获取文件内容(不需要通过io从文件系统或者网络中读取文件)
type File interface {
	Node

	Bytes() []byte
}

// 文件夹，可以通过It()函数获取到遍历文件的迭代器
type Dir interface {
	Node

	It() DirIterator
}

// 文件夹迭代器，可以获取当前文件夹下的文件/文件夹
type DirIterator interface {
	Next() bool

	Node() Node
}

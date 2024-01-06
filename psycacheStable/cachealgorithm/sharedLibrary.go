package cachealgorithm

// Lengthable 接口指明对象可以获取自身占有内存空间大小 以字节为单位
type Lengthable interface {
	Len() int
}

// OnEliminated 当key-value被淘汰时 执行的处理函数
type OnEliminated func(key string, value Lengthable)

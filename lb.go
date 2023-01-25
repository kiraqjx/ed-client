package edclient

import "sync/atomic"

type Lb struct {
	nodes []*NodeInfo
	count int64
	size  int
}

func NewLb(nodes []*NodeInfo) *Lb {
	return &Lb{
		nodes: nodes,
		size:  len(nodes),
		count: 0,
	}
}

func (lb *Lb) Lb() *NodeInfo {
	value := atomic.AddInt64(&lb.count, 1)
	index := value % int64(lb.size)
	return lb.nodes[index]
}

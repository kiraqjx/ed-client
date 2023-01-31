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

func NewLbFromMap(nodesMap map[string]*NodeInfo) *Lb {
	nodes := make([]*NodeInfo, 0, len(nodesMap))
	for _, v := range nodesMap {
		nodes = append(nodes, v)
	}
	return NewLb(nodes)
}

func (lb *Lb) ChangeNodes(nodes []*NodeInfo) {
	lb.nodes = nodes
}

func (lb *Lb) ChangeNodesFromMap(nodesMap map[string]*NodeInfo) {
	nodes := make([]*NodeInfo, len(nodesMap))
	for _, v := range nodesMap {
		nodes = append(nodes, v)
	}
	lb.nodes = nodes
}

func (lb *Lb) Lb() *NodeInfo {
	value := atomic.AddInt64(&lb.count, 1)
	index := value % int64(lb.size)
	return lb.nodes[index]
}

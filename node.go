package edclient

import "encoding/json"

type NodeInfo struct {
	Server string
	Tag    map[string]string
}

func kvToNodeInfo(value []byte) (*NodeInfo, error) {
	nodeInfo := &NodeInfo{}
	err := json.Unmarshal(value, nodeInfo)
	if err != nil {
		return nil, err
	}
	return nodeInfo, nil
}

func nodeInfoToByte(nodeInfo NodeInfo) ([]byte, error) {
	return json.Marshal(nodeInfo)
}

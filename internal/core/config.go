package core

type NodeCfg struct {
	Id         string
	Addr       string
	Peers      []string
	StorageDir string
}

func NewNodeCfg() *NodeCfg {
	return &NodeCfg{}
}

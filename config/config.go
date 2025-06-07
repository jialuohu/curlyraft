package config

type NodeCfg struct {
	Id         string
	NetAddr    string
	Peers      string
	StorageDir string
}

func NewNodeCfg(Id string, NetAddr string, Peers string, StorageDir string) NodeCfg {
	return NodeCfg{
		Id,
		NetAddr,
		Peers,
		StorageDir,
	}
}

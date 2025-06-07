package api

import (
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/internal/core"
)

func NewRaftNode(sm curlyraft.StateMachine) *core.Node {
	return core.NewNode(core.NewNodeCfg(), &sm)
}

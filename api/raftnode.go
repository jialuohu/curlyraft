package api

import (
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/config"
	"github.com/jialuohu/curlyraft/internal/core"
)

func NewServer(cfg config.NodeCfg, sm curlyraft.StateMachine) (*core.RaftCore, error) {
	rc := core.NewRaftCore(cfg, sm)
	err := rc.Start()
	return rc, err
}

func RunServer(rc *core.RaftCore) error {
	return rc.Run()
}

func StopServer(rc *core.RaftCore) error {
	rc.Stop()
	return nil
}

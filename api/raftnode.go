package api

import (
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/config"
	"github.com/jialuohu/curlyraft/internal/core"
	"time"
)

func NewServer(cfg config.NodeCfg, sm curlyraft.StateMachine) (*core.RaftCore, error) {
	rc := core.NewRaftCore(cfg, &sm)
	err := rc.Start()
	return rc, err
}

func RunServer(rc *core.RaftCore) {
	//TODO implement me
	time.Sleep(1 * time.Second)
}

func StopServer(rc *core.RaftCore) {
	rc.Stop()
}

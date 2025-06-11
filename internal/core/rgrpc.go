package core

import (
	"context"
	"github.com/jialuohu/curlyraft/internal/clog"
	"github.com/jialuohu/curlyraft/internal/proto/gateway"
	"log"
)

func (rg *RaftGateway) AppendCommand(ctx context.Context, request *gateway.AppendCommandRequest) (*gateway.AppendCommandResponse, error) {
	cmd := request.GetCommand()
	log.Printf("%s Got command: %s\n", clog.CBlueRg("AppendCommand"), string(cmd))
	if err := rg.rc.propose(cmd); err != nil {
		return &gateway.AppendCommandResponse{Applied: false}, err
	}
	return &gateway.AppendCommandResponse{Applied: true}, nil
}

func (rg *RaftGateway) Check(ctx context.Context, request *gateway.HealthCheckRequest) (*gateway.HealthCheckResponse, error) {
	service := request.GetService()
	if service == "run" {
		return &gateway.HealthCheckResponse{Status: gateway.HealthCheckResponse_SERVING}, nil
	} else {
		return &gateway.HealthCheckResponse{Status: gateway.HealthCheckResponse_NOT_SERVING}, nil
	}
}

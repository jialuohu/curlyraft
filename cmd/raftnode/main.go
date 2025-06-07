package main

import (
	"github.com/jialuohu/curlyraft/api"
	"github.com/jialuohu/curlyraft/config"
	"log"
)

type smDummy struct{}

func (s smDummy) Apply(command []byte) (result []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (s smDummy) Snapshot() (snapshot []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (s smDummy) Restore(snapshot []byte) (err error) {
	//TODO implement me
	panic("implement me")
}

func main() {
	sm := smDummy{}
	cfg := config.NewNodeCfg(
		"nodeA",
		"localhost:21001",
		"nodeB/localhost:21002,nodeC/localhost:21003",
		"storage/21001",
	)

	log.Println("[main] Server starting...")

	rc, err := api.NewServer(cfg, sm)
	if err != nil {
		log.Fatalf("[main] StartServer Error: %v\n", err)
	}

	log.Println("[main] Server running...")
	if err := api.RunServer(rc); err != nil {
		log.Fatalf("[main] RunServer Error: %v\n", err)
	}

	log.Println("[main] Server stopping...")
	if err := api.StopServer(rc); err != nil {
		log.Fatalf("[main] StopServer Error: %v\n", err)
	}

	log.Println("[main] Server stopped successfully...")
}

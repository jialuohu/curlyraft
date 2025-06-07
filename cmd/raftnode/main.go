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
		log.Fatalf("[main] Error: %v\n", err)
	}
	log.Println("[main] Server started successfully!")

	api.RunServer(rc)

	log.Println("[main] Server stopping...")
	api.StopServer(rc)
	log.Println("[main] Server stopped successfully...")
}

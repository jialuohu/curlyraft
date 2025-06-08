package main

import (
	"github.com/jialuohu/curlyraft/api"
	"github.com/jialuohu/curlyraft/config"
	"log"
	"sync"
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
	var wg sync.WaitGroup
	wg.Add(3)

	go server1()
	go server2()
	go server3()

	wg.Wait()
}

func server1() {
	sm := smDummy{}
	cfg := config.NewNodeCfg(
		"nodeA",
		"localhost:21001",
		"nodeB/localhost:21002,nodeC/localhost:21003",
		"storage/21001",
	)

	log.Println("[server1] Server starting...")

	rc, err := api.NewServer(cfg, sm)
	if err != nil {
		log.Fatalf("[server1] StartServer Error: %v\n", err)
	}

	log.Println("[server1] Server running...")
	if err := api.RunServer(rc); err != nil {
		log.Fatalf("[server1] RunServer Error: %v\n", err)
	}

	log.Println("[server1] Server stopping...")
	if err := api.StopServer(rc); err != nil {
		log.Fatalf("[server1] StopServer Error: %v\n", err)
	}

	log.Println("[server1] Server stopped successfully...")
}
func server2() {
	sm := smDummy{}
	cfg := config.NewNodeCfg(
		"nodeB",
		"localhost:21002",
		"nodeA/localhost:21001,nodeC/localhost:21003",
		"storage/21002",
	)

	log.Println("[server2] Server starting...")

	rc, err := api.NewServer(cfg, sm)
	if err != nil {
		log.Fatalf("[server2] StartServer Error: %v\n", err)
	}

	log.Println("[server2] Server running...")
	if err := api.RunServer(rc); err != nil {
		log.Fatalf("[server2] RunServer Error: %v\n", err)
	}

	log.Println("[server2] Server stopping...")
	if err := api.StopServer(rc); err != nil {
		log.Fatalf("[server2] StopServer Error: %v\n", err)
	}

	log.Println("[server2] Server stopped successfully...")
}
func server3() {
	sm := smDummy{}
	cfg := config.NewNodeCfg(
		"nodeC",
		"localhost:21003",
		"nodeB/localhost:21002,nodeA/localhost:21001",
		"storage/21003",
	)

	log.Println("[server3] Server starting...")

	rc, err := api.NewServer(cfg, sm)
	if err != nil {
		log.Fatalf("[server3] StartServer Error: %v\n", err)
	}

	log.Println("[server3] Server running...")
	if err := api.RunServer(rc); err != nil {
		log.Fatalf("[server3] RunServer Error: %v\n", err)
	}

	log.Println("[server3] Server stopping...")
	if err := api.StopServer(rc); err != nil {
		log.Fatalf("[server3] StopServer Error: %v\n", err)
	}

	log.Println("[server3] Server stopped successfully...")
}

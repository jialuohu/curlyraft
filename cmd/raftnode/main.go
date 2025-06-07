package main

import (
	"github.com/jialuohu/curlyraft/api"
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
	sm := &smDummy{}
	api.NewRaftNode(sm)
	log.Println("Done!")
}

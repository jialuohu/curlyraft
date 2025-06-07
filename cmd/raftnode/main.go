package main

import (
	"github.com/jialuohu/curlyraft/api"
	"log"
)

type smDummy struct{}

func (s smDummy) Apply(command []byte) (result []byte, err error) {
	return nil, nil
}

func main() {
	sm := &smDummy{}
	api.NewRaftNode(sm)
	log.Println("Done!")
}

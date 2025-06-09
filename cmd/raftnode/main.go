package main

import (
	"flag"
	"fmt"
	"github.com/jialuohu/curlyraft/api"
	"github.com/jialuohu/curlyraft/config"
	"log"
)

type smDummy struct{}

func (s smDummy) Apply(command []byte) ([]byte, error) { panic("TODO") }
func (s smDummy) Snapshot() ([]byte, error)            { panic("TODO") }
func (s smDummy) Restore(snapshot []byte) error        { panic("TODO") }

func main() {
	// 1) Declare flags
	id := flag.String("id", "", "this node's ID (e.g. nodeA)")
	listen := flag.String("addr", "", "this node's listen address (e.g. localhost:21001)")
	peersCSV := flag.String("peers", "", "comma-separated peers, each as ID/addr")
	storage := flag.String("data", "", "local storage directory (e.g. storage/21001)")

	flag.Parse()

	// 2) Validate
	if *id == "" || *listen == "" || *peersCSV == "" || *storage == "" {
		log.Fatalf("usage: %s --id ID --addr ADDR --peers ID1/ADDR1,ID2/ADDR2 --data DIR",
			flag.CommandLine.Name())
	}

	// 3) Build the config
	cfg := config.NewNodeCfg(
		*id,
		*listen,
		*peersCSV,
		*storage,
	)

	// 4) Prefix logs with node ID + timestamp flags
	log.SetPrefix(fmt.Sprintf("[%s] ", cfg.Id))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// 5) Instantiate your dummy state machine (or real one!)
	sm := smDummy{}

	// 6) Create the server
	rc, err := api.NewServer(cfg, sm)
	if err != nil {
		log.Fatalf("[%-6s] StartServer Error: %v", cfg.Id, err)
	}

	// 7) Run until interrupted
	log.Printf("[%s] Server running...", cfg.Id)
	if err := api.RunServer(rc); err != nil {
		log.Fatalf("[%s] RunServer Error: %v", cfg.Id, err)
	}
	log.Printf("[%s] Server stopping...", cfg.Id)

	if err := api.StopServer(rc); err != nil {
		log.Fatalf("[%s] StopServer Error: %v", cfg.Id, err)
	}
	log.Printf("[%s] Server stopped cleanly.", cfg.Id)
}

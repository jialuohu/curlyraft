package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/jialuohu/curlyraft/api"
	"github.com/jialuohu/curlyraft/config"
	"log"
	"os"
)

type smDummy struct{}

func (s smDummy) Apply(command []byte) ([]byte, error) { panic("TODO") }
func (s smDummy) Snapshot() ([]byte, error)            { panic("TODO") }
func (s smDummy) Restore(snapshot []byte) error        { panic("TODO") }

func main() {
	var c = color.New(color.FgCyan).Add(color.Bold)

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
	l := log.New(os.Stdout, c.Sprintf(fmt.Sprintf("[%s/%s] ", cfg.Id, cfg.NetAddr)), log.Ldate|log.Ltime|log.Lmicroseconds)
	// 5) Instantiate your dummy state machine (or real one!)
	sm := smDummy{}

	// 6) Create the server
	l.Println("Server Starting...")
	rc, err := api.NewServer(cfg, sm)
	if err != nil {
		l.Fatalf("[%-6s] StartServer Error: %v", cfg.Id, err)
	}

	// 7) Run until interrupted

	l.Println("Server running...")
	if err := api.RunServer(rc); err != nil {
		l.Fatalf("[%s] RunServer Error: %v\n", cfg.Id, err)
	}
	l.Println("Server stopping...")

	if err := api.StopServer(rc); err != nil {
		l.Fatalf("[%s] StopServer Error: %v", cfg.Id, err)
	}
	l.Println("Server stopped cleanly")
}

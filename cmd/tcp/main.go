package main

import (
	"log"
	"os"
	"sync"
	"github.com/zpatrick/go-config"
	"github.com/qframe/collector-tcp"
	"github.com/qframe/types/qchannel"
	"github.com/qframe/handler-tcp"
	"github.com/qframe/handler-log"
)

const (
	dockerHost = "unix:///var/run/docker.sock"
	dockerAPI = "v1.30"
)

func checkErr(name string, err error) {
	if err != nil {
		log.Printf("[EE] Failed to create plugin %s: %v", name, err)
		os.Exit(0)
	}
}

func Run(qChan qtypes_qchannel.QChan, cfg *config.Config, name string) {
	p, _ := qcollector_tcp.New(qChan, cfg, name)
	p.Run()
}


func main() {
	qChan := qtypes_qchannel.NewQChan()
	qChan.Broadcast()
	cfgMap := map[string]string{
		"log.level": "debug",
		"collector.in.bind-port": "10001",
		"collector.loop.bind-port": "10002",
		"handler.out.inputs": "in",
		"handler.out.port": "10002",
		"handler.log.inputs": "loop",
	}

	cfg := config.NewConfig(
		[]config.Provider{
			config.NewStatic(cfgMap),
		},
	)
	ptl, err := qcollector_tcp.New(qChan, cfg, "loop")
	checkErr("loop", err)
	go ptl.Run()
	// log handler
	phl, err := qhandler_log.New(qChan, cfg, "log")
	checkErr("log", err)
	go phl.Run()
	// tcp handler
	pht, err := qhandler_tcp.New(qChan, cfg, "out")
	checkErr("out", err)
	go pht.Run()
	// Start TCP collectors
	p, err := qcollector_tcp.New(qChan, cfg, "in")
	checkErr("in", err)
	go p.Run()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

package main

import (
	"flag"
	"fmt"

	raftoperation "github.com/europelee/raftplay/internal/raft_operation"
	"github.com/europelee/raftplay/pkg/election"
	"github.com/europelee/raftplay/pkg/utils"
)

var raftBindAddr = utils.NetAddr{IP: "127.0.0.1", Port: 1000}
var raftDataDir = "/tmp/raft_data"
var raftEnableSingle = true
var raftPeers utils.NetAddrList
var servePort uint

func init() {
	flag.Var(&raftBindAddr, "raftBindAddr", "set raft bind address")
	flag.StringVar(&raftDataDir, "raftDataDir", raftDataDir, "set raft data directory")
	flag.Var(&raftPeers, "raftPeers", "set raft peers, default null")
	flag.BoolVar(&raftEnableSingle, "raftEnableSingle", raftEnableSingle, "force enable raft single node")
	flag.UintVar(&servePort, "servePort", servePort, "set serve port")
}

func main() {
	fmt.Println("raftplay start")
	flag.Parse()
	fmt.Println(raftPeers, raftEnableSingle, raftBindAddr, raftDataDir)
	electionInst := election.New(raftBindAddr, raftDataDir, raftPeers, raftEnableSingle)
	go electionInst.Start()
	go func() {
		s := utils.NewAPIServer()
		hdl := raftoperation.CreateHandler(electionInst)
		hdl.Register(s)
		utils.ListenAndServeAPIServer(
			"0.0.0.0",
			servePort)
	}()
	select {}
}

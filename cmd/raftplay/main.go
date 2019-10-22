package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	raftoperation "github.com/europelee/raftplay/internal/raft_operation"
	"github.com/europelee/raftplay/pkg/election"
	"github.com/europelee/raftplay/pkg/utils"
)

var raftBindAddr = utils.NetAddr{IP: "127.0.0.1", Port: 1000}
var raftDataDir = "/tmp/raft_data"
var raftEnableSingle = false
var raftPeers utils.NetAddrList
var servePort uint
var joinAddr string

func init() {
	flag.Var(&raftBindAddr, "raftBindAddr", "set raft bind address")
	flag.StringVar(&raftDataDir, "raftDataDir", raftDataDir, "set raft data directory")
	flag.StringVar(&joinAddr, "joinAddr", joinAddr, "leader serve addr")
	flag.Var(&raftPeers, "raftPeers", "set raft peers, default null")
	flag.BoolVar(&raftEnableSingle, "raftEnableSingle", raftEnableSingle, "force enable raft single node")
	flag.UintVar(&servePort, "servePort", servePort, "set serve port")
}

func join() error {
	b, err := json.Marshal(map[string]string{"id": raftBindAddr.String()})
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/operations/addVoter", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}

	defer resp.Body.Close()

	return nil
}

func main() {
	fmt.Println("raftplay start")
	flag.Parse()
	fmt.Println(raftPeers, raftEnableSingle, raftBindAddr, raftDataDir)
	electionInst := election.New(raftBindAddr, raftDataDir, raftPeers, raftEnableSingle)
	go electionInst.Start()
	if joinAddr != "" {
		fmt.Printf("join %s......", joinAddr)
		join()
	}
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

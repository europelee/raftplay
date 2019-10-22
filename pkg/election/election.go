package election

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/europelee/raftplay/pkg/utils"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

//OperationType for raft peers
type OperationType int

const (
	// Remove remove raft peer
	Remove OperationType = iota
)

//NodeState follower, candidate, leader
type NodeState int

var (
	// ErrNotLeader is returned when a node attempts to execute a leader-only
	// operation.
	ErrNotLeader = errors.New("not leader")
)

const (
	//Follower value zero
	Follower NodeState = iota
	//Candidate value 1
	Candidate
	//Leader value 2
	Leader
)

func (e NodeState) String() string {
	switch e {
	case Follower:
		return "Follower"
	case Candidate:
		return "Follower"
	case Leader:
		return "Leader"
	default:
		return fmt.Sprintf("unknow[%d]", int(e))
	}
}

// Election leader election with raft
type Election struct {
	raft         *raft.Raft
	raftBindAddr utils.NetAddr
	raftDataDir  string
	raftPeers    utils.NetAddrList
	nodeCurStat  NodeState
	NodeStatCh   chan NodeState
	OperationCh  chan OperationType
	logger       *log.Logger
	enableSingle bool
}

// Config ?
type Config struct {
	Bind    string `json:"bind"`
	DataDir string `json:"data_dir"`
	Peers   string `json:"peers"`
}

type fsm struct {
}

func (f *fsm) Apply(*raft.Log) interface{} {
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (f *fsm) Restore(io.ReadCloser) error {
	return nil
}

var configFilePath string

// New return an Election instance
func New(raftBindAddr utils.NetAddr, raftDataDir string, raftPeers utils.NetAddrList, enableSingle bool) *Election {
	logger := log.New(os.Stderr, "[election] ", log.LstdFlags)
	return &Election{
		raftBindAddr: raftBindAddr,
		raftDataDir:  raftDataDir,
		raftPeers:    raftPeers,
		nodeCurStat:  Follower,
		NodeStatCh:   make(chan NodeState),
		logger:       logger,
		enableSingle: enableSingle}
}

func (p *Election) RecvOperation(op OperationType) error {
	p.OperationCh <- op
	return nil
}

func (p *Election) RemovePeer(id string) error {
	return nil
}

func (p *Election) TransferLeaderShip(id, address string) error {
	future := p.raft.LeadershipTransferToServer(
		raft.ServerID(id), raft.ServerAddress(address))
	if future.Error() != nil {
		err := future.Error()
		p.logger.Printf("TransferLeaderShip %v", err)
		return err
	}
	return nil
}

func (p *Election) RemoveSvr(id string) error {
	if p.raft.State() != raft.Leader {
		return ErrNotLeader
	}
	f := p.raft.RemoveServer(raft.ServerID(id), 0, 0)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	p.logger.Printf("node at %s remove successfully", id)
	return nil
}

// Join let peers join into cluster
func (p *Election) Join(id, addr string) error {
	if p.raft.State() != raft.Leader {
		return ErrNotLeader
	}
	configFuture := p.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		p.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(id) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, the no
			// join is actually needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(id) {
				p.logger.Printf("node %s at %s already member of cluster, ignoring join request",
					id, addr)
				return nil
			}
			// todo: remove id from config
			p.logger.Printf("todo: remove id from config")
		}
	}
	f := p.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	p.logger.Printf("node at %s joined successfully", addr)
	return nil
}

//Start start and maintain leader election and monitor
func (p *Election) Start() error {
	err := os.MkdirAll(p.raftDataDir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	newNode := !utils.PathExists(path.Join(p.raftDataDir, "raft_db"))
	/**
	if p.enableSingle && !newNode {
		err := os.Remove(path.Join(p.raftDataDir, "raft_db"))
		if err != nil {
			fmt.Println(err)
		}
		newNode = true
	}
	*/
	cfg := raft.DefaultConfig()
	cfg.LogOutput = os.Stdout
	cfg.LocalID = raft.ServerID(p.raftBindAddr.String())
	dbStore, err := raftboltdb.NewBoltStore(path.Join(p.raftDataDir, "raft_db"))
	if err != nil {
		log.Fatal(err)
	}
	fileStore, err := raft.NewFileSnapshotStore(p.raftDataDir, 1, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	trans, err := raft.NewTCPTransport(p.raftBindAddr.String(), nil, 3, 5*time.Second, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	r, err := raft.NewRaft(cfg, &fsm{}, dbStore, dbStore, fileStore, trans)
	if err != nil {
		log.Fatal(err)
	}
	p.raft = r

	if newNode && p.enableSingle {
		var bootstrapConfig raft.Configuration

		bootstrapConfig = raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(p.raftBindAddr.String()),
					Address:  raft.ServerAddress(p.raftBindAddr.String()),
				},
			},
		}

		/**
		if !p.enableSingle {
			// Add known peers to bootstrap
			for _, node := range p.raftPeers {

				if node.String() == p.raftBindAddr.String() {
					continue
				}

				bootstrapConfig.Servers = append(bootstrapConfig.Servers, raft.Server{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(node.String()),
					Address:  raft.ServerAddress(node.String()),
				})
			}
		}
		*/
		f := r.BootstrapCluster(bootstrapConfig)
		if err := f.Error(); err != nil {
			log.Fatalf("error bootstrapping: %s", err)
		}
	} else {
		fmt.Println("no need boostrap")
	}

	t := time.NewTicker(time.Duration(3) * time.Second)
	defer func() {
		t.Stop()
		close(p.NodeStatCh)
	}()
	joinSucc := false
	for {
		select {
		case <-t.C:
			future := r.VerifyLeader()
			if err = future.Error(); err != nil {
				fmt.Println("Node is a follower")
				if p.nodeCurStat == Leader {
					fmt.Println("Node stat change to:", Follower)
					select {
					case p.NodeStatCh <- Follower:
						fmt.Println("p.NodeStatCh <- Follower")
					default:
						fmt.Println("default")
					}

					p.nodeCurStat = Follower
				}

			} else {
				fmt.Println("Node is leader")
				if p.nodeCurStat == Follower {
					fmt.Println("Node stat change to:", Leader)
					select {
					case p.NodeStatCh <- Leader:
						fmt.Println("p.NodeStatCh <- Leader")
					default:
						fmt.Println("default")
					}
					p.nodeCurStat = Leader
				}
			}

			if p.nodeCurStat == Leader && joinSucc == false {
				joinFail := false
				for _, node := range p.raftPeers {
					if node.String() == p.raftBindAddr.String() {
						continue
					}
					// we take id and address as the same value
					joinErr := p.Join(node.String(), node.String())
					if joinErr != nil {
						fmt.Println("joinErr != nil", joinErr.Error())
						joinFail = true
					} else {
						continue
					}
				}
				if !joinFail {
					joinSucc = true
				}
			}

		}
	}
}

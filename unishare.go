package unishare

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/logger"
	sm "github.com/lni/dragonboat/v4/statemachine"
)

type UniShare struct {
	datadir   string
	sm        *StateMachine
	registers []*RegisterPasserFunc
}

type USHandler func(sm *StateMachine, result *sm.Entry) (sm.Result, error)

func (us *UniShare) RegisterHandler(cmdstruct any, ushandler USHandler) {
	us.registers = append(us.registers, &RegisterPasserFunc{
		CmdStruct: cmdstruct,
		Dofunc:    ushandler,
	})
}

func (us *UniShare) StartNode(cfg *ConfigServer) {

	replicaID := cfg.ServerID

	// addr := cfg.Address()
	// if len(addr) == 0 && replicaID != 1 && replicaID != 2 && replicaID != 3 {
	// 	fmt.Fprintf(os.Stderr, "node id must be 1, 2 or 3 when address is not specified\n")
	// 	os.Exit(1)
	// }
	// https://github.com/golang/go/issues/17393
	if runtime.GOOS == "darwin" {
		signal.Ignore(syscall.Signal(0xd))
	}
	initialMembers := make(map[uint64]string)

	// when joining a new node which is not an initial members, the initialMembers
	// map should be empty.
	// when restarting a node that is not a member of the initial nodes, you can
	// leave the initialMembers to be empty. we still populate the initialMembers
	// here for simplicity.

	for idx, v := range cfg.Cluster {
		// key is the ReplicaID, ReplicaID is not allowed to be 0
		// value is the raft address
		initialMembers[uint64(idx+1)] = v
	}

	// for simplicity, in this example program, addresses of all those 3 initial
	// raft members are hard coded. when address is not specified on the command
	// line, we assume the node being launched is an initial raft member.

	var nodeAddr = initialMembers[uint64(replicaID)]

	fmt.Fprintf(os.Stdout, "node address: %s\n", nodeAddr)
	// change the log verbosity
	logger.GetLogger("dragonboat").SetLevel(logger.ERROR)
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("raftpb").SetLevel(logger.ERROR)
	logger.GetLogger("logdb").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.ERROR)
	logger.GetLogger("transport").SetLevel(logger.ERROR)
	logger.GetLogger("grpc").SetLevel(logger.ERROR)
	// config for raft node
	// See GoDoc for all available options
	rc := config.Config{
		// ShardID and ReplicaID of the raft node
		ReplicaID: uint64(replicaID),
		ShardID:   128,

		ElectionRTT: 10,

		HeartbeatRTT: 1,
		CheckQuorum:  true,

		SnapshotEntries: 10,

		CompactionOverhead: 5,
	}
	datadir := filepath.Join(
		us.datadir,
		fmt.Sprintf("node_%d", replicaID))

	nhc := config.NodeHostConfig{

		WALDir: datadir,
		// NodeHostDir is where everything else is stored.
		NodeHostDir: datadir,
		// RTTMillisecond is the average round trip time between NodeHosts (usually
		// on two machines/vms), it is in millisecond. Such RTT includes the
		// processing delays caused by NodeHosts, not just the network delay between
		// two NodeHost instances.
		RTTMillisecond: 200,
		// RaftAddress is used to identify the NodeHost instance
		RaftAddress: nodeAddr,
	}

	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		panic(err)
	}

	if err := nh.StartReplica(initialMembers, false, NewStateMachine, rc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
		os.Exit(1)
	}

}

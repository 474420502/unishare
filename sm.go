package unishare

import (
	"context"
	"io"
	"sync"

	"github.com/474420502/passer"
	sm "github.com/lni/dragonboat/v4/statemachine"
)

// StateMachine实现了StateMachine接口,用于管理基于PriorityQueue的消息队列
type StateMachine struct {
	// 所属的Shard ID
	shardID uint64
	// Replica ID
	replicaID uint64

	// 互斥锁,保护队列Map的并发访问
	mu *sync.Mutex
	// 组名到队列的映射
	fsPasser *passer.Passer[sm.Result]
}

// // NewStateMachine creates and return a new ExampleStateMachine object.
func NewStateMachine(shardID uint64, replicaID uint64) sm.IStateMachine {
	mu := &sync.Mutex{}
	return &StateMachine{
		shardID:   shardID,
		replicaID: replicaID,

		mu: mu,
	}
}

type RegisterPasserFunc struct {
	CmdStruct any
	Dofunc    USHandler
}

var registers []*RegisterPasserFunc

var fsPasser *passer.Passer[sm.Result] = func() *passer.Passer[sm.Result] {
	fsPasser := passer.NewPasser[sm.Result]()

	for _, r := range registers {
		fsPasser.RegisterPasser(r.CmdStruct, func(ctx context.Context, obj any) (sm.Result, error) {
			return r.Dofunc(ctx.Value(ctxStateMachine{}).(*StateMachine), ctx.Value(ctxEntry{}).(*sm.Entry))
		})
	}

	return fsPasser
}()

// Lookup performs local lookup on the ExampleStateMachine instance. In this example,
// we always return the Count value as a little endian binary encoded byte
// slice.
func (s *StateMachine) Lookup(group interface{}) (item interface{}, err error) {
	return item, nil
}

type ctxEntry struct{}
type ctxStateMachine struct{}

// Update处理Entry中的更新命令
// Update updates the object using the specified committed raft entry.
func (s *StateMachine) Update(e sm.Entry) (result sm.Result, err error) {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, ctxEntry{}, &e)
	ctx = context.WithValue(ctx, ctxStateMachine{}, s)
	return s.fsPasser.ExecuteWithBytes(ctx, e.Cmd)
}

// SaveSnapshot saves the current IStateMachine state into a snapshot using the
// specified io.Writer object.
func (s *StateMachine) SaveSnapshot(w io.Writer,
	fc sm.ISnapshotFileCollection, done <-chan struct{}) error {
	// as shown above, the only state that can be saved is the Count variable
	// there is no external file in this IStateMachine example, we thus leave
	// the fc untouched

	s.mu.Lock()
	defer s.mu.Unlock()

	// return gob.NewEncoder(w).Encode(&s.queues)
	return nil
}

// RecoverFromSnapshot recovers the state using the provided snapshot.
func (s *StateMachine) RecoverFromSnapshot(r io.Reader,
	files []sm.SnapshotFile,
	done <-chan struct{}) error {
	// restore the Count variable, that is the only state we maintain in this
	// example, the input files is expected to be empty

	// err := gob.NewDecoder(r).Decode(&s.queues)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Close closes the IStateMachine instance. There is nothing for us to cleanup
// or release as this is a pure in memory data store. Note that the Close
// method is not guaranteed to be called as node can crash at any time.
func (s *StateMachine) Close() error { return nil }

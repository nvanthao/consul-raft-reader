package app

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"go.etcd.io/bbolt"
)

var logTypes = map[raft.LogType]string{
	raft.LogCommand:              "LogCommand",
	raft.LogNoop:                 "LogNoop",
	raft.LogAddPeerDeprecated:    "LogAddPeerDeprecated",
	raft.LogRemovePeerDeprecated: "LogRemovePeerDeprecated",
	raft.LogBarrier:              "LogBarrier",
	raft.LogConfiguration:        "LogConfiguration",
}

type Store struct {
	store      *raftboltdb.BoltStore
	FirstIndex uint64
	LastIndex  uint64
}

type RaftLog struct {
	Term    uint64
	Index   uint64
	Type    string
	MsgType string
}

type kv struct {
	Key   string
	Value int
}

// This function will read the Raft log at given index
func (s *Store) Read(index uint64) error {
	if index < s.FirstIndex || index > s.LastIndex {
		return fmt.Errorf("index %d out of range [%d, %d]", index, s.FirstIndex, s.LastIndex)
	}

	var log raft.Log
	err := s.store.GetLog(index, &log)

	if err != nil {
		return fmt.Errorf("unable to read log at index %d", index)
	}

	fmt.Printf("Index: %d \n", log.Index)
	fmt.Printf("Term: %d \n", log.Term)
	fmt.Printf("Log Type: %s \n", logTypes[log.Type])
	if log.Type == raft.LogCommand {
		fmt.Printf("Message Type: %s\n", getMessageType(log.Data[0]))
	}
	if len(log.Data) > 1 {
		fmt.Println("Data:")
		fmt.Println(getLogCommand(log.Data))
	}

	return nil
}

func (s *Store) Print(start, end uint64) {
	logs := s.getLogs(start, end)
	for _, log := range logs {
		fmt.Printf("Index: %d ", log.Index)
		fmt.Printf("Term: %d ", log.Term)
		fmt.Printf("Log Type: %s ", log.Type)
		if log.MsgType != "" {
			fmt.Printf("Message Type: %s ", log.MsgType)
		}
		fmt.Println()
	}
}

func (s *Store) Stats() {
	fmt.Printf("First Index: %d \n", s.FirstIndex)
	fmt.Printf("Last Index: %d \n", s.LastIndex)
	fmt.Printf("Current Term: %s \n", s.getBucketValue("CurrentTerm"))
	fmt.Printf("Last Vote Term: %s \n", s.getBucketValue("LastVoteTerm"))
	fmt.Printf("Last Vote Candidate: %s \n", s.getBucketValue("LastVoteCand"))
	fmt.Printf("=== COUNT MESSAGE TYPES === \n")
	logs := s.getLogs(s.FirstIndex, s.LastIndex)
	counts := countMsgType(logs)
	ss := sortMapByValue(counts)
	for _, s := range ss {
		fmt.Printf("%s: %d \n", s.Key, s.Value)
	}
}

func NewStore(file string) (*Store, error) {
	store, err := raftboltdb.New(raftboltdb.Options{
		BoltOptions: &bbolt.Options{
			NoFreelistSync: true,
		},
		Path: filepath.Join(file),
	})
	if err != nil {
		return nil, err
	}

	FirstIndex, _ := store.FirstIndex()
	LastIndex, _ := store.LastIndex()

	return &Store{
		store:      store,
		FirstIndex: FirstIndex,
		LastIndex:  LastIndex}, nil
}

// this function is trying to guess the command type from the log data
// and default to a string(data) if it can't guess
func getLogCommand(data []byte) string {
	var st map[string]interface{}

	// first we decode the data into our map[string]interface{}
	if err := structs.Decode(data[1:], &st); err != nil {
		return string(data[1:])
	}

	// then we encode it to a pretty json string
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return string(data[1:])
	}

	return string(b)
}

// get logs from a range
func (s *Store) getLogs(start, end uint64) (logs []RaftLog) {
	for i := start; i <= end; i++ {
		var log raft.Log
		err := s.store.GetLog(i, &log)

		// we don't worry much about not being able to read a log here
		if err != nil {
			continue
		}
		var msgType string
		if log.Type == raft.LogCommand {
			msgType = getMessageType(log.Data[0])
		}

		logs = append(logs, RaftLog{
			Term:    log.Term,
			Index:   log.Index,
			Type:    logTypes[log.Type],
			MsgType: msgType,
		})
	}
	return
}

func getMessageType(msgType byte) string {
	if msgType == 134 {
		return "CoordinateUpdate"
	}
	return structs.MessageType(msgType).String()
}

func (s *Store) getBucketValue(key string) string {
	value, err := s.store.Get([]byte(key))
	if err != nil {
		return "Unknown"
	}
	return string(value)
}

func countMsgType(logs []RaftLog) map[string]int {
	memo := make(map[string]int)
	for _, log := range logs {
		if log.MsgType != "" {
			memo[log.MsgType]++
		}
	}
	return memo
}

func sortMapByValue(m map[string]int) []kv {
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss
}

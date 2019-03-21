package chubby

// Provide distributed consensus services.
// Some code adapted from freno:
// https://github.com/github/freno/blob/master/go/group/raft.go

import (
	"fmt"
	"github.com/hashicorp/raft"
)

var store *Store

func Setup(inmem bool) (*Store, error) {
	store = New(inmem)

	if err := store.Open(); err != nil {
		return nil, fmt.Errorf(
			"failed to open Raft store: %s",
			err.Error())
	}

	return store, nil
}

// Private method for retrieving Raft instance
func getRaft() *raft.Raft {
	return store.raft
}

// Returns if this node is the current Raft leader
func IsLeader() bool {
	return GetState() == raft.Leader
}

// Returns the identity of the Raft leader
func GetLeader() raft.ServerAddress {
	return getRaft().Leader()
}

// Get the state of this node: Follower, Candidate, Leader, or Shutdown
func GetState() raft.RaftState {
	return getRaft().State()
}

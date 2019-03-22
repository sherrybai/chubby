// Provide distributed consensus services.
// Some code adapted from freno:
// https://github.com/github/freno/blob/master/go/group/raft.go
// TODO: get rid of this file -- doesn't really do anything right now

package chubby

import (
	"github.com/hashicorp/raft"
	"cos518project/chubby/store"
)

var s *store.Store

// Private method for retrieving Raft instance
func getRaft() *raft.Raft {
	return s.raft
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

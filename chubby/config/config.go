// Configuration for a Raft node.
//
// Adapted from leto config file:
// https://github.com/yongman/leto/blob/master/config/config.go

package config

type Config struct {
	Listen   string
	RaftDir  string
	RaftBind string
	Join     string
	NodeID   string
	InMem	 bool
}

func NewConfig(listen, raftDir, raftBind, nodeId, join string, inmem bool) *Config {
	return &Config{
		Listen:   listen,
		RaftDir:  raftDir,
		RaftBind: raftBind,
		NodeID:   nodeId,
		Join:     join,
		InMem:    inmem,
	}
}
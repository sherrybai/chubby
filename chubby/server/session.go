package server

import (
    "github.com/hashicorp/raft"
    "net"
)


type Session struct {
    trans               *raft.NetworkTransport
    currConn            *net.Conn
    raftServer          raft.ServerAddress
}
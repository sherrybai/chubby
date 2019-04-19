package session

import (
    "net"
    "time"
    "fmt"
    "errors"
    "bufio"
    "github.com/hashicorp/raft"
    "github.com/hashicorp/go-msgpack/codec"
)


type Session struct {
    trans               *raft.NetworkTransport
    currConn            *net.Conn
    raftServer         ServerAddress
}
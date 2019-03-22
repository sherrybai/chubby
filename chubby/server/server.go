// Define Chubby server application.
// Adapted from Leto server file:
// https://github.com/yongman/leto/blob/master/server/server.go

package server

import (
	"cos518project/chubby/config"
	"cos518project/chubby/store"
	"fmt"
	"log"
	"net"
	"os"
)

type App struct {
	listener net.Listener

	// wrapper and manager for db instance
	store *store.Store

	logger *log.Logger
}

func NewApp(conf *config.Config) *App {
	var err error
	app := &App{}

	// Create a new logger.
	app.logger = log.New(os.Stderr, "[server] ", log.LstdFlags)

	// Create a new store.
	app.store = store.New(conf.RaftDir, conf.RaftBind, conf.InMem)

	// Open the store.
	bootstrap := conf.Join == ""
	err = app.store.Open(bootstrap, conf.NodeID)

	if !bootstrap {
		// TODO: figure out how to join nodes to existing clusters
	}

	// Listen for client connections.
	app.listener, err = net.Listen("tcp", conf.Listen)
	app.logger.Printf("server listen in %s", conf.Listen)
	if err != nil {
		fmt.Println(err.Error())
	}

	return app
}

func (app *App) Run() {
	// Accept connections.
	for {
		select {
		default:
			// Accept new client connection.
			conn, err := app.listener.Accept()
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			// Handle connection.
			// ClientHandler(conn, app)
		}
	}
}

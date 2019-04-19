// Define Chubby server application.

// Adapted from Leto server file:
// https://github.com/yongman/leto/blob/master/server/server.go

// Copyright (C) 2018 YanMing <yming0221@gmail.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package server

import (
	"cos518project/chubby/config"
	"cos518project/chubby/store"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
)

// No choice but to make this variable package-level :(
var app *App

type App struct {
	listener net.Listener

	// wrapper and manager for db instance
	store *store.Store

	logger *log.Logger

}

type Lock struct {
	path string
	exclusive_owner string
	shared_owners []string
	value string
}

func Run(conf *config.Config) {
	var err error

	// Init app struct.
	app = &App{}

	// Create a new logger.
	app.logger = log.New(os.Stderr, "[server] ", log.LstdFlags)

	// Create a new store.
	app.store = store.New(conf.RaftDir, conf.RaftBind, conf.InMem)

	// Initialize a map
	app.file_handlers = make(map[int]string)

	// Open the store.
	bootstrap := conf.Join == ""
	err = app.store.Open(bootstrap, conf.NodeID)
	if err != nil {
		log.Fatal(err)
	}

	if !bootstrap {
		// Set up TCP connection.
		client, err := rpc.Dial("tcp", conf.Join)
		if err != nil {
			log.Fatal(err)
		}

		app.logger.Printf("set up connection to %s", conf.Join)

		var req JoinRequest
		var resp EmptyResponse

		req.RaftAddr = conf.RaftBind
		req.NodeID = conf.NodeID

		err = client.Call("Handler.Join", req, &resp)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Listen for client connections.
	handler := new(Handler)
	err = rpc.Register(handler)

	app.listener, err = net.Listen("tcp", conf.Listen)
	app.logger.Printf("server listen in %s", conf.Listen)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Accept connections.
	rpc.Accept(app.listener)
}

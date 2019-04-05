package server

//import (
//	//"bufio"
//	"bytes"
//	"cos518project/chubby/store"
//	"errors"
//	"io"
//	"strings"
//
//	//"io"
//	"log"
//	"net"
//	"os"
//	//"strings"
//)
//

//type Command struct {
//	cmd  string
//	args [][]byte
//}
//
//type Client struct {
//	app   *App
//	store *store.Store
//
//	// The command this client is processing.
//	command *Command
//
//	buf bytes.Buffer
//
//	conn net.Conn
//
//	logger *log.Logger
//}
//
//func ClientHandler(conn net.Conn, app *App) {
//	c := newClient(app)
//
//	c.conn = conn
//
//	go c.connHandler()
//}
//
//func newClient(app *App) *Client {
//	client := &Client{
//		app:    app,
//		store:  app.store,
//		logger: log.New(os.Stderr, "[client] ", log.LstdFlags),
//	}
//	return client
//}


//func (c *Client) connHandler() {
//	defer func(c *Client) {
//		c.conn.Close()
//	}(c)
//
//	for {
//		c.command = &Command{
//			cmd: "",
//			args: nil
//		}
//
//		req, err := c.rReader.ReadString('\n')
//		if err != nil && err != io.EOF {
//			c.logger.Println(err.Error())
//			return
//		} else if err != nil {
//			return
//		}
//		err = c.handleRequest(req)
//		if err != nil && err != io.EOF {
//			c.logger.Println(err.Error())
//			return
//		}
//	}
//}
//
//func (c *Client) handleRequest(req string) error {
//	if len(req) == 0 {
//		c.cmd = ""
//		c.args = nil
//	} else {
//		c.cmd = strings.ToLower(string(req[0]))
//
//	}
//}

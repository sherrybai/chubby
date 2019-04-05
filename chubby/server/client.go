package server

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type Command struct {
	cmd  string
	args [][]byte
}

type Client struct {
	app   *App
	store *store.Store

	// request is processing
	cmd  string
	args [][]byte

	buf bytes.Buffer

	conn net.Conn

	logger *log.Logger

	rWriter *Writer
	rReader *Reader
}

func ClientHandler(conn net.Conn, app *App) {
	c := newClient(app)

	c.conn = conn
	// connection buffer setting

	br := bufio.NewReader(conn)
	c.rReader = br

	bw := bufio.NewWriter(conn)
	c.rWriter = bw

	go c.connHandler()
}

func newClient(app *App) *Client {
	client := &Client{
		app:    app,
		store:  app.store,
		logger: log.New(os.Stderr, "[client] ", log.LstdFlags),
	}
	return client
}


func (c *Client) connHandler() {
	def func(c *Client) {
		c.conn.Close()
	}(c)

	for {
		c.cmd = ""
		c.args = nil

		req, err := c.rReader.ReadString('\n')
		if err != nil && err != io.EOF {
			c.logger.Println(err.Error())
			return
		} else if err != nil {
			return
		}
		err = c.handleRequest(req)
		if err != nil && err != io.EOF {
			c.logger.Println(err.Error())
			return
		}
	}
}

function (c *Client) handleRequest(req string) error {
	if len(req) == 0 {
		c.cmd = ""
		c.args = nil
	} else {
		c.cmd = strings.ToLower(string(req[0]))

	}
}


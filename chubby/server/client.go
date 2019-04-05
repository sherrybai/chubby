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

func (c *Client) Resp(resp string) error {
	var err error = nil

	err = c.rWriter.WriteString(resp)
	return err
}


func (c *Client) FlushResp(resp string) error {
	err := c.Resp(resp)
	if err != nil {
		return err
	}
	return c.rWriter.Flush()
}


function (c *Client) handleRequest(req string) error {
	if len(req) == 0 {
		c.cmd = ""
		c.args = nil
	} else {
		c.cmd = strings.ToLower(string(req))
		c.args = req
	}

	var (
		err error
		v string
	)

	c.logger.Printf("process %s command", c.cmd)

	switch c.cmd {
	case "acquire":
		if v, err = c.handleAcquire(); err == nil {
			c.FlushResp("OK")
		} else {
			c.FlushResp("Fail to acquire")
		}
	case "release":
		if v, err = c.handleRelease(); err == nil {
			c.FlushResp("OK")
		} else {
			c.FlushResp("Fail to release")
		}
	case "delete":
		if v, err = c.handleDelete(); err == nil {
			c.FlushResp("OK")
		} else {
			c.FlushResp("Fail to delete")
		}
	case "create":
		if v, err = c.handleCreate(); err == nil {
			c.FlushResp("OK")
		} else {
			c.FlushResp("Fail to create")
		}
	default:
		err = ErrCmdNotSupport
	}
	if err != nil {
		c.FlushResp(err)
	}

	return err
}


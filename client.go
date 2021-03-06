// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomudnet

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type client_state uint

const (
	cl_state_connected  client_state = 0
	cl_state_ready      client_state = 1
	cl_state_closing    client_state = 2
	cl_state_waiting_in client_state = 3
	cl_state_ready_out  client_state = 4
	cl_state_ready_in   client_state = 5
	cl_state_sleeping   client_state = 6
	cl_state_waiting    client_state = 99
	INPUT_BUFFER_SIZE   uint64       = 12 * 1024
	OUTPUT_BUFFER_SIZE  uint64       = 24 * 1024
)

var (
	ErrConnectionNilOrClosed = errors.New("read connection: closed or nil")
	clientHandler            *client_handler
)

// Client represents a client connection.
type Client struct {
	conn          net.Conn
	state         client_state
	bufferIn      []byte
	bufferOut     []byte
	pipeline      *Pipeline
	delimitterOut byte
	delimitterIn  byte
}

// newClient creates a new Client and sets the Pipe of all the ChannelHandlers
// in p to p.
func newClient(conn net.Conn, p *Pipeline) *Client {
	cl := new(Client)
	cl.conn = conn
	cl.pipeline = p
	cl.bufferIn = make([]byte, INPUT_BUFFER_SIZE)
	cl.bufferOut = make([]byte, OUTPUT_BUFFER_SIZE)
	cl.pipeline.owner = cl
	cl.state = cl_state_connected

	for e := cl.pipeline.channelHandlers.Front(); e != nil; e = e.Next() {
		if cMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for _, handler := range cMap {
				handler.SetPipeline(cl.pipeline)
			}
		}
	}

	return cl
}

// Close sends a signal to the client handler to start the termination of this client connection.
func (cl *Client) Close() {
	go func() {
		if DEBUG {
			log.Printf("Closing signal received for %s", cl)
		}

		cl.state = cl_state_closing
	}()
}

// String is the readable string representation of this Client.
func (cl *Client) String() string {
	if cl.conn != nil {
		return cl.conn.RemoteAddr().String()
	} else {
		return fmt.Sprintf("<closed connection>@%v", &cl)
	}
}

// sendMessage sends the provided Message to the client connection and waits
// for it to complete, then it informs the Pipeline that the Message was sent.
func (cl *Client) sendMessage(msg Message, delimitter byte) {
	cl.bufferOut = msg.Bytes()
	cl.delimitterOut = delimitter
	clientHandler.processOutput(cl)
	cl.pipeline.messageSent(msg, cl)

}

// client_handler is the placeholder for all client connections and is also responsible
// for handling client events like, connecting, closing, processing input and output, etc.
type client_handler struct {
	clientList *list.List
	running    bool
}

// newClientHandler creates a new client_handler.
func newClientHandler() *client_handler {
	clientHandler = new(client_handler)
	clientHandler.clientList = list.New()
	return clientHandler
}

// newConnection creates a new Client with the given connection and the initial
// Pipeline.
func (ch *client_handler) newConnection(conn net.Conn, pipe *Pipeline) *Client {
	cl := newClient(conn, pipe)
	ch.clientList.PushFront(cl)
	cl.state = cl_state_connected
	return cl
}

// dispatchClientHandler is the main routine for handling client events.
func (ch *client_handler) dispatchClientHandler() (err error) {
	defer func() {
		currentServerState <- state_cl_handler_stopped
	}()

	if DEBUG {
		log.Printf("Client handler started...")
	}
	ch.running = true
	for ch.running {
		for e := ch.clientList.Front(); e != nil; e = e.Next() {
			if cl, ok := e.Value.(*Client); ok {
				switch cl.state {
				case cl_state_connected:

					if DEBUG {
						log.Printf("New pipe opened: %s", cl.pipeline)
					}

					go cl.pipeline.opened(cl)
					cl.state = cl_state_ready_in
				case cl_state_ready:
					go cl.pipeline.send(cl, cl.bufferIn, DIR_UPSTREAM)
					cl.state = cl_state_ready_in
				case cl_state_closing:
					if DEBUG {
						log.Printf("Closing client :", cl.conn.RemoteAddr())
					}
					cl.pipeline.closing(cl)
					// this is forcibly closing the connection, although in
					// Server.startListening() all accepted connections are deferred
					// to close
					cl.conn.Close()
					ch.clientList.Remove(e)
					go cl.pipeline.closed(cl)
				case cl_state_ready_in:
					go ch.processInput(cl)
				case cl_state_ready_out:
					go ch.processOutput(cl)
				default:
					//sleeping
				}
			}
		}

		time.Sleep(time.Millisecond * 10)
	}

	if DEBUG {
		log.Printf("Client handler stopped.")
	}
	return
}

// processInput reads from waits for input from cl and copies it to it's buffer.
func (ch *client_handler) processInput(cl *Client) (err error) {
	cl.state = cl_state_sleeping
	if cl.conn != nil {
		if DEBUG {
			log.Printf("Waiting for input from %s.", cl)
		}
		buf := make([]byte, INPUT_BUFFER_SIZE)
		if count, err := cl.conn.Read(buf); err == nil {
			cl.bufferIn = buf[:count]

			if DEBUG {
				log.Printf("%s RCVD: (%d)%v", cl.conn.RemoteAddr(), count, cl.bufferIn)
			}
			cl.state = cl_state_ready
		} else {
			log.Printf("ERROR: read from %s, reason: %s", cl, err.Error())
			if err.Error() == "EOF" {
				// inform the pipeline
				cl.pipeline.closed(cl)
			}
		}
	} else {
		err = ErrConnectionNilOrClosed
	}
	if err != nil {
		log.Printf("ERROR: Reading from %s: %s", cl, err.Error())
	}
	return
}

// processOutput writes to cl's output buffer to it's network file descriptor.
func (ch *client_handler) processOutput(cl *Client) (err error) {
	if cl.conn != nil {
		if count, err := cl.conn.Write(cl.bufferOut); err == nil {
			if DEBUG {
				log.Printf("%s SEND: (%d)%v", cl.conn.RemoteAddr(), count, cl.bufferOut)
			}
		} else {
			if DEBUG {
				log.Printf("ERROR: Writing to %s: %s", cl, err.Error())
			}
		}
	} else {
		err = ErrConnectionNilOrClosed
	}

	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}

	return
}

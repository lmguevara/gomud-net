// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomudnet

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

// The Server state type.
type server_state uint

const (
	// The Server is starting.
	state_server_starting server_state = 0
	// The listers has started.
	state_listeners_started server_state = 1
	// The client_handler has started.
	state_cl_handler_started server_state = 2
	// The server is stopping.
	state_server_stopping server_state = 3
	// The client_handler has stopped.
	state_cl_handler_stopped server_state = 4
	// The listeners has stopped.
	state_listeners_stopped server_state = 5
	// The Server has completely stopped. Outside of loop.
	state_server_stopped server_state = 6
)

var (
	ErrNilOrEmptyPorts            = errors.New("Nil or empty ports")
	ErrNilOrEmptyHosts            = errors.New("Nil or empty hosts")
	ErrNilOrEmptyPipelineFactory  = errors.New("Nil or empty pipeline factory")
	ErrNoValidHostPortCombination = errors.New("No valid host/port combination started")

	currentServerState chan server_state
)

// ServerEventHandler provides functions that are called to handle specfic events
// in Server's lifetime.
type ServerEventHandler interface {
	// StartingListener handles the event when Server is preparing to listen on
	// host and port.
	StartingListener(string, uint)

	// StartedListener handles the event when Server has established the listener
	// on host and port.
	StartedListener(string, uint)

	// ListenersStarted handles the event when all the Server's valid listeners
	// have started.
	//
	// This event can also consider the Server to be running.
	ListenersStarted()

	// ClosingListener handles the event when the provided listener is preparing
	// to close.
	ClosingListener(net.Listener)

	// ClosedListener handles the event when the provided listener is completely
	// closed.
	ClosedListener(net.Listener)

	// ListenersClosed handles the event when all the Server's previously opened
	// listeners are all closed.
	ListenersClosed()

	// StoppedCompletely handles the event when Server has completelt stopped.
	StoppedCompletely()

	// ErrorEncountered handles the event when Server encouters an error.
	ErrorEncountered(error)
}

// The main server
type Server struct {
	ports           []uint
	hosts           []string
	listeners       []net.Listener
	pipelineFactory *PipelineFactory
	clientHandler   *client_handler
	eventHandlers   []ServerEventHandler
}

// Creates a new Server with the given parameters.
func NewServer(ports []uint, hosts []string, pipelineFactory *PipelineFactory) *Server {
	s := new(Server)
	s.ports = ports
	s.hosts = hosts
	s.pipelineFactory = pipelineFactory
	s.eventHandlers = make([]ServerEventHandler, 0)
	currentServerState = make(chan server_state)

	return s
}

// Close sends a signal to the server to start the termination process. It will then
// wait for the signal that the server has completely topped.
func (s *Server) Close() {
	go func() {
		currentServerState <- state_server_stopping
	}()
	stopped := false
	for !stopped {
		select {
		case state := <-currentServerState:
			if state == state_server_stopped {
				stopped = true
			} else {
				currentServerState <- state
			}
		default:
		}

		time.Sleep(time.Millisecond)
	}

	if DEBUG {
		log.Printf("Server closed.")
	}
	for _, h := range s.eventHandlers {
		h.StoppedCompletely()
	}
}

// Start runs this Server with the current configuration.
//
// This will bind to the hosts and ports combination provided by NewServer() and
// start listening for connections. If blocking is true this will run the server
// loop on the same routine, preventing it from returning until the signal to
// stop is received. Otherwise the server loop will be in a separate go routine.
func (s *Server) Start(blocking bool) (err error) {

	if s.ports == nil || len(s.ports) == 0 {
		return ErrNilOrEmptyPorts
	}

	if s.hosts == nil || len(s.hosts) == 0 {
		return ErrNilOrEmptyHosts
	}

	if s.pipelineFactory == nil || s.pipelineFactory.channelHandlers.Len() == 0 {
		return ErrNilOrEmptyPipelineFactory
	}
	s.listeners = make([]net.Listener, len(s.hosts)*len(s.ports))
	for _, h := range s.hosts {
		for _, p := range s.ports {

			for _, handler := range s.eventHandlers {
				handler.StartingListener(h, p)
			}
			go s.startListening(h, p)
		}
	}

	//wait for all listeners to start
	<-currentServerState

	//check for at least one listener online
	found := false
	for _, ln := range s.listeners {
		if ln != nil {
			found = true
			break
		}
	}

	if !found {
		return ErrNoValidHostPortCombination
	}

	//dispatch the client handler
	go func() {
		s.clientHandler = newClientHandler()
		err = s.clientHandler.dispatchClientHandler()
	}()

	if err != nil {
		return err
	}

	if DEBUG {
		log.Printf("Server started with no errors")
	}

	if !blocking {
		go func() {
			s.wait()
		}()
	} else {
		s.wait()
	}
	return err
}

func (s *Server) AddEventHandler(h ServerEventHandler) {
	s.eventHandlers = append(s.eventHandlers, h)
}

func (s *Server) RemoveEventHandler(h ServerEventHandler) {
	i := 0
	for ; i < len(s.eventHandlers); i++ {
		if s.eventHandlers[i] == h {
			break
		}
	}
	temp := s.listeners[:i-1]
	temp = append(temp, s.listeners[i+1:]...)
	s.listeners = temp
}

// wait is the main server loop
func (s *Server) wait() {
	//the main server loop
	running := true
	for running {
		select {
		case state := <-currentServerState:
			if state == state_server_stopping {
				//stop client handler
				s.clientHandler.running = false
			} else if state == state_cl_handler_stopped {
				//close listeners
				for _, ln := range s.listeners {

					for _, h := range s.eventHandlers {
						h.ClosingListener(ln)
					}
					ln.Close()
				}
			} else if state >= state_listeners_stopped {
				running = false
			} else {
				currentServerState <- state
			}
		default:
		}
		time.Sleep(time.Millisecond)
	}
	if DEBUG {
		log.Printf("Server stopped.")
	}
	currentServerState <- state_server_stopped

}

// startListening starts listening to the provided host and port.
func (s *Server) startListening(host string, port uint) (err error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))

	if err == nil {
		if DEBUG {
			log.Printf("Listening on %s", ln.Addr())
		}

		for _, h := range s.eventHandlers {
			h.StartedListener(host, port)
		}

		ctr := 0
		for i := range s.listeners {
			ctr++
			if s.listeners[i] == nil {
				s.listeners[i] = ln
				break
			}
		}

		if ctr >= len(s.listeners) {
			// we're done starting the valid listeners
			// inform Start() about it
			currentServerState <- state_listeners_started
			if DEBUG {
				log.Printf("Listeners started.")
			}

			for _, h := range s.eventHandlers {
				h.ListenersStarted()
			}
		}

		if ctr == 1 {
			defer func() {
				//everything should be close now
				if DEBUG {
					log.Printf("All Listeners closed.")
				}
				// wait() will be waiting for this signal after issuin Close()
				// to all listeners
				currentServerState <- state_listeners_stopped

				for _, h := range s.eventHandlers {
					h.ListenersClosed()
				}
			}()
		}

		running := true
		for running {
			if conn, err := ln.Accept(); err == nil {
				if DEBUG {
					log.Printf("Connection received from: %s", conn.RemoteAddr())
				}
				pipe := NewPipeline()
				pipe.channelHandlers.PushFrontList(s.pipelineFactory.channelHandlers)
				cl := s.clientHandler.newConnection(conn, pipe)

				defer func() {
					if DEBUG {
						log.Printf("Closing %s", conn.RemoteAddr())
					}
					conn.Close()
					pipe.closed(cl)
				}()

				select {
				case state := <-currentServerState:
					if state >= state_server_stopping {
						running = false
					} else {
						//resend unhandled state
						currentServerState <- state
					}
				default:
				}
			} else {
				for _, h := range s.eventHandlers {
					h.ErrorEncountered(err)
				}
				running = false
			}
		}
		if DEBUG {
			log.Printf("%s close.", ln.Addr())
		}

		for _, h := range s.eventHandlers {
			h.ClosedListener(ln)
		}
	} else {
		for _, h := range s.eventHandlers {
			h.ErrorEncountered(err)
		}
	}

	return
}

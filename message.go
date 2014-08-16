// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomudnet

import "fmt"

// This interface defines the methods required for an object to be able to move
// through a pipeline.
type Message interface {
	// The string equivalent of the message content.
	String() string

	// The slice of bytes of the message content.
	Bytes() []byte

	// Copies the message content to p and returns the number of bytes copied.
	Read(p []byte) (n int, err error)

	// Copies the p to the message content and returns the number of bytes copied.
	Write(p []byte) (n int, err error)

	// The source of the message. This can be the Client or one of the ChannelHandlers
	// in the pipeline, which represents the Server.
	Source() interface{}

	// Destination returns the reciever of this Message. This can be the Client or
	// one of the ChannelHandlers in the Pipeline, which represents the Server.
	Destination() interface{}

	// The readable string representation of this Message.
	ToString() string
}

// A simple implementation of Message interface.
type message struct {
	buffer      []byte
	source      interface{}
	destination interface{}
}

// Creates a default implementation of Message interface.
func NewMessage(source interface{}, destination interface{}, b []byte) Message {
	m := new(message)
	m.buffer = b
	m.source = source
	m.destination = destination

	return m
}

// Creates a new message with a Client source.
func NewClientMessage(cl *Client, b []byte) Message {
	m := new(message)
	m.source = cl
	m.buffer = b

	return m
}

func (m *message) String() (s string) {
	s = ""
	if m.buffer != nil && len(m.buffer) > 0 {
		s = string(m.buffer)
	}

	return
}

func (m *message) Bytes() []byte {
	return m.buffer
}

func (m *message) Read(p []byte) (n int, err error) {
	return copy(p, m.buffer), nil
}

func (m *message) Write(p []byte) (n int, err error) {
	return copy(m.buffer, p), nil
}

func (m *message) Source() interface{} {
	return m.source
}

func (m *message) Destination() interface{} {
	return m.destination
}

func (m *message) ToString() string {
	return fmt.Sprintf("Message@%v: source: %v, destination: %v", &m, m.source, m.destination)
}

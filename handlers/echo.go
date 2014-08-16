// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	"fmt"
	"gomudnet"
)

// EchoChannelHandler is an implementation of ChannelHandler which echoes the
// message back to the sender.
type EchoChannelHandler struct {
	pipeline *gomudnet.Pipeline
}

// NewEchoChannelHandler creates a new EchoChannelHandler.
func NewEchoChannelHandler() gomudnet.ChannelHandler {
	dch := new(EchoChannelHandler)

	return dch
}

func (dch *EchoChannelHandler) Receive(msg gomudnet.Message, from gomudnet.StreamDirection) (newMsg gomudnet.Message) {
	str := msg.String() + "\n"
	newMsg = gomudnet.NewMessage(dch, msg.Source(), []byte(str))
	return
}

func (dch *EchoChannelHandler) Sent(message gomudnet.Message, cl *gomudnet.Client) error {
	return nil
}

func (dch *EchoChannelHandler) Pipeline() *gomudnet.Pipeline {
	return dch.pipeline
}

func (dch *EchoChannelHandler) SetPipeline(pipeline *gomudnet.Pipeline) {
	dch.pipeline = pipeline
}

func (dch *EchoChannelHandler) Opened(cl *gomudnet.Client) error {
	strBuf := new(bytes.Buffer)
	strBuf.WriteString("Welcome!")
	dch.pipeline.SendDownstream(dch, gomudnet.NewMessage(dch, cl, strBuf.Bytes()))
	return nil
}

func (dch *EchoChannelHandler) Closing(cl *gomudnet.Client) error {
	return nil
}

func (dch *EchoChannelHandler) Closed(cl *gomudnet.Client) error {
	return nil
}

func (dch *EchoChannelHandler) String() string {
	return fmt.Sprintf("EchoChannelHandler")
}

func (dch *EchoChannelHandler) MessageDelimitter() byte {
	return '\n'
}

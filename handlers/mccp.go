// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"gomudnet"
	"gomudnet/handlers/zlib"
	"time"
)

const (
	telnet_code_IAC   byte = 255
	telnet_code_SB    byte = 250
	telnet_code_SE    byte = 240
	telnet_code_WILL  byte = 251
	telnet_code_WONT  byte = 252
	telnet_code_DO    byte = 253
	telnet_code_DONT  byte = 254
	telenet_code_MCCP byte = 86
)

var (
	mccpStartMessage gomudnet.Message
)

// MCCPHandler is an implementation of ChannelHandler which provides negotation
// and support for MCCP (MUD Compression Control Protocol).
//
// It utilises the ZlibHandler for the atual compression of the Message content.
type MCCPHandler struct {
	pipeline *gomudnet.Pipeline
}

func NewMCCPHandler() gomudnet.ChannelHandler {
	mccp := new(MCCPHandler)
	return mccp
}

func (mccp *MCCPHandler) Receive(msg gomudnet.Message, from gomudnet.StreamDirection) (newMsg gomudnet.Message) {
	b := msg.Bytes()
	newMsg = msg
	if b[0] == telnet_code_IAC && b[1] == telnet_code_DO && b[2] == telenet_code_MCCP {

		mccpStartMessage = gomudnet.NewMessage(mccp, msg.Source(), []byte{telnet_code_IAC, telnet_code_SB, telenet_code_MCCP, telnet_code_IAC, telnet_code_SE})
		mccp.pipeline.SendDownstream(mccp, mccpStartMessage)

		// I still don't know why this is needed, but without this the client
		// would suddenly close after sending the MCCP start code.
		time.Sleep(time.Second / 10) // give some time for the client to setup
		newMsg = nil
	}

	return newMsg
}

func (mccp *MCCPHandler) Sent(message gomudnet.Message, cl *gomudnet.Client) error {
	if mccpStartMessage != nil && mccpStartMessage == message {
		mccp.pipeline.AddLast("zlib", zlib.NewZlibChannelHandler())
		mccp.Pipeline().Unlock(mccp)
	}
	return nil
}

func (mccp *MCCPHandler) Pipeline() *gomudnet.Pipeline {
	return mccp.pipeline
}

func (mccp *MCCPHandler) SetPipeline(pipeline *gomudnet.Pipeline) {
	mccp.pipeline = pipeline
}

func (mccp *MCCPHandler) Opened(cl *gomudnet.Client) error {
	msg := gomudnet.NewMessage(mccp, cl, []byte{telnet_code_IAC, telnet_code_WILL, telenet_code_MCCP})
	mccp.pipeline.SendDownstream(mccp, msg)
	//lock the pipeline
	mccp.pipeline.Lock(mccp)
	return nil
}

func (mccp *MCCPHandler) Closing(cl *gomudnet.Client) error {
	return nil
}

func (mccp *MCCPHandler) Closed(cl *gomudnet.Client) error {
	return nil
}

func (mccp *MCCPHandler) MessageDelimitter() byte {
	return 0
}

func (m *MCCPHandler) String() string {
	return fmt.Sprintf("MCCPHandler@%v", &m)
}

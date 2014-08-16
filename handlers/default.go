// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package provides some useful implementations of ChannelHandler.
package handlers

import (
	"fmt"
	"gomudnet"
)

// This serves as a sample implementation of ChannnelHandler
type DefaultChannelHandler struct {
	pipeline *gomudnet.Pipeline
	name     string
}

// Returns a new DefaultChannelHandler
func NewDefaultChannelHandler(name string) gomudnet.ChannelHandler {
	dch := new(DefaultChannelHandler)
	dch.name = name

	return dch
}

func (dch *DefaultChannelHandler) Receive(msg gomudnet.Message, from gomudnet.StreamDirection) (newMsg gomudnet.Message) {
	return nil
}

func (dch *DefaultChannelHandler) Sent(message gomudnet.Message, cl *gomudnet.Client) error {
	return nil
}

func (dch *DefaultChannelHandler) Pipeline() *gomudnet.Pipeline {
	return dch.pipeline
}

func (dch *DefaultChannelHandler) SetPipeline(pipeline *gomudnet.Pipeline) {
	dch.pipeline = pipeline
}

func (dch *DefaultChannelHandler) Opened(cl *gomudnet.Client) error {
	return nil
}

func (dch *DefaultChannelHandler) Closing(cl *gomudnet.Client) error {
	return nil
}

func (dch *DefaultChannelHandler) Closed(cl *gomudnet.Client) error {
	return nil
}

func (dch *DefaultChannelHandler) String() string {
	return fmt.Sprintf("DefaultChannelHandler=%s", dch.name)
}

func (dch *DefaultChannelHandler) MessageDelimitter() byte {
	return '\n'
}

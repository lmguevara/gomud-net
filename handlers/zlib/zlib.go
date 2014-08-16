// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package provides an implementaion of ChannelHandler which compresses and
// decompress the raw message using the compress/zlib.
package zlib

import (
	"compress/zlib"
	"gomudnet"
)

type ZlibChannel struct {
	pipe *gomudnet.Pipeline
}

func NewZlibChannelHandler() *ZlibChannel {
	zChan := new(ZlibChannel)

	return zChan
}

func (this *ZlibChannel) Receive(msg gomudnet.Message, from gomudnet.StreamDirection) gomudnet.Message {
	if from == gomudnet.DIR_DOWNSTREAM {
		compressor := zlib.NewWriter(msg)
		defer compressor.Close()
		if _, err := compressor.Write(msg.Bytes()); err == nil {
			compressor.Flush()
		}
	} else if from == gomudnet.DIR_UPSTREAM {
		if dc, err := zlib.NewReader(msg); err == nil {
			defer dc.Close()
			dc.Read(msg.Bytes())
		}
	}

	return msg
}

func (this *ZlibChannel) Pipeline() *gomudnet.Pipeline {
	return this.pipe
}

func (this *ZlibChannel) SetPipeline(pipeline *gomudnet.Pipeline) {
	this.pipe = pipeline
}

func (this *ZlibChannel) Opened(cl *gomudnet.Client) error {
	return nil
}

func (this *ZlibChannel) Closed(cl *gomudnet.Client) error {
	return nil
}

func (this *ZlibChannel) Closing(cl *gomudnet.Client) error {
	return nil
}

func (this *ZlibChannel) MessageDelimitter() byte {
	return 0
}

func (this *ZlibChannel) Sent(msg gomudnet.Message, cl *gomudnet.Client) error {
	return nil
}

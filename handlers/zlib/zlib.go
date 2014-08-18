// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package provides an implementaion of ChannelHandler which compresses and
// decompress the raw message using the compress/zlib.
package zlib

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"gomudnet"
)

type ZlibChannel struct {
	pipe   *gomudnet.Pipeline
	out    *zlib.Writer
	buffer *bytes.Buffer
}

func NewZlibChannelHandler() *ZlibChannel {
	zChan := new(ZlibChannel)
	zChan.buffer = new(bytes.Buffer)
	zChan.out, _ = zlib.NewWriterLevel(zChan.buffer, zlib.BestCompression)

	return zChan
}

func (this *ZlibChannel) Receive(msg gomudnet.Message, from gomudnet.StreamDirection) gomudnet.Message {
	if from == gomudnet.DIR_DOWNSTREAM {
		if _, err := this.out.Write(msg.Bytes()); err == nil {
			this.out.Flush()
			this.buffer.WriteTo(msg)
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

func (this *ZlibChannel) String() string {
	return fmt.Sprintf("ZLIBChannel@%v", &this)
}

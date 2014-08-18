// Copyright 2014 Lorenz Millan Guevara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gomudnet

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"log"
)

// StreamDirection indicates the flow direction of a Message through the Pipeline.
type StreamDirection uint

var (
	ErrNullOrEmptyName          = errors.New("Channel name is empty or nil")
	ErrNonExistingChannel       = errors.New("Channel does not exit")
	ErrNonExistingMarkedChannel = errors.New("Marked named channel does not exist")
)

const (
	// Direction from client to server.
	DIR_UPSTREAM StreamDirection = 0

	// Direction from server to client.
	DIR_DOWNSTREAM StreamDirection = 1
)

// ChannelHandler defines the methods used in handling the events in a Pipeline.
type ChannelHandler interface {
	// Receive handles the event when a Message is received or arrived at a certain
	// point in the Pipeline with the given StreamDirection and returns a Message.
	//
	// If the returned Message is the same Message that was received, it will
	// continue through.
	//
	// If the returned Message is a different or new Message it will be sent back
	// to the sender which can be another ChannelHandler in the Pipeline or the Client.
	//
	// If the returned Message is nil, it means it was "swallowed" or arrived at
	// dead-end point in the Pipeline.
	Receive(msg Message, from StreamDirection) Message

	// Sent handles the event when the raw data of Message was sent to the client.
	//
	// This can be used to check if a Message from a certain source was sent
	// successfully to the client.
	Sent(msg Message, cl *Client) error

	// Opened handles the event when a new connection is received.
	Opened(cl *Client) error

	// Closing handles the event when a connection is closing.
	Closing(cl *Client) error

	// Closed handles the event when a connections has actually closed.
	Closed(cl *Client) error

	// Pipeline returns the Pipeline this ChannelHandler belongs to.
	Pipeline() *Pipeline

	// SetPipeline sets the Pipeline this ChannelHandler will belong to.
	SetPipeline(pipeline *Pipeline)

	// MessageDelimitter returns the delimitter of the Message which this
	// ChannelHandler will receive.
	//
	// This will tell the Pipeline to "truncate" the Message content at the first
	// occurence of the delimitter before passing it to this ChannelHandler.
	MessageDelimitter() byte
}

// PipelineFactory serves as a placeholder for the initial ChannelHandlers that a new
// Pipeline will have.
type PipelineFactory struct {
	channelHandlers *list.List
}

// Creates a new PipelineFactory.
func NewPipelineFactory() *PipelineFactory {
	pf := new(PipelineFactory)
	pf.channelHandlers = list.New()

	return pf
}

// Returns the list holding the ChannelHandlers.
func (p *PipelineFactory) ChannelHandlers() *list.List {
	return p.channelHandlers
}

// Adds the named ChannelHandler to the top of the list.
func (p *PipelineFactory) AddFirst(name string, ch ChannelHandler) {
	entry := map[string]ChannelHandler{name: ch}
	p.channelHandlers.PushFront(entry)
}

// Adds the named ChannelHandler to the bottom of the list.
func (p *PipelineFactory) AddLast(name string, ch ChannelHandler) {
	entry := map[string]ChannelHandler{name: ch}
	p.channelHandlers.PushBack(entry)
}

// Adds the name ChannelHander in the list, positioning it before the ChannelHandler
// named 'mark'.
func (p *PipelineFactory) AddBefore(mark string, name string, ch ChannelHandler) (err error) {
	entry := map[string]ChannelHandler{name: ch}
	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for k, _ := range chMap {
				if k == name {
					p.channelHandlers.InsertBefore(entry, e)
					return
				}
			}
		}
	}

	return ErrNonExistingMarkedChannel
}

// Adds the name ChannelHander in the list, positioning it after the ChannelHandler
// named 'mark'.
func (p *PipelineFactory) AddAfter(mark string, name string, ch ChannelHandler) (err error) {
	entry := map[string]ChannelHandler{name: ch}
	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for k, _ := range chMap {
				if k == name {
					p.channelHandlers.InsertAfter(entry, e)
					return
				}
			}
		}
	}

	return ErrNonExistingMarkedChannel

}

// Pipeline is where Messages flow from the Client through all the ChannelHandlers
// until one othe them sends a new or the same Message back.
//
// Pipeline takes on the perpective of the Client, meaning a Message with a DIR_UPSTREAM
// direction is considered flowing from the Client to the ChannelHandlers, and a Message
// wit a DIR_DOWNSTREAM direction is considered flowing from the ChannelHandlers to the
// Client.
//
// This represents the idea of "uploading" to the ChannelHandlers and "downloading" to
// the Client.
//
// The sequence of the ChannelHandlers in the list is very important to the flow of the
// Message. When a Client sends a Message, it will be be received by the first entry in
// the list and starts flowing upstream (First to Last). When a message flowing
// downstream (Last to First) reaches the top of the list, it's raw data will then be
// sent back to the Client.
type Pipeline struct {
	channelHandlers *list.List
	owner           *Client
	lockedBy        ChannelHandler
}

// Creates a new Pipeline.
func NewPipeline() *Pipeline {
	p := new(Pipeline)
	p.channelHandlers = list.New()

	return p
}

// Lock is used to direct the Message flow exclusively to the locking ChannelHandler.
//
// A Pipeline can only be locked if there is no existing lock.
func (p *Pipeline) Lock(ch ChannelHandler) {
	if p.lockedBy == nil {
		p.lockedBy = ch
	}
}

// Unlock is used to remove the exclusive Message flow to a the locking ChannelHandler.
//
// A Pipeline can only be unlocked by the locking ChannelHandler.
func (p *Pipeline) Unlock(ch ChannelHandler) {
	if p.lockedBy == ch {
		p.lockedBy = nil
	}
}

// Returns the ChannelHandler list.
func (p *Pipeline) ChannelHandlers() *list.List {
	return p.channelHandlers
}

// Adds the named ChannelHandler to the top of the list.
func (p *Pipeline) AddFirst(name string, ch ChannelHandler) {
	entry := map[string]ChannelHandler{name: ch}
	p.channelHandlers.PushFront(entry)
	ch.SetPipeline(p)
}

// Adds the named ChannelHandler to the bottom of the list.
func (p *Pipeline) AddLast(name string, ch ChannelHandler) {
	entry := map[string]ChannelHandler{name: ch}
	p.channelHandlers.PushBack(entry)
	ch.SetPipeline(p)
}

// Adds the name ChannelHander in the list, positioning it before the ChannelHandler
// named 'mark'.
func (p *Pipeline) AddBefore(mark string, name string, ch ChannelHandler) (err error) {
	entry := map[string]ChannelHandler{name: ch}
	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for k, _ := range chMap {
				if k == name {
					p.channelHandlers.InsertBefore(entry, e)
					ch.SetPipeline(p)
					return
				}
			}
		}
	}

	return ErrNonExistingMarkedChannel
}

// Adds the name ChannelHander in the list, positioning it after the ChannelHandler
// named 'mark'.
func (p *Pipeline) AddAfter(mark string, name string, ch ChannelHandler) (err error) {
	entry := map[string]ChannelHandler{name: ch}
	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for k, _ := range chMap {
				if k == name {
					p.channelHandlers.InsertAfter(entry, e)
					ch.SetPipeline(p)
					return
				}
			}
		}
	}

	return ErrNonExistingMarkedChannel

}

// SendUpstream sends the Message upstream with the ChannelHandler as the sender.
func (p *Pipeline) SendUpstream(sender ChannelHandler, msg Message) error {
	return p.deliverMessage(sender, msg, DIR_UPSTREAM)
}

// SendDownstream sends the Message downstream with the ChannelHandler as the sender.
func (p *Pipeline) SendDownstream(sender ChannelHandler, msg Message) error {
	return p.deliverMessage(sender, msg, DIR_DOWNSTREAM)
}

// The main Message flow routine.
func (p *Pipeline) deliverMessage(sender interface{}, msg Message, direction StreamDirection) (err error) {
	var receiver ChannelHandler

	if DEBUG {
		log.Printf("sender: %s, msg: %s, dir: %d", sender, msg.ToString(), direction)
		log.Printf("Pipe: %s", p)
	}
	//sender is a ChannelHandler, find the receiver based on the direction
	if source, ok := sender.(ChannelHandler); ok {
		for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
			if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
				for _, v := range chMap {
					if v == source {
						if direction == DIR_UPSTREAM {
							if e.Next() != nil {
								if r, ok := e.Next().Value.(map[string]ChannelHandler); ok {
									for _, rChan := range r {
										receiver = rChan
									}
								}
							}
						} else {
							if e.Prev() != nil {
								if r, ok := e.Prev().Value.(map[string]ChannelHandler); ok {
									for _, rChan := range r {
										receiver = rChan
									}
								}
							}

						}
						break
					}
				}
			}
		}
	} else if _, ok = msg.Source().(*Client); ok && msg.Destination() == nil {
		// Messages from Clients will always have no destination because, the only
		// meaningful data for a Client is the byte slice of the Message, thus they
		// will not know it's source and to whom they will send back.
		//
		// It is only logical that all messages from Clients are going upstream
		// and will always be received by the first ChannelHandler
		if direction == DIR_UPSTREAM {
			if r, ok := p.channelHandlers.Front().Value.(map[string]ChannelHandler); ok {
				for _, rChan := range r {
					receiver = rChan
				}
			}
		}
	}

	if receiver == nil {
		// When a receiver is nil, it means we have traversed through the pipeline
		// and all the ChannelHandlers relayed the message.
		//
		// Thus, is a Message traveling downstream was related by all ChannelHandlers
		// it could only mean send the byte slice to the Client.
		if direction == DIR_DOWNSTREAM {
			if cl, ok := msg.Destination().(*Client); ok {
				if chSender, ok := sender.(ChannelHandler); ok {
					if DEBUG {
						log.Printf("%s SENDING  TO CLIENT: (%d)%v", chSender, len(msg.Bytes()), msg.Bytes())
					}
					cl.sendMessage(msg, chSender.MessageDelimitter())
				}
			} else if _, ok = msg.Destination().(ChannelHandler); ok {
				p.deliverMessage(sender, msg, direction)
			}
		}
	} else {
		if p.lockedBy == nil || p.lockedBy == receiver {
			//p.delimitMessage(msg, receiver.MessageDelimitter())
			nMsg := receiver.Receive(msg, direction)

			if nMsg != nil {
				if nMsg == msg {
					// it's the same message, send it to the next ChannelHandler
					p.deliverMessage(receiver, nMsg, direction)
				} else {
					// return to sender
					newDir := DIR_UPSTREAM
					if direction == DIR_UPSTREAM {
						newDir = DIR_DOWNSTREAM
					} else {
						newDir = DIR_UPSTREAM
					}
					p.deliverMessage(receiver, nMsg, newDir)
				}
			} else {
				log.Printf("WARN: Message returned by %s is nil", receiver)
			}
		} else {
			if DEBUG {
				log.Printf("WARN: %s cannot handle Receive, pipeline is locked by:%s", receiver, p.lockedBy)
			}
		}
	}

	return
}

// The string representation of this Pipeline.
func (p *Pipeline) String() string {
	strBuf := new(bytes.Buffer)
	strBuf.WriteString(fmt.Sprintf("owner: %s, channel_handers: ", p.owner.conn.RemoteAddr()))

	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if cMap, ok := e.Value.(map[string]ChannelHandler); ok {
			strBuf.WriteString(fmt.Sprintf("%s", cMap))
		}
	}
	return strBuf.String()
}

// Informs all ChannelHandlers that a msg was sent to cl.
func (p *Pipeline) messageSent(msg Message, cl *Client) {

	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for _, ch := range chMap {
				if p.lockedBy == nil || p.lockedBy == ch {
					ch.Sent(msg, cl)
				} else {
					if DEBUG {
						log.Printf("WARN: %s cannot handle Sent, pipeline is locked by:%s", ch, p.lockedBy)
					}
				}
			}
		}
	}

}

// Informs all ChannelHandlers that a 'cl' has connected.
func (p *Pipeline) opened(cl *Client) {

	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for _, ch := range chMap {
				if p.lockedBy == nil || p.lockedBy == ch {
					ch.Opened(cl)
				} else {
					if DEBUG {
						log.Printf("WARN: %s cannot handle Opened, pipeline is locked by:%s", ch, p.lockedBy)
					}
				}
			}
		}
	}
}

// Informs all ChannelHandlers that 'cl' is closing.
func (p *Pipeline) closing(cl *Client) {

	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for _, ch := range chMap {
				if p.lockedBy == nil || p.lockedBy == ch {
					ch.Closing(cl)
				} else {
					if DEBUG {
						log.Printf("WARN: %s cannot handle Closing, pipeline is locked by:%s", ch, p.lockedBy)
					}
				}
			}
		}
	}
}

// Informs all ChannelHandlers that 'cl' has closed.
func (p *Pipeline) closed(cl *Client) {

	for e := p.channelHandlers.Front(); e != nil; e = e.Next() {
		if chMap, ok := e.Value.(map[string]ChannelHandler); ok {
			for _, ch := range chMap {
				if p.lockedBy == nil || p.lockedBy == ch {
					ch.Closed(cl)
				} else {
					if DEBUG {
						log.Printf("WARN: %s cannot handle Closed, pipeline is locked by:%s", ch, p.lockedBy)
					}
				}
			}
		}
	}
}

// Sends a message to the pipeline.
func (p *Pipeline) send(cl *Client, b []byte, from StreamDirection) {
	msg := NewClientMessage(cl, b)
	p.deliverMessage(cl, msg, from)
}

// Delimit the msg's raw data using the given delimitter.
func (p *Pipeline) delimitMessage(msg Message, delimitter byte) {
	b := msg.Bytes()
	tmp := bytes.Split(b, []byte{delimitter})
	msg.Write(tmp[0])
}

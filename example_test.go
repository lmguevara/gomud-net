package gomudnet_test

import (
	"fmt"
	"netex"
)

func ExampleNewMessage() {
	msg := NewMessage(source, destination, []byte("Hello World"))
	fmt.Println(msg.String())

	// Output:
	// Hello World
}

func ExampleMessage() {
	msg := NewMessage([]byte("New Message!"))
	fmt.Println(msg.String())

	// Output:
	// New Message!
}

func ExampleChannelHandler_receive() {
	//create a new message
	nMsg := netex.NewMessage(handler, client, msg.Bytes())

	//send it back
	return nMsg

}

func ExampleServer() {
	//non-blocking example
	pf := netex.NewPipelineFactory()
	dh := channels.NewDefaultHandler("default")
	pf.AddFirst(dh)

	//initialize the server
	serv := netex.NewServer([]byte{4000}, []string{"localhost"}, pf)
	err := serv.Start(false)
	//this will return, since the main server loop is started in a diffrent routine.
}

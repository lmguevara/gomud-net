package gomudnet_test

import (
	"fmt"
	"gomudnet"
	"gomudnet/handlers"
)

func source() gomudnet.ChannelHandler {
	return nil
}

func destination() *gomudnet.Client {
	return nil
}

func ExampleNewMessage() {
	msg := gomudnet.NewMessage(source(), destination(), []byte("Hello World"))
	fmt.Println(msg.String())

	// Output:
	// Hello World
}

func ExampleMessage() {
	msg := gomudnet.NewMessage(source(), destination(), []byte("New Message!"))
	fmt.Println(msg.String())

	// Output:
	// New Message!
}

func ExampleServer() {
	//non-blocking example
	pf := gomudnet.NewPipelineFactory()
	dh := handlers.NewDefaultChannelHandler("default")
	pf.AddFirst("default", dh)

	//initialize the server
	serv := gomudnet.NewServer([]uint{4000}, []string{"localhost"}, pf)
	serv.Start(false)
	//this will return, since the main server loop is started in a diffrent routine.
}

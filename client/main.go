package main

import "flag"

func main() {
	var clientId, serverAddr string
	flag.StringVar(&clientId, "client_id", "", "client id")
	flag.StringVar(&serverAddr, "server_addr", "", "server address")
	flag.Parse()
	c := NewClient(clientId, serverAddr)
	c.Run()
}

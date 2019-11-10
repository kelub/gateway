package main

import (
	"kelub/gateway/network"
)

func main() {
	server := &network.WSServer{
		Addr: "127.0.0.1:8585",
	}

	server.Start(server.Addr)
}

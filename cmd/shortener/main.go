package main

import (
	"shortener/internal/server"
)

func main() {
	var c config
	c.Set()
	serv := server.InitServer(c.Host, c.BaseURL)
	serv.RunServer()
}

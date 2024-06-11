package main

import (
	"shortener/internal/server"
)

func main() {
	var c config
	c.Set()
	if c.BaseURL != "/" {
		c.BaseURL = "/" + c.BaseURL + "/"
	}
	serv := server.InitServer(c.Host, c.BaseURL)
	serv.RunServer()
}

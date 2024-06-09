package main

import (
	Server "github.com/Madcow1313/shortener/internal/server"
)

func main() {
	serv := Server.InitServer("localhost:8080", "")
	serv.RunServer()
}

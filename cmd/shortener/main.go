package main

import "shortener/internal/server"

func main() {
	serv := server.InitServer("localhost:8080", "")
	serv.RunServer()
}

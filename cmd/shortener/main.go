package main

import (
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/mylogger"
	"shortener/internal/server"
)

func main() {
	var c config.Config
	c.SetConfigParameteres()

	file, err := os.OpenFile(c.URLStorage, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		mylogger.LogError(err)
		return
	}
	defer file.Close()

	serv := server.NewServer(c, file)
	serv.RunServer()
}

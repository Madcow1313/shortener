package main

import (
	"os"
	"shortener/internal/mylogger"
	"shortener/internal/server"
)

func main() {
	var c Config
	c.SetConfigParameteres()

	file, err := os.OpenFile(c.URLStorage, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		mylogger.LogError(err)
		return
	}
	defer file.Close()
	
	serv := server.NewServer(c.Host, c.BaseURL, file)
	serv.RunServer()
}

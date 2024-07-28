package main

import (
	"log"
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/server"
)

func main() {
	var c config.Config
	c.SetConfigParameteres()
	var file *os.File
	var err error
	if c.StorageType == config.File {
		file, err = os.OpenFile(c.URLStorage, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}

	serv := server.NewSimpleServer(c, file)

	switch c.StorageType {
	case config.Database:
		err = serv.CheckDBStorage()
	case config.File:
		err = serv.CheckFileStorage()
	default:
	}
	if err != nil {
		log.Fatal(err)
	}
	serv.RunServer()
}

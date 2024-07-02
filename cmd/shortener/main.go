package main

import (
	"fmt"
	"os"
	"shortener/internal/server"
)

func main() {
	var c config
	c.Set()
	file, err := os.OpenFile(c.URLStorage, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		fmt.Println(fmt.Errorf("can't open url storage file: %w", err))
		return
	}
	defer file.Close()
	serv := server.InitServer(c.Host, c.BaseURL, file)
	serv.RunServer()
}

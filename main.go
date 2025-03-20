package main

import (
	"log"
	"moshcast/mosh"
	"moshcast/utils"
)

func main() {
	server, err := mosh.NewServer()

	if err != nil {
		log.Println(err.Error())
		return
	}
	defer server.Close()

	utils.LoadPlugins(server.Options.Paths.Plugins)
	server.Start()
}

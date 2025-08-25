package main

import (
	"github.com/htranq/vortech-ome/internal/config"
	"github.com/htranq/vortech-ome/internal/server"
)

func main() {
	flags := config.ParseFlags()
	server.Run(flags)
}

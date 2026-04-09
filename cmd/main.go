package main

import (
	"authcenterapi/internal/app"
	"authcenterapi/util"
	"log"
)

func main() {
	cfg, err := util.LoadConfig("./")
	if err != nil {
		log.Fatal(err)
	}
	app.Run(cfg)
}

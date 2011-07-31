package main

import (
	"log"
	"util"
)

type config struct {
	ContentsFolder, WebFolder string
	AdminUsers []string
	SMTPServer string
	HTTPServeAddr string
}

var Conf config

func loadConfig() {
	log.Printf("Loading config file...")
	util.LoadJSONFile(ConfigFile, &Conf)
}

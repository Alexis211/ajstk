package config

import (
	"log"
	"util"
)

type config struct {
	ContentsFolder, UserDataFolder, WebFolder string
	AdminUsers map[string]struct{
		Email, Password string
	}
	SMTPServer string
	HTTPServeAddr string
}

var Conf config

func LoadConfig(file string) {
	log.Printf("Loading config file...")
	util.LoadJSONFile(file, &Conf)
}

package main

import (
	"main/config"
	"contents"
	"dic"
	"study"
	"webs"
)

const ConfigFile = "./config.json"

func main() {
	config.LoadConfig(ConfigFile)

	contents.LoadDataFolder()
	go dic.LoadDictionaries()	//do it in background, it's ok
	study.Startup()

	webs.LoadWebFilesDir()
	webs.Serve()
}

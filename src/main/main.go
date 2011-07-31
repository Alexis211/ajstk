package main

import (
	"contents"
	//"study"
	"webs"
)

const ConfigFile = "./config.json"

func main() {
	loadConfig()

	contents.LoadDataFolder(Conf.ContentsFolder)
	//study.CheckAdminUsers()

	webs.LoadWebFilesDir(Conf.WebFolder)
	webs.Serve(Conf.HTTPServeAddr)
}

package main

import (
	"discocheckbot/api"
	"discocheckbot/config"
	"log"
	"os" //todo log in file, not stdout
)

func main() {
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	log.Println("starting bot...")

	config, err := config.NewConfigReader("./config.json") //for debug, by default "config.json"
	if err != nil {
		log.Fatalln(err)
	}
	dcbot, err := NewDiscoCheckBot(config)
	if err != nil {
		log.Fatalln(err)
	}
	bot := api.NewBot(config, log, dcbot)
	if bot != nil {
		bot.ListenForUpdates()
	}

	log.Fatalln("bot terminated")
}

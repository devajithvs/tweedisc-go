package main

import (
	"tweedisc-go/bot"
	"tweedisc-go/config"
)

var log = config.Log

func main() {

	bot.Start()

	<-make(chan struct{})
	return
}

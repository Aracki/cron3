package main

import (
	"github.com/robfig/cron"
	"log"
)

func main() {

	c := cron.New()
	c.Start()
	if err := c.AddFunc("*/3 * * * *", func() { log.Println("Run every 3 second") }); err != nil {
		log.Fatal(err.Error())
	}

	select {}
}

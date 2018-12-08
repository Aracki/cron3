package main

import (
	"fmt"
	"github.com/robfig/cron"
	"log"
	"os/exec"
)

func main() {

	c := cron.New()
	c.Start()
	if err := c.AddFunc("*/3 * * * *", func() {
		if err := mongoDump(); err != nil {
			log.Fatal(err.Error())
		}
	}); err != nil {
		log.Fatal("Cannot parse cron spec:" ,err.Error())
	}

	select {}
}

func mongoDump() error {

	cmd := exec.Command("mongodump")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%s", fmt.Sprint(err),string(out))
	}

	fmt.Println("mongodump executed")
	return nil
}

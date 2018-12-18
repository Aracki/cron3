package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

func mongoDump() error {

	cmd := exec.Command("mongodump")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%s", fmt.Sprint(err), string(out))
	}

	log.Println("mongodump executed")
	return nil
}

func cronFunc() error {
	if err := mongoDump(); err != nil {
		log.Fatal(err.Error())
	}

	log.Println("uploading to s3...")
	svc := initS3()
	if key, err := uploadToS3(*svc); err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("uploaded %s", key)
	}

	return nil
}

func main() {

	// reading configurations from config.yml
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("fatal error config file: ", err)
	}

	c := cron.New()
	c.Start()
	if err := c.AddFunc("*/3 * * * *", func() {
		log.Println(cronFunc())
	}); err != nil {
		log.Fatal("cannot parse cron spec:", err.Error())
	}

	select {}
}

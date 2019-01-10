package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

var svc *s3.S3

// mongoDump() will create .bson file used for backups
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

	log.Println("executing mongodump command...")
	if err := mongoDump(); err != nil {
		log.Fatal(err.Error())
	}

	// generate key for file based on current date
	// eg. 2019/January/3
	year, month, day := time.Now().Date()
	key := strconv.Itoa(year) + "/" + month.String() + "/" + strconv.Itoa(day)

	log.Println("uploading new backup to s3...")
	if err := uploadToS3(svc, key); err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("uploaded %s", key)
	}

	log.Println("deleting old backups from s3...")
	if key, err := deleteFromS3(svc, key); err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("successfully deleted object with key=%s \n", key)
	}

	return nil
}

func main() {

	// init s3 service
	svc = initS3()

	// reading configurations from config.yml
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("fatal error config file: ", err)
	}

	log.Println("starting new cron job...")
	c := cron.New()
	c.Start()
	if err := c.AddFunc(viper.GetString("cron_time"), func() {
		log.Println(cronFunc())
	}); err != nil {
		log.Fatal("cannot parse cron spec:", err.Error())
	}

	select {}
}

package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

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

func cronFunc(now time.Time) {

	log.Println(".............")
	log.Println("starting cron")
	if err := mongoDump(); err != nil {
		log.Fatal(err.Error())
	}

	// generate key for file based on current date
	// eg. 2019/January/3
	year, month, day := now.Date()
	key := strconv.Itoa(year) + "/" + month.String() + "/" + strconv.Itoa(day) + ".bson"

	// S3 methods are safe to use concurrently. It is not safe to
	// modify mutate any of the struct's properties though.
	svc := initS3()

	log.Println("uploading new backup to s3")
	if err := uploadToS3(svc, key); err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("uploaded %s", key)
	}

	if err := deleteFromS3(svc, key); err != nil {
		log.Println(err.Error())
	} else {
		log.Printf("successfully deleted old backups")
	}
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

	log.Println("starting new cron job...")
	c := cron.New()
	c.Start()
	if err := c.AddFunc(viper.GetString("cron_time"), func() {
		cronFunc(time.Now())
	}); err != nil {
		log.Fatal("cannot parse cron spec:", err.Error())
	}

	select {}
}

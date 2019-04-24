package main

import (
	"github.com/spf13/viper"
	"log"
	"testing"
	"time"
)

// TestJanuaryBackups fakes every backup made for January in 3000 year.
func TestJanuaryBackups(t *testing.T) {

	// reading configurations from config.yml
	viper.SetConfigType("yaml")
	viper.SetConfigName("config_test")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("fatal error config file: ", err)
	}

	month := time.January
	days := 31

	for i := 1; i <= days; i++ {
		cronFunc(time.Date(3000, month, i, 0, 0, 0, 0, time.UTC))
		t.Logf("uploaded January %d", i)
	}
}

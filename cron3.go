package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/robfig/cron"
	"github.com/spf13/viper"

	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

const fileName = "cron3.go"

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

func mongoDump() error {

	cmd := exec.Command("mongodump")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%s", fmt.Sprint(err), string(out))
	}

	log.Println("mongodump executed")
	return nil
}

// Uploads mongo dump file to s3 bucket
func uploadToS3(svc s3.S3) (key string, err error) {

	ctx := context.Background()
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	bucketName := viper.GetString("bucket_name")
	year, month, day := time.Now().Date()
	dstKey := strconv.Itoa(year) + "/" + month.String() + "/" + strconv.Itoa(day)

	// Uploads the object to S3. The Context will interrupt the request if the
	// timeout expires.
	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(dstKey),
		Body:   f,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			// If the SDK can determine the request or retry delay was canceled
			// by a context the CanceledErrorCode error code will be returned.
			return "", errors.Wrap(err, "upload canceled due to timeout")
		}
		return "", err
	}

	log.Printf("%s uploaded to %s\n", dstKey, bucketName)
	return dstKey, nil
}

func initS3() *s3.S3 {

	creds := credentials.NewStaticCredentials(
		viper.GetString("access_key"),
		viper.GetString("secret_key"),
		"",
	)

	sess := session.Must(session.NewSession())
	return s3.New(sess, &aws.Config{
		Credentials: creds,
		Region:      aws.String(endpoints.UsEast1RegionID),
	})
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

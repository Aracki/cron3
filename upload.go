package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Uploads mongo dump file to s3 bucket
func uploadToS3(svc s3.S3) (key string, err error) {

	ctx := context.Background()
	f, err := os.Open(viper.GetString("file_name"))
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

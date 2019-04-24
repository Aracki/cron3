package main

import (
	"context"
	"log"
	"os"
	"path"

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

// uploadToS3 will upload mongo dump file to s3 bucket with destination key
func uploadToS3(svc *s3.S3, dstKey string) (err error) {

	ctx := context.Background()
	f, err := os.Open(viper.GetString("file_name"))
	if err != nil {
		return err
	}
	defer f.Close()

	bucketName := viper.GetString("bucket_name")

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
			return errors.Wrap(err, "upload canceled due to timeout")
		}
		return err
	}

	return nil
}

// deleteFromS3 will delete other backups for the same month
func deleteFromS3(svc *s3.S3, key string) (err error) {

	ctx := context.Background()
	bucketName := viper.GetString("bucket_name")

	// return whole put but last element
	prefix := path.Dir(key)

	log.Println("listing objects with prefix:", prefix)
	objects, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return err
	}

	if len(objects.Contents) == 0 {
		return errors.New("bucket is empty")
	}
	if len(objects.Contents) ==1 {
		return errors.New("there is only one backup in this month")
	}

	for i, o := range objects.Contents {
		if *o.Key != key {
			lastKey := objects.Contents[i].Key

			_, err := svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    lastKey,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
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

	log.Printf("%s uploaded to %s\n", dstKey, bucketName)
	return nil
}

// deleteFromS3 will delete backup if there are more than 3 backups for that specific month
func deleteFromS3(svc *s3.S3, key string) (deletedKey string, err error) {

	ctx := context.Background()
	bucketName := viper.GetString("bucket_name")

	// return whole put but last element
	prefix := path.Dir(key)

	log.Println("listing objects with prefix: ", prefix)
	objects, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return "", err
	}

	if len(objects.Contents) == 0 {
		return "", errors.New("bucket already empty")
	}

	if len(objects.Contents) > 3 {
		firstKey := objects.Contents[0].Key

		dltObj, err := svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    firstKey,
		})
		if err != nil {
			return "", err
		}

		return dltObj.String(), err
	} else {
		return "", errors.New(
			fmt.Sprintf("bucket with prefix=%s have no more than 3 objects", prefix))
	}
}

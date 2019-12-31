package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"io"
	"log"
	"time"
)

type uploader struct {
	glacier   *glacier.Glacier
	vaultName string
}

func NewUploader(cfg Config) *uploader {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cfg.Glacier.Id,
			cfg.Glacier.Secret,
			"",
		),
		Region: &cfg.Glacier.Region,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	return &uploader{
		glacier:   glacier.New(sess),
		vaultName: cfg.Glacier.VaultName,
	}
}

func (u uploader) Upload(buf io.ReadSeeker) error {
	result, err := u.glacier.UploadArchive(&glacier.UploadArchiveInput{
		VaultName: &u.vaultName,
		Body:      buf,
		ArchiveDescription: aws.String(fmt.Sprintf("Home assistant backup - %s", time.Now().Format(time.RFC3339))),
	})
	if err != nil {
		log.Println("Error uploading archive.", err)
		return err
	}
	log.Println("archive uploaded", result)


	return nil
}

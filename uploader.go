package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type uploader struct {
	glacier    *glacier.Glacier
	db         *fileDB
	vaultName  string
	archiveTTL time.Duration
}

func NewUploader(cfg Config, db *fileDB) *uploader {
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
		db:         db,
		glacier:    glacier.New(sess),
		vaultName:  cfg.Glacier.VaultName,
		archiveTTL: cfg.ArchiveTTL,
	}
}

func (u uploader) Upload(buf io.ReadSeeker) error {
	result, err := u.glacier.UploadArchive(&glacier.UploadArchiveInput{
		VaultName:          aws.String(u.vaultName),
		Body:               buf,
		ArchiveDescription: aws.String(fmt.Sprintf("Home assistant backup - %s", time.Now().Format(time.RFC3339))),
	})
	if err != nil {
		log.Println("Error uploading archive.", err)
		return err
	}

	err = u.db.WriteRecord(dbRecord{
		archiveID: aws.StringValue(result.ArchiveId),
		bucket:    u.vaultName,
		createdAt: time.Now(),
	})
	if err != nil {
		log.Println("Error saving archive data record", err)
	}

	log.Println("archive uploaded", result)

	return nil
}

func (u uploader) Cleanup() error {
	if u.archiveTTL == 0 {
		return errors.New("archive ttl is not set, skipping cleanup")
	}

	var cleanedRecords []dbRecord
	removeUntil := time.Now().Add(-u.archiveTTL)
	records, err := u.db.ReadRecords()
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.createdAt.After(removeUntil) {
			cleanedRecords = append(cleanedRecords, *record)
			continue

		}

		_, err := u.glacier.DeleteArchive(&glacier.DeleteArchiveInput{
			VaultName: &u.vaultName,
			ArchiveId: aws.String(record.archiveID),
		})
		if err != nil {
			return err
		}
	}

	if len(cleanedRecords) != len(records) {
		err = u.db.WriteRecords(cleanedRecords)
		if err != nil {
			return err
		}
	}

	return nil
}

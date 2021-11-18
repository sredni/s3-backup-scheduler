package main

import (
	"encoding/csv"
	"os"
	"sync"
	"time"
)

const (
	archivesDBFilename = "archives.csv"
)

type dbRecord struct {
	archiveID string
	bucket    string
	createdAt time.Time
}

func fromDB(data []string) *dbRecord {
	createdAt, err := time.Parse(time.RFC3339, data[2])
	if err != nil {
		return nil
	}

	return &dbRecord{
		archiveID: data[0],
		bucket:    data[1],
		createdAt: createdAt,
	}
}

func (r dbRecord) toDB() []string {
	return []string{
		r.archiveID,
		r.bucket,
		r.createdAt.Format(time.RFC3339),
	}
}

type fileDB struct {
	file *os.File
	lock sync.Mutex
}

func NewFileDB() *fileDB {
	file, _ := os.OpenFile(archivesDBFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	return &fileDB{
		file: file,
	}
}

func (db *fileDB) writeRecord(record dbRecord) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	w := csv.NewWriter(db.file)
	defer w.Flush()

	err := w.Write(record.toDB())
	if err != nil {
		return err
	}

	return nil
}

func (db *fileDB) readRecords() ([]*dbRecord, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var records []*dbRecord
	r := csv.NewReader(db.file)
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		records = append(records, fromDB(line))
	}

	return records, nil
}

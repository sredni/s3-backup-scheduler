package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
)

type archiver struct {
	filePaths []string
}

func NewArchiver(cfg Config) *archiver {
	return &archiver{
		filePaths: cfg.PathsToBackup,
	}
}

func (a archiver) Create(buf io.Writer) error {
	zw := gzip.NewWriter(buf)
	tw := tar.NewWriter(zw)

	for _, path := range a.filePaths {
		err := a.compressPath(path, tw)
		if err != nil {
			log.Println("Unable to tar", path, err)
			return err
		}
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zw.Close(); err != nil {
		return err
	}

	return nil
}

func (a archiver) compressPath(path string, tw *tar.Writer) error {
	paths, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	for _, p := range paths {
		err := filepath.Walk(p, func(file string, fi os.FileInfo, err error) error {
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			header.Name = filepath.ToSlash(file)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchiver_Create(t *testing.T) {
	expectedPaths := []string{
		"test/dir",
		"test/dir/file.txt",
		"test/dir/file2.txt",
		"test/test.txt",
		"test/wildcard/file.txt",
		"test/wildcard/file2.txt",
	}
	excludedFiles := []string{
		"test/wildcard/file.yaml",
	}
	var actualPaths []string
	var buf bytes.Buffer
	arch := NewArchiver(Config{
		PathsToBackup: []string{
			"test/dir",
			"test/test.txt",
			"test/wildcard/*.txt",
		},
	})

	err := arch.Create(&buf)
	require.NoError(t, err)

	zr, err := gzip.NewReader(&buf)
	tr := tar.NewReader(zr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		actualPaths = append(actualPaths, header.Name)
	}

	require.Equal(t, expectedPaths, actualPaths)
	for _, file := range excludedFiles {
		require.NotContains(t, actualPaths, file)
	}
}
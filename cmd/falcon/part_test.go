package main

import (
	"path/filepath"
	"testing"
)

func TestCalculateParts(t *testing.T) {
	url := "http://foo.bar/file"
	parts := calculateParts(int64(10), 100, url)
	if len(parts) != 10 {
		t.Fatalf("parts length should be 10")
	}
	if parts[0].URL != url {
		t.Fatalf("part url was wrong")
	}

	dir := filepath.Join(GetUserHome(), appHome, tempFolderName(url), "file.part0")
	if parts[0].Path != dir {
		t.Fatalf("part path was wrong")
	}
	if parts[0].RangeFrom != 0 && parts[0].RangeTo != 10 {
		t.Fatalf("part range was wrong")
	}
}

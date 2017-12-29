package main

import (
	"testing"
)

func TestGetFilenameFromURL(t *testing.T) {
	filename := filenameFromURL("https://storage.googleapis.com/golang/go1.9.2.darwin-amd64.pkg")
	if filename != "go1.9.2.darwin-amd64.pkg" {
		t.Fatalf("filename was wrong")
	}
}

func TestGetFilenameFromURLWithSpecialChars(t *testing.T) {
	filename := filenameFromURL("https://d.pcs.baidu.com/file/bab5cf0974c5d1a864e0823f86216800?fid=1560318009-250528-1048376041497695&time=15")

	if filename != "bab5cf0974c5d1a864e0823f86216800" {
		t.Fatalf("filename was wrong")
	}
}

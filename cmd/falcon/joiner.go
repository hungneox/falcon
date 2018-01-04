package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"

	"github.com/fatih/color"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// JoinFile combine a list of files into single file
func JoinFile(files []string, out string) error {
	//Sort with file name or we will join files with wrong order
	sort.Strings(files)
	fmt.Println("Start joining")
	var bar *pb.ProgressBar
	prefix := "Joining"

	if runtime.GOOS != "windows" {
		prefix = color.GreenString(prefix)
	}

	bar = pb.StartNew(len(files)).Prefix(prefix)

	outf, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer outf.Close()

	for _, f := range files {
		if err = copy(f, outf); err != nil {
			return err
		}
		bar.Increment()
	}

	bar.Finish()

	return nil
}

func copy(from string, to io.Writer) error {
	f, err := os.OpenFile(from, os.O_RDONLY, 0600)
	HandleError(err)
	defer f.Close()
	if err != nil {
		return err
	}
	io.Copy(to, f)
	return nil
}

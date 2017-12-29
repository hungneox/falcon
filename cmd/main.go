package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

func main() {
	var conn = runtime.NumCPU()

	var cmdGet = &cobra.Command{
		Use:   "get [url]",
		Short: "Download the given url",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			get(args[0], conn)
		},
	}

	var cmdResume = &cobra.Command{
		Use:   "resume [task]",
		Short: "Resume an unfinished task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: Implement resume task")
		},
	}

	var cmdTask = &cobra.Command{
		Use:   "tasks",
		Short: "Listing all unfinished tasks",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: Implement task listing")
		},
	}

	var rootCmd = &cobra.Command{Use: "falcon"}
	rootCmd.AddCommand(cmdGet, cmdResume, cmdTask)
	rootCmd.Execute()
}

func get(url string, conn int) {
	var err error
	var fileChan = make(chan string, int64(conn))
	var doneChan = make(chan bool, int64(conn))
	var errorChan = make(chan error, 1)
	var files = make([]string, 0)
	var filename = FilenameFromURL(url)

	downloader := NewHttpDownloader(url, int64(conn))

	go downloader.Start(doneChan, fileChan, errorChan)

	for {
		select {
		case file := <-fileChan:
			files = append(files, file)
		case err := <-errorChan:
			HandleError(err)
		case <-doneChan:
			err = JoinFile(files, filename)
			HandleError(err)
			err = os.RemoveAll(GetValidFolderPath(url))
			HandleError(err)
			return
		}
	}
}

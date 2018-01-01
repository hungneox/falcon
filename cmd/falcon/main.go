package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	var conn int

	var cmdGet = &cobra.Command{
		Use:   "get [url]",
		Short: "Download the given url",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			get(args[0], conn, nil)
		},
	}

	cmdGet.Flags().IntVarP(&conn, "connection", "c", runtime.NumCPU(), "The number of connections")

	//@TODO Bug when resume 3/4 parts
	var cmdResume = &cobra.Command{
		Use:   "resume [task]",
		Short: "Resume an unfinished task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resume(args[0], conn)
		},
	}

	var cmdTask = &cobra.Command{
		Use:   "tasks",
		Short: "Listing all unfinished tasks",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			listTasks()
		},
	}

	var rootCmd = &cobra.Command{Use: "falcon"}
	rootCmd.AddCommand(cmdGet, cmdResume, cmdTask)
	rootCmd.Execute()
}

func get(url string, conn int, state *State) {
	var err error
	fileChan := make(chan string, int64(conn))
	doneChan := make(chan bool, int64(conn))
	errorChan := make(chan error, 1)

	files := make([]string, 0)
	parts := make([]Part, 0)
	filename := FilenameFromURL(url)
	signalChan := make(chan os.Signal, 1)
	stateChan := make(chan Part, 1)
	interupted := false
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var downloader *HttpDownloader
	if state == nil {
		downloader = NewHttpDownloader(url, int64(conn), make([]Part, 0))
	} else {
		downloader = NewHttpDownloader(url, int64(conn), state.Parts)
	}

	go downloader.Start(doneChan, fileChan, errorChan, signalChan, stateChan)

	for {
		select {
		case file := <-fileChan:
			files = append(files, file)
		case err := <-errorChan:
			HandleError(err)
		case part := <-stateChan:
			parts = append(parts, part)
			interupted = true
		case <-doneChan:
			if interupted && downloader.resumable {
				fmt.Printf("Interrupted, saving state ... \n")
				s := &State{URL: url, Parts: parts}
				err = s.Save()
				HandleError(err)
				return
			}
			err = JoinFile(files, filename)
			HandleError(err)
			err = os.RemoveAll(GetValidFolderPath(url))
			HandleError(err)
			return
		}
	}
}

func resume(task string, conn int) {
	state, err := LoadState(task)
	HandleError(err)
	get(state.URL, conn, state)
}

func listTasks() {
	files, err := ioutil.ReadDir(filepath.Join(GetUserHome(), appHome))

	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		fmt.Println("There is no unfinished task")
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
}

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	client http.Client
)

var (
	acceptRangeHeader   = "Accept-Ranges"
	contentLengthHeader = "Content-Length"
)

/*
HttpDownloader struct
*/
type HttpDownloader struct {
	url        string
	file       string
	totalParts int64
	length     int64
	parts      []Part
	resumable  bool
}

// NewHttpDownloader constructor
func NewHttpDownloader(url string, connections int64, parts []Part) *HttpDownloader {
	downloader := new(HttpDownloader)
	header := downloader.getHeader(url)
	var resumable = true
	//print out host info
	downloader.printHostInfo(url)

	// CheckHTTPHeader Check if target url response
	// contains Accept-Ranges or Content-Length headers
	contentLength := header.Get(contentLengthHeader)
	acceptRange := header.Get(acceptRangeHeader)

	if contentLength == "" {
		fmt.Printf("Response header doesn't contain Content-Length, fallback to 1 connection\n")
		contentLength = "1" //set 1 because of progress bar not accept 0 length
		connections = 1
	}

	if acceptRange == "" {
		fmt.Printf("Response header doesn't contain Accept-Ranges, fallback to 1 connection\n")
		connections = 1
		resumable = false
	}

	fmt.Printf("Start download with %d connections \n", connections)

	length, err := strconv.ParseInt(contentLength, 10, 64)
	HandleError(err)

	downloader.url = url
	downloader.file = FilenameFromURL(url)
	downloader.totalParts = int64(connections)
	downloader.length = length
	downloader.resumable = resumable

	if len(parts) == 0 {
		downloader.parts = calculateParts(int64(connections), length, url)
	} else {
		downloader.parts = parts
	}

	return downloader
}

func (d HttpDownloader) printHostInfo(url string) {
	parsed, err := neturl.Parse(url)
	ips, err := net.LookupIP(parsed.Host)
	HandleError(err)

	ipstr := FilterIPV4(ips)
	fmt.Printf("Resolve ip: %s\n", strings.Join(ipstr, " | "))
}

func (d HttpDownloader) getHeader(url string) *http.Header {
	if IsValidURL(url) == false {
		fmt.Printf("Invalid url\n")
		os.Exit(1)
	}

	req, err := http.NewRequest("GET", url, nil)
	HandleError(err)

	resp, err := client.Do(req)
	HandleError(err)

	return &resp.Header
}

func (d HttpDownloader) initProgressbars() []*pb.ProgressBar {
	var bars []*pb.ProgressBar
	bars = make([]*pb.ProgressBar, 0)
	var prefix string
	for i, part := range d.parts {
		prefix = fmt.Sprintf("%s-%d", d.file, i)
		if runtime.GOOS != "windows" {
			prefix = color.YellowString(prefix)
		}
		newbar := pb.New64(part.RangeTo - part.RangeFrom).SetUnits(pb.U_BYTES).Prefix(prefix)
		bars = append(bars, newbar)
	}
	return bars
}

// Start downloading
func (d HttpDownloader) Start(doneChan chan bool, fileChan chan string, errorChan chan error, signalChan chan os.Signal, stateChan chan Part) {
	var ws sync.WaitGroup
	var barPool *pb.Pool
	var err error
	bars := d.initProgressbars()
	barPool, err = pb.StartPool(bars...)
	errorChan <- err
	defer barPool.Stop()

	for i, p := range d.parts {
		ws.Add(1)
		go func(i int64, part Part) {
			defer ws.Done()
			// send file path to file channel
			fileChan <- part.Path
			// get response for current part
			ranges := fmt.Sprintf("bytes=%d-%d", part.RangeFrom, part.RangeTo)
			req, err := http.NewRequest("GET", part.URL, nil)
			errorChan <- err

			req.Header.Add("Range", ranges)
			resp, err := client.Do(req)
			errorChan <- err
			defer resp.Body.Close()

			var bar *pb.ProgressBar
			bar = bars[i]
			// open part.path for writing
			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			defer f.Close()
			errorChan <- err

			var writer io.Writer
			writer = io.MultiWriter(f, bar)
			current := int64(0)
			for {
				select {
				case <-signalChan:
					stateChan <- Part{URL: d.url, Path: part.Path, RangeFrom: current + part.RangeFrom, RangeTo: part.RangeTo}
					return
				default:
					written, err := io.CopyN(writer, resp.Body, 100)
					current += written
					if err != nil {
						bar.Finish()
						return
					}
				}
			}
		}(int64(i), p)
	} //end for
	ws.Wait()
	doneChan <- true
}

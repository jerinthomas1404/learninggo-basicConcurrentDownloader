package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Download struct {
	Url           string
	TargetPath    string
	TotalSections int
}

func main() {

	startTime := time.Now()
	d := Download{
		Url:           "https://www.dropbox.com/s/n56joixupj3ivv2/SampleVideo_1280x720_30mb.mp4?dl=1",
		TargetPath:    "final.mp4",
		TotalSections: 10,
	}

	err := d.Do()
	if err != nil {
		log.Fatalf("An error occurred while downloading the file:%s\n", err)
	}
	fmt.Printf("Download completed in %v in seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d Download) Do() error {

	fmt.Println("Making connection")

	r, err := d.getNewRequest("HEAD")
	if err != nil {
		log.Fatalf("An error occured while getting the request")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatalf("An error occured while getting the request")
	}

	fmt.Printf("Got %v\n", resp.StatusCode)

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		log.Fatalf("An error occured while getting the content size")
	}
	fmt.Printf("Size is %v bytes\n", size)

	return nil

}

func (d Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(
		method,
		d.Url,
		nil,
	)
	if err != nil {
		log.Fatalf("An error occurred while creating a new request: %s \n", err)
	}

	r.Header.Set("User-Agent", "Silly Download Manager v0.01")

	return r, nil
}

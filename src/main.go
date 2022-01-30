package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
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

	var sections = make([][2]int, d.TotalSections)
	eachSize := size / d.TotalSections
	fmt.Printf("Size of each section: %v \n", eachSize)

	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < d.TotalSections-1 {
			sections[i][1] = sections[i][0] + eachSize
		} else {
			sections[i][1] = size - 1
		}

	}

	// fmt.Println(sections)

	var wg sync.WaitGroup
	for i, s := range sections {

		i, s := i, s
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = d.downloadSection(i, s)

			if err != nil {
				panic(err)
			}

		}()
	}
	wg.Wait()

	err = d.mergeFiles(sections)
	if err != nil {
		log.Fatalf("An error occurred while merging the sections:%v\n", err)
	}
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

func (d Download) downloadSection(i int, s [2]int) error {

	r, err := d.getNewRequest("GET")
	if err != nil {
		log.Fatalf("An error occurred while creating GET request for downloading %v\n", err)
	}

	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s[0], s[1]))

	resp, err := http.DefaultClient.Do(r)

	if err != nil {
		log.Fatalf("An error occurred while receiving the response %v\n", err)
	}

	fmt.Printf("Downloaded %v bystes for section %v:%v\n", resp.Header.Get("Content-Length"), i, s)

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("An error occurred while reading the response %v\n", err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("section-%v.tmp", i), b, os.ModePerm)
	if err != nil {
		log.Fatalf("An error occurred while writing the file %v\n", err)
	}
	return nil
}

func (d Download) mergeFiles(sections [][2]int) error {
	f, err := os.OpenFile(d.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		tmpFileName := fmt.Sprintf("section-%v.tmp", i)
		b, err := ioutil.ReadFile(tmpFileName)
		if err != nil {
			return err
		}
		n, err := f.Write(b)
		if err != nil {
			return err
		}
		err = os.Remove(tmpFileName)
		if err != nil {
			return err
		}
		fmt.Printf("%v bytes merged\n", n)
	}
	return nil
}

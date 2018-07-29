package main

import (
	"io"
	"net/http"
	"os"

	"fmt"
	"sync"

	"errors"

	"github.com/mariomac/manana"
)

// Example of Asynchronous application where multiple files are downloaded at the same time, and the result of the
// download is handled asynchronously

func main() {
	downloads := []struct {
		dstFile string
		url     string
	}{
		{"ubuntu18.metalink", "http://releases.ubuntu.com/18.04/ubuntu-18.04.1-desktop-amd64.metalink"},
		{"ubuntu18.torrent", "http://releases.ubuntu.com/18.04/ubuntu-18.04.1-desktop-amd64.iso.torrent"},
		{"ubuntu30.iso", "http://releases.ubuntu.com/30.04/ubuntu-30.iso"}, // We expect a 404 error here
	}

	wg := sync.WaitGroup{}
	wg.Add(len(downloads))
	for _, d := range downloads {
		file := d.dstFile
		durl := d.url

		fmt.Println("Downloading", file, "...")

		// We can handle asynchronously a synchronous function
		f := manana.Do(func() (interface{}, error) {
			err := Download(file, durl)
			return nil, err
		})

		// Registering success/fail callbacks for the Future
		f.OnSuccess(func(_ interface{}) {
			fmt.Println("File", file, "successfully downloaded")
			wg.Done()
		})
		f.OnFail(func(err error) {
			fmt.Println("Error with", file, ":", err)
			wg.Done()
		})
	}

	fmt.Println("Waiting to complete downloads...")
	wg.Wait()
}

// Download is synchronous
func Download(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	// Write the body to file in background, to allow
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

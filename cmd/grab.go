package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/dustin/go-humanize"
)

func tDownload(url string) (err error) {
	// Get the data
	fmt.Printf("downloading %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("response=%+v\n", resp.Body)

	memWriter := BandwidthReader{}

	// Writer the body to file
	wrote, err := io.Copy(&memWriter, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("wrote=%d\n", wrote)
	fmt.Printf("%+v\n", memWriter)

	return nil
}

func grab1(url string) {
	fmt.Printf("Initializing...\n")
	respch, err := grab.GetAsync(".", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, err)
		os.Exit(1)
	}

	fmt.Printf("Starting...\n")
	resp := <-respch

	start := time.Now().UTC()
	for !resp.IsComplete() {
		// fmt.Printf("\033[1AProgress %d / %d bytes (%d%%)\033[K\n", resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
		bw := float64(resp.BytesTransferred()) / time.Since(start).Seconds()
		fmt.Printf("%f - %f - %d/%d %s\n", resp.Duration().Seconds(), time.Since(start).Seconds(), resp.BytesTransferred(), resp.Size, humanize.Bytes(uint64(bw*8)))
		time.Sleep(200 * time.Millisecond)
	}

	// clear progress line
	fmt.Printf("\n")

	// check for errors
	if resp.Error != nil {
		fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", url, resp.Error)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded to ./%s\n", resp.Filename)
}

func grab2(url string) {
	// create a custom client
	client := grab.NewClient()
	client.UserAgent = "fast-cli"

	// create request for each URL given on the command line
	reqs := make([]*grab.Request, 0)
	// for _, url := range os.Args[1:] {
	for i := 0; i < 3; i++ {
		req, err := grab.NewRequest(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		reqs = append(reqs, req)
	}

	// start file downloads, 3 at a time
	fmt.Printf("Downloading %d files...\n", len(reqs))
	respch := client.DoBatch(3, reqs...)

	// Download single url
	// req, err := grab.NewRequest(url)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%v\n", err)
	// 	os.Exit(1)
	// }
	//
	// // start file downloads
	// client.Do(req)

	// start a ticker to update progress every 200ms
	t := time.NewTicker(200 * time.Millisecond)

	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// clear lines
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}

			// update completed downloads
			for i, resp := range responses {
				if resp != nil && resp.IsComplete() {
					// print final result
					if resp.Error != nil {
						fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Error)
					} else {
						fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
					}

					// mark completed
					responses[i] = nil
					completed++
				}
			}

			// update downloads in progress
			inProgress = 0
			for _, resp := range responses {
				if resp != nil {
					inProgress++
					fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
				}
			}
		}
	}

	t.Stop()

	fmt.Printf("%d files successfully downloaded.\n", len(reqs))

}

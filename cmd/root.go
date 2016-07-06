// Copyright © 2016 Gus Esquivel <gesquive@gmail.com>

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gesquive/fast-cli/fast"
	"github.com/gesquive/fast-cli/format"
	"github.com/gesquive/fast-cli/meters"
	"github.com/spf13/cobra"
)

var displayVersion string
var cfgFile string
var useHTTPS bool
var showVersion bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fast-cli",
	Short: "Estimates your current internet download speed",
	Long: `Estimates your current internet download speed using Netflix's fast.com service.

fast-cli caclulates this estimate by performing a series of downloads from Netflix's fast.com servers.`,
	Run: run,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	displayVersion = version
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	//TODO: Use https by default
	RootCmd.PersistentFlags().BoolVarP(&useHTTPS, "use-https", "s", false, "Use HTTPS when connecting")
	RootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Display the version number and exit")
	//TODO: Allow to estimate using time or size
}

func initConfig() {
}

func run(cmd *cobra.Command, args []string) {
	//TODO: Implement better logging and debug messages
	if showVersion {
		fmt.Println(displayVersion)
		os.Exit(0)
	}
	count := uint64(3)
	fmt.Printf("Estimating current download speed\n")
	urls := fast.GetDlUrls(count)
	// fmt.Printf("%+v\n", urls)

	if len(urls) == 0 {
		fmt.Printf("Using fallback endpoint\n")
		urls = append(urls, fast.GetDefaultURL())
	}

	err := calculateBandwidth(urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func calculateBandwidth(urls []string) (err error) {
	// fmt.Printf("downloading %s\n", urls)
	client := &http.Client{}
	count := uint64(len(urls))

	primaryBandwidthReader := meters.BandwidthMeter{}
	bandwidthMeter := meters.BandwidthMeter{}
	ch := make(chan *copyResults, 1)
	// var requests = make([]http.Request, count)
	bytesToRead := uint64(0)
	completed := uint64(0)

	for i := uint64(0); i < count; i++ {
		// Create the HTTP request
		request, err := http.NewRequest("GET", urls[i], nil)
		if err != nil {
			return err
		}
		request.Header.Set("User-Agent", displayVersion)

		// Get the HTTP Response
		response, err := client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		// Set information for the leading index
		if i == 0 {
			// Try to get content length
			contentLength := response.Header.Get("Content-Length")
			calculatedLength, err := strconv.Atoi(contentLength)
			if err != nil {
				calculatedLength = 26214400
			}
			bytesToRead = uint64(calculatedLength)

			tapMeter := io.TeeReader(response.Body, &primaryBandwidthReader)
			go asyncCopy(i, ch, &bandwidthMeter, tapMeter)
		} else {
			// Start reading
			go asyncCopy(i, ch, &bandwidthMeter, response.Body)
		}

	}

	// fmt.Printf("bytes=%d\n", bytesToRead)
	// fmt.Printf("totalBytes=%d\n", totalBytes)

	for {
		select {
		case results := <-ch:
			if results.err != nil {
				fmt.Fprintf(os.Stdout, "\n%v\n", results.err)
				os.Exit(1)
			}

			fmt.Printf("\r%s - %s",
				format.BitsPerSec(bandwidthMeter.Bandwidth()),
				format.Percent(primaryBandwidthReader.BytesRead(), bytesToRead))
			completed++
			// if completed >= count {
			fmt.Printf("  \n")
			fmt.Printf("Completed in %.1f seconds\n", bandwidthMeter.Duration().Seconds())
			return nil
			// }
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("\r%s - %s",
				format.BitsPerSec(bandwidthMeter.Bandwidth()),
				format.Percent(primaryBandwidthReader.BytesRead(), bytesToRead))
		}
	}
}

type copyResults struct {
	index        uint64
	bytesWritten uint64
	err          error
}

func asyncCopy(index uint64, channel chan *copyResults, writer io.Writer, reader io.Reader) {
	bytesWritten, err := io.Copy(writer, reader)
	channel <- &copyResults{index, uint64(bytesWritten), err}
}

func sumArr(array []uint64) (sum uint64) {
	for i := 0; i < len(array); i++ {
		sum = sum + array[i]
	}
	return
}

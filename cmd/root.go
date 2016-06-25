// Copyright © 2016 Gus Esquivel <gesquive@gmail.com>

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var cfgFile string
var useHTTPS bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fast-cli",
	Short: "Estimates your current internet download speed",
	Long: `Estimates your current internet download speed using Netflix's fast.com service.

fast-cli caclulates this estimage by performing a series of downloads from Netflix's fast.com servers`,
	Run: run,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.fast-cli.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&useHTTPS, "use-https", "s", false, "Use HTTPS when connecting")
	//TODO: Add version flag
	//TODO: Allow to estimate using time or size
}

func initConfig() {
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("Estimating current download speed\n")
	url := "http://api.fast.com/netflix/speedtest?https=false"
	if useHTTPS {
		url = "https://api.fast.com/netflix/speedtest?https=true"
	}

	err := calculateBandwidth(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func calculateBandwidth(url string) (err error) {
	// fmt.Printf("downloading %s\n", url)
	client := &http.Client{}

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "fast-cli")

	// Get the HTTP Response
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Try to get the content length
	contentLength := resp.Header.Get("Content-Length")
	calculatedLength, err := strconv.Atoi(contentLength)
	if err != nil {
		calculatedLength = 26214400
	}
	bytesToRead := uint64(calculatedLength)

	// Start reading
	bandwidthReader := BandwidthReader{}
	ch := make(chan *copyResults, 1)

	go func() {
		bytesWritten, err := io.Copy(&bandwidthReader, resp.Body)
		ch <- &copyResults{uint64(bytesWritten), err}
	}()

	for {
		select {
		case results := <-ch:
			if results.err != nil {
				fmt.Fprintf(os.Stdout, "\n%v\n", results.err)
				os.Exit(1)
			}
			fmt.Printf("\r%s - %s  \n",
				fmtBitsPerSec(bandwidthReader.Bandwidth()),
				fmtPercent(bandwidthReader.BytesRead(), bytesToRead))
			fmt.Printf("Completed in %.1f seconds\n", bandwidthReader.Duration().Seconds())
			return nil
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("\r%s - %s",
				fmtBitsPerSec(bandwidthReader.Bandwidth()),
				fmtPercent(bandwidthReader.BytesRead(), bytesToRead))
		}
	}
}

type copyResults struct {
	bytesWritten uint64
	err          error
}

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

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.fast-cli.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&useHTTPS, "use-https", "s", false, "Use HTTPS when connecting")
	RootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Display the version number and exit")
	//TODO: Allow to estimate using time or size
}

func initConfig() {
}

func run(cmd *cobra.Command, args []string) {
	if showVersion {
		fmt.Println(displayVersion)
		os.Exit(0)
	}

	fmt.Printf("Estimating current download speed\n")
	urls := fast.GetDlUrls(1)
	fmt.Printf("%+v\n", urls)

	if len(urls) == 0 {
		urls = append(urls, fast.GetDefaultURL())
	}

	err := calculateBandwidth(urls[0])
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
	req.Header.Set("User-Agent", displayVersion)

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
	bandwidthMeter := meters.BandwidthMeter{}
	ch := make(chan *copyResults, 1)

	go func() {
		bytesWritten, err := io.Copy(&bandwidthMeter, resp.Body)
		ch <- &copyResults{uint64(bytesWritten), err}
	}()

	// TODO: Need to add ability to dl 3 files at time
	for {
		select {
		case results := <-ch:
			if results.err != nil {
				fmt.Fprintf(os.Stdout, "\n%v\n", results.err)
				os.Exit(1)
			}
			fmt.Printf("\r%s - %s  \n",
				format.BitsPerSec(bandwidthMeter.Bandwidth()),
				format.Percent(bandwidthMeter.BytesRead(), bytesToRead))
			fmt.Printf("Completed in %.1f seconds\n", bandwidthMeter.Duration().Seconds())
			return nil
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("\r%s - %s",
				format.BitsPerSec(bandwidthMeter.Bandwidth()),
				format.Percent(bandwidthMeter.BytesRead(), bytesToRead))
		}
	}
}

type copyResults struct {
	bytesWritten uint64
	err          error
}

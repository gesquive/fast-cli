// Copyright © 2016 Gus Esquivel <gesquive@gmail.com>

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var cfgFile string
var useHTTPS bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fast-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.fast-cli.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&useHTTPS, "use-https", "s", false, "Use HTTPS when connecting")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
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

// BandwidthReader counts the number of bytes written to it.
type BandwidthReader struct {
	bytesRead uint64 // Total # of bytes transferred
	start     time.Time
	lastRead  time.Time
}

// Write implements the io.Writer interface.
//
// Always completes and never returns an error.
func (br *BandwidthReader) Write(p []byte) (int, error) {
	br.lastRead = time.Now().UTC()
	n := len(p)
	br.bytesRead += uint64(n)
	if br.start.IsZero() {
		br.start = br.lastRead
	}
	// bandwidth := float64(float64(br.bytesRead) / time.Since(br.start).Seconds())
	// fmt.Printf("\r%s        ", humanize.SI(bandwidth, "bps"))
	// fmt.Printf("Read %d bytes for a total of %d\n", n, br.Total)

	return n, nil
}

// Bandwidth returns the current bandwidth
func (br *BandwidthReader) Bandwidth() (bytesPerSec float64) {
	deltaSecs := br.lastRead.Sub(br.start).Seconds()
	bytesPerSec = float64(br.bytesRead) / deltaSecs
	return
}

// BytesRead returns the number of bytes read by this BandwidthReader
func (br *BandwidthReader) BytesRead() (bytes uint64) {
	bytes = br.bytesRead
	return
}

// Start records the start time
func (br *BandwidthReader) Start() {
	br.start = time.Now().UTC()
}

type copyResults struct {
	bytesWritten uint64
	err          error
}

func calculateBandwidth(url string) (err error) {
	// fmt.Printf("downloading %s\n", url)
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "fast-cli")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

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
			// fmt.Printf("\n%d received\n", results.bytesWritten)
			fmt.Printf("\r%s - %s  \n",
				fmtBitsPerSec(bandwidthReader.Bandwidth()),
				fmtPercent(bandwidthReader.BytesRead(), 26214400))
			fmt.Printf("Complete\n")
			return nil
		case <-time.After(100 * time.Millisecond):

			fmt.Printf("\r%s - %s  ",
				fmtBitsPerSec(bandwidthReader.Bandwidth()),
				fmtPercent(bandwidthReader.BytesRead(), 26214400))
		}
	}
}

func fmtBitsPerSec(bytes float64) string {
	prettySize, prettyUnit := humanize.ComputeSI(bytes * 8)
	return fmt.Sprintf("%7.2f %sbps", prettySize, prettyUnit)
}

func fmtBytes(bytes uint64) string {
	prettySize, prettyUnit := humanize.ComputeSI(float64(bytes))
	return fmt.Sprintf("%3.f %sB", prettySize, prettyUnit)
}

func fmtPercent(current uint64, total uint64) string {
	return fmt.Sprintf("%5.1f%%", float64(current)/float64(total)*100)
}

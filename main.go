package main

import "os"
import "fmt"
import "io"
import "net/http"
import "strconv"
import "time"

import "github.com/spf13/cobra"
import "github.com/gesquive/cli"
import "github.com/gesquive/fast-cli/fast"
import "github.com/gesquive/fast-cli/format"
import "github.com/gesquive/fast-cli/meters"

var version = "v0.2.8"
var dirty = ""
var displayVersion string

var cfgFile string
var logDebug bool
var notHTTPS bool
var simpleProgress bool
var showVersion bool

//RootCmd is the only command
var RootCmd = &cobra.Command{
	Use:   "fast-cli",
	Short: "Estimates your current internet download speed",
	Long:  `fast-cli estimates your current internet download speed by performing a series of downloads from Netflix's fast.com servers.`,
	Run:   run,
}

func main() {
	displayVersion = fmt.Sprintf("fast-cli %s%s",
		version,
		dirty)
	Execute(displayVersion)
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	displayVersion = version
	RootCmd.SetHelpTemplate(fmt.Sprintf("%s\nVersion:\n  github.com/gesquive/%s\n",
		RootCmd.HelpTemplate(), displayVersion))
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initLog)

	RootCmd.PersistentFlags().BoolVarP(&notHTTPS, "no-https", "n", false, "Do not use HTTPS when connecting")
	RootCmd.PersistentFlags().BoolVarP(&simpleProgress, "simple", "s", false, "Only display the result, no dynamic progress bar")
	RootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Display the version number and exit")
	RootCmd.PersistentFlags().BoolVarP(&logDebug, "debug", "D", false, "Write debug messages to console")

	RootCmd.PersistentFlags().MarkHidden("debug")
	//TODO: Allow to estimate using time or size
}

func initLog() {
	cli.SetPrintLevel(cli.LevelInfo)
	if logDebug {
		cli.SetPrintLevel(cli.LevelDebug)
	}
	if notHTTPS {
		cli.Debugln("Not using HTTPS")
	} else {
		cli.Debugln("Using HTTPS")
	}
}

func run(cmd *cobra.Command, args []string) {
	if showVersion {
		cli.Infoln(displayVersion)
		os.Exit(0)
	}
	count := uint64(3)
	fast.UseHTTPS = !notHTTPS
	urls := fast.GetDlUrls(count)
	cli.Debugf("Got %d from fast service\n", len(urls))

	if len(urls) == 0 {
		cli.Warnf("Using fallback endpoint\n")
		urls = append(urls, fast.GetDefaultURL())
	}

	err := calculateBandwidth(urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func calculateBandwidth(urls []string) (err error) {
	client := &http.Client{}
	count := uint64(len(urls))

	primaryBandwidthReader := meters.BandwidthMeter{}
	bandwidthMeter := meters.BandwidthMeter{}
	ch := make(chan *copyResults, 1)
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
			cli.Debugf("Download Size=%d\n", bytesToRead)

			tapMeter := io.TeeReader(response.Body, &primaryBandwidthReader)
			go asyncCopy(i, ch, &bandwidthMeter, tapMeter)
		} else {
			// Start reading
			go asyncCopy(i, ch, &bandwidthMeter, response.Body)
		}

	}

	if !simpleProgress {
		cli.Infof("Estimating current download speed\n")
	}
	for {
		select {
		case results := <-ch:
			if results.err != nil {
				fmt.Fprintf(os.Stdout, "\n%v\n", results.err)
				os.Exit(1)
			}

			completed++
			if !simpleProgress {
				fmt.Printf("\r%s - %s",
					format.BitsPerSec(bandwidthMeter.Bandwidth()),
					format.Percent(primaryBandwidthReader.BytesRead(), bytesToRead))
				fmt.Printf("  \n")
				fmt.Printf("Completed in %.1f seconds\n", bandwidthMeter.Duration().Seconds())
			} else {
				fmt.Printf("%s\n", format.BitsPerSec(bandwidthMeter.Bandwidth()))
			}
			return nil
		case <-time.After(100 * time.Millisecond):
			if !simpleProgress {
				fmt.Printf("\r%s - %s",
					format.BitsPerSec(bandwidthMeter.Bandwidth()),
					format.Percent(primaryBandwidthReader.BytesRead(), bytesToRead))
			}
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

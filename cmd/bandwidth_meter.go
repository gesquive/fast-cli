package cmd

import "time"

// BandwidthMeter counts the number of bytes written to it over time.
type BandwidthMeter struct {
	bytesRead uint64
	start     time.Time
	lastRead  time.Time
}

// Write implements the io.Writer interface.
func (br *BandwidthMeter) Write(p []byte) (int, error) {
	// Always completes and never returns an error.
	br.lastRead = time.Now().UTC()
	n := len(p)
	br.bytesRead += uint64(n)
	if br.start.IsZero() {
		br.start = br.lastRead
	}

	return n, nil
}

// Start records the start time
func (br *BandwidthMeter) Start() {
	br.start = time.Now().UTC()
}

// Bandwidth returns the current bandwidth
func (br *BandwidthMeter) Bandwidth() (bytesPerSec float64) {
	deltaSecs := br.lastRead.Sub(br.start).Seconds()
	bytesPerSec = float64(br.bytesRead) / deltaSecs
	return
}

// BytesRead returns the number of bytes read by this BandwidthMeter
func (br *BandwidthMeter) BytesRead() (bytes uint64) {
	bytes = br.bytesRead
	return
}

// Duration returns the current duration
func (br *BandwidthMeter) Duration() (duration time.Duration) {
	duration = br.lastRead.Sub(br.start)
	return
}

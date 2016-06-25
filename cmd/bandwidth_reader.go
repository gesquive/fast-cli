package cmd

import "time"

// BandwidthReader counts the number of bytes written to it.
type BandwidthReader struct {
	bytesRead uint64
	start     time.Time
	lastRead  time.Time
}

// Write implements the io.Writer interface.
func (br *BandwidthReader) Write(p []byte) (int, error) {
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
func (br *BandwidthReader) Start() {
	br.start = time.Now().UTC()
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

// Duration returns the current duration
func (br *BandwidthReader) Duration() (duration time.Duration) {
	duration = br.lastRead.Sub(br.start)
	return
}

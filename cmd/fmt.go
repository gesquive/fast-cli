package cmd

import "fmt"
import "github.com/dustin/go-humanize"

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

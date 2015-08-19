package main

import (
	"os"

	"github.com/cheggaaa/pb"
)

// NewProgressBar initializes new progress bar based on size of file
func NewProgressBar(file *os.File) *pb.ProgressBar {
	fi, err := file.Stat()

	total := int64(0)
	if err == nil {
		total = fi.Size()
	}

	bar := pb.New64(total)
	bar.SetUnits(pb.U_BYTES)
	return bar
}

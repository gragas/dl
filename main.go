package main

import (
	"fmt"
	"os"

	"github.com/gragas/dl/download"
)

const (
	PATH     = "output.mp4"
	URL      = "https://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4"
	ROUTINES = 1
)

func main() {
	if err := download.Download(PATH, URL, ROUTINES); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

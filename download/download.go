// download package exposes one function "Download"
// which can download a file
package download

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const MAX_TRIES = 8

// Download downloads the resource at the given URL
// to the given path, and attempts to spread the work
// across a given number of goroutines
func Download(path, url string, routines int) error {

	start := time.Now()
	fmt.Printf("GET %v, response...", url)
	// issue a GET request
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	fmt.Println(response.Status)
	fmt.Printf("-- Downloading to %v with %v goroutine(s)...", path, routines)

	var buf []byte

	h := Header(response.Header)
	allowRange := h.AcceptRanges()
	numBytes, err := h.ContentLength()
	if err != nil || !allowRange || routines == 1 {
		// if the server doesn't allow range downloads,
		// use a single thread
		var b bytes.Buffer
		if _, err := io.Copy(&b, response.Body); err != nil {
			fmt.Printf("failure! %v elapsed\n", time.Since(start))
			return err
		}
		buf = b.Bytes()
	} else {
		// if the server allows ranges for the resource, and we know
		// the content length, download in parallel
		buf, err = downloadPar(url, numBytes, routines)
		if err != nil {
			fmt.Printf("failure! %v elapsed\n", time.Since(start))
			return err
		}
	}

	checksum, err := h.MD5()
	if err == nil {
		// if there was an MD5 checksum in the header,
		// verify the integrity of the file with it
		ok := verifyMD5(buf, checksum)
		if !ok {
			fmt.Printf("failure! %v elapsed\n", time.Since(start))
			return fmt.Errorf("MD5 checksums do not match")
		}
		fmt.Printf("MD5 checksums match! ")
	}
	

	// create the file to write into
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("failure! %v elapsed\n", time.Since(start))
		return err
	}
	_, err = f.Write(buf)
	if err != nil {
		fmt.Printf("failure! %v elapsed\n", time.Since(start))
		return err
	}
	fmt.Printf("Success!\n%v elapsed\n", time.Since(start))
	return nil
}

// downloadPar uses $routines goroutines to download the resource
// at the given URL to the given file
func downloadPar(url string, numBytes, routines int) (buf []byte, err error) {
	// create a buf for the goroutines to write into
	buf = make([]byte, numBytes)
	sliceLen := numBytes / routines
	remaining := numBytes % routines

	// create slices for the return values of each call to downloadSlice
	bytesRead, errs := make([]int, numBytes), make([]error, numBytes)

	// create a channel to use as a barrier
	// and to signify which goroutine finished (the int is an index)
	b := make(chan int)

	// spin up the goroutines
	for i := 0; i < routines; i++ {
		start := i * sliceLen
		end := start + sliceLen
		if i == routines-1 {
			// make sure we get ALL the bytes at the end
			end += remaining
		}
		go func() {
			bytesRead[i], errs[i] = downloadSlice(buf, url, start, end)
			// once downloadSlice returns, let the other side of the
			// channel know that goroutine $i has finished
			b <- i
		}()
	}

	// wait for them all to finish
	tries := 0
	numFinished := 0
	for numFinished < routines {
		i := <-b
		if errs[i] != nil {
			tries += 1
			if tries > MAX_TRIES {
				return nil, fmt.Errorf("exceeded max tries")
			}
		}
		// if the whole slice wasn't downloaded,
		// spin up a goroutine to download the rest of it
		if i == routines-1 {
			if bytesRead[i] < sliceLen+remaining {
				// deal with the pesky extra bytes at the end
				go func() {
					start := sliceLen*i + bytesRead[i]
					end := sliceLen*(i+1) + remaining
					n, err := downloadSlice(buf, url, start, end)
					bytesRead[i] += n
					errs[i] = err
					b <- i
				}()
			} else {
				numFinished += 1
			}
		} else {
			if bytesRead[i] < sliceLen {
				go func() {
					start := sliceLen*i + bytesRead[i]
					end := sliceLen * (i + 1)
					n, err := downloadSlice(buf, url, start, end)
					bytesRead[i] += n
					errs[i] = err
					b <- i
				}()
			} else {
				numFinished += 1
			}
		}
	}

	return buf, nil
}

// download slice downloads the range [start-end] of the
// given url and copies the download bytes into the given
// []byte starting at $start. It returns the total number
// of bytesWritten or an error
func downloadSlice(dst []byte, url string, start, end int) (bytesWritten int, err error) {

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	request.Header.Add("Range", "bytes="+strconv.Itoa(start)+"-"+strconv.Itoa(end-1))
	client := new(http.Client)
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	// download the range of bytes into the specified slice of the buffer
	return io.ReadFull(response.Body, dst[start:end])
}

// verifyMD5 returns true if the checksum of the provided buf
// is equal to the provided checksum, otherwise false
func verifyMD5(buf []byte, checksum string) bool {
	bufChecksum := md5.Sum(buf) // make addressable
	return checksum == string(bufChecksum[:])
}

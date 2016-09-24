// download package exposes one function "Download"
// which can download a file
package download

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type Header map[string][]string

// AcceptRanges returns True if the given header
// specifies that ranges are accepted, otherwise false
func (h Header) AcceptRanges() bool {
	acceptRanges, ok := h["Accept-Ranges"]
	if !ok {
		return false
	}
	if len(acceptRanges) != 1 {
		return false
	}
	return acceptRanges[0] != "none"
}

// ContentLength returns the content length specified
// by the given header, or an error if the content length
// could not be determined
func (h Header) ContentLength() (numBytes int, err error) {
	contentLength, ok := h["Content-Length"]
	if !ok {
		return 0, fmt.Errorf("Content-Length not found in header")
	}
	if len(contentLength) != 1 {
		return 0, fmt.Errorf("could not parse Content-Length")
	}
	return strconv.Atoi(contentLength[0])
}

func (h Header) MD5() (checksum string, err error) {
	eTag, ok := h["ETag"]
	if ok {
		return eTag[0], nil
	}
	xGoogHash, ok := h["X-Goog-Hash"]
	if !ok {
		return "", fmt.Errorf("x-goog-hash not found in header")
	}
	for _, s := range xGoogHash {
		if strings.HasPrefix(s, "md5=") {
			decoded, err := base64.StdEncoding.DecodeString(s[4:])
			if err != nil {
				return "", err
			}
			return string(decoded), nil
		}
	}
	return "", fmt.Errorf("md5 not found in header")
}

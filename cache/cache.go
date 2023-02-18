package cache

import (
	"bytes"
	"fmt"
	"github.com/DQNEO/apple-dictionary/extracter/raw"
	"io"
	"os"
)

const DEFAULT_PATH = "noad.cache"

const DLMT = "\t"

func SaveEntries(w io.Writer, entries []*raw.Entry) {
	for _, e := range entries {
		fmt.Fprintf(w, "%s%s%s\n", e.Title, DLMT, e.Body)
	}
}

func LoadFromCacheFile(path string) []*raw.Entry {
	var r []*raw.Entry
	contents, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	lines := bytes.Split(contents, []byte{'\n'})
	for _, line := range lines[:] {
		if len(line) == 0 {
			// Possibly end of file
			continue
		}
		ttlBytes, rawBody, found := bytes.Cut(line, []byte(DLMT))
		if !found {
			panic("failed to Cut:" + (string(line)))
		}
		title := string(ttlBytes)
		e := &raw.Entry{
			Title: title,
			Body:  rawBody,
		}
		r = append(r, e)
	}
	return r
}

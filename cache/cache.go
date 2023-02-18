package cache

import (
	"fmt"
	"github.com/DQNEO/apple-dictionary/extracter/raw"
	"io"
)

const DLMT = "\t"

func SaveEntries(w io.Writer, entries []*raw.Entry) {
	for _, e := range entries {
		fmt.Fprintf(w, "%s%s%s\n", e.Title, DLMT, e.Body)
	}
}

func EntryToText(title string, body []byte) string {
	return fmt.Sprintf("%s%s%s\n", title, DLMT, body)
}

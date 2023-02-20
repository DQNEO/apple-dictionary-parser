package extracter

import (
	"bytes"
	"compress/zlib"
	"github.com/DQNEO/apple-dictionary-parser/extracter/raw"
	"io"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func parseBinaryFile(filePath string) [][]byte {
	var chunks [][]byte

	data, err := os.ReadFile(filePath)
	check(err)
	br := bytes.NewReader(data)
	_, err = br.Seek(108, io.SeekCurrent) // skip non-zlib data in the head of the file
	check(err)
	for {
		var header = make([]byte, 2)
		_, err = br.Read(header)
		if err == io.EOF {
			return chunks
		}
		check(err)
		if header[0] == 0x78 && header[1] == 0xda { // check if it's a zlib magic header
			br.UnreadByte()
			br.UnreadByte()
			r, err := zlib.NewReader(br)
			check(err)
			buf, err := io.ReadAll(r)
			check(err)
			chunks = append(chunks, buf)
			check(err)
			r.Close()
			_, err = br.Seek(12, io.SeekCurrent) // skip magic 12 bytes
			if err == io.EOF {
				// This does not happen in the current version of NOAD file
				panic("Unexpected EOF")
			}
			check(err)
		} else {
			return chunks
		}
	}
}

func parseChunk(buf []byte) [][]byte {
	var entries [][]byte
	buf = buf[4:]
	for {
		idx := bytes.IndexByte(buf, '\n')
		if idx > -1 {
			entry := buf[0:idx]
			entries = append(entries, entry)
			if idx+5 >= len(buf) {
				return entries
			}
			buf = buf[idx+5:]
		} else {
			return entries
		}
	}
}

const titleStartMarker = `d:title="`

func parseEntry(entry []byte) *raw.Entry {
	titleStart := bytes.Index(entry, []byte(titleStartMarker)) + len(titleStartMarker)
	titleLen := bytes.Index(entry[titleStart:], []byte(`"`))
	title := entry[titleStart : titleStart+titleLen]

	return &raw.Entry{
		Title: string(title),
		Body:  entry,
	}
}

var LastTitle = "°"

func ParseBinaryFile(filePath string) []*raw.Entry {
	var entries []*raw.Entry
	chunks := parseBinaryFile(filePath)
	for _, chunk := range chunks {
		rawEntries := parseChunk(chunk)
		for _, rawEntry := range rawEntries {
			e := parseEntry(rawEntry)
			entries = append(entries, e)
			if e.Title == LastTitle {
				return entries
			}
		}
	}
	panic("internal error (probably last title does not match what we expect")
}

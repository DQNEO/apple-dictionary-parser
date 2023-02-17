// This program parses a dictionary file in MacOS
// Basic technique is inspired by this article : https://fmentzer.github.io/posts/2020/dictionary/

// Usage:
//
//	NOAD_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"
//	go run main.go $NOAD_FILE > noad.raw.txt
package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
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

type Entry struct {
	Title string
	Body  []byte
}

const titleStartMarker = `d:title="`

func parseEntry(entry []byte) *Entry {
	titleStart := bytes.Index(entry, []byte(titleStartMarker)) + len(titleStartMarker)
	titleLen := bytes.Index(entry[titleStart:], []byte(`"`))
	title := entry[titleStart : titleStart+titleLen]

	return &Entry{
		Title: string(title),
		Body:  entry,
	}
}

func ParseBinaryFile(filePath string) []*Entry {
	var entries []*Entry
	chunks := parseBinaryFile(filePath)
	for _, chunk := range chunks {
		rawEntries := parseChunk(chunk)
		for _, rawEntry := range rawEntries {
			e := parseEntry(rawEntry)
			entries = append(entries, e)
			if e.Title == "°" { // last title
				return entries
			}
		}
	}
	panic("internal error")
}

const DLMT = "\t"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please specify a dictionary body file\n")
		os.Exit(1)
	}
	var filePath = os.Args[1]
	entries := ParseBinaryFile(filePath)
	for _, e := range entries {
		fmt.Printf("%s%s%s\n", e.Title, DLMT, e.Body)
	}
}

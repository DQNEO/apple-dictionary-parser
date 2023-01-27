// This program parses a dictionary file in MacOS
// Basic technique is inspired by this article : https://fmentzer.github.io/posts/2020/dictionary/

// Usage:
//
//	NOAD_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/7e66d0bf940535a6ed4e0b6b29b6879cecc18477.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"
//	go run main.go $NOAD_FILE > /tmp/noad.txt
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

func parseBinaryFile(filePath string) []*bytes.Buffer {
	var chunks []*bytes.Buffer

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
			var buf = new(bytes.Buffer)
			r, err := zlib.NewReader(br)
			check(err)
			_, err = io.Copy(buf, r)
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

func parseChunk(chunk *bytes.Buffer) [][]byte {
	var entries [][]byte
	chunk.Next(4)
	buf := chunk.Bytes()
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
const bodyStartMarker = `" class="entry">`

func parseEntry(entry []byte) *Entry {
	titleStart := bytes.Index(entry, []byte(titleStartMarker)) + len(titleStartMarker)
	titleLen := bytes.Index(entry[titleStart:], []byte(`"`))
	title := entry[titleStart : titleStart+titleLen]
	//bodyStart := bytes.Index(entry, []byte(bodyStartMarker)) + len(bodyStartMarker)
	//body := entry[bodyStart : len(entry)-len("</d:entry>")]
	e := &Entry{
		Title: string(title),
		Body:  entry,
	}
	return e
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please specify a dictionary body file\n")
		os.Exit(1)
	}
	var filePath = os.Args[1]
	chunks := parseBinaryFile(filePath)
	for _, chunk := range chunks {
		rawEntries := parseChunk(chunk)
		for _, rawEntry := range rawEntries {
			e := parseEntry(rawEntry)
			fmt.Printf("%s:::%s\n", e.Title, e.Body)
			if e.Title == "°" { // last title
				return
			}
		}
	}
}

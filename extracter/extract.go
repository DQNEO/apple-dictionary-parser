package extracter

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/DQNEO/apple-dictionary-parser/extracter/raw"
	"io"
	"os"
	"unsafe"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Logic is borrowed from here: https://gist.github.com/josephg/5e134adf70760ee7e49d?permalink_comment_id=4554558#gistcomment-4554558
func parseBinaryFile(filePath string) [][]byte {
	var chunks [][]byte

	data, err := os.ReadFile(filePath)
	check(err)
	br := bytes.NewReader(data)
	_, err = br.Seek(0x40, io.SeekStart)
	check(err)
	fmt.Printf("initial 0x40 bytes skipped\n")
	// read the entire binary size
	var limitMarker = make([]byte, 4, 4)
	_, err = br.Read(limitMarker)
	check(err)
	var limitP = (*int32)(unsafe.Pointer(&limitMarker[0]))
	fmt.Printf("limit = %d\n", *limitP)

	_, err = br.Seek(0x60, io.SeekStart)
	check(err)
	var entryId int
	var blockIdx int
	for {
		pos, err := br.Seek(0, io.SeekCurrent)
		check(err)
		if pos >= int64(*limitP)+0x40 {
			fmt.Printf("=====  read all blocks\n")
			return chunks
		}
		blockIdx++
		fmt.Printf("=====  blockIdx %d, position %d\n", blockIdx, pos)
		var blockSizeBin = make([]byte, 4, 4)
		_, err = br.Read(blockSizeBin)
		if err == io.EOF {
			fmt.Printf("reached EOF\n")
			return chunks
		}
		var sz *int32 = (*int32)(unsafe.Pointer(&blockSizeBin[0]))
		if *sz == 0 {
			panic("block size is zero")
		}
		fmt.Printf("  block size = %d\n", *sz)
		var body = make([]byte, *sz)
		_, err = br.Read(body)
		check(err)

		btsr := bytes.NewReader(body[8:])
		r, err := zlib.NewReader(btsr)
		check(err)
		buf, err := io.ReadAll(r)
		check(err)
		chunkPos := 0
		for chunkPos < len(buf) {
			entryId++
			var chunkSize *int32 = (*int32)(unsafe.Pointer(&buf[chunkPos]))
			chunkPos += 4
			fmt.Printf("[entryId]: chunkSize = [%d]:%d\n", entryId, *chunkSize)
			entry := buf[chunkPos : chunkPos+int(*chunkSize)]
			chunks = append(chunks, entry)
			chunkPos += int(*chunkSize)
		}
		r.Close()
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
